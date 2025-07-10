#version 300 es

#define MAX_LIGHTS 10

precision mediump float;
precision highp sampler2DArray;

in vec2 v_texcoord;
in vec3 v_normal;
in vec3 v_frag_pos;
// Note: light.position should be in light space
// The position should be multiplied by view matrix
// Could be done on server instead?
in mat3 v_view;
flat in int v_tex_index;

uniform sampler2D u_palette;

struct Material {
    sampler2DArray diffuse;
    sampler2DArray specular;
    float shininess;
};

struct DirectionalLight {
    vec3 direction;

    vec3 ambient;
    vec3 diffuse;
    vec3 specular;

    float intensity;
};

struct PointLight {
    vec3 position;

    vec3 ambient;
    vec3 diffuse;
    vec3 specular;

    float radius;
    float intensity;
};

struct SpotLight {
    vec3 position;
    vec3 direction;

    vec3 ambient;
    vec3 diffuse;
    vec3 specular;

    float cut_off;
    float outer_cut_off;
};

uniform Material u_material;

uniform DirectionalLight u_dir_light;

uniform int u_point_light_count;
uniform int u_spot_light_count;

uniform PointLight u_point_light[MAX_LIGHTS];
uniform SpotLight u_spot_light[MAX_LIGHTS];

out vec4 fragColor;

vec4 ComputeDirectionalLight(in DirectionalLight light, vec4 diffuse_color, vec4 specular_color, vec3 norm);
vec4 ComputePointLight(in PointLight light, vec4 diffuse_color, vec4 specular_color, vec3 norm);
vec4 ComputeSpotLight(in SpotLight light, vec4 diffuse_color, vec4 specular_color, vec3 norm);

float sqr(float x);
float attenuate_cusp(float distance, float radius, float max_intensity, float falloff);

void main() {
    float index = texture(u_material.diffuse, vec3(v_texcoord, v_tex_index)).a;
    vec4 diffuse_color = texture(u_palette, vec2(index, 0));
    vec4 specular_color = vec4(1.0f, 1.0f, 1.0f, 1.0f) * texture(u_material.specular, vec3(v_texcoord, v_tex_index)).a;

    vec3 norm = normalize(v_normal);

    vec4 color = ComputeDirectionalLight(u_dir_light, diffuse_color, specular_color, norm);

    for(int i = 0; i < u_point_light_count; i++) {
        color += ComputePointLight(u_point_light[i], diffuse_color, specular_color, norm);
    }
    for(int i = 0; i < u_spot_light_count; i++) {
        color += ComputeSpotLight(u_spot_light[i], diffuse_color, specular_color, norm);
    }

    // Depth
    // float depth = 1.0f - (gl_FragCoord.z / gl_FragCoord.w) * .5f;
    // color = vec4(depth, depth, depth, 1.0f);

    // Apply Gamma correction
    float gamma = 2.2f;
    color.rgb = pow(color.rgb, vec3(1.0f / gamma));

    fragColor = color;
}

vec4 ComputeDirectionalLight(DirectionalLight light, vec4 diffuse_color, vec4 specular_color, vec3 norm) {
    vec4 color;

    vec3 light_dir = normalize(-light.direction);
    vec3 view_dir = normalize(-v_frag_pos);
    vec3 half_dir = normalize(light_dir + view_dir);

    // Ambient light
    vec4 ambient = vec4(light.ambient, 1.0f) * diffuse_color;
    color += ambient;

    // Diffuse light
    float diff = max(dot(norm, light_dir), 0.0f);
    vec4 diffuse = vec4(light.diffuse * diff, 1.0f) * diffuse_color;
    diffuse *= light.intensity;
    color += diffuse;

    // Specular light
    float spec = pow(max(dot(norm, half_dir), 0.0f), u_material.shininess);
    vec4 specular = vec4(light.specular * spec, 1.0f) * specular_color;
    specular *= light.intensity;
    color += specular;

    return color;
}

vec4 ComputePointLight(PointLight light, vec4 diffuse_color, vec4 specular_color, vec3 norm) {
    vec4 color;

    vec3 light_dir = normalize(v_view * light.position - v_frag_pos);
    vec3 view_dir = normalize(-v_frag_pos);
    vec3 half_dir = normalize(light_dir + view_dir);

    float distance = length(light_dir);
    float attenuation = attenuate_cusp(distance, light.radius, light.intensity, 1.0f);

    // Ambient light
    vec4 ambient = vec4(light.ambient, 1.0f) * diffuse_color;
    ambient *= attenuation;
    color += ambient;

    // Diffuse light
    float diff = max(dot(norm, light_dir), 0.0f);
    vec4 diffuse = vec4(light.diffuse * diff, 1.0f) * diffuse_color;
    diffuse *= attenuation;
    color += diffuse;

    // Specular light
    float spec = pow(max(dot(norm, half_dir), 0.0f), u_material.shininess);
    vec4 specular = vec4(light.specular * spec, 1.0f) * specular_color;
    specular *= attenuation;
    color += specular;

    return color;
}

vec4 ComputeSpotLight(SpotLight light, vec4 diffuse_color, vec4 specular_color, vec3 norm) {
    vec4 color;

    vec3 light_dir = normalize(v_view * light.position - v_frag_pos);
    vec3 view_dir = normalize(-v_frag_pos);
    vec3 half_dir = normalize(light_dir + view_dir);

    float theta = dot(light_dir, normalize(-light.direction));
    float epsilon = light.cut_off - light.outer_cut_off;
    float intensity = clamp((theta - light.outer_cut_off) / epsilon, 0.0f, 1.0f);

    // Ambient light
    vec4 ambient = vec4(light.ambient, 1.0f) * diffuse_color;
    ambient *= intensity;
    color += ambient;

    // Diffuse light
    float diff = max(dot(norm, light_dir), 0.0f);
    vec4 diffuse = vec4(light.diffuse * diff, 1.0f) * diffuse_color;
    diffuse *= intensity;
    color += diffuse;

    // Specular light
    float spec = pow(max(dot(norm, half_dir), 0.0f), u_material.shininess);
    vec4 specular = vec4(light.specular * spec, 1.0f) * specular_color;
    specular *= intensity;
    color += specular;

    return color;
}

// float ShadowCalculation(vec4 fragPosLightSpace)
// {
//     // perform perspective divide
//     vec3 projCoords = fragPosLightSpace.xyz / fragPosLightSpace.w;
//     // transform to [0,1] range
//     projCoords = projCoords * 0.5 + 0.5;
//     // get closest depth value from light's perspective (using [0,1] range fragPosLight as coords)
//     float closestDepth = texture(shadowMap, projCoords.xy).r; 
//     // get depth of current fragment from light's perspective
//     float currentDepth = projCoords.z;
//     // check whether current frag pos is in shadow
//     float shadow = currentDepth > closestDepth  ? 1.0 : 0.0;

//     return shadow;
// } 

float sqr(float x) {
    return x * x;
}

float attenuate_cusp(float distance, float radius, float max_intensity, float falloff) {
    float s = distance / radius;

    if(s >= 1.0f)
        return 0.0f;

    float s2 = sqr(s);

    return max_intensity * sqr(1.0f - s2) / (1.0f + falloff * s);
}
