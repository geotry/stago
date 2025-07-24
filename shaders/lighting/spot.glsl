struct SpotLight {
  vec3 position;
  vec3 direction;

  vec3 ambient;
  vec3 diffuse;
  vec3 specular;

  float cut_off;
  float outer_cut_off;
};

// Sampler in struct/uniform block seems unstable
uniform sampler2DShadow u_spot_light_shadow_map;

vec4 ComputeSpotLight(SpotLight light, int index, mat4 view, vec4 diffuse_color, vec4 specular_color, float shininess, vec3 norm, vec3 fragPos, vec4 lightFragPos) {
  vec4 color;

  vec3 light_dir = normalize(vec3(view * vec4(light.position, 1.0)) - fragPos);
  vec3 view_dir = normalize(-fragPos);
  vec3 half_dir = normalize(light_dir + view_dir);

  float theta = dot(light_dir, normalize(-light.direction));
  float epsilon = light.cut_off - light.outer_cut_off;
  float intensity = clamp((theta - light.outer_cut_off) / epsilon, 0.0, 1.0);

  // Ambient light
  vec4 ambient = vec4(light.ambient, 1.0) * diffuse_color;
  ambient *= intensity;

  // Diffuse light
  float diff = max(dot(norm, light_dir), 0.0);
  vec4 diffuse = vec4(light.diffuse * diff, 1.0) * diffuse_color;
  diffuse *= intensity;

  // Specular light
  float spec = pow(max(dot(norm, half_dir), 0.0), shininess);
  vec4 specular = vec4(light.specular * spec, 1.0) * specular_color;
  specular *= intensity;

  float shadow = ShadowCalculation(u_spot_light_shadow_map, lightFragPos, max(0.002 * (1.0 - dot(norm, light_dir)), 0.002));
  color = ambient + (shadow * diffuse) + (specular * shadow);

  return color;
}