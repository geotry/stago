#version 300 es

precision mediump float;
precision highp sampler2DArray;

in vec2 v_texcoord;
in vec3 v_normal;
in vec3 v_frag_pos;
flat in int v_tex_index;

uniform sampler2D u_palette;

// Light
uniform vec3 u_view_pos; // Position of the camera: use VP matrix instead with 0,0,0

struct Material {
    sampler2DArray diffuse;
    sampler2DArray specular;
    // sampler2DArray normal;
    float shininess;
};

struct Light {
    vec3 position;
    vec3 direction;

    // vec4 vector; // vec4: w = 0 is directional light, w = 1 is position

    vec3 ambient;
    vec3 diffuse;
    vec3 specular;

    // Attenuation (Point light)
    float constant;
    float linear;
    float quadratic;
};

uniform Material u_material;
uniform Light u_light;

out vec4 fragColor;

void main() {
    float index = texture(u_material.diffuse, vec3(v_texcoord, v_tex_index)).a;
    vec4 diffuse_color = texture(u_palette, vec2(index, 0));
    vec4 specular_color = vec4(1.0f, 1.0f, 1.0f, 1.0f) * texture(u_material.specular, vec3(v_texcoord, v_tex_index)).a;

    vec3 light_dir;
    float attenuation;

    // if(u_light.vector.w == 0.0f) {
    //     // Directional light
    //     light_dir = normalize(-u_light.vector.xyz);
    //     attenuation = 1.0f;
    // }
    // if(u_light.vector.w == 1.0f) {
    // Point light
    light_dir = normalize(u_light.position.xyz - v_frag_pos);
    float distance = length(light_dir);
    attenuation = 1.0f / (u_light.constant + u_light.linear * distance + u_light.quadratic * (distance * distance));

    vec3 view_dir = normalize(u_view_pos - v_frag_pos);

    // Ambient light
    vec4 ambient = vec4(u_light.ambient, 1.0f) * diffuse_color;
    ambient *= attenuation;

    // Diffuse light
    vec3 norm = normalize(v_normal);
    float diff = max(dot(norm, light_dir), 0.0f);
    vec4 diffuse = vec4(u_light.diffuse * diff, 1.0f) * diffuse_color;
    diffuse *= attenuation;

    // Specular light
    vec3 reflect_dir = reflect(-light_dir, norm);
    float spec = pow(max(dot(view_dir, reflect_dir), 0.0f), u_material.shininess);
    vec4 specular = vec4(u_light.specular * spec, 1.0f) * specular_color;
    specular *= attenuation;

    vec4 color = ambient + diffuse + specular;

    // Depth
    // float depth = 1.0f - (gl_FragCoord.z / gl_FragCoord.w) * 0.01f;
    // color = color * vec4(depth, depth, depth, 1.0f);

    fragColor = color;
}
