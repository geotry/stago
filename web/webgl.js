const { Format } = require("./pb/frame_pb.js");
const { readSceneObjectBuffer, SceneObjectBlock } = require("./encoding.js");

/**
 * @type {number}
 */
let lastFrame, skippedFrames = 0;

/**
 * @type {Uint8Array}
 */
let texture;

const vertexShaderObjectSrc = `
#version 300 es

#define PALETTE_SIZE 256
#define MAX_OBJECTS 64
#define OBJECT_VERTICES 6

precision mediump float;
precision highp sampler2DArray;

flat out int object_index;
flat out int tex_index;
out vec2 v_texcoord;

uniform vec2 u_tex_dim;
uniform vec2 u_texture_size[PALETTE_SIZE];

uniform int u_object_texture[MAX_OBJECTS];
uniform mat4 u_transform[MAX_OBJECTS];

void main() {
  vec4 a_position = vec4(0.0, 0.0, 1.0, 1.0);

  // Generate two tris to create a rectangle
  switch (gl_VertexID % OBJECT_VERTICES) {
    case 0:
      a_position.xy = vec2(1.0, 1.0);
      break;
    case 1:
      a_position.xy = vec2(-1.0, 1.0);
      break;
    case 2:
      a_position.xy = vec2(-1.0, -1.0);
      break;
    case 3:
      a_position.xy = vec2(1.0, 1.0);
      break;
    case 4:
      a_position.xy = vec2(-1.0, -1.0);
      break;
    case 5:
      a_position.xy = vec2(1.0, -1.0);
      break;
  }
  
  // Compute object index from gl_VertexID
  object_index = int(floor(float(gl_VertexID)/float(OBJECT_VERTICES)));
  
  tex_index = u_object_texture[object_index];
  vec2 tex_size = u_texture_size[tex_index].xy;
  // Normalize coords to have (0,0) on top left
  v_texcoord = a_position.xy * vec2(0.5, -0.5) + 0.5;
  // Scale with texture size
  v_texcoord.x *= (tex_size.x/u_tex_dim.x);
  v_texcoord.y *= (tex_size.y/u_tex_dim.y);
  
  // Apply texture aspect ratio
  if (tex_size.x > tex_size.y) {
    a_position.x *= tex_size.x / tex_size.y;
  } else if (tex_size.x < tex_size.y) {
    a_position.y *= tex_size.y / tex_size.x;
  }

  gl_Position = u_transform[object_index] * a_position;
  // gl_Position = a_position * u_transform[object_index];
}`;

const fragmentShaderObjectSrc = `
#version 300 es

precision mediump float;
precision highp sampler2DArray;

in vec2 v_texcoord;
flat in int tex_index;

out vec4 fragColor;

uniform sampler2DArray u_image;
uniform sampler2D u_palette;

void main() {
    float index = texture(u_image, vec3(v_texcoord, tex_index)).a;
    fragColor = texture(u_palette, vec2(index, 0));
}`;

/**
 * Compile a new shader.
 *
 * @param {WebGLRenderingContext} gl 
 * @param {string} sourceCode 
 * @param {gl.VERTEX_SHADER|gl.FRAGMENT_SHADER} type 
 * @returns 
 */
const createShader = (gl, sourceCode, type) => {
  const shader = gl.createShader(type);
  gl.shaderSource(shader, sourceCode);
  gl.compileShader(shader);

  if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
    const info = gl.getShaderInfoLog(shader);
    throw new Error(`Could not compile WebGL program. \n\n${info}`);
  }
  return shader;
};

/**
 * Setup shaders.
 *
 * @param {WebGLRenderingContext} gl 
 * @param {WebGLProgram} program 
 */
const setupShaders = (gl, program, vertexShader, fragmentShader) => {
  gl.attachShader(program, createShader(gl, vertexShader.trim(), gl.VERTEX_SHADER));
  gl.attachShader(program, createShader(gl, fragmentShader.trim(), gl.FRAGMENT_SHADER));
  gl.linkProgram(program);
};

/**
 * @type {WebGLProgram}
 */
let activeProgram;

/**
 * Create a WebGL program by compiling its shaders, and returns api to update uniforms.
 * 
 * @param {WebGLRenderingContext} gl 
 * @returns 
 */
const createProgram = (gl) => {
  const program = gl.createProgram();
  setupShaders(gl, program, vertexShaderObjectSrc, fragmentShaderObjectSrc);

  const imageLoc = gl.getUniformLocation(program, "u_image");
  const paletteLoc = gl.getUniformLocation(program, "u_palette");
  const textureSizeLoc = gl.getUniformLocation(program, "u_texture_size");
  const textureLoc = gl.getUniformLocation(program, "u_object_texture");
  const textureDimensionLoc = gl.getUniformLocation(program, "u_tex_dim");
  // const scaleLoc = gl.getUniformLocation(program, "u_scale");
  // const rotateLoc = gl.getUniformLocation(program, "u_rotate");
  // const translateLoc = gl.getUniformLocation(program, "u_translate");
  const transformLoc = gl.getUniformLocation(program, "u_transform");
  // const resolutionLoc = gl.getUniformLocation(program, "u_resolution");

  // const IndexLocation = gl.getAttribLocation(program, "a_");
  // const buffer = new ArrayBuffer();
  // const vbo = gl.createBuffer();
  // gl.bindBuffer(gltexture.ARRAY_BUFFER, vbo);
  // gl.bufferData(gl.ARRAY_BUFFER, buffer, gl.STATIC_DRAW);

  // API to set uniforms on shaders
  const api = {
    use() {
      if (activeProgram !== program) {
        gl.useProgram(program);
        activeProgram = program;
      }
    },
    setPaletteTextureIndex(index) {
      gl.uniform1i(paletteLoc, index);
    },
    setObjectTextureIndex(index) {
      gl.uniform1i(imageLoc, index);
    },
    setTextureDimension(width, height) {
      gl.uniform2f(textureDimensionLoc, width, height);
    },
    setTexturesSize(textureSizes) {
      gl.uniform2fv(textureSizeLoc, textureSizes);
    },
    /** @param {Int32Array} textures */
    setObjectTextures(textures) {
      gl.uniform1iv(textureLoc, textures);
    },
    /** @param {Float32Array} matrices */
    setTransform(matrices) {
      gl.uniformMatrix4fv(transformLoc, false, matrices);
    },
  };

  return api;
};

// const programSetUniform = (program, string, ) => {

// };

/**
 * Setup a unit quad to draw on
 *
 * @param {WebGLRenderingContext} gl 
 */
const setupQuad = (gl) => {
  const positions = [
    1, 1, -1,
    1, -1, -1,
    1, 1, -1,
    -1, 1, -1,
  ];

  const vertBuffer = gl.createBuffer();
  gl.bindBuffer(gl.ARRAY_BUFFER, vertBuffer);
  gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(positions), gl.STATIC_DRAW);
  gl.enableVertexAttribArray(0);
  gl.vertexAttribPointer(0, 2, gl.FLOAT, false, 0, 0);

  return vertBuffer;
};

/**
 * Create a texture {size} pixels wide with all colors in RGB.
 *
 * @param {WebGLRenderingContext} gl 
 * @param {number} index
 * @param {number} size
 * @param {Uint8Array} palette
 */
const createPalette = (gl, index, size, palette) => {
  const buffer = new Uint8Array(size * 4);

  for (let i = 0; i < buffer.length; i = i + 4) {
    if (palette[i] !== undefined) {
      buffer[i] = palette[i];
      buffer[i + 1] = palette[i + 1];
      buffer[i + 2] = palette[i + 2];
      buffer[i + 3] = palette[i + 3];
    } else {
      buffer[i] = 0;
      buffer[i + 1] = 0;
      buffer[i + 2] = 0;
      buffer[i + 3] = 255;
    }
  }

  gl.activeTexture(index);

  const paletteTex = gl.createTexture();
  gl.bindTexture(gl.TEXTURE_2D, paletteTex);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST);
  gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, size, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE, buffer);

  return paletteTex;
};

/**
 * Create the main texture with a single channel.
 *
 * @param {WebGLRenderingContext} gl
 * @param {number} width
 * @param {number} height
 * @returns 
 */
const setupTexture = (gl, width, height) => {
  texture = new Uint8Array(width * height);

  gl.activeTexture(gl.TEXTURE0);

  const tex = gl.createTexture();
  gl.bindTexture(gl.TEXTURE_2D, tex);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST);
  gl.texImage2D(gl.TEXTURE_2D, 0, gl.ALPHA, width, height, 0, gl.ALPHA, gl.UNSIGNED_BYTE, texture);

  return tex;
};

/**
 * Create a texture array from a uint8 buffer.
 *
 * @param {WebGLRenderingContext} gl
 * @param {number} index
 * @param {number} width
 * @param {number} height
 * @param {Uint8Array} buffer
 * @returns 
 */
const createTextureArray = (gl, index, width, height, buffer) => {
  const size = buffer.length / (width * height);

  gl.activeTexture(index);

  const tex = gl.createTexture();
  gl.bindTexture(gl.TEXTURE_2D_ARRAY, tex);
  gl.pixelStorei(gl.UNPACK_ALIGNMENT, 1);
  gl.texParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
  gl.texParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
  gl.texParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_MIN_FILTER, gl.NEAREST);
  gl.texParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_MAG_FILTER, gl.NEAREST);
  gl.texImage3D(gl.TEXTURE_2D_ARRAY, 0, gl.ALPHA, width, height, size, 0, gl.ALPHA, gl.UNSIGNED_BYTE, buffer);

  return tex;
};

/**
 * @param {number} i
 * @param {GLenum} target
 * @param {WebGLTexture} texture 
 */
const useTexture = (i, target, texture) => {
  gl.activeTexture(i);
  gl.bindTexture(target, texture);
}

/**
 * @type {WebGLRenderingContext}
 */
let gl;

/**
 * @type {number}
 */
let _width;
/**
 * @type {number}
 */
let _height;

/**
 * @type {WebGLProgram}
 */
let program;

let vertexBuffer, paletteTex, mainTex;

/**
 * Setup WebGL context.
 *
 * @param {OffscreenCanvas} canvas 
 * @param {Uint8Array} palette 
 * @param {Uint8Array} texture 
 * @param {number} width 
 * @param {number} height 
 */
const setupWebGl = (canvas, palette, width, height) => {
  if (gl) {
    gl.deleteBuffer(vertexBuffer);
    gl.deleteTexture(paletteTex);
    gl.deleteTexture(mainTex);
    gl.deleteProgram(program);
  }

  _width = width;
  _height = height;
  lastFrame = 0;
  skippedFrames = 0;

  gl = canvas.getContext("webgl", { alpha: false, antialias: false });
  program = createProgram(gl);

  vertexBuffer = setupQuad(gl);
  paletteTex = createPalette(gl, palette, 256);
  mainTex = setupTexture(gl, width, height);
};

/**
 * Setup WebGL2 context.
 *
 * @param {OffscreenCanvas} canvas 
 * @param {Uint8Array} palette 
 * @param {Uint8Array} texture 
 * @param {number} width 
 * @param {number} height 
 */
const setupWebGl2 = (canvas) => {
  // _width = canvas.width;
  // _height = canvas.height;
  lastFrame = 0;
  skippedFrames = 0;

  const gl = canvas.getContext("webgl2", { antialias: false });
  gl.enable(gl.BLEND)
  gl.blendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);

  return gl;
};

/**
 * Render a single frame.
 *
 * @param {proto.pubs.Frame} frame 
 * @returns {number}
 */
const renderFrame = (frame) => {
  if (frame.number < lastFrame) {
    lastFrame = frame.number;
    skippedFrames++;
    return lastFrame;
  }

  lastFrame = frame.number;

  /**
   * @type {Uint8Array}
   */
  const data = frame.data;

  switch (frame.format) {
    case Format.RAW:
      texture = data;
      break;
    case Format.RLE: {
      // first 3 bytes header is position of first block
      const offset = 3;
      // first 3 bytes is block length, next byte is color palette
      const stride = 4;
      let pos = data[0] << 16 | data[1] << 8 | data[2];
      for (let i = offset; i < data.length; i = i + stride) {
        const length = data[i] << 16 | data[i + 1] << 8 | data[i + 2];
        for (let p = pos; p < pos + length; p++) {
          texture[p] = data[i + 3];
        }
        pos = pos + length;
      }
      break;
    }
  }

  // Render the new texture
  gl.activeTexture(gl.TEXTURE0);
  gl.texSubImage2D(gl.TEXTURE_2D, 0, 0, 0, _width, _height, gl.ALPHA, gl.UNSIGNED_BYTE, texture);
  gl.drawArrays(gl.TRIANGLES, 0, 6);

  return lastFrame;
};

const MAX_OBJECTS = 1024;
const MAX_OBJECTS_PER_BATCH = 64;

/**
 * 
 * @param {OffscreenCanvas} canvas 
 */
export const createContext = (canvas) => {
  const gl = setupWebGl2(canvas);
  const program = createProgram(gl);
  program.use();

  gl.activeTexture(gl.TEXTURE0);
  gl.clearColor(.0, .0, .0, .0);
  gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);

  const mTextures = Array(MAX_OBJECTS).fill(0);
  const mTransform = Array(MAX_OBJECTS * 16).fill(0);

  const uTextures = new Int32Array(MAX_OBJECTS_PER_BATCH).fill(0);
  const uTransform = new Float32Array(MAX_OBJECTS_PER_BATCH * 16).fill(0);

  return {
    resize(width, height) {
      canvas.width = width;
      canvas.height = height;
      gl.viewport(0, 0, gl.canvas.width, gl.canvas.height);
    },

    /**
     * @param {ArrayBuffer} buffer 
     */
    createPalette(buffer) {
      createPalette(gl, gl.TEXTURE1, 256, new Uint8Array(buffer));
      program.setPaletteTextureIndex(1);
    },

    /**
     * @param {ArrayBuffer} buffer 
     */
    createTexture(buffer) {
      const view = new DataView(buffer);

      // Textures index
      // x: width, y: height
      const uTexturesIndex = Array(256 * 2).fill(0);

      const texW = view.getUint16(0, false);
      const texH = view.getUint16(2, false);
      const texCount = (buffer.byteLength - 4) / (texW * texH * 4);

      const texture = new Uint8Array(texCount * texW * texH);
      for (let i = 0; i < texCount; ++i) {
        const offset = 4 + i * (texW * texH) + i * 4;
        const tw = view.getUint16(offset, false);
        const th = view.getUint16(offset + 2, false);
        const pixels = new Uint8Array(buffer.slice(offset + 4, offset + 4 + texW * texH));
        for (let p = 0; p < pixels.length; ++p) {
          texture[texW * texH * i + p] = pixels[p];
        }
        uTexturesIndex[i * 2] = tw;
        uTexturesIndex[i * 2 + 1] = th;
      }

      createTextureArray(
        gl,
        gl.TEXTURE0,
        texW,
        texH,
        texture,
      );

      program.setObjectTextureIndex(0);
      program.setTexturesSize(uTexturesIndex);
      program.setTextureDimension(texW, texH);
    },

    /**
     * @param {ArrayBuffer} buffer 
     */
    render(buffer) {
      const view = new DataView(buffer);
      const objectCount = view.getUint16(0, false);

      readSceneObjectBuffer(view, 2, (type, block) => {
        switch (type) {
          case SceneObjectBlock.TEXTURE:
            mTextures[block.index] = block.textureIndex;
            break;
          case SceneObjectBlock.MATRIX:
            for (let j = 0; j < 16; ++j) {
              mTransform[(block.index * 16) + j] = block.matrix[j];
            }
            break;
        }
      })

      if (objectCount > 0) {
        // Do one draw call per batch
        const batches = Math.ceil(objectCount / MAX_OBJECTS_PER_BATCH);

        for (let b = 0; b < batches; ++b) {
          const batchObjectCount = b < batches - 1 ? MAX_OBJECTS_PER_BATCH : objectCount % MAX_OBJECTS_PER_BATCH;
          for (let i = 0; i < batchObjectCount; ++i) {
            const objectIndex = (b * MAX_OBJECTS_PER_BATCH) + i;
            uTextures[i] = mTextures[objectIndex];
            for (let j = 0; j < 16; ++j) {
              uTransform[(i * 16) + j] = mTransform[(objectIndex * 16) + j];
            }
          }

          program.setObjectTextures(uTextures);
          program.setTransform(uTransform);

          gl.drawArrays(gl.TRIANGLES, 0, batchObjectCount * 6);
        }
      }
    }
  };
};