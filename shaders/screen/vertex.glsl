#version 300 es

// A simple shader to render a fullscreen quad with a texture (for debugging frame buffers for instance)

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