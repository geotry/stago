// https://webgpu.github.io/webgpu-samples/?sample=shadowMapping#vertexShadow.wgsl
export const createShadowMappingShader = () => {
  const source = /* wgsl */`
    @group(0) @binding(0) var<uniform> lightViewProjection : mat4x4f;
    @group(1) @binding(0) var<uniform> model : mat4x4f;

    @vertex
    fn vertexMain(@location(0) position : vec3f) -> @builtin(position) vec4f {
      return lightViewProjection * model * vec4f(position, 1.0);
    }
  `;

  return { name: "shadow_mapping", source };
};
