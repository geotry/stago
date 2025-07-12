#version 300 es

#define MAX_LIGHTS 10
#define EPSILON 0.00001

#include common.glsl

precision mediump float;
precision highp sampler2DArray;
precision mediump sampler2DShadow;
precision mediump sampler2DArrayShadow;

in vec2 v_texcoord;
in vec3 v_normal;
in vec3 v_frag_pos;
in vec4 v_frag_pos_directional_light_space;
in vec4 v_frag_pos_spot_light_space[10];

flat in int v_spot_light_count;

flat in int v_tex_index;
// Note: light.position should be in light space
// The position should be multiplied by view matrix
// Could be done on server instead?
in mat4 v_view;

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
uniform sampler2DShadow u_dir_light_shadow_map;

uniform int u_point_light_count;

uniform PointLight u_point_light[MAX_LIGHTS];
uniform SpotLight u_spot_light[MAX_LIGHTS];
// uniform sampler2DArrayShadow u_spot_light_shadow_map;
uniform sampler2DShadow u_spot_light_shadow_map;

out vec4 fragColor;

// float PCF(vec3 projCoords, float bias);
// float PCFSampled(vec3 projCoords, float bias);
vec4 ComputeDirectionalLight(in DirectionalLight light, vec4 diffuse_color, vec4 specular_color, vec3 norm, vec4 lightFragPos);
vec4 ComputePointLight(in PointLight light, vec4 diffuse_color, vec4 specular_color, vec3 norm);
vec4 ComputeSpotLight(in SpotLight light, int index, vec4 diffuse_color, vec4 specular_color, vec3 norm, vec4 lightFragPos);

float ShadowCalculation(in sampler2DShadow shadowMap, vec4 fragPosLightSpace, float bias);
float ShadowCalculation(in sampler2DArrayShadow shadowMap, int index, vec4 fragPosLightSpace, float bias);

float sqr(float x);
float attenuate_cusp(float distance, float radius, float max_intensity, float falloff);

float shadow_filter(sampler2DShadow shadowMap, vec3 uv_shadowMap, vec2 shadowMapSize) {
    // return texture(u_shadow_map, uv_shadowMap);
    float result = 0.0f;

    for(int x = -3; x <= 3; x++) {
        for(int y = -3; y <= 3; y++) {
            float x_l = (uv_shadowMap.x - float(x) / float(shadowMapSize.x));
            float y_l = (uv_shadowMap.y - float(y) / float(shadowMapSize.y));
            vec3 lookup = vec3(x_l, y_l, uv_shadowMap.z);
            result += texture(shadowMap, lookup); //get(x,y);
        }
    }

    return result / 49.0f;
}
float shadow_filter(sampler2DArrayShadow shadowMap, int index, vec3 uv_shadowMap, vec2 shadowMapSize) {
    // return texture(shadowMap, vec4(uv_shadowMap, float(index)));
    float result = 0.0f;

    for(int x = -3; x <= 3; x++) {
        for(int y = -3; y <= 3; y++) {
            float x_l = (uv_shadowMap.x - float(x) / float(shadowMapSize.x));
            float y_l = (uv_shadowMap.y - float(y) / float(shadowMapSize.y));
            vec3 lookup = vec3(x_l, y_l, uv_shadowMap.z);
            result += texture(shadowMap, vec4(lookup, float(index)));
        }
    }

    return result / 49.0f;
}
void main() {
    float index = texture(u_material.diffuse, vec3(v_texcoord, v_tex_index)).a;
    vec4 diffuse_color = texture(u_palette, vec2(index, 0));
    vec4 specular_color = vec4(1.0f, 1.0f, 1.0f, 1.0f) * texture(u_material.specular, vec3(v_texcoord, v_tex_index)).a;

    vec3 norm = normalize(v_normal);

    vec4 color = vec4(0.0f, 0.0f, 0.0f, 1.0f);

    color += ComputeDirectionalLight(u_dir_light, diffuse_color, specular_color, norm, v_frag_pos_directional_light_space);

    for(int i = 0; i < u_point_light_count; i++) {
        color += ComputePointLight(u_point_light[i], diffuse_color, specular_color, norm);
    }
    for(int i = 0; i < v_spot_light_count; i++) {
        color += ComputeSpotLight(u_spot_light[i], i, diffuse_color, specular_color, norm, v_frag_pos_spot_light_space[i]);
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

vec4 ComputeDirectionalLight(DirectionalLight light, vec4 diffuse_color, vec4 specular_color, vec3 norm, vec4 lightFragPos) {
    vec4 color;

    vec3 light_dir = normalize(-light.direction);
    vec3 view_dir = normalize(-v_frag_pos);
    vec3 half_dir = normalize(light_dir + view_dir);
    // float shadow = ShadowCalculation(v_frag_pos_directional_light_space, max(0.05f * (1.0f - dot(norm, light_dir)), 0.005f)); 
    // float shadow = ShadowCalculation(v_frag_pos_directional_light_space, 0.0001f); 
    // float shadow = ShadowCalculation(v_frag_pos_directional_light_space, 0.0f); 
    float shadow = ShadowCalculation(u_dir_light_shadow_map, lightFragPos, max(0.002f * (1.0f - dot(norm, light_dir)), 0.002f));

    // if(shadow == 1.0f) {
    //     return vec4(0.0f, 0.0f, 1.0f, 1.0f);
    // }

    // Ambient light
    vec4 ambient = vec4(light.ambient, 1.0f) * diffuse_color;

    // Diffuse light
    float diff = max(dot(norm, light_dir), 0.0f);
    vec4 diffuse = vec4(light.diffuse * diff, 1.0f) * diffuse_color;
    diffuse *= light.intensity;

    // Specular light
    float spec = pow(max(dot(norm, half_dir), 0.0f), u_material.shininess);
    vec4 specular = vec4(light.specular * spec, 1.0f) * specular_color;
    specular *= light.intensity;

    color = ambient + (shadow * diffuse) + (specular * shadow);

    return color;
}

vec4 ComputePointLight(PointLight light, vec4 diffuse_color, vec4 specular_color, vec3 norm) {
    vec4 color;

    vec3 light_dir = normalize(vec3(v_view * vec4(light.position, 1.0f)) - v_frag_pos);
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

vec4 ComputeSpotLight(SpotLight light, int index, vec4 diffuse_color, vec4 specular_color, vec3 norm, vec4 lightFragPos) {
    vec4 color;

    vec3 light_dir = normalize(vec3(v_view * vec4(light.position, 1.0f)) - v_frag_pos);
    vec3 view_dir = normalize(-v_frag_pos);
    vec3 half_dir = normalize(light_dir + view_dir);
    // float shadow = ShadowCalculation(u_spot_light_shadow_map, index, lightFragPos, max(0.002f * (1.0f - dot(norm, light_dir)), 0.002f));
    float shadow = ShadowCalculation(u_spot_light_shadow_map, lightFragPos, max(0.002f * (1.0f - dot(norm, light_dir)), 0.002f));

    float theta = dot(light_dir, normalize(-light.direction));
    float epsilon = light.cut_off - light.outer_cut_off;
    float intensity = clamp((theta - light.outer_cut_off) / epsilon, 0.0f, 1.0f);

    // Ambient light
    vec4 ambient = vec4(light.ambient, 1.0f) * diffuse_color;
    ambient *= intensity;

    // Diffuse light
    float diff = max(dot(norm, light_dir), 0.0f);
    vec4 diffuse = vec4(light.diffuse * diff, 1.0f) * diffuse_color;
    diffuse *= intensity;

    // Specular light
    float spec = pow(max(dot(norm, half_dir), 0.0f), u_material.shininess);
    vec4 specular = vec4(light.specular * spec, 1.0f) * specular_color;
    specular *= intensity;

    color = ambient + (shadow * diffuse) + (specular * shadow);
    // if(shadow > 0.0f) {
    //     color = vec4(1.0f, 0.0f, 0.0f, 1.0f);
    // }

    return color;
}

float ShadowCalculation(in sampler2DShadow shadowMap, vec4 fragPosLightSpace, float bias) {
    vec3 projCoords = fragPosLightSpace.xyz / fragPosLightSpace.w;
    projCoords = projCoords * 0.5f + 0.5f;

    // float closestDepth = texture(shadowMap, projCoords); 
    // // get depth of current fragment from light's perspective
    // float currentDepth = projCoords.z;
    // // check whether current frag pos is in shadow
    // float shadow = currentDepth > closestDepth ? 0.0f : 1.0f;
    // return shadow;

    vec2 shadowMapSize = vec2(textureSize(shadowMap, 0));
    // return textureProj(shadowMap, fragPosLightSpace);
    // return textureProjOffset(shadowMap, vec4(projCoords, fragPosLightSpace.w), ivec2(.5f));
    // return textureProj(shadowMap, vec4(projCoords, fragPosLightSpace.w), 0.5f);

    // return texture(shadowMap, projCoords);

    float shadowCoeff;

    // BEGIN PCF
    float sum = 0.0f;
    vec2 duv;

    for(float pcf_x = -1.5f; pcf_x <= 1.5f; pcf_x += 1.f) {
        for(float pcf_y = -1.5f; pcf_y <= 1.5f; pcf_y += 1.f) {
            duv = vec2(pcf_x / float(shadowMapSize.x), pcf_y / float(shadowMapSize.y));
            sum += shadow_filter(shadowMap, vec3(projCoords.xy, projCoords.z) + vec3(duv, 0.0f), shadowMapSize);
        }
    }

    sum = sum / 16.0f;

    shadowCoeff = projCoords.z - sum;
    shadowCoeff = 1.0f - (smoothstep(0.000f, 0.085f, shadowCoeff));

    // PCF 2
    // vec2 poissonDisk[4] = vec2[](vec2(-0.94201624f, -0.39906216f), vec2(0.94558609f, -0.76890725f), vec2(-0.094184101f, -0.92938870f), vec2(0.34495938f, 0.29387760f));
    // for(int i = 0; i < 4; i++) {
    //     if(texture(shadowMap, vec3(projCoords.xy + poissonDisk[i] / 700.0f, projCoords.z)) < projCoords.z) {
    //         shadowCoeff -= 0.2f;
    //     }
    // }

    // VSM
    // float distance = projCoords.z;
    // float mean = shadow_filter(shadowMap, projCoords, shadowMapSize);
    // float depth = texture(shadowMap, projCoords);
    // // note: normally depth_2 is stored in shadow map (in green channel)
    // float depth_2 = pow(depth, 2.0f);
    // float dx = dFdx(depth);
    // float dy = dFdy(depth);
    // depth_2 = depth_2 + 0.5f * (dx * dx + dy * dy);
    // float variance = depth_2 - pow(mean, 2.00f);
    // variance = max(variance, 0.005f);

    // float p = smoothstep(distance - 0.02f, distance, mean);
    // float d = distance - mean;

    // float p_max = linstep(0.2f, 1.0f, variance / (variance + d * d));
    // shadowCoeff = clamp(max(p, p_max), 0.0f, 1.0f);

    // shadowCoeff = 0 = shadow
    // shadowCoeff = 1 = no shadow
    return shadowCoeff;
    // To debug
    // if(projCoords.z > zShadowMap + EPSILON) {
    //     return 1.0f;
    // }

    // return PCF(projCoords, bias);
    // return PCFSampled(projCoords, bias);
    // return PCFSampled(projCoords, bias);

    // ivec2 shadowMapSize = textureSize(u_shadow_map, 0);

    // float xOffset = 1.0f / float(shadowMapSize.x);
    // float yOffset = 1.0f / float(shadowMapSize.y);
    // float factor = 0.0f;

    // for(int y = -1; y <= 1; y++) {
    //     for(int x = -1; x <= 1; x++) {
    //         vec2 offsets = vec2(float(x) * xOffset, float(y) * yOffset);
    //         vec3 uvc = vec3(projCoords.xy + offsets, projCoords.z + EPSILON);
    //         factor += texture(u_shadow_map, uvc);
    //     }
    // }

    // return (0.5f + (factor / 18.0f));

    // vec3 biased = vec3(projCoords.xy, projCoords.z - bias);

    // return texture(u_shadow_map, biased);

    // for(int i = 0; i < samples; i++) {
    //     // vec3 biased = vec3(projCoords.xy, projCoords.z - bias);
    //     vec3 biased = vec3(projCoords.xy + vec2(sampleOffsetDirections[i]) / shadowSpread, projCoords.z - bias);
    //     // vec3 biased = vec3(projCoords.xy + sampleOffsetDirections[i] * diskRadius, projCoords.z - bias);
    //     // vec3 biased = vec3(projCoords + sampleOffsetDirections[i] * diskRadius);
    //     // float litPercent = texture(u_shadow_map, vec3(projCoords.xy + adjacentPixels[i % 5] / shadowSpread, projCoords.z), 0.5f);

    //     // float litPercent = texture(u_shadow_map, vec3(projCoords.xy + vec2(sampleOffsetDirections[i % 5]) / shadowSpread, projCoords.z));
    //     float litPercent = texture(u_shadow_map, biased);
    //     visibility += litPercent;

    //     // float litPercent = texture(u_shadow_map, vec3(projCoords.xy + adjacentPixels[i % 5] / shadowSpread, projCoords.z), bias);
    //     // float litPercent = texture(u_shadow_map, vec3(projCoords.xy + adjacentPixels[i % 5] / shadowSpread, projCoords.z));
    //     // visibility *= max(litPercent, 0.8f);
    // }

    // return visibility / float(samples);
}

float ShadowCalculation(in sampler2DArrayShadow shadowMap, int index, vec4 fragPosLightSpace, float bias) {
    vec3 projCoords = fragPosLightSpace.xyz / fragPosLightSpace.w;
    projCoords = projCoords * 0.5f + 0.5f;

    vec2 shadowMapSize = vec2(textureSize(shadowMap, 0));

    float shadowCoeff = 0.0f;

    // BEGIN PCF
    float sum = 0.0f;
    vec2 duv;

    for(float pcf_x = -1.5f; pcf_x <= 1.5f; pcf_x += 1.f) {
        for(float pcf_y = -1.5f; pcf_y <= 1.5f; pcf_y += 1.f) {
            duv = vec2(pcf_x / float(shadowMapSize.x), pcf_y / float(shadowMapSize.y));
            sum += shadow_filter(shadowMap, index, vec3(projCoords.xy, projCoords.z) + vec3(duv, 0.0f), shadowMapSize);
        }
    }

    sum = sum / 16.0f;

    shadowCoeff = projCoords.z - sum;
    shadowCoeff = 1.0f - (smoothstep(0.000f, 0.085f, shadowCoeff));
    return shadowCoeff;
}

// float PCF(vec3 projCoords, float bias) {
//     float shadow = 0.0f;
//     vec2 texelSize = 1.0f / vec2(textureSize(u_shadow_map, 0));
//     for(int x = -1; x <= 1; ++x) {
//         for(int y = -1; y <= 1; ++y) {
//             float pcfDepth = texture(u_shadow_map, vec3(projCoords.xy + vec2(x, y) * texelSize, projCoords.z), bias);
//             shadow += projCoords.z > pcfDepth ? 1.0f : 0.0f;
//         }
//     }
//     shadow /= 9.0f;

//     // keep the shadow at 0.0 when outside the far_plane region of the light's frustum.
//     if(projCoords.z > 1.0f) {
//         shadow = 0.0f;
//     }

//     return 1.0f - shadow;
// }

float shadowSpread = 2100.0f;
int samples = 20;
vec3 sampleOffsetDirections[20] = vec3[](vec3(1, 1, 1), vec3(1, -1, 1), vec3(-1, -1, 1), vec3(-1, 1, 1), vec3(1, 1, -1), vec3(1, -1, -1), vec3(-1, -1, -1), vec3(-1, 1, -1), vec3(1, 1, 0), vec3(1, -1, 0), vec3(-1, -1, 0), vec3(-1, 1, 0), vec3(1, 0, 1), vec3(-1, 0, 1), vec3(1, 0, -1), vec3(-1, 0, -1), vec3(0, 1, 1), vec3(0, -1, 1), vec3(0, -1, -1), vec3(0, 1, -1));

// float PCFSampled(vec3 uvShadowMap, float bias) {
//     float visibility = 1.0f;
//     for(int i = 0; i < samples; i++) {
//         // vec3 biased = vec3(uvShadowMap.xy, uvShadowMap.z - bias);
//         vec3 biased = vec3(uvShadowMap.xy + vec2(sampleOffsetDirections[i]) / shadowSpread, uvShadowMap.z - bias);
//         // vec3 biased = vec3(uvShadowMap.xy + sampleOffsetDirections[i] * diskRadius, uvShadowMap.z - bias);
//         // vec3 biased = vec3(uvShadowMap + sampleOffsetDirections[i] * diskRadius);
//         // float litPercent = texture(u_shadow_map, vec3(uvShadowMap.xy + adjacentPixels[i % 5] / shadowSpread, uvShadowMap.z), 0.5f);

//         // float litPercent = texture(u_shadow_map, vec3(uvShadowMap.xy + vec2(sampleOffsetDirections[i % 5]) / shadowSpread, uvShadowMap.z));
//         float litPercent = texture(u_shadow_map, biased);
//         visibility += litPercent;

//         // float litPercent = texture(u_shadow_map, vec3(uvShadowMap.xy + adjacentPixels[i % 5] / shadowSpread, uvShadowMap.z), bias);
//         // float litPercent = texture(u_shadow_map, vec3(uvShadowMap.xy + adjacentPixels[i % 5] / shadowSpread, uvShadowMap.z));
//         // visibility *= max(litPercent, 0.8f);
//     }

//     return visibility / float(samples);
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
