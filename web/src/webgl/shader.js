
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

  gl.attachShader(program, createShader(gl, shader.vertexShaderSource, gl.VERTEX_SHADER));
  gl.attachShader(program, createShader(gl, shader.fragmentShaderSource, gl.FRAGMENT_SHADER));
  gl.linkProgram(program);
  gl.useProgram(program);

  return program;
};

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