const { LightsStruct } = require("./lights.wgsl.js");
// Read: https://toji.dev/webgpu-best-practices/dynamic-shader-construction

const gammaCorrection = /* wgsl */`
fn gammaCorrection(color: vec4f) -> vec4f {
  return vec4f(color.rgb * (1/2.2), color.a);
}
`;

export const createStandardShader = (opts = { lightCount: 10 }) => {
  const lightCount = opts.lightCount ?? 10;

  const source = /* wgsl */`
    struct Camera {
      view: mat4x4f,
      projection: mat4x4f,
    }

    struct Model {
      model : mat4x4f,
      normalModel : mat4x4f,
      tint: vec4f,
    }

    ${LightsStruct}

    struct VertexInput {
      @location(0) position : vec3f,
      @location(1) uv : vec2f,
      @location(2) normal : vec3f,
    }

    struct VertexOutput {
      @builtin(position) position : vec4f,
      @location(1) texCoords : vec2f,
      @location(3) normal : vec3f,
      @location(4) fragPos : vec3f,
      @location(5) viewPos : vec3f,
      @location(6) shadowPos : vec3f,
      @location(7) tint : vec4f,
    }
    
    override shadowDepthTextureSize : f32;
    override debug : bool = false;
    
    @group(0) @binding(0) var<uniform> camera : Camera;
    @group(0) @binding(1) var<uniform> lights: Lights;
    @group(0) @binding(2) var textureSampler: sampler;
    @group(0) @binding(3) var shadowSampler: sampler_comparison;
    @group(0) @binding(4) var shadowMap: texture_depth_2d;
    @group(0) @binding(5) var spotLightShadowMaps: texture_depth_2d_array;

    @group(1) @binding(0) var diffuseTexture: texture_2d<f32>;
    @group(1) @binding(1) var specularTexture: texture_2d<f32>;
    @group(1) @binding(2) var<uniform> shininess: f32;
    
    @group(2) @binding(0) var<uniform> object : Model;

    @vertex
    fn vertexMain(input : VertexInput) -> VertexOutput {
      let position = (camera.projection * camera.view * object.model * vec4f(input.position, 1));

      var output : VertexOutput;
      output.position = position;
      output.texCoords = input.uv;
      output.normal = (object.normalModel * vec4f(input.normal, 1)).xyz;
      output.fragPos = (object.model * vec4f(input.position, 1)).xyz;
      output.viewPos = (camera.view * vec4f(0, 0, 0, 1)).xyz;
      output.tint = object.tint;

      // Convert fragment light position to [0, 1]
      let directionalLightPos = lights.directionalLight.viewProjectionMatrix * object.model * vec4f(input.position, 1.0);
      output.shadowPos = vec3f((directionalLightPos.xy / directionalLightPos.w) * vec2f(0.5, -0.5) + vec2f(0.5), directionalLightPos.z / directionalLightPos.w);

      return output;
    }

    @fragment
    fn fragmentMain(input: VertexOutput) -> @location(0) vec4f {
      var color = vec4f(0, 0, 0, 0);
      let diffuse = textureSample(diffuseTexture, textureSampler, input.texCoords);
      let specular = vec4f(1) * textureSample(specularTexture, textureSampler, input.texCoords).r;
      var normal = normalize(input.normal);

      // Lights
      color += directionalLightColor(lights.directionalLight, input.viewPos, input.fragPos, input.shadowPos, normal, diffuse, specular);
      for (var i = 0; i < lights.pointLightCount; i++) {
        color += pointLightColor(lights.pointLights[i], input.viewPos, input.fragPos, normal, diffuse, specular);
      }
      for (var i = 0; i < lights.spotLightCount; i++) {
        let spotLightPos = lights.spotLights[i].viewProjectionMatrix * vec4f(input.fragPos, 1);
        let spotLightShadowPos = vec3f((spotLightPos.xy/ spotLightPos.w) * vec2f(0.5, -0.5) + vec2f(0.5), spotLightPos.z / spotLightPos.w);
        color += spotLightColor(lights.spotLights[i], i, input.viewPos, input.fragPos, spotLightShadowPos, normal, diffuse, specular);
      }

      // To show only the tint of the object
      // if (color.a > 0) {
      //   color = input.tint;
      // }
      color *= input.tint;

      color = gammaCorrection(color);

      // let depth = 1-(input.fragPos.z / input.fragPos.w);
      // color = vec4f(depth, depth, depth, 1);
      // color = input.fragPos;
      // color = vec4(normalize(input.normal), 1);

      // return vec4f(normal, 1);

      return color;
    }

    fn directionalLightColor(
      light: DirectionalLight,
      viewPos: vec3f,
      fragPos: vec3f,
      shadowPos: vec3f,
      normal: vec3f,
      diffuseColor: vec4f,
      specularColor: vec4f,
    ) -> vec4f {
      let lightDir = normalize(-light.direction);
      let viewDir = normalize(viewPos - fragPos);
      let halfDir = normalize(lightDir + viewDir);

      let ambient = vec4f(light.ambient, 1) * diffuseColor;

      let diff = max(dot(normal, lightDir), 0);
      var diffuse = vec4f(light.diffuse * diff, 1) * diffuseColor;
      diffuse = diffuse * light.intensity;

      let spec = pow(max(dot(normal, halfDir), 0), shininess);
      var specular = vec4f(light.specular * spec, 1) * specularColor;
      specular = specular * light.intensity;

      // Shadow
      // Percentage-closer filtering. Sample texels in the region
      // to smooth the result.
      var visibility = 0.0;
      let oneOverShadowDepthTextureSize = 1.0 / shadowDepthTextureSize;
      for (var y = -1; y <= 1; y++) {
        for (var x = -1; x <= 1; x++) {
          let offset = vec2f(vec2(x, y)) * oneOverShadowDepthTextureSize;
          visibility += textureSampleCompare(shadowMap, shadowSampler, shadowPos.xy + offset, shadowPos.z - 0.002);
        }
      }
      visibility /= 9.0;
      
      // Outside of shadow map
      if (fragPos.z > 80) {
        visibility = 1.0;
      }

      // let lambertFactor = max(dot(normalize(vec3f(-300, 300, -300) - fragPos), normal), 0.0);
      let lightingFactor = min(visibility * 1.0, 1.0);

      return ambient + (diffuse * lightingFactor) + (specular * lightingFactor);
    }

    fn pointLightColor(
      light: PointLight,
      viewPos: vec3f,
      fragPos: vec3f,
      normal: vec3f,
      diffuseColor: vec4f,
      specularColor: vec4f,
    ) -> vec4f {
      let lightVector = light.position - fragPos;
      let lightDir = normalize(lightVector);
      let viewDir = normalize(viewPos - fragPos);
      let halfDir = normalize(lightDir + viewDir); 
      
      let distance = length(lightVector);
      let attenuation = attenuateCusp(distance, light.radius, light.intensity, 1.0);

      let diff = max(dot(normal, lightDir), 0);
      var diffuse = vec4f(light.diffuse * diff, 1) * diffuseColor;
      diffuse = diffuse * attenuation;
      diffuse = diffuse * light.intensity;

      let spec = pow(max(dot(normal, halfDir), 0), shininess);
      var specular = vec4f(light.specular * spec, 1) * specularColor;
      specular = specular * attenuation;
      specular = specular * light.intensity;

      // todo: shadow

      //

      return diffuse + specular;
    }

    fn spotLightColor(
      light: SpotLight,
      index: i32,
      viewPos: vec3f,
      fragPos: vec3f,
      shadowPos: vec3f,
      normal: vec3f,
      diffuseColor: vec4f,
      specularColor: vec4f,
    ) -> vec4f {
      let lightVector = light.position - fragPos;
      let lightDir = normalize(lightVector);
      let viewDir = normalize(viewPos - fragPos);
      let halfDir = normalize(lightDir + viewDir);

      let distance = length(lightVector);
      let attenuation = attenuateCusp(distance, light.range, light.intensity, 1.0);
      let theta = dot(lightDir, normalize(-light.direction));
      let epsilon = light.cutOff - light.outerCutOff;
      let intensity = clamp((theta - light.outerCutOff) / epsilon, 0.0, 1.0);

      let diff = max(dot(normal, lightDir), 0);
      var diffuse = vec4f(light.diffuse * diff, 1) * diffuseColor;
      diffuse = diffuse * intensity * attenuation * light.intensity;

      let spec = pow(max(dot(normal, halfDir), 0), shininess);
      var specular = vec4f(light.specular * spec, 1) * specularColor;
      specular = specular * intensity * attenuation * light.intensity;

      // Shadow
      // Percentage-closer filtering. Sample texels in the region
      // to smooth the result.
      var visibility = 0.0;
      let oneOverShadowDepthTextureSize = 1.0 / shadowDepthTextureSize;
      for (var y = -1; y <= 1; y++) {
        for (var x = -1; x <= 1; x++) {
          let offset = vec2f(vec2(x, y)) * oneOverShadowDepthTextureSize;
          visibility += textureSampleCompare(spotLightShadowMaps, shadowSampler, shadowPos.xy + offset, index, shadowPos.z - 0.005);
        }
      }
      visibility /= 9.0;

      let lambertFactor = max(dot(normalize(light.position - fragPos), normal), 0.0);
      let lightingFactor = min(visibility * lambertFactor, 1.0);

      return (diffuse * lightingFactor) + (specular * lightingFactor);
    }

    fn attenuateCusp(distance : f32, radius : f32, maxIntensity : f32, fallOff : f32) -> f32 {
      let s = distance / radius;
      if (s >= 1.0) {
        return 0.0;
      }
      let s2 = sqrt(s);
      return maxIntensity * sqrt(1.0 - s2) / (1.0 + fallOff * s);
    }

    ${gammaCorrection}
  `;

  return { name: "standard", source };
};
