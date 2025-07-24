/* Generated file, DO NOT EDIT! */

/**
 * GLSL vertex shader (source: shaders/standard/vertex.glsl)
 */
const VERTEX_SRC = `
#version 300 es
#define MAX_LIGHTS 10
precision mediump float;
precision highp sampler2DArray;
uniform Camera {
    mat4 view;
    mat4 projection;
} camera;
in vec3 a_position;
in vec2 a_uv;
in vec3 a_normal;
in mat4 a_model;
uniform int u_tex_index;
uniform mat4 u_directional_light_space;
uniform mat4 u_spot_light_space[MAX_LIGHTS];
uniform int u_spot_light_count;
flat out int v_tex_index;
out vec2 v_texcoord;
out vec3 v_normal;
out vec3 v_frag_pos;
out vec4 v_frag_pos_directional_light_space;
flat out int v_spot_light_count;
out vec4 v_frag_pos_spot_light_space[MAX_LIGHTS];
void main() {
    vec4 position = vec4(a_position, 1.0f);
    mat4 viewModel = camera.view * a_model;
    v_texcoord = a_uv;
    v_tex_index = u_tex_index;
    v_normal = normalize(mat3(transpose(inverse(viewModel))) * a_normal);
    v_frag_pos = vec3(viewModel * position);
    v_frag_pos_directional_light_space = u_directional_light_space * a_model * position;
    v_spot_light_count = u_spot_light_count;
    for(int i = 0; i < u_spot_light_count; i++) {
        v_frag_pos_spot_light_space[i] = u_spot_light_space[i] * a_model * position;
    }
    gl_Position = camera.projection * viewModel * position;
}
`.trim();
/**
 * GLSL fragment shader (source: shaders/standard/fragment.glsl)
 */
const FRAGMENT_SRC = `
#version 300 es
#define MAX_LIGHTS 10
#define EPSILON 0.00001
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
uniform sampler2DShadow u_spot_light_shadow_map;
uniform sampler2DShadow u_directional_light_shadow_map;
uniform sampler2D u_palette;
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
struct DirectionalLight {
  vec3 direction;
  vec3 ambient;
  vec3 diffuse;
  vec3 specular;
  float intensity;
};
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
out vec4 fragColor;
float shadow_filter(sampler2DShadow shadowMap, vec3 uv_shadowMap, vec2 shadowMapSize) {
  float result = 0.0;
  for(int x = -3; x <= 3; x++) {
    for(int y = -3; y <= 3; y++) {
      float x_l = (uv_shadowMap.x - float(x) / float(shadowMapSize.x));
      float y_l = (uv_shadowMap.y - float(y) / float(shadowMapSize.y));
      vec3 lookup = vec3(x_l, y_l, uv_shadowMap.z);
      result += texture(shadowMap, lookup); //get(x,y);
    }
  }
  return result / 49.0;
}
float shadow_filter(sampler2DArrayShadow shadowMap, int index, vec3 uv_shadowMap, vec2 shadowMapSize) {
  float result = 0.0;
  for(int x = -3; x <= 3; x++) {
    for(int y = -3; y <= 3; y++) {
      float x_l = (uv_shadowMap.x - float(x) / float(shadowMapSize.x));
      float y_l = (uv_shadowMap.y - float(y) / float(shadowMapSize.y));
      vec3 lookup = vec3(x_l, y_l, uv_shadowMap.z);
      result += texture(shadowMap, vec4(lookup, float(index)));
    }
  }
  return result / 49.0;
}
float ShadowCalculation(in sampler2DShadow shadowMap, vec4 fragPosLightSpace, float bias) {
  vec3 projCoords = fragPosLightSpace.xyz / fragPosLightSpace.w;
  projCoords = projCoords * 0.5 + 0.5;
  vec2 shadowMapSize = vec2(textureSize(shadowMap, 0));
  float shadowCoeff;
  float sum = 0.0;
  vec2 duv;
  for(float pcf_x = -1.5; pcf_x <= 1.5; pcf_x += 1.) {
    for(float pcf_y = -1.5; pcf_y <= 1.5; pcf_y += 1.) {
      duv = vec2(pcf_x / float(shadowMapSize.x), pcf_y / float(shadowMapSize.y));
      sum += shadow_filter(shadowMap, vec3(projCoords.xy, projCoords.z) + vec3(duv, 0.0), shadowMapSize);
    }
  }
  sum = sum / 16.0;
  shadowCoeff = projCoords.z - sum;
  shadowCoeff = 1.0 - (smoothstep(0.000, 0.085, shadowCoeff));
  return shadowCoeff;
}
float ShadowCalculation(in sampler2DArrayShadow shadowMap, int index, vec4 fragPosLightSpace, float bias) {
  vec3 projCoords = fragPosLightSpace.xyz / fragPosLightSpace.w;
  projCoords = projCoords * 0.5 + 0.5;
  vec2 shadowMapSize = vec2(textureSize(shadowMap, 0));
  float shadowCoeff = 0.0;
  float sum = 0.0;
  vec2 duv;
  for(float pcf_x = -1.5; pcf_x <= 1.5; pcf_x += 1.) {
    for(float pcf_y = -1.5; pcf_y <= 1.5; pcf_y += 1.) {
      duv = vec2(pcf_x / float(shadowMapSize.x), pcf_y / float(shadowMapSize.y));
      sum += shadow_filter(shadowMap, index, vec3(projCoords.xy, projCoords.z) + vec3(duv, 0.0), shadowMapSize);
    }
  }
  sum = sum / 16.0;
  shadowCoeff = projCoords.z - sum;
  shadowCoeff = 1.0 - (smoothstep(0.000, 0.085, shadowCoeff));
  return shadowCoeff;
}
vec4 ComputeDirectionalLight(in DirectionalLight light, vec4 diffuse_color, vec4 specular_color, float shininess, vec3 norm, vec3 fragPos, vec4 lightFragPos) {
  vec4 color;
  vec3 light_dir = normalize(-light.direction);
  vec3 view_dir = normalize(-fragPos);
  vec3 half_dir = normalize(light_dir + view_dir);
  vec4 ambient = vec4(light.ambient, 1.0) * diffuse_color;
  float diff = max(dot(norm, light_dir), 0.0);
  vec4 diffuse = vec4(light.diffuse * diff, 1.0) * diffuse_color;
  diffuse *= light.intensity;
  float spec = pow(max(dot(norm, half_dir), 0.0), shininess);
  vec4 specular = vec4(light.specular * spec, 1.0) * specular_color;
  specular *= light.intensity;
  float shadow = ShadowCalculation(u_directional_light_shadow_map, lightFragPos, max(0.002 * (1.0 - dot(norm, light_dir)), 0.002));
  float d = 1.0 - (gl_FragCoord.z / gl_FragCoord.w) * .25;
  if(d < 0.0) {
    shadow = 1.0;
  }
  color = ambient + (shadow * diffuse) + (specular * shadow);
  return color;
}
vec4 ComputeSpotLight(SpotLight light, int index, mat4 view, vec4 diffuse_color, vec4 specular_color, float shininess, vec3 norm, vec3 fragPos, vec4 lightFragPos) {
  vec4 color;
  vec3 light_dir = normalize(vec3(view * vec4(light.position, 1.0)) - fragPos);
  vec3 view_dir = normalize(-fragPos);
  vec3 half_dir = normalize(light_dir + view_dir);
  float theta = dot(light_dir, normalize(-light.direction));
  float epsilon = light.cut_off - light.outer_cut_off;
  float intensity = clamp((theta - light.outer_cut_off) / epsilon, 0.0, 1.0);
  vec4 ambient = vec4(light.ambient, 1.0) * diffuse_color;
  ambient *= intensity;
  float diff = max(dot(norm, light_dir), 0.0);
  vec4 diffuse = vec4(light.diffuse * diff, 1.0) * diffuse_color;
  diffuse *= intensity;
  float spec = pow(max(dot(norm, half_dir), 0.0), shininess);
  vec4 specular = vec4(light.specular * spec, 1.0) * specular_color;
  specular *= intensity;
  float shadow = ShadowCalculation(u_spot_light_shadow_map, lightFragPos, max(0.002 * (1.0 - dot(norm, light_dir)), 0.002));
  color = ambient + (shadow * diffuse) + (specular * shadow);
  return color;
}
float sqr(float x) {
  return x * x;
}
float attenuate_cusp(float distance, float radius, float max_intensity, float falloff) {
  float s = distance / radius;
  if(s >= 1.0)
    return 0.0;
  float s2 = sqr(s);
  return max_intensity * sqr(1.0 - s2) / (1.0 + falloff * s);
}
vec4 ComputePointLight(PointLight light, int index, mat4 view, vec4 diffuse_color, vec4 specular_color, float shininess, vec3 norm, vec3 fragPos) {
  vec4 color;
  vec3 light_dir = normalize(vec3(view * vec4(light.position, 1.0)) - fragPos);
  vec3 view_dir = normalize(-fragPos);
  vec3 half_dir = normalize(light_dir + view_dir);
  float distance = length(light_dir);
  float attenuation = attenuate_cusp(distance, light.radius, light.intensity, 1.0);
  vec4 ambient = vec4(light.ambient, 1.0) * diffuse_color;
  ambient *= attenuation;
  color += ambient;
  float diff = max(dot(norm, light_dir), 0.0);
  vec4 diffuse = vec4(light.diffuse * diff, 1.0) * diffuse_color;
  diffuse *= attenuation;
  color += diffuse;
  float spec = pow(max(dot(norm, half_dir), 0.0), shininess);
  vec4 specular = vec4(light.specular * spec, 1.0) * specular_color;
  specular *= attenuation;
  color += specular;
  return color;
}
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
    float gamma = 2.2f;
    color.rgb = pow(color.rgb, vec3(1.0f / gamma));
    fragColor = color;
}
`.trim();

/**
 * @param {WebGL2RenderingContext} gl
 * @param {WebGLProgram} program
 * @returns
 */
const createUniforms = (gl, program) => {
  const locs = {
    [`u_dir_light.direction`]: gl.getUniformLocation(program, "u_dir_light.direction"),
    [`u_dir_light.ambient`]: gl.getUniformLocation(program, "u_dir_light.ambient"),
    [`u_dir_light.diffuse`]: gl.getUniformLocation(program, "u_dir_light.diffuse"),
    [`u_dir_light.specular`]: gl.getUniformLocation(program, "u_dir_light.specular"),
    [`u_dir_light.intensity`]: gl.getUniformLocation(program, "u_dir_light.intensity"),
    [`u_directional_light_shadow_map`]: gl.getUniformLocation(program, "u_directional_light_shadow_map"),
    [`u_directional_light_space`]: gl.getUniformLocation(program, "u_directional_light_space"),
    [`u_material.diffuse`]: gl.getUniformLocation(program, "u_material.diffuse"),
    [`u_material.specular`]: gl.getUniformLocation(program, "u_material.specular"),
    [`u_material.shininess`]: gl.getUniformLocation(program, "u_material.shininess"),
    [`u_palette`]: gl.getUniformLocation(program, "u_palette"),
    [`u_point_light[0].position`]: gl.getUniformLocation(program, "u_point_light[0].position"),
    [`u_point_light[0].ambient`]: gl.getUniformLocation(program, "u_point_light[0].ambient"),
    [`u_point_light[0].diffuse`]: gl.getUniformLocation(program, "u_point_light[0].diffuse"),
    [`u_point_light[0].specular`]: gl.getUniformLocation(program, "u_point_light[0].specular"),
    [`u_point_light[0].radius`]: gl.getUniformLocation(program, "u_point_light[0].radius"),
    [`u_point_light[0].intensity`]: gl.getUniformLocation(program, "u_point_light[0].intensity"),
    [`u_point_light[1].position`]: gl.getUniformLocation(program, "u_point_light[1].position"),
    [`u_point_light[1].ambient`]: gl.getUniformLocation(program, "u_point_light[1].ambient"),
    [`u_point_light[1].diffuse`]: gl.getUniformLocation(program, "u_point_light[1].diffuse"),
    [`u_point_light[1].specular`]: gl.getUniformLocation(program, "u_point_light[1].specular"),
    [`u_point_light[1].radius`]: gl.getUniformLocation(program, "u_point_light[1].radius"),
    [`u_point_light[1].intensity`]: gl.getUniformLocation(program, "u_point_light[1].intensity"),
    [`u_point_light[2].position`]: gl.getUniformLocation(program, "u_point_light[2].position"),
    [`u_point_light[2].ambient`]: gl.getUniformLocation(program, "u_point_light[2].ambient"),
    [`u_point_light[2].diffuse`]: gl.getUniformLocation(program, "u_point_light[2].diffuse"),
    [`u_point_light[2].specular`]: gl.getUniformLocation(program, "u_point_light[2].specular"),
    [`u_point_light[2].radius`]: gl.getUniformLocation(program, "u_point_light[2].radius"),
    [`u_point_light[2].intensity`]: gl.getUniformLocation(program, "u_point_light[2].intensity"),
    [`u_point_light[3].position`]: gl.getUniformLocation(program, "u_point_light[3].position"),
    [`u_point_light[3].ambient`]: gl.getUniformLocation(program, "u_point_light[3].ambient"),
    [`u_point_light[3].diffuse`]: gl.getUniformLocation(program, "u_point_light[3].diffuse"),
    [`u_point_light[3].specular`]: gl.getUniformLocation(program, "u_point_light[3].specular"),
    [`u_point_light[3].radius`]: gl.getUniformLocation(program, "u_point_light[3].radius"),
    [`u_point_light[3].intensity`]: gl.getUniformLocation(program, "u_point_light[3].intensity"),
    [`u_point_light[4].position`]: gl.getUniformLocation(program, "u_point_light[4].position"),
    [`u_point_light[4].ambient`]: gl.getUniformLocation(program, "u_point_light[4].ambient"),
    [`u_point_light[4].diffuse`]: gl.getUniformLocation(program, "u_point_light[4].diffuse"),
    [`u_point_light[4].specular`]: gl.getUniformLocation(program, "u_point_light[4].specular"),
    [`u_point_light[4].radius`]: gl.getUniformLocation(program, "u_point_light[4].radius"),
    [`u_point_light[4].intensity`]: gl.getUniformLocation(program, "u_point_light[4].intensity"),
    [`u_point_light[5].position`]: gl.getUniformLocation(program, "u_point_light[5].position"),
    [`u_point_light[5].ambient`]: gl.getUniformLocation(program, "u_point_light[5].ambient"),
    [`u_point_light[5].diffuse`]: gl.getUniformLocation(program, "u_point_light[5].diffuse"),
    [`u_point_light[5].specular`]: gl.getUniformLocation(program, "u_point_light[5].specular"),
    [`u_point_light[5].radius`]: gl.getUniformLocation(program, "u_point_light[5].radius"),
    [`u_point_light[5].intensity`]: gl.getUniformLocation(program, "u_point_light[5].intensity"),
    [`u_point_light[6].position`]: gl.getUniformLocation(program, "u_point_light[6].position"),
    [`u_point_light[6].ambient`]: gl.getUniformLocation(program, "u_point_light[6].ambient"),
    [`u_point_light[6].diffuse`]: gl.getUniformLocation(program, "u_point_light[6].diffuse"),
    [`u_point_light[6].specular`]: gl.getUniformLocation(program, "u_point_light[6].specular"),
    [`u_point_light[6].radius`]: gl.getUniformLocation(program, "u_point_light[6].radius"),
    [`u_point_light[6].intensity`]: gl.getUniformLocation(program, "u_point_light[6].intensity"),
    [`u_point_light[7].position`]: gl.getUniformLocation(program, "u_point_light[7].position"),
    [`u_point_light[7].ambient`]: gl.getUniformLocation(program, "u_point_light[7].ambient"),
    [`u_point_light[7].diffuse`]: gl.getUniformLocation(program, "u_point_light[7].diffuse"),
    [`u_point_light[7].specular`]: gl.getUniformLocation(program, "u_point_light[7].specular"),
    [`u_point_light[7].radius`]: gl.getUniformLocation(program, "u_point_light[7].radius"),
    [`u_point_light[7].intensity`]: gl.getUniformLocation(program, "u_point_light[7].intensity"),
    [`u_point_light[8].position`]: gl.getUniformLocation(program, "u_point_light[8].position"),
    [`u_point_light[8].ambient`]: gl.getUniformLocation(program, "u_point_light[8].ambient"),
    [`u_point_light[8].diffuse`]: gl.getUniformLocation(program, "u_point_light[8].diffuse"),
    [`u_point_light[8].specular`]: gl.getUniformLocation(program, "u_point_light[8].specular"),
    [`u_point_light[8].radius`]: gl.getUniformLocation(program, "u_point_light[8].radius"),
    [`u_point_light[8].intensity`]: gl.getUniformLocation(program, "u_point_light[8].intensity"),
    [`u_point_light[9].position`]: gl.getUniformLocation(program, "u_point_light[9].position"),
    [`u_point_light[9].ambient`]: gl.getUniformLocation(program, "u_point_light[9].ambient"),
    [`u_point_light[9].diffuse`]: gl.getUniformLocation(program, "u_point_light[9].diffuse"),
    [`u_point_light[9].specular`]: gl.getUniformLocation(program, "u_point_light[9].specular"),
    [`u_point_light[9].radius`]: gl.getUniformLocation(program, "u_point_light[9].radius"),
    [`u_point_light[9].intensity`]: gl.getUniformLocation(program, "u_point_light[9].intensity"),
    [`u_point_light_count`]: gl.getUniformLocation(program, "u_point_light_count"),
    [`u_spot_light[0].position`]: gl.getUniformLocation(program, "u_spot_light[0].position"),
    [`u_spot_light[0].direction`]: gl.getUniformLocation(program, "u_spot_light[0].direction"),
    [`u_spot_light[0].ambient`]: gl.getUniformLocation(program, "u_spot_light[0].ambient"),
    [`u_spot_light[0].diffuse`]: gl.getUniformLocation(program, "u_spot_light[0].diffuse"),
    [`u_spot_light[0].specular`]: gl.getUniformLocation(program, "u_spot_light[0].specular"),
    [`u_spot_light[0].cut_off`]: gl.getUniformLocation(program, "u_spot_light[0].cut_off"),
    [`u_spot_light[0].outer_cut_off`]: gl.getUniformLocation(program, "u_spot_light[0].outer_cut_off"),
    [`u_spot_light[1].position`]: gl.getUniformLocation(program, "u_spot_light[1].position"),
    [`u_spot_light[1].direction`]: gl.getUniformLocation(program, "u_spot_light[1].direction"),
    [`u_spot_light[1].ambient`]: gl.getUniformLocation(program, "u_spot_light[1].ambient"),
    [`u_spot_light[1].diffuse`]: gl.getUniformLocation(program, "u_spot_light[1].diffuse"),
    [`u_spot_light[1].specular`]: gl.getUniformLocation(program, "u_spot_light[1].specular"),
    [`u_spot_light[1].cut_off`]: gl.getUniformLocation(program, "u_spot_light[1].cut_off"),
    [`u_spot_light[1].outer_cut_off`]: gl.getUniformLocation(program, "u_spot_light[1].outer_cut_off"),
    [`u_spot_light[2].position`]: gl.getUniformLocation(program, "u_spot_light[2].position"),
    [`u_spot_light[2].direction`]: gl.getUniformLocation(program, "u_spot_light[2].direction"),
    [`u_spot_light[2].ambient`]: gl.getUniformLocation(program, "u_spot_light[2].ambient"),
    [`u_spot_light[2].diffuse`]: gl.getUniformLocation(program, "u_spot_light[2].diffuse"),
    [`u_spot_light[2].specular`]: gl.getUniformLocation(program, "u_spot_light[2].specular"),
    [`u_spot_light[2].cut_off`]: gl.getUniformLocation(program, "u_spot_light[2].cut_off"),
    [`u_spot_light[2].outer_cut_off`]: gl.getUniformLocation(program, "u_spot_light[2].outer_cut_off"),
    [`u_spot_light[3].position`]: gl.getUniformLocation(program, "u_spot_light[3].position"),
    [`u_spot_light[3].direction`]: gl.getUniformLocation(program, "u_spot_light[3].direction"),
    [`u_spot_light[3].ambient`]: gl.getUniformLocation(program, "u_spot_light[3].ambient"),
    [`u_spot_light[3].diffuse`]: gl.getUniformLocation(program, "u_spot_light[3].diffuse"),
    [`u_spot_light[3].specular`]: gl.getUniformLocation(program, "u_spot_light[3].specular"),
    [`u_spot_light[3].cut_off`]: gl.getUniformLocation(program, "u_spot_light[3].cut_off"),
    [`u_spot_light[3].outer_cut_off`]: gl.getUniformLocation(program, "u_spot_light[3].outer_cut_off"),
    [`u_spot_light[4].position`]: gl.getUniformLocation(program, "u_spot_light[4].position"),
    [`u_spot_light[4].direction`]: gl.getUniformLocation(program, "u_spot_light[4].direction"),
    [`u_spot_light[4].ambient`]: gl.getUniformLocation(program, "u_spot_light[4].ambient"),
    [`u_spot_light[4].diffuse`]: gl.getUniformLocation(program, "u_spot_light[4].diffuse"),
    [`u_spot_light[4].specular`]: gl.getUniformLocation(program, "u_spot_light[4].specular"),
    [`u_spot_light[4].cut_off`]: gl.getUniformLocation(program, "u_spot_light[4].cut_off"),
    [`u_spot_light[4].outer_cut_off`]: gl.getUniformLocation(program, "u_spot_light[4].outer_cut_off"),
    [`u_spot_light[5].position`]: gl.getUniformLocation(program, "u_spot_light[5].position"),
    [`u_spot_light[5].direction`]: gl.getUniformLocation(program, "u_spot_light[5].direction"),
    [`u_spot_light[5].ambient`]: gl.getUniformLocation(program, "u_spot_light[5].ambient"),
    [`u_spot_light[5].diffuse`]: gl.getUniformLocation(program, "u_spot_light[5].diffuse"),
    [`u_spot_light[5].specular`]: gl.getUniformLocation(program, "u_spot_light[5].specular"),
    [`u_spot_light[5].cut_off`]: gl.getUniformLocation(program, "u_spot_light[5].cut_off"),
    [`u_spot_light[5].outer_cut_off`]: gl.getUniformLocation(program, "u_spot_light[5].outer_cut_off"),
    [`u_spot_light[6].position`]: gl.getUniformLocation(program, "u_spot_light[6].position"),
    [`u_spot_light[6].direction`]: gl.getUniformLocation(program, "u_spot_light[6].direction"),
    [`u_spot_light[6].ambient`]: gl.getUniformLocation(program, "u_spot_light[6].ambient"),
    [`u_spot_light[6].diffuse`]: gl.getUniformLocation(program, "u_spot_light[6].diffuse"),
    [`u_spot_light[6].specular`]: gl.getUniformLocation(program, "u_spot_light[6].specular"),
    [`u_spot_light[6].cut_off`]: gl.getUniformLocation(program, "u_spot_light[6].cut_off"),
    [`u_spot_light[6].outer_cut_off`]: gl.getUniformLocation(program, "u_spot_light[6].outer_cut_off"),
    [`u_spot_light[7].position`]: gl.getUniformLocation(program, "u_spot_light[7].position"),
    [`u_spot_light[7].direction`]: gl.getUniformLocation(program, "u_spot_light[7].direction"),
    [`u_spot_light[7].ambient`]: gl.getUniformLocation(program, "u_spot_light[7].ambient"),
    [`u_spot_light[7].diffuse`]: gl.getUniformLocation(program, "u_spot_light[7].diffuse"),
    [`u_spot_light[7].specular`]: gl.getUniformLocation(program, "u_spot_light[7].specular"),
    [`u_spot_light[7].cut_off`]: gl.getUniformLocation(program, "u_spot_light[7].cut_off"),
    [`u_spot_light[7].outer_cut_off`]: gl.getUniformLocation(program, "u_spot_light[7].outer_cut_off"),
    [`u_spot_light[8].position`]: gl.getUniformLocation(program, "u_spot_light[8].position"),
    [`u_spot_light[8].direction`]: gl.getUniformLocation(program, "u_spot_light[8].direction"),
    [`u_spot_light[8].ambient`]: gl.getUniformLocation(program, "u_spot_light[8].ambient"),
    [`u_spot_light[8].diffuse`]: gl.getUniformLocation(program, "u_spot_light[8].diffuse"),
    [`u_spot_light[8].specular`]: gl.getUniformLocation(program, "u_spot_light[8].specular"),
    [`u_spot_light[8].cut_off`]: gl.getUniformLocation(program, "u_spot_light[8].cut_off"),
    [`u_spot_light[8].outer_cut_off`]: gl.getUniformLocation(program, "u_spot_light[8].outer_cut_off"),
    [`u_spot_light[9].position`]: gl.getUniformLocation(program, "u_spot_light[9].position"),
    [`u_spot_light[9].direction`]: gl.getUniformLocation(program, "u_spot_light[9].direction"),
    [`u_spot_light[9].ambient`]: gl.getUniformLocation(program, "u_spot_light[9].ambient"),
    [`u_spot_light[9].diffuse`]: gl.getUniformLocation(program, "u_spot_light[9].diffuse"),
    [`u_spot_light[9].specular`]: gl.getUniformLocation(program, "u_spot_light[9].specular"),
    [`u_spot_light[9].cut_off`]: gl.getUniformLocation(program, "u_spot_light[9].cut_off"),
    [`u_spot_light[9].outer_cut_off`]: gl.getUniformLocation(program, "u_spot_light[9].outer_cut_off"),
    [`u_spot_light_count`]: gl.getUniformLocation(program, "u_spot_light_count"),
    [`u_spot_light_shadow_map`]: gl.getUniformLocation(program, "u_spot_light_shadow_map"),
    [`u_spot_light_space[0]`]: gl.getUniformLocation(program, "u_spot_light_space[0]"),
    [`u_spot_light_space[1]`]: gl.getUniformLocation(program, "u_spot_light_space[1]"),
    [`u_spot_light_space[2]`]: gl.getUniformLocation(program, "u_spot_light_space[2]"),
    [`u_spot_light_space[3]`]: gl.getUniformLocation(program, "u_spot_light_space[3]"),
    [`u_spot_light_space[4]`]: gl.getUniformLocation(program, "u_spot_light_space[4]"),
    [`u_spot_light_space[5]`]: gl.getUniformLocation(program, "u_spot_light_space[5]"),
    [`u_spot_light_space[6]`]: gl.getUniformLocation(program, "u_spot_light_space[6]"),
    [`u_spot_light_space[7]`]: gl.getUniformLocation(program, "u_spot_light_space[7]"),
    [`u_spot_light_space[8]`]: gl.getUniformLocation(program, "u_spot_light_space[8]"),
    [`u_spot_light_space[9]`]: gl.getUniformLocation(program, "u_spot_light_space[9]"),
    [`u_tex_index`]: gl.getUniformLocation(program, "u_tex_index"),
  };
  const u_dir_light = {
    direction: {
      /**
       * Set the value of uniform `u_dir_light.direction`.
       *
       * @param {number} x
       * @param {number} y
       * @param {number} z
       */
      set(x, y, z) {
        gl.uniform3f(locs[`u_dir_light.direction`],x, y, z);
      },
      /**
       * Returns the value of uniform `u_dir_light.direction`.
       *
       * @returns {Float32Array}
       */
      get() {
        return gl.getUniform(program, locs[`u_dir_light.direction`]);
      },
    },
    ambient: {
      /**
       * Set the value of uniform `u_dir_light.ambient`.
       *
       * @param {number} x
       * @param {number} y
       * @param {number} z
       */
      set(x, y, z) {
        gl.uniform3f(locs[`u_dir_light.ambient`],x, y, z);
      },
      /**
       * Returns the value of uniform `u_dir_light.ambient`.
       *
       * @returns {Float32Array}
       */
      get() {
        return gl.getUniform(program, locs[`u_dir_light.ambient`]);
      },
    },
    diffuse: {
      /**
       * Set the value of uniform `u_dir_light.diffuse`.
       *
       * @param {number} x
       * @param {number} y
       * @param {number} z
       */
      set(x, y, z) {
        gl.uniform3f(locs[`u_dir_light.diffuse`],x, y, z);
      },
      /**
       * Returns the value of uniform `u_dir_light.diffuse`.
       *
       * @returns {Float32Array}
       */
      get() {
        return gl.getUniform(program, locs[`u_dir_light.diffuse`]);
      },
    },
    specular: {
      /**
       * Set the value of uniform `u_dir_light.specular`.
       *
       * @param {number} x
       * @param {number} y
       * @param {number} z
       */
      set(x, y, z) {
        gl.uniform3f(locs[`u_dir_light.specular`],x, y, z);
      },
      /**
       * Returns the value of uniform `u_dir_light.specular`.
       *
       * @returns {Float32Array}
       */
      get() {
        return gl.getUniform(program, locs[`u_dir_light.specular`]);
      },
    },
    intensity: {
      /**
       * Set the value of uniform `u_dir_light.intensity`.
       *
       * @param {number} value
       */
      set(value) {
        gl.uniform1f(locs[`u_dir_light.intensity`],value);
      },
      /**
       * Returns the value of uniform `u_dir_light.intensity`.
       *
       * @returns {number}
       */
      get() {
        return gl.getUniform(program, locs[`u_dir_light.intensity`]);
      },
    },

  };


  const u_directional_light_shadow_map = {
      /**
       * Set the value of uniform `u_directional_light_shadow_map`.
       *
       * @param {number} value
       */
      set(value) {
        gl.uniform1i(locs[`u_directional_light_shadow_map`],value);
      },
      /**
       * Returns the value of uniform `u_directional_light_shadow_map`.
       *
       * @returns {number}
       */
      get() {
        return gl.getUniform(program, locs[`u_directional_light_shadow_map`]);
      },

  };


  const u_directional_light_space = {
      /**
       * Set the value of uniform `u_directional_light_space`.
       *
       * @param {Float32Array} matrix
       * @param {boolean} transpose
       */
      set(matrix, transpose = false) {
        gl.uniformMatrix4fv(locs[`u_directional_light_space`], transpose, matrix);
      },
      /**
       * Returns the value of uniform `u_directional_light_space`.
       *
       * @returns {Float32Array}
       */
      get() {
        return gl.getUniform(program, locs[`u_directional_light_space`]);
      },

  };


  const u_material = {
    diffuse: {
      /**
       * Set the value of uniform `u_material.diffuse`.
       *
       * @param {number} value
       */
      set(value) {
        gl.uniform1i(locs[`u_material.diffuse`],value);
      },
      /**
       * Returns the value of uniform `u_material.diffuse`.
       *
       * @returns {number}
       */
      get() {
        return gl.getUniform(program, locs[`u_material.diffuse`]);
      },
    },
    specular: {
      /**
       * Set the value of uniform `u_material.specular`.
       *
       * @param {number} value
       */
      set(value) {
        gl.uniform1i(locs[`u_material.specular`],value);
      },
      /**
       * Returns the value of uniform `u_material.specular`.
       *
       * @returns {number}
       */
      get() {
        return gl.getUniform(program, locs[`u_material.specular`]);
      },
    },
    shininess: {
      /**
       * Set the value of uniform `u_material.shininess`.
       *
       * @param {number} value
       */
      set(value) {
        gl.uniform1f(locs[`u_material.shininess`],value);
      },
      /**
       * Returns the value of uniform `u_material.shininess`.
       *
       * @returns {number}
       */
      get() {
        return gl.getUniform(program, locs[`u_material.shininess`]);
      },
    },

  };


  const u_palette = {
      /**
       * Set the value of uniform `u_palette`.
       *
       * @param {number} value
       */
      set(value) {
        gl.uniform1i(locs[`u_palette`],value);
      },
      /**
       * Returns the value of uniform `u_palette`.
       *
       * @returns {number}
       */
      get() {
        return gl.getUniform(program, locs[`u_palette`]);
      },

  };


  const u_point_light = new Array(10).fill(undefined).map((_, u_point_lighti) => ({
    position: {
      /**
       * Set the value of uniform `u_point_light[${u_point_lighti}].position`.
       *
       * @param {number} x
       * @param {number} y
       * @param {number} z
       */
      set(x, y, z) {
        gl.uniform3f(locs[`u_point_light[${u_point_lighti}].position`],x, y, z);
      },
      /**
       * Returns the value of uniform `u_point_light[${u_point_lighti}].position`.
       *
       * @returns {Float32Array}
       */
      get() {
        return gl.getUniform(program, locs[`u_point_light[${u_point_lighti}].position`]);
      },
    },
    ambient: {
      /**
       * Set the value of uniform `u_point_light[${u_point_lighti}].ambient`.
       *
       * @param {number} x
       * @param {number} y
       * @param {number} z
       */
      set(x, y, z) {
        gl.uniform3f(locs[`u_point_light[${u_point_lighti}].ambient`],x, y, z);
      },
      /**
       * Returns the value of uniform `u_point_light[${u_point_lighti}].ambient`.
       *
       * @returns {Float32Array}
       */
      get() {
        return gl.getUniform(program, locs[`u_point_light[${u_point_lighti}].ambient`]);
      },
    },
    diffuse: {
      /**
       * Set the value of uniform `u_point_light[${u_point_lighti}].diffuse`.
       *
       * @param {number} x
       * @param {number} y
       * @param {number} z
       */
      set(x, y, z) {
        gl.uniform3f(locs[`u_point_light[${u_point_lighti}].diffuse`],x, y, z);
      },
      /**
       * Returns the value of uniform `u_point_light[${u_point_lighti}].diffuse`.
       *
       * @returns {Float32Array}
       */
      get() {
        return gl.getUniform(program, locs[`u_point_light[${u_point_lighti}].diffuse`]);
      },
    },
    specular: {
      /**
       * Set the value of uniform `u_point_light[${u_point_lighti}].specular`.
       *
       * @param {number} x
       * @param {number} y
       * @param {number} z
       */
      set(x, y, z) {
        gl.uniform3f(locs[`u_point_light[${u_point_lighti}].specular`],x, y, z);
      },
      /**
       * Returns the value of uniform `u_point_light[${u_point_lighti}].specular`.
       *
       * @returns {Float32Array}
       */
      get() {
        return gl.getUniform(program, locs[`u_point_light[${u_point_lighti}].specular`]);
      },
    },
    radius: {
      /**
       * Set the value of uniform `u_point_light[${u_point_lighti}].radius`.
       *
       * @param {number} value
       */
      set(value) {
        gl.uniform1f(locs[`u_point_light[${u_point_lighti}].radius`],value);
      },
      /**
       * Returns the value of uniform `u_point_light[${u_point_lighti}].radius`.
       *
       * @returns {number}
       */
      get() {
        return gl.getUniform(program, locs[`u_point_light[${u_point_lighti}].radius`]);
      },
    },
    intensity: {
      /**
       * Set the value of uniform `u_point_light[${u_point_lighti}].intensity`.
       *
       * @param {number} value
       */
      set(value) {
        gl.uniform1f(locs[`u_point_light[${u_point_lighti}].intensity`],value);
      },
      /**
       * Returns the value of uniform `u_point_light[${u_point_lighti}].intensity`.
       *
       * @returns {number}
       */
      get() {
        return gl.getUniform(program, locs[`u_point_light[${u_point_lighti}].intensity`]);
      },
    },

  }));

  const u_point_light_count = {
      /**
       * Set the value of uniform `u_point_light_count`.
       *
       * @param {number} value
       */
      set(value) {
        gl.uniform1i(locs[`u_point_light_count`],value);
      },
      /**
       * Returns the value of uniform `u_point_light_count`.
       *
       * @returns {number}
       */
      get() {
        return gl.getUniform(program, locs[`u_point_light_count`]);
      },

  };


  const u_spot_light = new Array(10).fill(undefined).map((_, u_spot_lighti) => ({
    position: {
      /**
       * Set the value of uniform `u_spot_light[${u_spot_lighti}].position`.
       *
       * @param {number} x
       * @param {number} y
       * @param {number} z
       */
      set(x, y, z) {
        gl.uniform3f(locs[`u_spot_light[${u_spot_lighti}].position`],x, y, z);
      },
      /**
       * Returns the value of uniform `u_spot_light[${u_spot_lighti}].position`.
       *
       * @returns {Float32Array}
       */
      get() {
        return gl.getUniform(program, locs[`u_spot_light[${u_spot_lighti}].position`]);
      },
    },
    direction: {
      /**
       * Set the value of uniform `u_spot_light[${u_spot_lighti}].direction`.
       *
       * @param {number} x
       * @param {number} y
       * @param {number} z
       */
      set(x, y, z) {
        gl.uniform3f(locs[`u_spot_light[${u_spot_lighti}].direction`],x, y, z);
      },
      /**
       * Returns the value of uniform `u_spot_light[${u_spot_lighti}].direction`.
       *
       * @returns {Float32Array}
       */
      get() {
        return gl.getUniform(program, locs[`u_spot_light[${u_spot_lighti}].direction`]);
      },
    },
    ambient: {
      /**
       * Set the value of uniform `u_spot_light[${u_spot_lighti}].ambient`.
       *
       * @param {number} x
       * @param {number} y
       * @param {number} z
       */
      set(x, y, z) {
        gl.uniform3f(locs[`u_spot_light[${u_spot_lighti}].ambient`],x, y, z);
      },
      /**
       * Returns the value of uniform `u_spot_light[${u_spot_lighti}].ambient`.
       *
       * @returns {Float32Array}
       */
      get() {
        return gl.getUniform(program, locs[`u_spot_light[${u_spot_lighti}].ambient`]);
      },
    },
    diffuse: {
      /**
       * Set the value of uniform `u_spot_light[${u_spot_lighti}].diffuse`.
       *
       * @param {number} x
       * @param {number} y
       * @param {number} z
       */
      set(x, y, z) {
        gl.uniform3f(locs[`u_spot_light[${u_spot_lighti}].diffuse`],x, y, z);
      },
      /**
       * Returns the value of uniform `u_spot_light[${u_spot_lighti}].diffuse`.
       *
       * @returns {Float32Array}
       */
      get() {
        return gl.getUniform(program, locs[`u_spot_light[${u_spot_lighti}].diffuse`]);
      },
    },
    specular: {
      /**
       * Set the value of uniform `u_spot_light[${u_spot_lighti}].specular`.
       *
       * @param {number} x
       * @param {number} y
       * @param {number} z
       */
      set(x, y, z) {
        gl.uniform3f(locs[`u_spot_light[${u_spot_lighti}].specular`],x, y, z);
      },
      /**
       * Returns the value of uniform `u_spot_light[${u_spot_lighti}].specular`.
       *
       * @returns {Float32Array}
       */
      get() {
        return gl.getUniform(program, locs[`u_spot_light[${u_spot_lighti}].specular`]);
      },
    },
    cut_off: {
      /**
       * Set the value of uniform `u_spot_light[${u_spot_lighti}].cut_off`.
       *
       * @param {number} value
       */
      set(value) {
        gl.uniform1f(locs[`u_spot_light[${u_spot_lighti}].cut_off`],value);
      },
      /**
       * Returns the value of uniform `u_spot_light[${u_spot_lighti}].cut_off`.
       *
       * @returns {number}
       */
      get() {
        return gl.getUniform(program, locs[`u_spot_light[${u_spot_lighti}].cut_off`]);
      },
    },
    outer_cut_off: {
      /**
       * Set the value of uniform `u_spot_light[${u_spot_lighti}].outer_cut_off`.
       *
       * @param {number} value
       */
      set(value) {
        gl.uniform1f(locs[`u_spot_light[${u_spot_lighti}].outer_cut_off`],value);
      },
      /**
       * Returns the value of uniform `u_spot_light[${u_spot_lighti}].outer_cut_off`.
       *
       * @returns {number}
       */
      get() {
        return gl.getUniform(program, locs[`u_spot_light[${u_spot_lighti}].outer_cut_off`]);
      },
    },

  }));

  const u_spot_light_count = {
      /**
       * Set the value of uniform `u_spot_light_count`.
       *
       * @param {number} value
       */
      set(value) {
        gl.uniform1i(locs[`u_spot_light_count`],value);
      },
      /**
       * Returns the value of uniform `u_spot_light_count`.
       *
       * @returns {number}
       */
      get() {
        return gl.getUniform(program, locs[`u_spot_light_count`]);
      },

  };


  const u_spot_light_shadow_map = {
      /**
       * Set the value of uniform `u_spot_light_shadow_map`.
       *
       * @param {number} value
       */
      set(value) {
        gl.uniform1i(locs[`u_spot_light_shadow_map`],value);
      },
      /**
       * Returns the value of uniform `u_spot_light_shadow_map`.
       *
       * @returns {number}
       */
      get() {
        return gl.getUniform(program, locs[`u_spot_light_shadow_map`]);
      },

  };


  const u_spot_light_space = new Array(10).fill(undefined).map((_, u_spot_light_spacei) => ({
      /**
       * Set the value of uniform `u_spot_light_space[${u_spot_light_spacei}]`.
       *
       * @param {Float32Array} matrix
       * @param {boolean} transpose
       */
      set(matrix, transpose = false) {
        gl.uniformMatrix4fv(locs[`u_spot_light_space[${u_spot_light_spacei}]`], transpose, matrix);
      },
      /**
       * Returns the value of uniform `u_spot_light_space[${u_spot_light_spacei}]`.
       *
       * @returns {Float32Array}
       */
      get() {
        return gl.getUniform(program, locs[`u_spot_light_space[${u_spot_light_spacei}]`]);
      },

  }));

  const u_tex_index = {
      /**
       * Set the value of uniform `u_tex_index`.
       *
       * @param {number} value
       */
      set(value) {
        gl.uniform1i(locs[`u_tex_index`],value);
      },
      /**
       * Returns the value of uniform `u_tex_index`.
       *
       * @returns {number}
       */
      get() {
        return gl.getUniform(program, locs[`u_tex_index`]);
      },

  };


  return {
    u_dir_light,
    u_directional_light_shadow_map,
    u_directional_light_space,
    u_material,
    u_palette,
    u_point_light,
    u_point_light_count,
    u_spot_light,
    u_spot_light_count,
    u_spot_light_shadow_map,
    u_spot_light_space,
    u_tex_index,
  };
};

/**
 * The standard shader program.
 */
export const StandardShader = {
  name: "standard",
  vertex: VERTEX_SRC,
  fragment: FRAGMENT_SRC,
  createUniforms,
};
