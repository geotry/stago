const { createContext } = require("./context.js");
const { createScene } = require("./scene.js");
const { decodeBuffer, assertSceneLight, assertTexture, assertSceneLightDeleted, assertCamera, assertSceneNodeDeleted, assertSceneObject, assertSceneNode, TextureBuffer } = require("./decoder.js");

/**
 * @typedef {ReturnType<typeof createScene>} Scene
 * @typedef {ReturnType<typeof createContext>} Context
 * @typedef {{
 *  setup: (gl: WebGL2RenderingContext, context: Context) => void,
 *  update: (gl: WebGL2RenderingContext, scene: Scene, context: Context) => void,
 *  updateTexture: (gl: WebGL2RenderingContext, texture: TextureBuffer, context: Context) => void,
 * }} PipelineConfig
 */

/**
 * @template T 
 * @typedef {{
 *  createProgram: (gl: WebGL2RenderingContext) => WebGLProgram,
 *  createUniforms: (gl: WebGL2RenderingContext, program: WebGLProgram) => T
 * }} Shader<T>
 */

/**
 * @template T 
 * @typedef {{
 *  enabled?: boolean,
 *  setup: (gl: WebGL2RenderingContext, program: WebGLProgram, context: Context) => WebGLVertexArrayObject | undefined,
 *  render: (gl: WebGL2RenderingContext, program: WebGLProgram, uniforms: T, scene: Scene, context: Context) => void,
 * }} ShaderConfig<T>
 */

/**
 * Create a new general render pipeline.
 * 
 * @template T
 *
 * @param {WebGL2RenderingContext} gl
 * @param {PipelineConfig} config
 *
 * @returns 
 */
export const createPipeline = (gl, config) => {
  /**
   * Create a scene local to the pipeline to represent 3D objects, camera, and do interpolations.
   */
  const scene = createScene();

  const context = createContext(gl);

  /**
   * @template T
   * @type {{
   *  name: string,
   *  enabled: boolean,
   *  vao?: WebGLVertexArrayObject,
   *  render: ShaderConfig<T>["render"],
   *  setup: ShaderConfig<T>["setup"],
   *  uniforms: T,
   *  program: WebGLProgram
   * }[]}
   */
  const shaders = [];

  return {
    /**
     * Register a new shader to the pipeline.
     *
     * @template T
     * @param {{name: string, vertex: string, fragment: string, createUniforms: (gl: WebGL2RenderingContext, program: WebGLProgram) => T}} source 
     * @param {ShaderConfig<T>} shaderConfig 
     */
    addShader(source, shaderConfig) {
      if (shaders.findIndex(s => s.name === source.name) !== -1) {
        throw new Error(`Shader ${source.name} already exist`);
      }

      const program = gl.createProgram();

      const vertexShader = gl.createShader(gl.VERTEX_SHADER);
      gl.shaderSource(vertexShader, source.vertex);
      gl.compileShader(vertexShader);
      if (!gl.getShaderParameter(vertexShader, gl.COMPILE_STATUS)) {
        const info = gl.getShaderInfoLog(vertexShader);
        throw new Error(`Could not compile WebGL program. \n\n${info}`);
      }
      gl.attachShader(program, vertexShader);

      const fragmentShader = gl.createShader(gl.FRAGMENT_SHADER);
      gl.shaderSource(fragmentShader, source.fragment);
      gl.compileShader(fragmentShader);
      if (!gl.getShaderParameter(fragmentShader, gl.COMPILE_STATUS)) {
        const info = gl.getShaderInfoLog(fragmentShader);
        throw new Error(`Could not compile WebGL program. \n\n${info}`);
      }
      gl.attachShader(program, fragmentShader);

      gl.linkProgram(program);
      gl.useProgram(program);

      const uniforms = source.createUniforms(gl, program);

      shaders.push({
        name: source.name,
        enabled: shaderConfig.enabled ?? true,
        program,
        uniforms,
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
      if (config.setup) {
        config.setup(gl, context);
      }

      for (const shader of shaders) {
        if (shader.setup) {
          const vao = shader.setup(gl, shader.program, context);
          if (vao !== null) {
            shader.vao = vao;
          }
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
          scene.clear();
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
                  config.updateTexture(gl, block, context);
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
          if (context.frame === 0) {
            console.log("[pipeline] start render");
          }

          context.frame++;
          if (context.renderTime === 0) {
            context.renderTime = new Date().getTime();
          }
          context.deltaTime = new Date().getTime() - context.renderTime;
          context.renderTime = new Date().getTime();

          config.update(gl, scene, context);

          for (const shader of shaders) {
            if (!shader.enabled) {
              continue;
            }

            gl.useProgram(shader.program);
            if (shader.vao) {
              gl.bindVertexArray(shader.vao);
            }

            shader.render(gl, shader.program, shader.uniforms, scene, context);

            if (shader.vao) {
              gl.bindVertexArray(null);
            }
            gl.useProgram(null);
          }

          gl.flush();
        },
      };
    },
  };
};
