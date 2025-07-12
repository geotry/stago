/* Generated file, DO NOT EDIT! */

/**
 * GLSL vertex shader (source: shaders/simpleDepth/vertex.glsl)
 */
const VERTEX_SRC = `
#version 300 es
precision mediump float;
precision highp sampler2DArray;
in vec3 a_position;
in mat4 a_model;
uniform mat4 u_light_space;
void main() {
  vec4 position = vec4(a_position, 1.0f);
  gl_Position = u_light_space * a_model * position;
}
`.trim();
/**
 * GLSL fragment shader (source: shaders/simpleDepth/fragment.glsl)
 */
const FRAGMENT_SRC = `
#version 300 es
void main() {
}
`.trim();

/**
 * @param {WebGL2RenderingContext} gl
 * @param {WebGLProgram} program
 * @returns
 */
const createUniforms = (gl, program) => {
  const locs = {
    [`u_light_space`]: gl.getUniformLocation(program, "u_light_space"),
  };
  const u_light_space = {
    /**
     * Set the value of uniform `u_light_space`.
     *
     * @param {Float32Array} matrix
     * @param {boolean} transpose
     */
    set(matrix, transpose = false) {
      gl.uniformMatrix4fv(locs[`u_light_space`], transpose, matrix);
    },
    /**
     * Returns the value of uniform `u_light_space`.
     *
     * @returns {number}
     */
    get() {
      return gl.getUniform(program, locs[`u_light_space`]);
    },


  };


  return {
    u_light_space,
  };
};

/**
 * The simpleDepth shader program.
 */
export const SimpledepthShader = {
  name: "simpleDepth",
  vertex: VERTEX_SRC,
  fragment: FRAGMENT_SRC,
  createUniforms,
};
