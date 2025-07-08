const { readSceneObjectBuffer } = require("./encoding.js");
const { loadShaderProgram } = require("./webgl/shader.js");
const { createArrayBuffer, createDepthMapBuffer, createTexture } = require("./webgl/buffers.js");

const MAX_OBJECTS = 1024;
const MAX_VERTICES = MAX_OBJECTS * 12;

const createObjectStore = () => {
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
  const objectIndexVertices = new Map();

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


  const vertexBufferData = new Float32Array(MAX_VERTICES * 8);
  let vertexBufferChanged = false;
  /** @type {Float32Array[]} */
  const vertexBufferDataPosition = [];
  /** @type {Float32Array[]} */
  const vertexBufferDataUV = [];
  /** @type {Float32Array[]} */
  const vertexBufferDataNormal = [];
  for (let i = 0; i < MAX_VERTICES; ++i) {
    vertexBufferDataPosition.push(new Float32Array(vertexBufferData.buffer, i * 32, 3));
    vertexBufferDataUV.push(new Float32Array(vertexBufferData.buffer, 12 + i * 32, 2));
    vertexBufferDataNormal.push(new Float32Array(vertexBufferData.buffer, 20 + i * 32, 3));
  }

  /**
   * 
   * @param {number} objectId 
   * @param {Float32Array} vertices 
   * @param {Float32Array} uv 
   * @param {Float32Array} normals 
   */
  const updateVertexBufferVertices = (objectId, vertices, uv, normals) => {
    const objectIndex = getObjectIndex(objectId);
    let offset = 0;
    for (const [idx, vertexCount] of objectIndexVertices.entries()) {
      if (idx === objectIndex) {
        break;
      }
      offset += vertexCount;
    }
    objectIndexVertices.set(objectIndex, vertices.length / 3);
    objectVerticeCount.set(objectId, vertices.length / 3);

    for (let i = 0; i < vertices.length / 3; i++) {
      vertexBufferDataPosition[offset + i].set(vertices.slice(i * 3, i * 3 + 3));
    }
    for (let i = 0; i < uv.length / 2; i++) {
      vertexBufferDataUV[offset + i].set(uv.slice(i * 2, i * 2 + 2));
    }
    for (let i = 0; i < normals.length / 3; i++) {
      vertexBufferDataNormal[offset + i].set(normals.slice(i * 3, i * 3 + 3));
    }

    vertexBufferChanged = true;
  };

  // Create a model buffer data to store all model matrices
  const modelBufferData = new Float32Array(MAX_OBJECTS * MAX_OBJECTS * 16);
  let modelBufferChanged = false;
  // Create view to point to slices of model matrices indexed by object index
  /** @type {Float32Array[]} */
  const modelBufferDataView = [];
  for (let i = 0; i < MAX_OBJECTS; ++i) {
    modelBufferDataView.push(new Float32Array(modelBufferData.buffer, MAX_OBJECTS * i * 16 * 4, MAX_OBJECTS * 16));
  }

  const updateModelMatrix = (id, objectId, matrix) => {
    const objectIndex = getObjectIndex(objectId);
    const instanceId = getInstanceId(id, objectId);
    for (let i = 0; i < matrix.length; i++) {
      modelBufferDataView[objectIndex][instanceId * 16 + i] = matrix[i];
    }
    modelBufferChanged = true;
  };

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

  // Camera
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

  return {
    getObjectIndex,
    getObjectId,
    getInstanceId,
    getInstanceCount,
    objects: () => objectIdIndexMap,
    objectVerticeCount: (id) => objectVerticeCount.get(id),

    updateVertexBufferVertices,
    vertexBufferData,
    vertexBufferChanged: () => vertexBufferChanged,
    updateModelMatrix,
    modelBufferData: modelBufferDataView,
    modelBufferChanged: () => modelBufferChanged,
    updateTextureIndexBuffer,
    textureIndexBufferData,
    textureIndexBufferChanged: () => textureIndexBufferChanged,

    refresh() {
      vertexBufferChanged = false;
      modelBufferChanged = false;
      textureIndexBufferChanged = false;
    },

    updateCameraMatrix,
    updateObjectCameraIndex,
    getObjectCameraMatrix,

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

    debug() {
      // Debug buffers
      // if (frame === 1) {
      //   console.log(`DRAW_OBJECT ${objectId} x${instanceCount}`);
      //   for (let i = 0; i < verticesCount; ++i) {
      //     const vertexIndex = skipVertices + i;
      //     const positionX = vertexBufferData.at(vertexIndex * 5);
      //     const positionY = vertexBufferData.at(vertexIndex * 5 + 1);
      //     const positionZ = vertexBufferData.at(vertexIndex * 5 + 2);
      //     const uvX = vertexBufferData.at(vertexIndex * 5 + 3);
      //     const uvY = vertexBufferData.at(vertexIndex * 5 + 4);

      //     const textureIndex = textureIndexBufferData.at(objectIndex);

      //     console.log(`DRAW_VERTEX index=${vertexIndex} position=(${positionX}, ${positionY}, ${positionZ}) uv=(${uvX}, ${uvY}) texture_index=${textureIndex}`);
      //   }
      //   console.log(`=======================`);
      // }
    }
  };
};

/**
 * @type {WebGLProgram}
 */
let activeProgram;

const createStandardProgram = async (gl) => {
  const program = await loadShaderProgram(gl, "standard");

  return wrapProgram(gl, program);
}

const createDepthProgram = async (gl) => {
  const program = await loadShaderProgram(gl, "depth");

  return wrapProgram(gl, program);
}


/**
 * Create a WebGL program by compiling its shaders, and returns api to update uniforms.
 * 
 * @param {WebGL2RenderingContext} gl 
 * @param {WebGLProgram} program 
 * @returns 
 */
const wrapProgram = (gl, program) => {
  const uniforms = new Map();
  const getUniform = (name) => {
    let uniform = uniforms.get(name);
    if (!uniform) {
      uniform = gl.getUniformLocation(program, name);
      uniforms.set(name, uniform);
    }
    return uniform;
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
    {
      name: "a_normal",
      type: gl.FLOAT,
      size: 3,
    },
  ]);

  const modelBuffer = createArrayBuffer(gl, program, MAX_OBJECTS, gl.DYNAMIC_DRAW, [
    {
      name: "a_model",
      type: gl.FLOAT,
      size: 4,
      repeat: 4,
      instance: 1,
    },
  ]);

  const textureIndexBuffer = createArrayBuffer(gl, program, MAX_OBJECTS, gl.STATIC_DRAW, [
    {
      name: "a_tex_index",
      type: gl.INT,
      size: 1,
      instance: 1,
    },
  ]);

  const api = {
    getUniform,
    use() {
      if (activeProgram !== program) {
        gl.useProgram(program);
        activeProgram = program;
      }
    },

    /**
     * Bind a texture index to a uniform sampler.
     *
     * @param {string} name 
     * @param {number} index 
     */
    bindTexture(name, index) {
      const loc = getUniform(name);
      if (loc) {
        gl.uniform1i(loc, index);
      }
    },

    /**
     * Update the vertex buffer object, set uniforms.
     * 
     * @param store
     * @param {number} frame 
     */
    render(store, frame) {
      if (activeProgram !== program) {
        gl.useProgram(program);
        activeProgram = program;
      }

      gl.bindVertexArray(vao);

      if (store.vertexBufferChanged()) {
        gl.bindBuffer(gl.ARRAY_BUFFER, vertexBuffer);
        gl.bufferSubData(gl.ARRAY_BUFFER, 0, store.vertexBufferData);
      }
      if (store.textureIndexBufferChanged()) {
        gl.bindBuffer(gl.ARRAY_BUFFER, textureIndexBuffer);
        gl.bufferSubData(gl.ARRAY_BUFFER, 0, store.textureIndexBufferData);
      }

      gl.uniform1f(getUniform("u_material.shininess"), 32.0);

      gl.uniform3f(getUniform("u_view_pos"), 0, 0, 0);

      // Directional light
      gl.uniform3f(getUniform("u_dir_light.direction"), -0.2, -1.0, -0.3);
      gl.uniform3f(getUniform("u_dir_light.ambient"), 0.002, 0.002, 0.002);
      gl.uniform3f(getUniform("u_dir_light.diffuse"), 0.5, 0.5, 0.5);
      gl.uniform3f(getUniform("u_dir_light.specular"), 1.0, 1.0, 1.0);
      gl.uniform1f(getUniform("u_dir_light.intensity"), 0.1);

      gl.uniform1i(getUniform("u_point_light_count"), 1);
      gl.uniform3f(getUniform("u_point_light[0].position"), Math.abs(Math.cos(frame / 200) * 100), -8, Math.abs(Math.sin(frame / 200) * 100));
      gl.uniform3f(getUniform("u_point_light[0].ambient"), 0.002, 0.002, 0.002);
      gl.uniform3f(getUniform("u_point_light[0].diffuse"), 0, 0, 1.0);
      gl.uniform3f(getUniform("u_point_light[0].specular"), 0, 0, 1.0);
      gl.uniform1f(getUniform("u_point_light[0].radius"), 10.0);
      gl.uniform1f(getUniform("u_point_light[0].intensity"), 1.0);

      // Point light
      // const frequency = 50;
      // // Spot
      // gl.uniform3f(getUniform("u_light.position"), 0, 5, 0);
      // gl.uniform3f(getUniform("u_light.direction"), .2, -.4, .6);
      // gl.uniform1f(getUniform("u_light.cut_off"), Math.cos(12.5 * Math.PI / 180));
      // gl.uniform1f(getUniform("u_light.outer_cut_off"), Math.cos(17.5 * Math.PI / 180));

      // Iterate on each scene object, and make one drawCallinstanced for each
      for (const [objectId, objectIndex] of store.objects()) {
        const instanceCount = store.getInstanceCount(objectId);
        if (instanceCount > 0) {
          // Bind the slice of model buffer data for this object index
          gl.bindBuffer(gl.ARRAY_BUFFER, modelBuffer);
          gl.bufferSubData(gl.ARRAY_BUFFER, 0, store.modelBufferData[objectIndex]);

          // todo: light depth shader should use the light source view matrix instead
          gl.uniformMatrix4fv(getUniform("u_view"), false, store.getObjectCameraMatrix(objectId));
          gl.uniform1i(getUniform("u_tex_index"), store.textureIndexBufferData.at(objectIndex));

          let skipVertices = 0;
          for (let i = 0; i < objectIndex; ++i) {
            skipVertices += store.objectVerticeCount(store.getObjectId(i));
          }

          const verticesCount = store.objectVerticeCount(objectId);

          gl.drawArraysInstanced(gl.TRIANGLES, skipVertices, verticesCount, instanceCount);
        }
      }
    },
  };

  return api;
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
  return gl;
};

/**
 * 
 * @param {OffscreenCanvas} canvas 
 */
export const createContext = async (canvas) => {
  const gl = setupWebGl2(canvas);
  gl.enable(gl.DEPTH_TEST);
  gl.depthFunc(gl.LESS);

  gl.enable(gl.BLEND)
  gl.blendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);

  gl.clearColor(.0, .0, .0, .0);
  gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);

  const std = await createStandardProgram(gl);
  const depth = await createDepthProgram(gl);

  std.use();

  const SHADOW_SIZE = 1024;

  const [depthMapBuffer, depthMap] = createDepthMapBuffer(gl, SHADOW_SIZE);

  const store = createObjectStore();

  let renderCount = 0;

  return {
    reset() {
      store.reset();
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
          createTexture(gl, t.role, t.format, t.width, t.depth, t.pixels);
          switch (t.role) {
            case 0:
              std.bindTexture("u_material.diffuse", t.role);
              break;
            case 1:
              std.bindTexture("u_palette", t.role);
              break;
            case 2:
              std.bindTexture("u_material.specular", t.role);
              break;
            case 3:
              std.bindTexture("u_material.normal", t.role);
              break;
          }
        },
        onSceneObjectUpdated(o) {
          store.updateVertexBufferVertices(o.id, o.vertices, o.uv, o.normals);
          store.updateTextureIndexBuffer(o.id, o.textureIndex);
          store.updateObjectCameraIndex(o.id, o.isUI ? 0 : 1);
        },
        onCameraUpdated(c) {
          store.updateCameraMatrix(0, c.ortho);
          store.updateCameraMatrix(1, c.perspective);
        },
        onSceneObjectInstanceUpdated(i) {
          store.updateModelMatrix(i.id, i.objectId, i.model);
        }
      });

      // Render depth map texture for shadows
      depth.use();
      gl.viewport(0, 0, SHADOW_SIZE, SHADOW_SIZE);
      gl.bindFramebuffer(gl.FRAMEBUFFER, depthMapBuffer);
      gl.clear(gl.DEPTH_BUFFER_BIT);
      // todo: use light view matrix
      depth.render(store, renderCount);
      gl.bindFramebuffer(gl.FRAMEBUFFER, null);

      // Render normal scene
      std.use();
      gl.viewport(0, 0, gl.canvas.width, gl.canvas.height);
      std.render(store, renderCount);

      store.refresh();
    },
  };
};