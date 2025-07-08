#version 300 es

precision mediump float;

// Attributes
in vec3 a_position;
in mat4 a_model;

// Uniforms
uniform mat4 u_view;

// uniform mat4 lightSpaceMatrix;
// uniform mat4 model;

void main() {
  gl_Position = u_view * a_model * vec4(a_position, 1.0f);
}