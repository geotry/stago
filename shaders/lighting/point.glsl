struct PointLight {
  vec3 position;

  vec3 ambient;
  vec3 diffuse;
  vec3 specular;

  float radius;
  float intensity;
};

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

    // Ambient light
  vec4 ambient = vec4(light.ambient, 1.0) * diffuse_color;
  ambient *= attenuation;
  color += ambient;

    // Diffuse light
  float diff = max(dot(norm, light_dir), 0.0);
  vec4 diffuse = vec4(light.diffuse * diff, 1.0) * diffuse_color;
  diffuse *= attenuation;
  color += diffuse;

    // Specular light
  float spec = pow(max(dot(norm, half_dir), 0.0), shininess);
  vec4 specular = vec4(light.specular * spec, 1.0) * specular_color;
  specular *= attenuation;
  color += specular;

  return color;
}