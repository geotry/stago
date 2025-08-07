const { createScene } = require("./scene.js");
const { decodeBuffer, assertSceneLight, assertTexture, assertSceneLightDeleted, assertCamera, assertSceneNodeDeleted, assertSceneObject, assertSceneNode, TextureBuffer } = require("./decoder.js");

/**
 * @typedef {ReturnType<typeof createScene>} Scene
 * @typedef {{
 *  update: (scene: Scene) => void,
 *  updateTexture: (texture: TextureBuffer) => void,
 * }} WebGPUPipelineConfig
 */

/**
 * @typedef {{
 *  frame: number,
 *  view: GPUTextureView,
 *  depthView: GPUTextureView, 
 * }} WebGPURenderContext
 */

/**
 * @typedef {{
 *  enabled?: boolean,
 *  setup: (shader: GPUShaderModule) => GPURenderPipeline,
 *  render: (renderPipeline: GPURenderPipeline, encoder: GPUCommandEncoder, scene: Scene, renderContext: WebGPURenderContext) => void,
 * }} WebGPUShaderConfig
 */

/**
 * Create a new general render pipeline for WebGPU.
 * 
 * @param {GPUCanvasContext} context
 * @param {GPUDevice} device
 * @param {GPUTextureFormat} format
 * @param {WebGPUPipelineConfig} config
 *
 * @returns 
 */
export const createWebGPUPipeline = (context, device, format, config) => {
  /**
   * Create a scene local to the pipeline to represent 3D objects, camera, and do interpolations.
   */
  let scene = createScene();

  let frame = 0;
  let renderTime = 0;
  let deltaTime = 0;

  /**
   * @type {{
   *  name: string,
   *  enabled: boolean,
   *  render: WebGPUShaderConfig["render"],
   *  setup: WebGPUShaderConfig["setup"],
   *  shader: GPUShaderModule,
   *  renderPipeline: GPURenderPipeline,
   * }[]}
   */
  const shaders = [];

  return {
    /**
     * Register a new shader to the pipeline.
     *
     * @param {{name: string, source: string}} source 
     * @param {WebGPUShaderConfig} shaderConfig 
     */
    addShader(source, shaderConfig) {
      if (shaders.findIndex(s => s.name === source.name) !== -1) {
        throw new Error(`Shader ${source.name} already exist`);
      }

      const shader = device.createShaderModule({
        label: `${source.name} shader`,
        code: source.source,
      });

      shaders.push({
        name: source.name,
        enabled: shaderConfig.enabled ?? true,
        shader,
        setup: shaderConfig.setup,
        render: shaderConfig.render,
      });

      console.log(`[pipeline] added shader ${source.name}`);
    },

    /**
     * Enable a shader in the pipeline. The shader will be used on next render calls.
     *
     * @param {string} name 
     */
    enableShader(name) {
      const index = shaders.findIndex((s) => s.shader.name === name);
      if (index !== -1) {
        shaders[index].enabled = true;
        console.log(`[pipeline] enabled shader ${name}`);
      }
    },

    /**
     * Disable a shader in the pipeline. The shader will be skipped on next render calls.
     *
     * @param {string} name 
     */
    disableShader(name) {
      const index = shaders.findIndex((s) => s.shader.name === shader);
      if (index !== -1) {
        shaders[index].enabled = false;
        console.log(`[pipeline] disabled shader ${name}`);
      }
    },

    /**
     * Finalize the pipeline and returns the render function.
     *
     * @returns
     */
    end() {
      let depthTexture = device.createTexture({
        size: [context.canvas.width, context.canvas.height],
        format: 'depth24plus',
        usage: GPUTextureUsage.RENDER_ATTACHMENT,
        sampleCount: 4,
      });
      let depthTargetView = depthTexture.createView();

      // Create render target with multisampling
      // todo: move this in setup?
      let renderTarget = device.createTexture({
        size: [context.canvas.width, context.canvas.height],
        sampleCount: 4,
        format,
        usage: GPUTextureUsage.RENDER_ATTACHMENT,
      });
      let renderTargetView = renderTarget.createView();

      for (const shader of shaders) {
        if (shader.setup) {
          const renderPipeline = shader.setup(shader.shader);
          shader.renderPipeline = renderPipeline;
        }
        console.log(`[pipeline] shader ${shader.name} ready`);
      }

      console.log(`[pipeline] ready 🚀`);

      return {
        /**
         * Reset the pipeline, clear objects.
         */
        reset() {
          console.log("[pipeline] reset");
          scene = createScene();
          frame = 0;
          renderTime = 0;
        },
        /**
         * Resize the canvas viewport.
         *
         * @param {number} width 
         * @param {number} height 
         */
        resize(width, height) {
          if (depthTexture) {
            depthTexture.destroy();
          }
          depthTexture = device.createTexture({
            size: [width, height],
            format: 'depth24plus',
            usage: GPUTextureUsage.RENDER_ATTACHMENT,
            sampleCount: 4,
          });
          depthTargetView = depthTexture.createView();

          if (renderTarget) {
            renderTarget.destroy();
          }
          // todo: move in pipeline.createRenderTarget(width, height);
          renderTarget = device.createTexture({
            size: [width, height],
            sampleCount: 4,
            format,
            usage: GPUTextureUsage.RENDER_ATTACHMENT,
          });
          renderTargetView = renderTarget.createView();
        },
        /**
         * Handle a new message from the server.
         *
         * @param {ArrayBuffer} buffer 
         */
        update(buffer) {
          scene.update();

          for (const block of decodeBuffer(buffer)) {
            switch (true) {
              case assertTexture(block): {
                if (config.updateTexture) {
                  config.updateTexture(block);
                }
                break;
              }
              case assertCamera(block): {
                scene.updateCamera(block.id, block);
                break;
              }
              case assertSceneObject(block): {
                scene.updateObject(block.id, {
                  space: block.space,
                  texCoords: block.uv,
                  normals: block.normals,
                  vertices: block.vertices,
                  material: {
                    diffuse: block.textureIndex,
                    specular: block.textureIndex,
                    shininess: block.shininess,
                  },
                });
                break;
              }
              case assertSceneNode(block): {
                scene.updateNode(block.id, {
                  objectId: block.objectId,
                  model: block.model,
                });
                break;
              }
              case assertSceneNodeDeleted(block): {
                scene.deleteNode(block.id);
                break;
              }
              case assertSceneLight(block): {
                scene.updateLight(block.id, {
                  type: block.type,
                  lightSpace: block.viewSpace,
                  ambient: { x: block.ambientR, y: block.ambientG, z: block.ambientB },
                  diffuse: { x: block.diffuseR, y: block.diffuseG, z: block.diffuseB },
                  specular: { x: block.specularR, y: block.specularG, z: block.specularB },
                  direction: { x: block.directionX, y: block.directionY, z: block.directionZ },
                  position: { x: block.posX, y: block.posY, z: block.posZ },
                  outerCutOff: block.outerCutOff,
                  radius: block.radius,
                });
                break;
              }
              case assertSceneLightDeleted(block): {
                scene.deleteLight(block.id);
                break;
              }
            }
          }
        },

        /**
         * Render a new frame.
         */
        render() {
          if (frame === 0) {
            console.log("[pipeline] start render");
          }

          frame++;

          if (renderTime === 0) {
            renderTime = new Date().getTime();
          }
          deltaTime = new Date().getTime() - renderTime;
          renderTime = new Date().getTime();

          config.update(scene);

          const encoder = device.createCommandEncoder();

          for (const shader of shaders) {
            if (!shader.enabled) {
              continue;
            }
            shader.render(shader.renderPipeline, encoder, scene, {
              frame,
              view: renderTargetView,
              depthView: depthTargetView,
            });
          }

          device.queue.submit([encoder.finish()]);

          if (frame === 1) {
            console.log("frame: ", frame);
            console.log("render time:", new Date().getTime() - renderTime);
            console.log("shaders: ", shaders.length);
            console.log("camera: ", scene.getCamera());
            console.log("light: ", scene.getDirectionalLight());
            console.log("objects: ", scene.listObjects());
            console.log("nodes: ", Array.from(scene.listObjects()).map(o => scene.listNodes(o)).flat());
          }
        },
      };
    },
  };
};
