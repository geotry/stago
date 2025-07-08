const { typeByteSize, typeIsInt } = require("./utils.js");

/**
 * Create a new array buffer with attributes and returns an API to update it.
 *
 * @param {WebGL2RenderingContext} gl
 * @param {WebGLProgram} program
 * @param {number} size the object size of the buffer
 * @param {number} instances the maximum number instances of data this buffer can hold 
 * @param {number} usage
 * @param {Object[]} attributes
 * @param {string} attributes.name
 * @param {number} attributes.size
 * @param {number} attributes.type
 * @param {boolean} attributes.normalized
 * @param {number} attributes.repeat
 * @param {number} attributes.instance
 */
export const createArrayBuffer = (gl, program, size, instances, usage, attributes) => {
  const stride = attributes.reduce((prev, curr) => prev + typeByteSize(gl, curr.type) * curr.size * (curr.repeat ?? 1), 0);
  const components = attributes.reduce((prev, curr) => prev + curr.size * (curr.repeat ?? 1), 0);

  const buffer = gl.createBuffer();
  gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
  gl.bufferData(gl.ARRAY_BUFFER, size * stride, usage);

  console.log(`ARRAY_BUFFER of size ${size}x${stride} = ${size * stride} bytes (${size * stride * instances} bytes)`);

  let offset = 0;
  for (const attr of attributes) {
    const loc = gl.getAttribLocation(program, attr.name);
    const iterations = attr.repeat ?? 1;
    for (let i = 0; i < iterations; ++i) {
      const attrLoc = loc + i;
      if (attrLoc >= 0) {
        console.log(`attribute ${attr.name} (loc=${attrLoc}) size=${attr.size} type=${attr.type} stride=${stride} offset=${offset}`);
        gl.enableVertexAttribArray(attrLoc);
        if (typeIsInt(gl, attr.type) && !attr.normalized) {
          gl.vertexAttribIPointer(attrLoc, attr.size, attr.type, stride, offset);
        } else {
          gl.vertexAttribPointer(attrLoc, attr.size, attr.type, attr.normalized, stride, offset);
        }
        offset += attr.size * typeByteSize(gl, attr.type);
        if (attr.instance) {
          gl.vertexAttribDivisor(attrLoc, attr.instance);
        }
      }
    }
  }

  // Create the buffer data
  const bufferDataSize = size * components * instances;
  const bufferData = typeIsInt(gl, attributes[0].type) ? new Int32Array(bufferDataSize) : new Float32Array(bufferDataSize);

  // For each instance, for each attribute, create a view of the original buffer data
  /** @type {Record<string, Float32Array[]>[]} */
  const attributeIndexedData = [];
  /** @type {Float32Array[]} */
  const indexedBufferData = [];
  for (let i = 0; i < instances; i++) {
    /** @type {Record<string, Float32Array[]>} */
    const attributeData = Object.fromEntries(attributes.map(attr => [attr.name, []]));
    let attributeOffset = i * size * stride;
    for (const attr of attributes) {
      const attrSize = attr.size * (attr.repeat ?? 1);
      for (let j = 0; j < size; ++j) {
        if (typeIsInt(gl, attr.type)) {
          attributeData[attr.name].push(new Int32Array(bufferData.buffer, attributeOffset + j * stride, attrSize));
        } else {
          attributeData[attr.name].push(new Float32Array(bufferData.buffer, attributeOffset + j * stride, attrSize));
        }
      }
      attributeOffset += attrSize * 4;
    }
    attributeIndexedData.push(attributeData);
    // Index buffer data by instance
    if (typeIsInt(gl, attributes[0].type)) {
      indexedBufferData.push(new Int32Array(bufferData.buffer, i * size * stride, size * components));
    } else {
      indexedBufferData.push(new Float32Array(bufferData.buffer, i * size * stride, size * components));
    }
  }

  /** @type {Map<number, number>} */
  let bufferUpdateStartOffset = new Map();
  /** @type {Map<number, number>} */
  let bufferUpdateEndOffset = new Map();
  /** @type {Map<number, Map<string, [number, number]>>} */
  const indexOffsets = new Map();

  /**
   * 
   * @param {number} index the buffer index to update
   * @param {number|string} ref string or number that uniquely identify an object in buffer
   * @param {Record<string, ArrayLike<number>[]>[]} data set of attributes data to update for this object
   */
  const bufferSetRef = (index, ref, data) => {
    // Keep track of objectId offset and attrs internally to compute the offset

    // Each buffer index has its own offsets
    let refOffsetSizes = indexOffsets.get(index);
    if (!refOffsetSizes) {
      refOffsetSizes = new Map();
      indexOffsets.set(index, refOffsetSizes);
    }

    // Offset should be stored by index
    let offsetSize = refOffsetSizes.get(ref);
    if (!offsetSize) {
      // First time this object is updated, create an offset for it
      const offsetsSizes = Array.from(refOffsetSizes.values());
      const lastOffsetSize = offsetsSizes.length > 0 ? offsetsSizes[offsetsSizes.length - 1] : [0, 0];
      offsetSize = [lastOffsetSize[0] + lastOffsetSize[1], data.length];
      refOffsetSizes.set(ref, offsetSize);
    }

    // Vertex offsets (1 = 1 * stride)
    const startOffset = offsetSize[0];
    const endOffset = offsetSize[0] + offsetSize[1];

    // Update global offset and size to know the slice of buffer to update
    if (!bufferUpdateStartOffset.has(index) || startOffset < bufferUpdateStartOffset.get(index)) {
      bufferUpdateStartOffset.set(index, startOffset);
    }
    if (!bufferUpdateEndOffset.has(index) || endOffset > bufferUpdateEndOffset.get(index)) {
      bufferUpdateEndOffset.set(index, endOffset);
    }

    // Update the actual buffer
    // Each entry in data is a vertex with its attributes
    for (const [i, item] of data.entries()) {
      // attributeIndexedData[index][attr][0] = first offset of attribute in buffer
      // attributeIndexedData[index]["a_position"][0] = [x, y, z];
      // attributeIndexedData[index]["a_position"][1] = [x, y, z];
      for (const [attr, values] of Object.entries(item)) {
        try {
          attributeIndexedData[index][attr][startOffset + i].set(values);
        } catch (err) {
          console.error(attr, startOffset + i, attributeIndexedData[bufIndex][attr][startOffset + i], values, err);
        }
      }
    }
  };

  /**
   * Send the internal buffer to GPU if it has changed.
   * 
   * @param {number} index the instance of the buffer to update (default: 0)
   */
  const updateBuffer = (index = 0) => {
    const startOffset = bufferUpdateStartOffset.get(index);
    if (startOffset !== undefined) {
      const endOffset = bufferUpdateEndOffset.get(index);
      gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
      gl.bufferSubData(gl.ARRAY_BUFFER, startOffset * stride, indexedBufferData[index].slice(startOffset * components, (startOffset + endOffset) * components));
      bufferUpdateStartOffset.delete(index);
      bufferUpdateEndOffset.delete(index);
    }
  };

  const readBuffer = (index, attr) => {
    if (attr !== undefined) {
      return attributeIndexedData[index][attr];
    }
    return indexedBufferData[index];
  };

  return { update: updateBuffer, set: bufferSetRef, read: readBuffer, };
};

/**
 * 
 * @param {WebGL2RenderingContext} gl
 * @param {number} size
 */
export const createDepthMapBuffer = (gl, size) => {
  const buffer = gl.createFramebuffer();

  const depthMap = gl.createTexture();
  gl.bindTexture(gl.TEXTURE_2D, depthMap);

  gl.texImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT24, size, size, 0, gl.DEPTH_COMPONENT, gl.UNSIGNED_INT, null);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT);

  gl.bindFramebuffer(gl.FRAMEBUFFER, buffer);
  gl.framebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, depthMap, 0);
  gl.drawBuffers([gl.NONE]);
  gl.readBuffer(gl.NONE);
  gl.bindFramebuffer(gl.FRAMEBUFFER, null);

  return [buffer, depthMap];
};

/**
 * Create a texture.
 *
 * @param {WebGL2RenderingContext} gl 
 * @param {number} index
 * @param {number} format
 * @param {number} width
 * @param {number} depth
 * @param {Uint8Array} pixels
 */
export const createTexture = (gl, index, format, width, depth, pixels) => {
  let pixelSize;
  let glFormat, glInternalFormat;

  switch (format) {
    case 0:
      pixelSize = 1;
      glFormat = gl.ALPHA;
      glInternalFormat = glFormat;
      break;
    case 1:
      pixelSize = 3;
      glFormat = gl.RGB;
      if (index === 0 || index === 1) {
        glInternalFormat = gl.SRGB;
      } else {
        glInternalFormat = glFormat;
      }
      break;
    case 2:
      pixelSize = 4;
      glFormat = gl.RGBA;
      if (index === 0 || index === 1) {
        glInternalFormat = gl.SRGB8_ALPHA8;
      } else {
        glInternalFormat = glFormat;
      }
      break;
  }

  const height = pixels.length / pixelSize / depth / width;

  gl.activeTexture(gl.TEXTURE0 + index);

  const texture = gl.createTexture();
  if (depth > 1) {
    gl.bindTexture(gl.TEXTURE_2D_ARRAY, texture);
    gl.pixelStorei(gl.UNPACK_ALIGNMENT, 1);
    gl.texParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
    gl.texParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
    gl.texParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_MIN_FILTER, gl.NEAREST);
    gl.texParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_MAG_FILTER, gl.NEAREST);
    gl.texImage3D(gl.TEXTURE_2D_ARRAY, 0, glInternalFormat, width, height, depth, 0, glFormat, gl.UNSIGNED_BYTE, pixels);
  } else {
    gl.bindTexture(gl.TEXTURE_2D, texture);
    gl.pixelStorei(gl.UNPACK_ALIGNMENT, 1);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST);
    gl.texImage2D(gl.TEXTURE_2D, 0, glInternalFormat, width, height, 0, glFormat, gl.UNSIGNED_BYTE, pixels);
  }

  return texture;
};
