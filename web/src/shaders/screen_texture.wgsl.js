export const createScreenTextureShader = () => {
  const source = /* wgsl */`
struct VertexOut {
  @builtin(position) position: vec4f,
  @location(0) uv: vec2f,
}

override textureWidth: f32;
override textureHeight: f32;

// @group(0) @binding(0) var textureSampler : sampler;
@group(0) @binding(0) var depthTexture : texture_depth_2d;

@vertex
fn vertexMain(@builtin(vertex_index) vertexIndex : u32) -> VertexOut {
  var out : VertexOut;
  switch (vertexIndex) {
    case 0: {
      out.position = vec4f(1.0, -1.0, 0.0, 1.0);
      out.uv = vec2f(1.0, 1.0);
      break;
    }
    case 1: {
      out.position = vec4f(1.0, 1.0, 0.0, 1.0);
      out.uv = vec2f(1.0, 0.0);
      break;
    }
    case 2: {
      out.position = vec4f(-1.0, 1.0, 0.0, 1.0);
      out.uv = vec2f(0.0, 0.0);
      break;
    }
    case 3: {
      out.position = vec4f(1.0, -1.0, 0.0, 1.0);
      out.uv = vec2f(1.0, 1.0);
      break;
    }
    case 4: {
      out.position = vec4f(-1.0, 1.0, 0.0, 1.0);
      out.uv = vec2f(0.0, 0.0);
      break;
    }
    case 5: {
      out.position = vec4f(-1.0, -1.0, 0.0, 1.0);
      out.uv = vec2f(0.0, 1.0);
      break;
    }
    default: {
      break;
    }
  }
  return out;
}
@fragment
fn fragmentMain(input : VertexOut) -> @location(0) vec4f {
  var texCoords = vec2u(input.uv * vec2f(textureWidth, textureHeight));
  let depth = textureLoad(depthTexture, texCoords, 0);
  return vec4f(1-depth, 1-depth, 1-depth, 1);
}
`;

  return { name: "screen_texture", source };
};
