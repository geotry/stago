#version 300 es

precision mediump float;
precision highp sampler2DArray;

// Attributes
in vec3 a_position;
in vec2 a_uv;
in vec3 a_normal;
in mat4 a_model;

// Uniforms
uniform int u_tex_index;
uniform mat4 u_view;
uniform mat4 u_projection;

// Output
flat out int v_tex_index;
out vec2 v_texcoord;
out vec3 v_normal;
out vec3 v_frag_pos;
out mat4 v_view;

void main() {
    vec4 position = vec4(a_position, 1.0f);
    mat4 viewModel = u_view * a_model;

    v_texcoord = a_uv;
    v_tex_index = u_tex_index;
    // Note: this is expansive, do it on CPU and put it in VBO
    v_normal = mat3(transpose(inverse(viewModel))) * a_normal;
    v_frag_pos = vec3(viewModel * position);
    v_view = u_view;

    gl_Position = u_projection * viewModel * position;
}