/* Generated file, DO NOT EDIT! */

/**
 * GLSL vertex shader (source: shaders/screen/vertex.glsl)
 */
const VERTEX_SRC = `
#version 300 es
precision mediump float;
out vec2 v_tex_coords;
void main() {
  switch(gl_VertexID) {
    case 0:
      gl_Position = vec4(1.0f, 1.0f, 0.0f, 1.0f);
      v_tex_coords = vec2(1.0f, 0.0f);
      break;
    case 1:
      gl_Position = vec4(-1.0f, 1.0f, 0.0f, 1.0f);
      v_tex_coords = vec2(0.0f, 0.0f);
      break;
    case 2:
      gl_Position = vec4(-1.0f, -1.0f, 0.0f, 1.0f);
      v_tex_coords = vec2(0.0f, 1.0f);
      break;
    case 3:
      gl_Position = vec4(1.0f, 1.0f, 0.0f, 1.0f);
      v_tex_coords = vec2(1.0f, 0.0f);
      break;
    case 4:
      gl_Position = vec4(-1.0f, -1.0f, 0.0f, 1.0f);
      v_tex_coords = vec2(0.0f, 1.0f);
      break;
    case 5:
      gl_Position = vec4(1.0f, -1.0f, 0.0f, 1.0f);
      v_tex_coords = vec2(1.0f, 1.0f);
      break;
  }
}
`.trim();
/**
 * GLSL fragment shader (source: shaders/screen/fragment.glsl)
 */
const FRAGMENT_SRC = `
#version 300 es
precision mediump float;
precision mediump sampler2D;
precision mediump sampler2DShadow;
precision mediump sampler2DArrayShadow;
precision mediump sampler2DArray;
in vec2 v_tex_coords;
out vec4 fragColor;
uniform sampler2DArray u_screen_texture;
void main() {
  float z = texture(u_screen_texture, vec3(v_tex_coords, 0.0f)).r;
  fragColor = vec4(z, z, z, 1.0f);
}
`.trim();

/**
 * @param {WebGL2RenderingContext} gl
 * @param {WebGLProgram} program
 * @returns
 */
const createUniforms = (gl, program) => {
  const locs = {
    [`u_screen_texture`]: gl.getUniformLocation(program, "u_screen_texture"),
  };
  const u_screen_texture = {
    /**
     * Set the value of uniform `u_screen_texture`.
     *
     * @param {number} value
     */
    set(value) {
      gl.uniform1i(locs[`u_screen_texture`],value);
    },
    /**
     * Returns the value of uniform `u_screen_texture`.
     *
     * @returns {number}
     */
    get() {
      return gl.getUniform(program, locs[`u_screen_texture`]);
    },


  };


  return {
    u_screen_texture,
  };
};

/**
 * The screen shader program.
 */
export const ScreenShader = {
  name: "screen",
  vertex: VERTEX_SRC,
  fragment: FRAGMENT_SRC,
  createUniforms,
};
