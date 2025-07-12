#version 300 es

precision mediump float;
precision highp sampler2DArray;

// Attributes
in vec3 a_position;
in mat4 a_model;

// Uniforms
uniform mat4 u_light_space;

void main() {
  vec4 position = vec4(a_position, 1.0f);
  gl_Position = u_light_space * a_model * position;
}