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
// uniform mat4 u_view;

uniform int u_camera_index;
uniform mat4 u_camera[2];

// Note: splitting the projection and view matrices can enable
// more interesting stuff with lights, like computing vectors in view space.
// uniform mat4 u_projection;

// Output
flat out int v_tex_index;
out vec2 v_texcoord;
out vec3 v_normal;
out vec3 v_frag_pos;

void main() {
    vec4 position = vec4(a_position, 1.0f);

    v_texcoord = a_uv;
    v_tex_index = u_tex_index;

    v_normal = a_normal;

    // Note: this is expansive, do it on CPU and put it in VBO
    v_normal = mat3(transpose(inverse(a_model))) * a_normal;

    v_frag_pos = vec3(a_model * position);

    mat4 view = u_camera[u_camera_index];

    gl_Position = view * a_model * position;
}