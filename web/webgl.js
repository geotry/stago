const { readSceneObjectBuffer, Block } = require("./encoding.js");

const MAX_OBJECTS = 1024;
const MAX_VERTICES = MAX_OBJECTS * 12;

const vertexShader = `
#version 300 es

#define MAX_OBJECTS 64
#define OBJECT_VERTICES 6

precision mediump float;
precision highp sampler2DArray;

// Attributes
in vec3 a_position;
in vec2 a_texuv;
in mat4 a_model;
in int a_tex_index;

// Uniforms
uniform int u_tex_index;
uniform mat4 u_view; // View matrix

// Output
flat out int v_tex_index;
out vec2 v_texcoord;

void main() {
  vec4 position = vec4(a_position, 1.0);

  v_texcoord = a_texuv;
  v_tex_index = a_tex_index; // Attribute is broken, uniform works
  v_tex_index = u_tex_index;

  gl_Position = u_view * a_model * position;
}`;

const fragmentShader = `
#version 300 es

precision mediump float;
precision highp sampler2DArray;

in vec2 v_texcoord;
flat in int v_tex_index;

out vec4 fragColor;

uniform sampler2DArray u_image;
// uniform sampler2DArray u_normal;
uniform sampler2D u_palette;

void main() {
    float index = texture(u_image, vec3(v_texcoord, v_tex_index)).a;
    vec4 color = texture(u_palette, vec2(index, 0));
    // vec3 normal = texture(u_normal, vec3(v_texcoord, v_tex_index)).rgb;
    // normal = normalize(normal * 2.0 - 1.0);

    float depth = 1.0  - (gl_FragCoord.z / gl_FragCoord.w) * .1;
    color = color * vec4(depth, depth, depth, 1.0);

    fragColor = color;
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
 * 
 * @param {WebGL2RenderingContext} gl 
 * @param {number} type 
 * @returns 
 */
const typeByteSize = (gl, type) => {
  switch (type) {
    case gl.BYTE:
    case gl.UNSIGNED_BYTE:
      return 1;
    case gl.UNSIGNED_SHORT:
    case gl.SHORT:
    case gl.HALF_FLOAT:
      return 2;
    case gl.FLOAT:
    case gl.INT:
    case gl.UNSIGNED_INT:
      return 4;
  }
  throw new Error(`Unknown type ${type}`);
};

/**
 * 
 * @param {WebGL2RenderingContext} gl 
 * @param {number} type 
 * @returns 
 */
const typeIsInt = (gl, type) => {
  switch (type) {
    case gl.BYTE:
    case gl.UNSIGNED_BYTE:
    case gl.UNSIGNED_SHORT:
    case gl.SHORT:
    case gl.INT:
    case gl.UNSIGNED_INT:
      return true;
  }
  return false;
};

/**
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
const createArrayBuffer = (gl, program, size, usage, attributes) => {
  const stride = attributes.reduce((prev, curr) => prev + typeByteSize(gl, curr.type) * curr.size * (curr.repeat ?? 1), 0);

  const buffer = gl.createBuffer();
  gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
  gl.bufferData(gl.ARRAY_BUFFER, size * stride, usage);

  console.log(`Created ARRAY_BUFFER of size ${size}x${stride} = ${size * stride}`);

  let offset = 0;
  for (const attr of attributes) {
    const loc = gl.getAttribLocation(program, attr.name);
    const iterations = attr.repeat ?? 1;
    for (let i = 0; i < iterations; ++i) {
      const attrLoc = loc + i;
      console.log(`attribute ${attr.name} size=${attr.size} type=${attr.type} stride=${stride} offset=${offset}`);
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

  return buffer;
};


/**
 * @type {WebGLProgram}
 */
let activeProgram;

/**
 * Create a WebGL program by compiling its shaders, and returns api to update uniforms.
 * 
 * @param {WebGL2RenderingContext} gl 
 * @returns 
 */
const createProgram = (gl) => {
  const program = gl.createProgram();

  gl.attachShader(program, createShader(gl, vertexShader.trim(), gl.VERTEX_SHADER));
  gl.attachShader(program, createShader(gl, fragmentShader.trim(), gl.FRAGMENT_SHADER));
  gl.linkProgram(program);

  const objectIdIndexMap = new Map();
  const objectIndexIdMap = new Map();
  const getObjectIndex = (objectId) => {
    if (!objectIdIndexMap.has(objectId)) {
      const index = objectIdIndexMap.size;
      objectIdIndexMap.set(objectId, index);
      objectIndexIdMap.set(index, objectId);
    }
    return objectIdIndexMap.get(objectId);
  };

  const getObjectId = (objectIndex) => {
    return objectIndexIdMap.get(objectIndex);
  };

  const objectVerticeCount = new Map();

  // Returns the associated instance id (gl_InstanceID)
  const instanceIdMap = new Map();
  const instanceObjectIdCounter = new Map();
  const getInstanceId = (id, objectId) => {
    if (!instanceIdMap.has(id)) {
      let instanceCounter = instanceObjectIdCounter.get(objectId) ?? 0;
      instanceCounter++;
      instanceObjectIdCounter.set(objectId, instanceCounter);
      instanceIdMap.set(id, instanceCounter - 1);
    }
    return instanceIdMap.get(id);
  };

  const getInstanceCount = (objectId) => {
    return instanceObjectIdCounter.get(objectId);
  };

  const vao = gl.createVertexArray();
  gl.bindVertexArray(vao);

  // Create buffers

  const vertexBuffer = createArrayBuffer(gl, program, MAX_VERTICES, gl.STATIC_DRAW, [
    {
      name: "a_position",
      type: gl.FLOAT,
      size: 3,
    },
    {
      name: "a_texuv",
      type: gl.FLOAT,
      size: 2,
    },
  ]);

  const vertexBufferData = new Float32Array(MAX_VERTICES * 5);
  /** @type {Float32Array[]} */
  const vertexBufferDataPositionView = [];
  /** @type {Float32Array[]} */
  const vertexBufferDataUVView = [];
  for (let i = 0; i < MAX_VERTICES; ++i) {
    vertexBufferDataPositionView.push(new Float32Array(vertexBufferData.buffer, i * 20, 3));
    vertexBufferDataUVView.push(new Float32Array(vertexBufferData.buffer, 12 + i * 20, 2));
  }

  let vertexBufferChanged = false;
  const objectIndexVertices = new Map();

  /**
   * 
   * @param {number} objectId 
   * @param {Float32Array} vertices 
   */
  const updateVertexBufferVertices = (objectId, vertices) => {
    const objectIndex = getObjectIndex(objectId);
    let offset = 0;
    for (const [idx, vertexCount] of objectIndexVertices.entries()) {
      if (idx === objectIndex) {
        break;
      }
      offset += vertexCount;
    }
    for (let i = 0; i < vertices.length / 3; i++) {
      vertexBufferDataPositionView[offset + i].set(vertices.slice(i * 3, i * 3 + 3));
    }
    if (!objectIndexVertices.has(objectIndex)) {
      objectIndexVertices.set(objectIndex, vertices.length / 3);
    }
    objectVerticeCount.set(objectId, vertices.length / 3);
    vertexBufferChanged = true;
  };

  const updateVertexBufferUV = (objectId, uv) => {
    const objectIndex = getObjectIndex(objectId);
    let offset = 0;
    for (const [idx, vertexCount] of objectIndexVertices.entries()) {
      if (idx === objectIndex) {
        break;
      }
      offset += vertexCount;
    }
    for (let i = 0; i < uv.length / 2; i++) {
      vertexBufferDataUVView[offset + i].set(uv.slice(i * 2, i * 2 + 2));
    }
    if (!objectIndexVertices.has(objectIndex)) {
      objectIndexVertices.set(objectIndex, uv.length / 2);
    }
    vertexBufferChanged = true;
  };

  const modelBuffer = createArrayBuffer(gl, program, MAX_OBJECTS, gl.DYNAMIC_DRAW, [
    {
      name: "a_model",
      type: gl.FLOAT,
      size: 4,
      repeat: 4,
      instance: 1,
    },
  ]);
  // Create a model buffer data to store all model matrices
  const modelBufferData = new Float32Array(MAX_OBJECTS * MAX_OBJECTS * 16);
  // Create view to point to slices of model matrices indexed by object index
  /** @type {Float32Array[]} */
  const modelBufferDataView = [];
  for (let i = 0; i < MAX_OBJECTS; ++i) {
    modelBufferDataView.push(new Float32Array(modelBufferData.buffer, MAX_OBJECTS * i * 16 * 4, MAX_OBJECTS * 16));
  }

  let modelBufferChanged = false;

  const updateModelMatrix = (id, objectId, matrix) => {
    const objectIndex = getObjectIndex(objectId);
    const instanceId = getInstanceId(id, objectId);
    for (let i = 0; i < matrix.length; i++) {
      modelBufferDataView[objectIndex][instanceId * 16 + i] = matrix[i];
    }
    modelBufferChanged = true;
  };

  const textureIndexBuffer = createArrayBuffer(gl, program, MAX_OBJECTS, gl.STATIC_DRAW, [
    {
      name: "a_tex_index",
      type: gl.INT,
      size: 1,
      instance: 1,
    },
  ]);

  const textureIndexBufferData = new Int32Array(MAX_OBJECTS);
  let textureIndexBufferChanged = false;

  /**
   * 
   * @param {number} id 
   * @param {number} textureIndex 
   */
  const updateTextureIndexBuffer = (id, textureIndex) => {
    textureIndexBufferData[getObjectIndex(id)] = textureIndex;
    textureIndexBufferChanged = true;
  };

  const uniforms = new Map();
  const getUniform = (name) => {
    let uniform = uniforms.get(name);
    if (!uniform) {
      uniform = gl.getUniformLocation(program, name);
      uniforms.set(name, uniform);
    }
    return uniform;
  };

  const cameraBufferData = new Float32Array(16 * 2);
  const cameraBufferDataView = [];
  for (let i = 0; i < 2; ++i) {
    cameraBufferDataView.push(new Float32Array(cameraBufferData.buffer, i * 16 * 4, 16));
  }
  const objectCameraIndex = new Map();
  const updateCameraMatrix = (cameraIndex, matrix) => {
    for (let i = 0; i < matrix.length; ++i) {
      cameraBufferDataView[cameraIndex][i] = matrix[i];
    }
  };
  const updateObjectCameraIndex = (objectId, cameraIndex) => {
    objectCameraIndex.set(objectId, cameraIndex);
  };
  const getObjectCameraMatrix = (objectId) => {
    const index = objectCameraIndex.get(objectId);
    if (cameraBufferDataView[index] === undefined) {
      console.error(`camera index not found`, index, objectId, cameraBufferDataView);
    }
    return cameraBufferDataView[index];
  };

  const api = {
    debug() { },
    use() {
      if (activeProgram !== program) {
        gl.useProgram(program);
        activeProgram = program;
      }
    },

    reset() {
      objectIdIndexMap.clear();
      objectIndexIdMap.clear();
      instanceIdMap.clear();
      instanceObjectIdCounter.clear();
      objectVerticeCount.clear();
      objectIndexVertices.clear();
      // objectCameraIndex.clear();
      // vertexBufferData.fill(0);
      // textureIndexBufferData.fill(0);
      // modelBufferData.fill(0);
    },

    /**
     * Bind a texture index to a uniform sampler.
     *
     * @param {string} name 
     * @param {number} index 
     */
    bindTexture(name, index) {
      gl.uniform1i(getUniform(name), index);
    },

    /**
     * Update an object.
     * 
     * @param {number} id 
     * @param {number} textureId 
     * @param {number} textureIndex 
     * @param {number} cameraIndex 
     * @param {Float32Array} vertices 
     * @param {Float32Array} uv
     */
    updateObject(id, textureId, textureIndex, cameraIndex, vertices, uv) {
      updateVertexBufferVertices(id, vertices);
      updateVertexBufferUV(id, uv);
      updateTextureIndexBuffer(id, textureIndex);
      updateObjectCameraIndex(id, cameraIndex);
    },

    /**
     * Update camera matrices.
     *
     * @param {Float32Array} ortho 
     * @param {Float32Array} perspective 
     */
    updateCamera(ortho, perspective) {
      updateCameraMatrix(0, ortho);
      updateCameraMatrix(1, perspective);
    },

    /**
     * Update an object instance.
     * 
     * @param {number} id 
     * @param {number} objectId 
     * @param {Float32Array} model
     */
    updateObjectInstance(id, objectId, model) {
      updateModelMatrix(id, objectId, model);
    },

    /**
     * Update the vertex buffer object, set uniforms.
     * 
     * @param {number} frame 
     */
    render(frame) {
      gl.bindVertexArray(vao);

      if (vertexBufferChanged) {
        gl.bindBuffer(gl.ARRAY_BUFFER, vertexBuffer);
        gl.bufferSubData(gl.ARRAY_BUFFER, 0, vertexBufferData);
        vertexBufferChanged = false;
      }
      if (textureIndexBufferChanged) {
        gl.bindBuffer(gl.ARRAY_BUFFER, textureIndexBuffer);
        gl.bufferSubData(gl.ARRAY_BUFFER, 0, textureIndexBufferData);
        textureIndexBufferChanged = false;
      }

      // Iterate on each scene object, and make one drawCallinstanced for each
      for (const [objectId, objectIndex] of objectIdIndexMap) {
        const instanceCount = getInstanceCount(objectId);
        if (instanceCount > 0) {
          // Bind the slice of model buffer data for this object index
          gl.bindBuffer(gl.ARRAY_BUFFER, modelBuffer);
          gl.bufferSubData(gl.ARRAY_BUFFER, 0, modelBufferDataView[objectIndex]);

          gl.uniformMatrix4fv(getUniform("u_view"), false, getObjectCameraMatrix(objectId));
          gl.uniform1i(getUniform("u_tex_index"), textureIndexBufferData.at(objectIndex));

          let skipVertices = 0;
          for (let i = 0; i < objectIndex; ++i) {
            skipVertices += objectVerticeCount.get(getObjectId(i));
          }

          const verticesCount = objectVerticeCount.get(objectId);

          gl.drawArraysInstanced(gl.TRIANGLES, skipVertices, verticesCount, instanceCount);

          // Debug buffers
          if (frame === 1) {
            console.log(`DRAW_OBJECT ${objectId} x${instanceCount}`);
            for (let i = 0; i < verticesCount; ++i) {
              const vertexIndex = skipVertices + i;
              const positionX = vertexBufferData.at(vertexIndex * 5);
              const positionY = vertexBufferData.at(vertexIndex * 5 + 1);
              const positionZ = vertexBufferData.at(vertexIndex * 5 + 2);
              const uvX = vertexBufferData.at(vertexIndex * 5 + 3);
              const uvY = vertexBufferData.at(vertexIndex * 5 + 4);

              const textureIndex = textureIndexBufferData.at(objectIndex);

              console.log(`DRAW_VERTEX index=${vertexIndex} position=(${positionX}, ${positionY}, ${positionZ}) uv=(${uvX}, ${uvY}) texture_index=${textureIndex}`);
            }
            console.log(`=======================`);
          }
        }
      }
    },
  };

  return api;
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
const createTexture = (gl, index, format, width, depth, pixels) => {
  let pixelSize;
  let glFormat;
  switch (format) {
    case 0:
      pixelSize = 1;
      glFormat = gl.ALPHA;
      break;
    case 1:
      pixelSize = 3;
      glFormat = gl.RGB;
      break;
    case 2:
      pixelSize = 4;
      glFormat = gl.RGBA;
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
    gl.texImage3D(gl.TEXTURE_2D_ARRAY, 0, glFormat, width, height, depth, 0, glFormat, gl.UNSIGNED_BYTE, pixels);
  } else {
    gl.bindTexture(gl.TEXTURE_2D, texture);
    gl.pixelStorei(gl.UNPACK_ALIGNMENT, 1);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST);
    gl.texImage2D(gl.TEXTURE_2D, 0, glFormat, width, height, 0, glFormat, gl.UNSIGNED_BYTE, pixels);
  }

  return texture;
};

/**
 * @type {WebGL2RenderingContext}
 */
let gl;

/**
 * Setup WebGL2 context.
 *
 * @param {OffscreenCanvas} canvas 
 */
const setupWebGl2 = (canvas) => {
  const gl = canvas.getContext("webgl2", { antialias: false });
  gl.enable(gl.BLEND)
  gl.blendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);

  gl.enable(gl.DEPTH_TEST);
  gl.depthFunc(gl.LESS);

  return gl;
};

/**
 * 
 * @param {OffscreenCanvas} canvas 
 */
export const createContext = (canvas) => {
  const gl = setupWebGl2(canvas);
  const program = createProgram(gl);
  program.use();

  gl.activeTexture(gl.TEXTURE1);
  gl.clearColor(.0, .0, .0, .0);
  gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);

  let renderCount = 0;

  return {
    program,
    reset() {
      program.reset();
    },
    resize(width, height) {
      canvas.width = width;
      canvas.height = height;
      gl.viewport(0, 0, gl.canvas.width, gl.canvas.height);
    },
    /**
     * @param {ArrayBuffer} buffer 
     */
    render(buffer) {
      renderCount++;
      const view = new DataView(buffer);

      readSceneObjectBuffer(view, renderCount, {
        onTextureUpdated(t) {
          const index = t.id - 1;
          createTexture(gl, index, t.format, t.width, t.depth, t.pixels);
          switch (index) {
            case 0:
              program.bindTexture("u_palette", index);
              break;
            case 1:
              program.bindTexture("u_image", index);
              break;
          }
        },
        onSceneObjectUpdated(o) {
          program.updateObject(o.id, o.textureId, o.textureIndex, o.isUI ? 0 : 1, o.vertices, o.uv);
        },
        onCameraUpdated(c) {
          program.updateCamera(c.ortho, c.perspective);
        },
        onSceneObjectInstanceUpdated(i) {
          program.updateObjectInstance(i.id, i.objectId, i.model);
        }
      });

      program.render(renderCount);
    }
  };
};