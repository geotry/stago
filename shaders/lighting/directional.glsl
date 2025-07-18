struct DirectionalLight {
  vec3 direction;
  vec3 ambient;
  vec3 diffuse;
  vec3 specular;
  float intensity;
  bool cast_shadow;
  mat4 view;
};

uniform sampler2DShadow u_directional_light_shadow_map;

vec4 ComputeDirectionalLight(in DirectionalLight light, vec4 diffuse_color, vec4 specular_color, float shininess, vec3 norm, vec3 fragPos, vec4 lightFragPos) {
  vec4 color;

  vec3 light_dir = normalize(-light.direction);
  vec3 view_dir = normalize(-fragPos);
  vec3 half_dir = normalize(light_dir + view_dir);

  // Ambient light
  vec4 ambient = vec4(light.ambient, 1.0) * diffuse_color;

  // Diffuse light
  float diff = max(dot(norm, light_dir), 0.0);
  vec4 diffuse = vec4(light.diffuse * diff, 1.0) * diffuse_color;
  diffuse *= light.intensity;

  // Specular light
  float spec = pow(max(dot(norm, half_dir), 0.0), shininess);
  vec4 specular = vec4(light.specular * spec, 1.0) * specular_color;
  specular *= light.intensity;

  if(light.cast_shadow) {
    float shadow = ShadowCalculation(u_directional_light_shadow_map, lightFragPos, max(0.002 * (1.0 - dot(norm, light_dir)), 0.002));
    color = ambient + (shadow * diffuse) + (specular * shadow);
  } else {
    color = ambient + diffuse + specular;
  }

  return color;
}