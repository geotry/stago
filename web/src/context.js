const { typeByteSize } = require("./utils.js");

/**
 * @param {WebGL2RenderingContext} gl 
 *
 * @returns 
 */
export const createContext = (gl) => {
  /**
   * @type {Map<string, WebGLBuffer>}
   */
  const buffers = new Map();
  /**
   * @type {Map<string, number>}
   */
  const bufferTargets = new Map();

  /**
   * @type {Map<string, WebGLTexture>}
   */
  const textures = new Map();
  /**
   * @type {Map<string, {target: number, format: number, type: number, unit: number, width: number, height: number}>}
   */
  const textureInfo = new Map();

  /**
   * Map of buffer attributes and their offset in their buffer.
   * @type {Map<string, Map<string, number>>}
   */
  const bufferOffsets = new Map();

  /**
   * Create and allocate data for a new buffer.
   *
   * @param {string} name 
   * @param {number} type 
   * @param {number} usage 
   * @param {{name: string, attributes: Record<string, {size: number, type: number}>}[]} blocks 
   */
  const createBuffer = (name, type, usage, blocks) => {
    if (buffers.has(name)) {
      throw new Error(`Buffer ${name} already exists`);
    }

    const offsetMap = new Map();
    bufferOffsets.set(name, offsetMap);

    // Create offsets for each block, and for each attribute
    // {block}
    // {block}.{attribute}

    let startOffset = 0;
    for (const block of blocks) {
      const blockSize = block.size ?? 1;
      let blockStride = 0;
      let attributeOffset = 0;
      offsetMap.set(`${block.name}`, startOffset);
      for (const [attrName, attr] of Object.entries(block.attributes)) {
        offsetMap.set(`${block.name}.${attrName}`, startOffset + attributeOffset);
        const attributeStride = attr.size * typeByteSize(gl, attr.type);
        attributeOffset += attributeStride;
        blockStride += attributeStride;
      }
      startOffset += blockStride * blockSize;
    }
    const size = startOffset;

    const buffer = gl.createBuffer();
    buffers.set(name, buffer);
    bufferTargets.set(name, type);

    gl.bindBuffer(type, buffer);
    gl.bufferData(type, size, usage);

    gl.bindBuffer(type, null);

    console.log(`[context] created buffer ${name} of size ${Math.ceil(size / 1024)}kb`);
  };

  /**
   * Create a new frame buffer.
   *
   * @param {string} name 
   * @param {number[]} drawBuffers 
   * @param {number} readBuffer 
   * @returns 
   */
  const createFrameBuffer = (name, drawBuffers = [gl.NONE], readBuffer = gl.NONE) => {
    if (buffers.has(name)) {
      throw new Error(`Buffer ${name} already exists`);
    }

    const buffer = gl.createFramebuffer();
    buffers.set(name, buffer);

    gl.bindFramebuffer(gl.FRAMEBUFFER, buffer);
    gl.drawBuffers(drawBuffers);
    gl.readBuffer(readBuffer);
    gl.bindFramebuffer(gl.FRAMEBUFFER, null);
  };

  /**
   * Bind a frame buffer with a texture.
   *
   * @param {string} name 
   * @param {string} textureName 
   * @param {number|undefined} textureLayer 
   */
  const bindFrameBufferTexture = (name, textureName, textureLayer) => {
    const buffer = buffers.get(name);
    if (!buffer) {
      throw new Error(`Buffer ${name} does not exist`);
    }
    const texture = textures.get(textureName);
    if (!texture) {
      throw new Error(`Texture ${textureName} does not exist`);
    }

    gl.bindFramebuffer(gl.FRAMEBUFFER, buffer);
    if (textureLayer === undefined) {
      gl.framebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, texture, 0);
    } else {
      gl.framebufferTextureLayer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, texture, 0, textureLayer);
    }
    if (gl.checkFramebufferStatus(gl.FRAMEBUFFER) !== gl.FRAMEBUFFER_COMPLETE) {
      throw new Error(`Failed to create framebuffer (error=${gl.checkFramebufferStatus(gl.GL_FRAMEBUFFER)})`);
    }
  };

  /**
   * Update the buffer data at offset pointed by path.
   *
   * @param {string} name
   * @param {string} block 
   * @param {number} offset 
   * @param {(Record<string, Float32Array>|Record<string, Float32Array>[])} data 
   */
  const updateBuffer = (name, block, offset, data) => {
    const buffer = buffers.get(name);
    if (!buffer) {
      throw new Error(`Buffer ${name} does not exist`);
    }

    const blockOffset = bufferOffsets.get(name)?.get(block);
    if (blockOffset === undefined) {
      throw new Error(`Block ${block} does not exist in buffer ${name}`);
    }

    if (Array.isArray(data) && data.length === 0) {
      return;
    }

    // Buffer data must match exactly the configuration in buffer, otherwise it the buffer data will be wrong.
    // Todo: write some kind of validation here
    const item = Array.isArray(data) ? data[0] : data;
    const attrs = Object.keys(item);
    if (attrs.length === 0) {
      return;
    }

    let byteSize = 0;
    const attributeOffsets = {};
    for (const attr of attrs) {
      byteSize += item[attr].byteLength;
      attributeOffsets[attr] = bufferOffsets.get(name).get(`${block}.${attr}`);
    }

    // if (size !== bufferSize.get(name)) {
    //   throw new Error(`Buffer data has invalid size: expected ${bufferSize.get(name)}, got ${size}`);
    // }

    // This is not performant and hacky
    // let minOffset;
    // let maxOffset;
    // const offsets = {};
    // for (const attr of attrs) {
    //   const attrOffset = bufferOffsets.get(name).get(`${path}.${attr}`);
    //   if (minOffset === undefined || attrOffset < minOffset) {
    //     minOffset = attrOffset;
    //   }
    //   if (maxOffset === undefined || attrOffset > maxOffset) {
    //     maxOffset = attrOffset;
    //   }
    //   offsets[attr] = attrOffset;
    // }

    // const size = maxOffset - minOffset;

    const startOffset = blockOffset + offset * byteSize;

    // console.log(`update buffer `, blockOffset, startOffset, attributeOffsets, byteSize, data);

    const target = bufferTargets.get(name);
    gl.bindBuffer(target, buffer);

    if (Array.isArray(data)) {
      const bufferData = new Float32Array(byteSize * data.length);
      for (const [i, d] of data.entries()) {
        for (const attr of attrs) {
          if (!d[attr]) {
            throw new Error(`Invalid data at index ${i}: missing ${attr}`);
          }
          bufferData.set(d[attr], i * (byteSize / 4) + attributeOffsets[attr] / 4);
        }
      }
      gl.bufferSubData(target, startOffset, bufferData);
    } else {
      const bufferData = new Float32Array(byteSize);
      for (const attr of attrs) {
        bufferData.set(data[attr], attributeOffsets[attr] / 4);
      }
      gl.bufferSubData(target, startOffset, bufferData);
    }

    gl.bindBuffer(target, null);
  };

  // /**
  //  * Bind a buffer with a program.
  //  *
  //  * @param {string} buffer 
  //  * @param {WebGLProgram} program 
  //  */
  // const bindBuffer = (buffer, program, attributes) => {

  //   gl.getAttribLocation(program, "a_model");

  // };

  // const vao = bindBuffers(program, {
  //   Vertex: {
  //     position: "a_position",
  //     texCoords: "a_uv",
  //     normals: "a_normal",
  //   },
  //   Model: {
  //     model: { attribute: "a_model", instanced: true },
  //   }
  // });

  // bindBuffer("Vertex", program, {
  //   position: "a_position",
  //   texCoords: "a_uv",
  //   normals: "a_normal",
  // });

  // bindBuffer("Model", program, {
  //   model: { attribute: "a_model", instanced: true },
  // });

  /**
   * Returns an active buffer.
   *
   * @param {string} name 
   * @returns 
   */
  const getBuffer = (name) => {
    const buffer = buffers.get(name);
    if (!buffer) {
      throw new Error(`Buffer ${name} does not exist`);
    }
    return buffer;
  };

  /**
   * Create a new texture.
   * 
   * @param {string} name 
   * @param {number} format 
   * @param {number} internalFormat 
   * @param {number} width 
   * @param {number} height 
   * @param {number} depth 
   * @param {Uint8Array|null} pixels
   */
  const createTexture = (name, format, internalFormat, width, height, depth, pixels) => {
    if (textures.has(name)) {
      throw new Error(`Texture ${name} already exist`);
    }
    if (textures.size >= gl.getParameter(gl.MAX_TEXTURE_IMAGE_UNITS)) {
      throw new Error(`Too many textures (${textures.size})`);
    }

    const target = depth > 1 ? gl.TEXTURE_2D_ARRAY : gl.TEXTURE_2D;
    const type = gl.UNSIGNED_BYTE;
    const unit = textures.size;

    const texture = gl.createTexture();
    gl.activeTexture(gl.TEXTURE0 + unit);
    gl.bindTexture(target, texture);
    if (target === gl.TEXTURE_2D_ARRAY) {
      gl.texImage3D(target, 0, internalFormat, width, height, depth, 0, format, type, pixels);
    } else {
      gl.texImage2D(target, 0, internalFormat, width, height, 0, format, type, pixels);
    }
    gl.pixelStorei(gl.UNPACK_ALIGNMENT, 1);
    gl.texParameteri(target, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
    gl.texParameteri(target, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
    gl.texParameteri(target, gl.TEXTURE_MIN_FILTER, gl.NEAREST);
    gl.texParameteri(target, gl.TEXTURE_MAG_FILTER, gl.NEAREST);

    textureInfo.set(name, {
      format,
      target,
      height,
      width,
      type,
      unit,
    });
    textures.set(name, texture);

    console.log(`[context] created texture ${name} ${unit} ${width}x${height}x${depth}`);
  };

  /**
   * Create a new depth texture.
   *
   * @param {string} name 
   * @param {number} width 
   * @param {number} height 
   * @param {number} depth 
   */
  const createDepthTexture = (name, width, height, depth) => {
    if (textures.has(name)) {
      throw new Error(`Texture ${name} already exist`);
    }
    if (textures.size >= gl.getParameter(gl.MAX_TEXTURE_IMAGE_UNITS)) {
      throw new Error(`Too many textures (${textures.size})`);
    }

    const format = gl.DEPTH_COMPONENT;
    const target = depth > 1 ? gl.TEXTURE_2D_ARRAY : gl.TEXTURE_2D;
    const type = gl.FLOAT;
    const unit = textures.size;

    const texture = gl.createTexture();
    gl.activeTexture(gl.TEXTURE0 + unit);
    gl.bindTexture(target, texture);
    if (target === gl.TEXTURE_2D_ARRAY) {
      gl.texImage3D(target, 0, gl.DEPTH_COMPONENT32F, width, height, depth, 0, format, type, null);
    } else {
      gl.texImage2D(target, 0, gl.DEPTH_COMPONENT32F, width, height, 0, format, type, null);
    }
    // gl.pixelStorei(gl.UNPACK_ALIGNMENT, 1);
    gl.texParameteri(target, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
    gl.texParameteri(target, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
    gl.texParameteri(target, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
    gl.texParameteri(target, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
    gl.texParameteri(target, gl.TEXTURE_COMPARE_MODE, gl.COMPARE_REF_TO_TEXTURE);

    textureInfo.set(name, {
      format,
      type,
      target: depth > 1 ? gl.TEXTURE_2D_ARRAY : gl.TEXTURE_2D,
      height,
      width,
      unit,
    });
    textures.set(name, texture);

    console.log(`[context] created depth texture ${name} ${unit} ${width}x${height}x${depth}`);
  };

  /**
   * Update texture data.
   *
   * @param {string} name 
   * @param {number} layer 
   * @param {Uint8Array} pixels 
   */
  const updateTexture = (name, layer, pixels) => {
    const texture = textures.get(name);
    if (!texture) {
      throw new Error(`Texture ${name} does not exist`);
    }

    const info = textureInfo.get(name);

    gl.activeTexture(gl.TEXTURE0 + info.unit);
    gl.bindTexture(target, texture);

    switch (info.target) {
      case gl.TEXTURE_2D:
        gl.texSubImage2D(target, 0, 0, 0, info.width, info.height, info.format, info.type, pixels);
        break;
      case gl.TEXTURE_2D_ARRAY:
        gl.texSubImage3D(target, 0, 0, 0, 0, info.width, info.height, layer, info.format, info.type, pixels);
        break;
    }

    gl.bindTexture(target, null);
  };

  /**
   * 
   * @param {string} name 
   * @return {number}
   */
  const getTextureIndex = (name) => {
    const info = textureInfo.get(name);
    if (!info) {
      throw new Error(`Texture ${name} does not exist`);
    }
    return info.unit;
  };

  const getAspectRatio = () => {
    return gl.canvas.width / gl.canvas.height;
  };

  return {
    frame: 0,
    deltaTime: 0,
    renderTime: 0,
    getAspectRatio,
    createBuffer,
    createFrameBuffer,
    bindFrameBufferTexture,
    updateBuffer,
    getBuffer,
    createTexture,
    createDepthTexture,
    updateTexture,
    getTextureIndex,
  };
};
