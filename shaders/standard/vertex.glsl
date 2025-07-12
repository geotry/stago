#version 300 es

#define MAX_LIGHTS 10

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

uniform mat4 u_directional_light_space;
uniform mat4 u_spot_light_space[MAX_LIGHTS];
uniform int u_spot_light_count;

// Output
flat out int v_tex_index;
out vec2 v_texcoord;
out vec3 v_normal;
out vec3 v_frag_pos;
out mat4 v_view;

out vec4 v_frag_pos_directional_light_space;

flat out int v_spot_light_count;
out vec4 v_frag_pos_spot_light_space[MAX_LIGHTS];

void main() {
    vec4 position = vec4(a_position, 1.0f);
    mat4 viewModel = u_view * a_model;

    v_texcoord = a_uv;
    v_tex_index = u_tex_index;
    v_view = u_view;
    // Note: this is expansive, do it on CPU and put it in VBO
    v_normal = mat3(transpose(inverse(viewModel))) * a_normal;
    v_frag_pos = vec3(viewModel * position);

    v_frag_pos_directional_light_space = u_directional_light_space * a_model * position;
    v_spot_light_count = u_spot_light_count;
    for(int i = 0; i < u_spot_light_count; i++) {
        v_frag_pos_spot_light_space[i] = u_spot_light_space[i] * a_model * position;
    }

    gl_Position = u_projection * viewModel * position;
}