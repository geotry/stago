export const LightsStruct = /* wgsl */`
struct Lights {
  directionalLight: DirectionalLight,
  pointLightCount : i32,
  pointLights : array<PointLight, 10>,
  spotLightCount : i32,
  spotLights : array<SpotLight, 10>,
}

struct DirectionalLight {
  viewProjectionMatrix: mat4x4f,
  direction : vec3f,
  ambient : vec3f,
  diffuse : vec3f,
  specular : vec3f,
  intensity : f32,
}

struct PointLight {
  position : vec3f,
  ambient : vec3f,
  diffuse : vec3f,
  specular : vec3f,
  radius : f32,
  intensity : f32,
}

struct SpotLight {
  viewProjectionMatrix: mat4x4f,
  position : vec3f,
  direction : vec3f,
  ambient : vec3f,
  diffuse : vec3f,
  specular : vec3f,
  cutOff : f32,
  outerCutOff : f32,
  range : f32,
  intensity : f32,
}
`;