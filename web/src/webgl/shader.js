const { createIndexMap } = require("./utils.js");
const { isUniform } = require("./uniforms.js");

/**
 * Compile a new shader.
 *
 * @param {WebGL2RenderingContext} gl 
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
 * Fetch the fragment and vertex shader text from external files.
 *
 * @param shaderName
 * @returns {Promise<{vertexShaderSource: string | null, fragmentShaderSource: string | null}>}
 */
const loadShader = async (shaderName) => {
  const results = {
    vertexShaderSource: null,
    fragmentShaderSource: null,
  };

  const vertexShaderPath = `/shaders/${shaderName}/vertex.glsl`;
  const fragmentShaderPath = `/shaders/${shaderName}/fragment.glsl`;

  let errors = [];
  await Promise.all([
    fetch(vertexShaderPath)
      .catch((e) => {
        errors.push(e);
      })
      .then(async (response) => {
        if (response.status === 200) {
          results.vertexShaderSource = await response.text();
        } else {
          errors.push(
            `Non-200 response for ${vertexShaderPath}.  ${response.status}:  ${response.statusText}`
          );
        }
      }),

    fetch(fragmentShaderPath)
      .catch((e) => errors.push(e))
      .then(async (response) => {
        if (response.status === 200) {
          results.fragmentShaderSource = await response.text();
        } else {
          errors.push(
            `Non-200 response for ${fragmentShaderPath}.  ${response.status}:  ${response.statusText}`
          );
        }
      }),
  ]);

  if (errors.length !== 0) {
    throw new Error(
      `Failed to fetch shader(s):\n${JSON.stringify(errors, (key, value) => {
        if (value?.constructor.name === 'Error') {
          return {
            name: value.name,
            message: value.message,
            stack: value.stack,
            cause: value.cause,
          };
        }
        return value;
      }, 2)}`
    );
  }
  return results;
};

/**
 * Load, compile and create a shader program.
 *
 * @param {WebGL2RenderingContext} gl 
 * @param {string} shaderName 
 * @returns 
 */
export const loadShaderProgram = async (gl, shaderName) => {
  const shader = await loadShader(shaderName);
  const program = gl.createProgram();

  if (shader.vertexShaderSource) {
    gl.attachShader(program, createShader(gl, shader.vertexShaderSource, gl.VERTEX_SHADER));
  }
  if (shader.fragmentShaderSource) {
    gl.attachShader(program, createShader(gl, shader.fragmentShaderSource, gl.FRAGMENT_SHADER));
  }

  gl.linkProgram(program);
  gl.useProgram(program);

  return program;
};

/**
 *
 * @param {WebGL2RenderingContext} gl 
 * @param {WebGLProgram} program 
 * @param {WebGLVertexArrayObject} vao
 * @param {any[]} globalBuffers
 * @param {any[]} instanceBuffers
 * @param {{}} uniforms
 * @returns 
 */
export const prepareProgram = (gl, program, vao, globalBuffers, instanceBuffers, uniforms) => {
  // Flatten all uniforms in a list
  const uniformList = [];
  const parseUniform = (obj) => {
    if (isUniform(obj)) {
      uniformList.push(obj);
    } else if (Array.isArray(obj)) {
      obj.forEach(item => parseUniform(item));
    } else {
      Object.values(obj).forEach(item => parseUniform(item));
    }
  };
  parseUniform(uniforms);

  // Index of objects in program in the order they are added
  const objectIndex = createIndexMap();
  const pointLightIndex = createIndexMap();
  const spotLightIndex = createIndexMap();

  /** @type {Map<number, Set<number>>} */
  const objects = new Map();
  const recordObject = (id, instanceId) => {
    const exists = objects.has(id);
    const instances = objects.get(id) ?? new Set();
    if (!instances.has(instanceId)) {
      instances.add(instanceId);
    }
    if (!exists) {
      objects.set(id, instances);
    }
  };

  const objectVerticesCount = new Map();
  const recordObjectVertices = (id, verticesCount) => {
    objectVerticesCount.set(objectIndex.get(id), verticesCount);
  };

  /**
   * Render objects handled by this program.
   *
   * @param {number} frame 
   */
  const render = (frame) => {
    gl.useProgram(program);
    gl.bindVertexArray(vao);

    for (const buffer of globalBuffers) {
      buffer.update();
    }
    for (const uniform of uniformList) {
      uniform.use();
    }

    for (const [objectId, instances] of objects) {
      const instanceCount = instances.size;
      if (instanceCount > 0) {
        const index = objectIndex.get(objectId);
        const verticesCount = objectVerticesCount.get(index);
        let skipVertices = 0;
        for (let i = 0; i < index; ++i) {
          skipVertices += objectVerticesCount.get(i);
        }
        for (const buffer of instanceBuffers) {
          buffer.update(index);
        }
        for (const uniform of uniformList) {
          uniform.use(objectId);
        }
        gl.drawArraysInstanced(gl.TRIANGLES, skipVertices, verticesCount, instanceCount);
      }
    }
  };

  return {
    recordObject,
    recordObjectVertices,
    objectIndex,
    pointLightIndex,
    spotLightIndex,
    render,
  };
};