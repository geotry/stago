/**
 * @typedef {{set: (data: T) => void}} UniformApi<T>
 * @template {string} T
 */

/**
 * @typedef {{set: (x: T, y: T) => void}} Uniform2Api<T>
 * @template {string} T
 */

/**
 * @typedef {{set: (x: T, y: T, z: T) => void}} Uniform3Api<T>
 * @template {string} T
 */

/**
 * 
 * @param {WebGL2RenderingContext} gl 
 * @param {WebGLProgram} program 
 * @param {string} name
 */
const getLocation = (gl, program, name) => {
  const loc = gl.getUniformLocation(program, name);
  if (loc === null) {
    throw new Error(`Uniform ${name} does not exist in shader program`);
  }
  return loc;
}

/**
 *
 * @param {WebGL2RenderingContext} gl 
 * @param {WebGLProgram} program 
 * @param {string} name
 * @returns {UniformApi<GLint>}
 */
const create1i = (gl, program, name) => {
  return createUniform(gl, program, name, (loc, value) => { gl.uniform1i(loc, value[0]); });
};

/**
 * 
 * @param {WebGL2RenderingContext} gl 
 * @param {WebGLProgram} program 
 * @param {string} name
 * @returns {UniformApi<GLfloat>}
 */
const create1f = (gl, program, name) => {
  return createUniform(gl, program, name, (loc, value) => { gl.uniform1f(loc, value[0]); });
};

/**
 * 
 * @param {WebGL2RenderingContext} gl 
 * @param {WebGLProgram} program 
 * @param {string} name
 * @returns {Uniform2Api<GLfloat>}
 */
const create2f = (gl, program, name) => {
  return createUniform(gl, program, name, (loc, value) => { gl.uniform2f(loc, value[0], value[1]); });
};

/**
 * 
 * @param {WebGL2RenderingContext} gl 
 * @param {WebGLProgram} program 
 * @param {string} name
 * @returns {Uniform3Api<GLfloat>}
 */
const create3f = (gl, program, name) => {
  return createUniform(gl, program, name, (loc, value) => { gl.uniform3f(loc, value[0], value[1], value[2]); });
};

/**
 * 
 * @param {WebGL2RenderingContext} gl 
 * @param {WebGLProgram} program 
 * @param {string} name
 * @returns {UniformApi<Float32List>}
 */
const createMatrix4fv = (gl, program, name) => {
  return createUniform(gl, program, name, (loc, value) => { gl.uniformMatrix4fv(loc, false, value[0]); });
};

const UniformSymbol = Symbol();

export const isUniform = (v) => typeof v === "object" && v !== null && v[UniformSymbol] !== undefined;

/**
 * 
 * @param {WebGL2RenderingContext} gl 
 * @param {WebGLProgram} program 
 * @param {string} name
 * @param {(loc: WebGLUniformLocation, value: any[]) => void} set
 * @returns
 */
const createUniform = (gl, program, name, set) => {
  const loc = getLocation(gl, program, name);
  const preparedValues = new Map();
  let tmpValue;

  return {
    [UniformSymbol]: name,
    name,
    set(...data) {
      tmpValue = data;
    },
    prepare(key, ...data) {
      preparedValues.set(key, data);
    },
    value(key) {
      if (key !== undefined) {
        return preparedValues.get(key);
      } else {
        return tmpValue;
      }
    },
    use(key) {
      if (key !== undefined) {
        if (preparedValues.has(key)) {
          set(loc, preparedValues.get(key));
        }
      } else if (tmpValue !== undefined) {
        set(loc, tmpValue);
        tmpValue = undefined;
      }
    },
  }
};

export const uniforms = {
  create1i,
  create1f,
  create2f,
  create3f,
  createMatrix4fv,
};
