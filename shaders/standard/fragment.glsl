#version 300 es

#define MAX_LIGHTS 10
#define EPSILON 0.00001

#include ../lighting/shadow.glsl
#include ../lighting/directional.glsl
#include ../lighting/spot.glsl
#include ../lighting/point.glsl

precision mediump float;
precision highp sampler2DArray;
precision mediump sampler2DShadow;
precision mediump sampler2DArrayShadow;

uniform Camera {
    mat4 view;
    mat4 projection;
} camera;

in vec2 v_texcoord;
in vec3 v_normal;
in vec3 v_frag_pos;
in vec4 v_frag_pos_directional_light_space;
in vec4 v_frag_pos_spot_light_space[10];

flat in int v_spot_light_count;

flat in int v_tex_index;

uniform sampler2D u_palette;

struct Material {
    sampler2DArray diffuse;
    sampler2DArray specular;
    float shininess;
};

uniform Material u_material;

uniform DirectionalLight u_dir_light;
uniform int u_point_light_count;

uniform PointLight u_point_light[MAX_LIGHTS];
uniform SpotLight u_spot_light[MAX_LIGHTS];

// uniform Lighting {
//     int spot_light_count;
//     SpotLight spot_lights[10];
// } light;

out vec4 fragColor;

void main() {
    float index = texture(u_material.diffuse, vec3(v_texcoord, v_tex_index)).a;
    vec4 diffuse_color = texture(u_palette, vec2(index, 0));
    vec4 specular_color = vec4(1.0f, 1.0f, 1.0f, 1.0f) * texture(u_material.specular, vec3(v_texcoord, v_tex_index)).a;

    vec4 color = vec4(0.0f, 0.0f, 0.0f, 1.0f);

    color += ComputeDirectionalLight(u_dir_light, diffuse_color, specular_color, u_material.shininess, v_normal, v_frag_pos, v_frag_pos_directional_light_space);

    for(int i = 0; i < u_point_light_count; i++) {
        color += ComputePointLight(u_point_light[i], i, camera.view, diffuse_color, specular_color, u_material.shininess, v_normal, v_frag_pos);
    }
    for(int i = 0; i < v_spot_light_count; i++) {
        color += ComputeSpotLight(u_spot_light[i], i, camera.view, diffuse_color, specular_color, u_material.shininess, v_normal, v_frag_pos, v_frag_pos_spot_light_space[i]);
    }

    // Depth
    // float depth = 1.0f - (gl_FragCoord.z / gl_FragCoord.w) * .5f;
    // color = vec4(depth, depth, depth, 1.0f);

    // DEBUG (no light)
    // color = diffuse_color;
    // color = vec4(1.0f, 0.0f, 0.0f, 1.0f);

    // Apply Gamma correction
    float gamma = 2.2f;
    color.rgb = pow(color.rgb, vec3(1.0f / gamma));

    fragColor = color;
}