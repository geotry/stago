const { createWebGPUPipeline } = require("./pipeline.js");
const { createStandardShader } = require("./shaders/standard_wgsl.js");
const { createShadowMappingShader } = require("./shaders/shadow_mapping.wgsl.js");
const { createScreenTextureShader } = require("./shaders/screen_texture.wgsl.js");
const { mat4, vec3 } = require("wgpu-matrix");
const { loadImageBitmap } = require("./utils.js");

const DEBUG = false;

/**
 * 
 * @param {OffscreenCanvas} canvas 
 * @param {webgpu} ctx
 */
export const createContext = async (canvas, ctx) => {
  let pipeline;

  switch (ctx) {
    case "webgpu": {
      if (!navigator.gpu) {
        throw new Error(`WebGPU not supported on this device.`);
      }

      const adapter = await navigator.gpu.requestAdapter();
      if (!adapter) {
        throw new Error("No appropriate GPUAdapter found.");
      }

      const device = await adapter.requestDevice();
      const context = canvas.getContext("webgpu");
      if (!context) {
        throw new Error("Could not get WebGPU context.");
      }
      const canvasFormat = navigator.gpu.getPreferredCanvasFormat();
      context.configure({ device: device, format: canvasFormat, alphaMode: "premultiplied" });

      pipeline = await webGPUPipeline(context, device, canvasFormat);
      break;
    }
    default:
      throw new Error(`Context ${context} invalid.`);
  }


  return {
    reset() {
      pipeline.reset();
    },
    /**
     * Set an option in pipeline.
     *
     * @param {string} option 
     * @param {string | number | boolean} value 
     */
    setOption(option, value) {
      pipeline.setOption(option, value);
    },
    /**
     * @param {number} width 
     * @param {number} height 
     */
    resize(width, height) {
      canvas.width = width;
      canvas.height = height;
      pipeline.resize(width, height);
    },
    /**
     * @param {ArrayBuffer} buffer 
     */
    handle(buffer) {
      pipeline.update(buffer);
    },
    render() {
      pipeline.render();
    },
  };
};

/**
 * @param {GPUCanvasContext} context
 * @param {GPUDevice} device
 * @param {GPUTextureFormat} format
 */
const webGPUPipeline = async (context, device, format) => {
  const vertexBufferData = new Float32Array(8 * 10000); // 10,000 vertices maximum
  const vertexBuffer = device.createBuffer({
    label: "Vertex buffer",
    size: vertexBufferData.byteLength,
    usage: GPUBufferUsage.VERTEX | GPUBufferUsage.COPY_DST,
  });
  /** @type {GPUVertexBufferLayout} */
  const vertexBufferLayout = {
    arrayStride: 32,
    attributes: [
      {
        // position
        format: "float32x3",
        offset: 0,
        shaderLocation: 0,
      },
      {
        // uv
        format: "float32x2",
        offset: 12,
        shaderLocation: 1,
      },
      {
        // normal
        format: "float32x3",
        offset: 20,
        shaderLocation: 2,
      },
    ],
  };

  // Must be multiple of 256 bytes, so 64*4 = 256
  // https://webgpufundamentals.org/webgpu/lessons/resources/wgsl-offset-computer.html#x=5d00000100c900000000000000003d888b0237284d3025f2381bcb288993d4e720509e6096aeab474111d6d9f5309f81fdd02fa8d6511000171b1dea5930bb70788363fa4a4a557a6c95cff0bb4dac4a98616495d1d4ca6e458c2088162fa63656624f6c8071840d7e206f680cf11b62cd9cb0e68d84b62bb8fa051e306b87ef30053f98e3c5b69fdbb794cd72ffff92a8c000
  const modelBufferData = new Float32Array(64 * 1000); // 1,000 instances maximum (256kb)
  const modelBuffer = device.createBuffer({
    label: "Model buffer",
    size: modelBufferData.byteLength,
    usage: GPUBufferUsage.UNIFORM | GPUBufferUsage.COPY_DST,
  });

  const cameraBufferData = new Float32Array(32);
  const cameraBuffer = device.createBuffer({
    label: "Camera buffer",
    size: cameraBufferData.byteLength,
    usage: GPUBufferUsage.UNIFORM | GPUBufferUsage.COPY_DST,
  });

  // Check: https://webgpufundamentals.org/webgpu/lessons/resources/wgsl-offset-computer.html#x=5d000001001d01000000000000003d888b0237284c233fe98faab4a8c11cec1c512fb375a14e643bbd3b156d0f1edc75994fb8867f3abfbad62e753ae6594342d0b60256a1f95e6cf72aa9c42ba0eb6e2e0e9b7ba2293639419adf3480e0c57739509e6197d1440e8a336d7db12c5d5751558b1d5ccb8a7a4fda99c639b45967fee79da3c2ae12bd14d449eaf1f63b5e46cf645e8295fcb80ddf78f79da73b47c8cf3c1a97c9cad6e88628a26ccfb1a4139fffb2d83d60
  // use library or do some parsing to compute alignment
  const lightBufferData = new ArrayBuffer(2560);
  const lightBuffer = device.createBuffer({
    label: "Light buffer",
    size: lightBufferData.byteLength,
    usage: GPUBufferUsage.UNIFORM | GPUBufferUsage.COPY_DST,
  });
  const directionalLightView = {
    byteOffset: 0,
    byteLength: 128,
    viewProjMatrix: new Float32Array(lightBufferData, 0, 16),
    direction: new Float32Array(lightBufferData, 64, 3),
    ambient: new Float32Array(lightBufferData, 80, 3),
    diffuse: new Float32Array(lightBufferData, 96, 3),
    specular: new Float32Array(lightBufferData, 112, 3),
    intensity: new Float32Array(lightBufferData, 124, 1),
  };
  const pointLightByteOffset = directionalLightView.byteOffset + directionalLightView.byteLength;
  const pointLightStride = 80; // Size in bytes between each element
  const pointLightCount = 10; // Maximum number of point lights
  const pointLightsView = {
    byteOffset: pointLightByteOffset,
    byteLength: 16 + pointLightStride * pointLightCount,
    count: new Uint32Array(lightBufferData, pointLightByteOffset, 1),
    pointLights: new Array(pointLightCount).fill(null).map((_, i) => ({
      position: new Float32Array(lightBufferData, pointLightByteOffset + i * pointLightStride + 16, 3),
      ambient: new Float32Array(lightBufferData, pointLightByteOffset + i * pointLightStride + 32, 3),
      diffuse: new Float32Array(lightBufferData, pointLightByteOffset + i * pointLightStride + 48, 3),
      specular: new Float32Array(lightBufferData, pointLightByteOffset + i * pointLightStride + 64, 3),
      radius: new Float32Array(lightBufferData, pointLightByteOffset + i * pointLightStride + 76, 1),
      intensity: new Float32Array(lightBufferData, pointLightByteOffset + i * pointLightStride + 80, 1),
    })),
  };
  const spotLightByteOffset = pointLightsView.byteOffset + pointLightsView.byteLength;
  const spotLightStride = 160; // Size in bytes between each element
  const spotLightCount = 10; // Maximum number of spot lights
  const spotLightsView = {
    byteOffset: spotLightByteOffset,
    byteLength: 16 + spotLightStride * spotLightCount,
    count: new Uint32Array(lightBufferData, spotLightByteOffset, 1),
    spotLights: new Array(spotLightCount).fill(null).map((_, i) => ({
      viewProjMatrix: new Float32Array(lightBufferData, spotLightByteOffset + i * spotLightStride + 16, 16),
      position: new Float32Array(lightBufferData, spotLightByteOffset + i * spotLightStride + 80, 3),
      direction: new Float32Array(lightBufferData, spotLightByteOffset + i * spotLightStride + 96, 3),
      ambient: new Float32Array(lightBufferData, spotLightByteOffset + i * spotLightStride + 112, 3),
      diffuse: new Float32Array(lightBufferData, spotLightByteOffset + i * spotLightStride + 128, 3),
      specular: new Float32Array(lightBufferData, spotLightByteOffset + i * spotLightStride + 144, 3),
      cutOff: new Float32Array(lightBufferData, spotLightByteOffset + i * spotLightStride + 156, 1),
      outerCutOff: new Float32Array(lightBufferData, spotLightByteOffset + i * spotLightStride + 160, 1),
      range: new Float32Array(lightBufferData, spotLightByteOffset + i * spotLightStride + 164, 1),
      intensity: new Float32Array(lightBufferData, spotLightByteOffset + i * spotLightStride + 168, 1),
    })),
  };

  const lightViewMatrixBufferData = new ArrayBuffer(256 + 10 * 256); // 256 bytes per light
  const lightViewMatrixBuffer = device.createBuffer({
    label: "Light view matrix buffer",
    size: lightViewMatrixBufferData.byteLength,
    usage: GPUBufferUsage.UNIFORM | GPUBufferUsage.COPY_DST,
  });
  const directionalLightMatrixView = new Float32Array(lightViewMatrixBufferData, 0, 16);
  // const pointLightsMatrixView = new Array(spotLightCount).fill(null).map((_, i) => new Float32Array(lightViewMatrixBufferData, 256 + (i * 256), 16));
  const spotLightsMatrixView = new Array(spotLightCount).fill(null).map((_, i) => new Float32Array(lightViewMatrixBufferData, 256 + (i * 256), 16));

  const materialBufferData = new Float32Array(64 * 100); // 100 materials maximum
  const materialBuffer = device.createBuffer({
    label: "Material buffer",
    size: materialBufferData.byteLength,
    usage: GPUBufferUsage.UNIFORM | GPUBufferUsage.COPY_DST,
  });

  const textureSampler = device.createSampler({
    magFilter: "nearest",
    minFilter: "nearest",
    addressModeU: "repeat",
    addressModeV: "repeat",
    addressModeW: "repeat",
  });

  const [diffuseSource, specularSource] = await Promise.all([
    loadImageBitmap("resources/images/diffuse.png"),
    loadImageBitmap("resources/images/specular.png"),
  ]);

  const diffuseTexture = device.createTexture({
    label: "resources/images/diffuse.png",
    size: [diffuseSource.width, diffuseSource.width, diffuseSource.height / diffuseSource.width],
    format: "rgba8unorm",
    usage: GPUTextureUsage.TEXTURE_BINDING | GPUTextureUsage.RENDER_ATTACHMENT | GPUTextureUsage.COPY_DST,
  });
  for (let i = 0; i < diffuseSource.height / diffuseSource.width; i++) {
    device.queue.copyExternalImageToTexture(
      { source: diffuseSource, origin: [0, i * diffuseSource.width] },
      { texture: diffuseTexture, origin: [0, 0, i], premultipliedAlpha: true },
      [diffuseSource.width, diffuseSource.width],
    );
  }

  const specularTexture = device.createTexture({
    label: "resources/images/specular.png",
    size: [specularSource.width, specularSource.width, specularSource.height / specularSource.width],
    format: "r8unorm",
    usage: GPUTextureUsage.TEXTURE_BINDING | GPUTextureUsage.RENDER_ATTACHMENT | GPUTextureUsage.COPY_DST,
  });
  for (let i = 0; i < specularSource.height / specularSource.width; i++) {
    device.queue.copyExternalImageToTexture(
      { source: specularSource, origin: [0, i * specularSource.width] },
      { texture: specularTexture, origin: [0, 0, i] },
      [specularSource.width, specularSource.width],
    );
  }

  // Create the depth texture for rendering/sampling the shadow map.
  const shadowDepthTexture = device.createTexture({
    label: "Shadow depth texture",
    size: [1024, 1024, 21],
    usage: GPUTextureUsage.RENDER_ATTACHMENT | GPUTextureUsage.TEXTURE_BINDING,
    format: 'depth32float',
  });
  const shadowDepthTextureView = shadowDepthTexture.createView({ baseArrayLayer: 0, arrayLayerCount: 1, dimension: "2d" });
  const spotLightShadowDepthTextureView = shadowDepthTexture.createView({ baseArrayLayer: 1, arrayLayerCount: 10, dimension: "2d-array" });
  const spotLightShadowDepthTextureViews = new Array(10).fill(null).map((_, i) => shadowDepthTexture.createView({ baseArrayLayer: i + 1, arrayLayerCount: 1, dimension: "2d" }));

  // Create global bind groups
  const perFrameBindGroupLayout = device.createBindGroupLayout({
    entries: [
      {
        binding: 0,
        visibility: GPUShaderStage.VERTEX | GPUShaderStage.FRAGMENT,
        buffer: { type: "uniform" },
      },
      {
        binding: 1,
        visibility: GPUShaderStage.VERTEX | GPUShaderStage.FRAGMENT,
        buffer: { type: "uniform" },
      },
      {
        binding: 2,
        visibility: GPUShaderStage.FRAGMENT,
        sampler: {},
      },
      {
        binding: 3,
        visibility: GPUShaderStage.FRAGMENT,
        sampler: { type: "comparison" },
      },
      {
        binding: 4,
        visibility: GPUShaderStage.FRAGMENT,
        texture: { sampleType: "depth" },
      },
      {
        binding: 5,
        visibility: GPUShaderStage.FRAGMENT,
        texture: { sampleType: "depth", viewDimension: "2d-array" },
      },
    ]
  });

  const shadowDepthBindGroupLayout = device.createBindGroupLayout({
    entries: [
      {
        binding: 0,
        visibility: GPUShaderStage.VERTEX | GPUShaderStage.FRAGMENT,
        buffer: { type: "uniform" },
      },
    ]
  });

  const perFrameBindGroup = device.createBindGroup({
    label: "Per frame bind group",
    layout: perFrameBindGroupLayout,
    entries: [
      {
        binding: 0,
        resource: { buffer: cameraBuffer },
      },
      {
        binding: 1,
        resource: { buffer: lightBuffer },
      },
      {
        binding: 2,
        resource: textureSampler,
      },
      {
        binding: 3,
        resource: device.createSampler({ compare: "less" })
      },
      {
        binding: 4,
        resource: shadowDepthTextureView,
      },
      {
        binding: 5,
        resource: spotLightShadowDepthTextureView,
      },
    ],
  });

  const directionalLightBindGroup = device.createBindGroup({
    label: "Directional light shadow map",
    layout: shadowDepthBindGroupLayout,
    entries: [
      {
        binding: 0,
        resource: { buffer: lightViewMatrixBuffer, offset: directionalLightMatrixView.byteOffset, size: directionalLightMatrixView.byteLength },
      },
    ],
  });

  const spotLightShadowMapBindGroups = spotLightsMatrixView.map(spotLight => device.createBindGroup({
    label: "Spot light shadow map",
    layout: shadowDepthBindGroupLayout,
    entries: [
      {
        binding: 0,
        resource: { buffer: lightViewMatrixBuffer, offset: spotLight.byteOffset, size: spotLight.byteLength },
      },
    ],
  }));

  const objectBindGroupLayout = device.createBindGroupLayout({
    entries: [
      {
        binding: 0,
        visibility: GPUShaderStage.FRAGMENT,
        texture: {},
      },
      {
        binding: 1,
        visibility: GPUShaderStage.FRAGMENT,
        texture: {},
      },
      {
        binding: 2,
        visibility: GPUShaderStage.FRAGMENT,
        buffer: { type: "uniform" },
      },
    ]
  });

  const nodeBindGroupLayout = device.createBindGroupLayout({
    entries: [
      // Model and normal matrices
      {
        binding: 0,
        visibility: GPUShaderStage.VERTEX,
        buffer: { type: "uniform" },
      },
    ],
  });

  const pipeline = createWebGPUPipeline(context, device, format, {
    update(scene) {
      const camera = scene.getCamera();
      if (camera) {
        cameraBufferData.set(camera.viewMatrix, 0);
        cameraBufferData.set(camera.projectionMatrix, 16);
        device.queue.writeBuffer(cameraBuffer, 0, cameraBufferData);
      }

      const directionalLight = scene.getDirectionalLight();
      if (directionalLight) {
        directionalLightMatrixView.set(directionalLight.viewProjectionMatrix);
        directionalLightView.viewProjMatrix.set(directionalLight.viewProjectionMatrix);
        directionalLightView.direction.set([directionalLight.direction.x, directionalLight.direction.y, directionalLight.direction.z]);
        directionalLightView.ambient.set([directionalLight.ambient.x, directionalLight.ambient.y, directionalLight.ambient.z]);
        directionalLightView.diffuse.set([directionalLight.diffuse.x, directionalLight.diffuse.y, directionalLight.diffuse.z]);
        directionalLightView.specular.set([directionalLight.specular.x, directionalLight.specular.y, directionalLight.specular.z]);
        directionalLightView.intensity.set([2]);
      }

      const pointLights = scene.listPointLights();
      pointLightsView.count.set([pointLights.length]);
      for (const [i, pointLight] of pointLights.entries()) {
        if (pointLightsView.pointLights[i]) {
          pointLightsView.pointLights[i].position.set([pointLight.position.x, pointLight.position.y, pointLight.position.z]);
          pointLightsView.pointLights[i].ambient.set([pointLight.ambient.x, pointLight.ambient.y, pointLight.ambient.z]);
          pointLightsView.pointLights[i].diffuse.set([pointLight.diffuse.x, pointLight.diffuse.y, pointLight.diffuse.z]);
          pointLightsView.pointLights[i].specular.set([pointLight.specular.x, pointLight.specular.y, pointLight.specular.z]);
          pointLightsView.pointLights[i].radius.set([pointLight.radius]);
          pointLightsView.pointLights[i].intensity.set([3]);
        }
      }

      const spotLights = scene.listSpotLights();
      spotLightsView.count.set([spotLights.length]);
      for (const [i, spotLight] of spotLights.entries()) {
        if (spotLightsView.spotLights[i]) {
          spotLightsMatrixView[i].set(spotLight.viewProjectionMatrix);
          spotLightsView.spotLights[i].viewProjMatrix.set(spotLight.viewProjectionMatrix);
          spotLightsView.spotLights[i].position.set([spotLight.position.x, spotLight.position.y, spotLight.position.z]);
          spotLightsView.spotLights[i].direction.set([spotLight.direction.x, spotLight.direction.y, spotLight.direction.z]);
          spotLightsView.spotLights[i].ambient.set([spotLight.ambient.x, spotLight.ambient.y, spotLight.ambient.z]);
          spotLightsView.spotLights[i].diffuse.set([spotLight.diffuse.x, spotLight.diffuse.y, spotLight.diffuse.z]);
          spotLightsView.spotLights[i].specular.set([spotLight.specular.x, spotLight.specular.y, spotLight.specular.z]);
          spotLightsView.spotLights[i].cutOff.set([spotLight.radius]);
          spotLightsView.spotLights[i].outerCutOff.set([spotLight.outerCutOff]);
          spotLightsView.spotLights[i].range.set([150]);
          spotLightsView.spotLights[i].intensity.set([2]);
        }
      }

      device.queue.writeBuffer(lightBuffer, 0, lightBufferData);
      device.queue.writeBuffer(lightViewMatrixBuffer, 0, lightViewMatrixBufferData);

      if (scene.didObjectChange()) {
        for (const object of scene.listNewObjects()) {
          if (object.vertexCount === 0) {
            continue;
          }
          // Update vertex buffer
          for (let i = 0; i < object.vertexCount; ++i) {
            vertexBufferData.set(object.vertices.slice(i * 3, i * 3 + 3), object.vertexOffset * 8 + i * 8);
            vertexBufferData.set(object.texCoords.slice(i * 2, i * 2 + 2), object.vertexOffset * 8 + i * 8 + 3);
            vertexBufferData.set(object.normals.slice(i * 3, i * 3 + 3), object.vertexOffset * 8 + i * 8 + 5);
          }

          // Update material buffer
          const materialBufferOffset = object.offset * 64;
          materialBufferData.set([object.material.shininess], materialBufferOffset);

          // Create a bind group for the object's material
          object.bindGroup = device.createBindGroup({
            label: `Object ${object.id}`,
            layout: objectBindGroupLayout,
            entries: [
              {
                binding: 0,
                resource: diffuseTexture.createView({ dimension: "2d", baseArrayLayer: object.material.diffuse, arrayLayerCount: 1 }),
              },
              {
                binding: 1,
                resource: specularTexture.createView({ dimension: "2d", baseArrayLayer: object.material.specular, arrayLayerCount: 1 }),
              },
              {
                binding: 2,
                resource: { buffer: materialBuffer, offset: materialBufferOffset * materialBufferData.BYTES_PER_ELEMENT, size: 256 },
              },
            ],
          });
        }

        device.queue.writeBuffer(vertexBuffer, 0, vertexBufferData);
        device.queue.writeBuffer(materialBuffer, 0, materialBufferData);
      }

      // Update model matrixes
      for (const object of scene.listObjects()) {
        // Skip non-world objects
        if (object.space !== 0 || object.vertexCount === 0) {
          continue;
        }
        const nodes = scene.listNodes(object);
        if (nodes.length === 0) {
          continue;
        }
        for (const node of nodes) {
          // Create inversed model matrix for normals
          const normalMatrix = mat4.identity();
          mat4.inverse(node.model, normalMatrix);
          mat4.transpose(normalMatrix, normalMatrix);
          modelBufferData.set(node.model, node.offset * 64);
          modelBufferData.set(normalMatrix, node.offset * 64 + 16);
          modelBufferData.set(new Float32Array([node.tint.r, node.tint.g, node.tint.b, node.tint.a]), node.offset * 64 + 32);

          if (!node.bindGroup) {
            node.bindGroup = device.createBindGroup({
              label: `Node ${node.id}`,
              layout: nodeBindGroupLayout,
              entries: [
                {
                  binding: 0,
                  resource: { buffer: modelBuffer, offset: modelBufferData.BYTES_PER_ELEMENT * node.offset * 64, size: 64 + 64 + 16 },
                },
                // {
                //   binding: 1,
                //   resource: { buffer: modelBuffer, offset: modelBufferData.BYTES_PER_ELEMENT * node.offset * 64, size: 128 },
                // },
              ],
            });
          }
        }
        device.queue.writeBuffer(modelBuffer, 0, modelBufferData);
      }
    },
  });

  pipeline.addShader(createShadowMappingShader(), {
    setup(module) {
      const renderPipeline = device.createRenderPipeline({
        label: "Shadow mapping shader",
        layout: device.createPipelineLayout({
          bindGroupLayouts: [
            shadowDepthBindGroupLayout,
            nodeBindGroupLayout,
          ],
        }),
        primitive: {
          topology: "triangle-list",
          cullMode: "front",
        },
        vertex: {
          module,
          entryPoint: "vertexMain",
          buffers: [vertexBufferLayout],
        },
        depthStencil: {
          depthWriteEnabled: true,
          depthCompare: 'less',
          format: 'depth32float',
        },
      });
      return renderPipeline;
    },
    render(renderPipeline, encoder, scene) {
      const shadowDepthRenderPasses = [
        {
          bindGroup: directionalLightBindGroup,
          view: shadowDepthTextureView
        },
      ];
      for (let i = 0; i < scene.listSpotLights().length; i++) {
        shadowDepthRenderPasses.push({
          bindGroup: spotLightShadowMapBindGroups[i],
          view: spotLightShadowDepthTextureViews[i]
        });
      }

      for (const { bindGroup, view } of shadowDepthRenderPasses) {
        const pass = encoder.beginRenderPass({
          colorAttachments: [],
          depthStencilAttachment: {
            view,
            depthClearValue: 1.0,
            depthLoadOp: 'clear',
            depthStoreOp: 'store',
          },
        });
        pass.setPipeline(renderPipeline);
        pass.setVertexBuffer(0, vertexBuffer);
        pass.setBindGroup(0, bindGroup);
        let nodeCount = 0;
        for (const object of scene.listObjects()) {
          if (object.space !== 0 || object.vertexCount === 0 || object.material.opaque === false) {
            continue;
          }
          const nodes = scene.listNodes(object);
          if (nodes.length === 0) {
            continue;
          }
          for (const node of nodes) {
            pass.setBindGroup(1, node.bindGroup);
            pass.draw(object.vertexCount, 1, object.vertexOffset);
            nodeCount++;
          }
        }
        pass.end();
      }
    },
  });

  pipeline.addShader(createScreenTextureShader(), {
    enabled: DEBUG === true,
    setup(module) {
      return device.createRenderPipeline({
        label: "Texture shader",
        layout: "auto",
        primitive: {
          topology: "triangle-list",
          cullMode: "none",
        },
        vertex: {
          module,
          entryPoint: "vertexMain",
        },
        multisample: {
          count: 4,
        },
        fragment: {
          module,
          entryPoint: "fragmentMain",
          targets: [{ format }],
          constants: {
            textureWidth: 1024,
            textureHeight: 1024,
          }
        },
      });
    },
    render(renderPipeline, encoder, scene, renderContext) {
      const debugTexture = shadowDepthTextureView;
      // const debugTexture = spotLightShadowDepthTextureViews[0];
      const textureBindGroup = device.createBindGroup({
        label: "Texture bind group",
        layout: renderPipeline.getBindGroupLayout(0),
        entries: [
          {
            binding: 0,
            resource: debugTexture,
          },
        ],
      });
      const pass = encoder.beginRenderPass({
        colorAttachments: [
          {
            view: renderContext.view,
            resolveTarget: context.getCurrentTexture().createView(),
            loadOp: "clear",
            clearValue: [0, 0, 0, 1],
            storeOp: "store",
          }
        ],
      });
      pass.setPipeline(renderPipeline);
      pass.setBindGroup(0, textureBindGroup);
      pass.draw(6);
      pass.end();
    }
  });

  pipeline.addShader(createStandardShader(), {
    enabled: DEBUG === false,
    setup(module) {
      const renderPipeline = device.createRenderPipeline({
        label: "Standard shader",
        layout: device.createPipelineLayout({
          bindGroupLayouts: [
            perFrameBindGroupLayout,
            objectBindGroupLayout,
            nodeBindGroupLayout,
          ],
        }),
        primitive: {
          topology: "triangle-list",
          cullMode: "back",
        },
        vertex: {
          module,
          entryPoint: "vertexMain",
          buffers: [vertexBufferLayout],
        },
        fragment: {
          module,
          entryPoint: "fragmentMain",
          targets: [{
            format,
            blend: {
              color: { operation: "add", srcFactor: 'one', dstFactor: 'one-minus-src-alpha' },
              alpha: { operation: "add", srcFactor: 'one', dstFactor: 'one-minus-src-alpha' },
            }
          }],
          constants: {
            shadowDepthTextureSize: 1024,
            debug: true,
          }
        },
        multisample: {
          count: 4,
        },
        depthStencil: {
          depthWriteEnabled: true,
          depthCompare: 'less',
          format: 'depth24plus',
        },
      });
      return renderPipeline;
    },
    render(renderPipeline, encoder, scene, renderContext) {
      const pass = encoder.beginRenderPass({
        colorAttachments: [
          {
            view: renderContext.view,
            resolveTarget: context.getCurrentTexture().createView(),
            loadOp: "clear",
            clearValue: [0.5, 0.5, 0.8, 1],
            storeOp: "store",
          }
        ],
        depthStencilAttachment: {
          view: renderContext.depthView,
          depthClearValue: 1.0,
          depthLoadOp: 'clear',
          depthStoreOp: 'store',
        }
      });

      pass.setPipeline(renderPipeline);
      pass.setVertexBuffer(0, vertexBuffer);
      pass.setBindGroup(0, perFrameBindGroup);

      let nodeCount = 0;
      for (const object of scene.listObjects()) {
        // Skip non-world objects and transparent objects
        if (object.space !== 0 || object.vertexCount === 0 || object.material.opaque === false) {
          continue;
        }
        const nodes = scene.listNodes(object);
        if (nodes.length === 0) {
          continue;
        }
        pass.setBindGroup(1, object.bindGroup);
        for (const node of nodes) {
          pass.setBindGroup(2, node.bindGroup);
          pass.draw(object.vertexCount, 1, object.vertexOffset);
          nodeCount++;
        }
      }
      pass.end();
    },
  });

  pipeline.addShader({ name: "Standard transparent", source: createStandardShader().source }, {
    enabled: DEBUG === false,
    setup(module) {
      const renderPipeline = device.createRenderPipeline({
        label: "Standard transparent shader",
        layout: device.createPipelineLayout({
          bindGroupLayouts: [
            perFrameBindGroupLayout,
            objectBindGroupLayout,
            nodeBindGroupLayout,
          ],
        }),
        primitive: {
          topology: "triangle-list",
          cullMode: "none",
        },
        vertex: {
          module,
          entryPoint: "vertexMain",
          buffers: [vertexBufferLayout],
        },
        fragment: {
          module,
          entryPoint: "fragmentMain",
          targets: [{
            format,
            blend: {
              color: { operation: "add", srcFactor: 'one', dstFactor: 'one-minus-src-alpha' },
              alpha: { operation: "add", srcFactor: 'one', dstFactor: 'one-minus-src-alpha' },
            }
          }],
          constants: {
            shadowDepthTextureSize: 1024,
            debug: true,
          }
        },
        multisample: {
          count: 4,
        },
        depthStencil: {
          depthWriteEnabled: false,
          depthCompare: 'less',
          format: 'depth24plus',
        },
      });
      return renderPipeline;
    },
    render(renderPipeline, encoder, scene, renderContext) {
      const pass = encoder.beginRenderPass({
        colorAttachments: [
          {
            view: renderContext.view,
            resolveTarget: context.getCurrentTexture().createView(),
            loadOp: "load",
            storeOp: "store",
          }
        ],
        depthStencilAttachment: {
          view: renderContext.depthView,
          depthReadOnly: true
        }
      });

      pass.setPipeline(renderPipeline);
      pass.setVertexBuffer(0, vertexBuffer);
      pass.setBindGroup(0, perFrameBindGroup);

      let nodeCount = 0;
      for (const object of scene.listObjects()) {
        // Skip non-world objects
        if (object.space !== 0 || object.vertexCount === 0 || object.material.opaque === true) {
          continue;
        }
        const nodes = scene.listNodes(object);
        if (nodes.length === 0) {
          continue;
        }
        pass.setBindGroup(1, object.bindGroup);
        for (const node of nodes) {
          pass.setBindGroup(2, node.bindGroup);
          pass.draw(object.vertexCount, 1, object.vertexOffset);
          nodeCount++;
        }
      }
      pass.end();
    },
  });

  return pipeline.end();
};
