const { typeByteSize, typeIsInt } = require("./utils.js");

/**
 * Create a new array buffer with attributes.
 *
 * @param {WebGL2RenderingContext} gl
 * @param {WebGLProgram} program
 * @param {number} size
 * @param {number} usage
 * @param {Object[]} attributes
 * @param {string} attributes.name
 * @param {number} attributes.size
 * @param {number} attributes.type
 * @param {boolean} attributes.normalized
 * @param {number} attributes.repeat
 * @param {number} attributes.instance
 */
export const createArrayBuffer = (gl, program, size, usage, attributes) => {
  const stride = attributes.reduce((prev, curr) => prev + typeByteSize(gl, curr.type) * curr.size * (curr.repeat ?? 1), 0);

  const buffer = gl.createBuffer();
  gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
  gl.bufferData(gl.ARRAY_BUFFER, size * stride, usage);

  console.log(`ARRAY_BUFFER of size ${size}x${stride} = ${size * stride}`);

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

  return buffer;
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
