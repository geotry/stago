const { m4 } = require("./matrix.js");
const { createPipeline } = require("./pipeline.js");
const { StandardShader } = require("./shaders/standard.js");
const { SimpledepthShader } = require("./shaders/simpleDepth.js");

/**
 * 
 * @param {OffscreenCanvas} canvas 
 */
export const createContext = (canvas) => {
  const gl = canvas.getContext("webgl2", { antialias: false });
  const pipeline = defaultPipeline(gl);

  return {
    reset() {
      // for (const shader of Object.values(shaders)) {
      //   shader.clear(true);
      // }
      // scene.clear();
    },
    resize(width, height) {
      canvas.width = width;
      canvas.height = height;
      gl.viewport(0, 0, gl.canvas.width, gl.canvas.height);
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
 * 
 * @param {WebGL2RenderingContext} gl 
 * @returns 
 */
const defaultPipeline = (gl) => {
  const pipeline = createPipeline(gl, {
    setup(gl, context) {
      gl.getExtension("GL_OES_standard_derivatives");

      gl.enable(gl.DEPTH_TEST);
      gl.depthFunc(gl.LESS);

      gl.enable(gl.BLEND)
      gl.blendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);

      gl.clearColor(.0, .0, .0, .0);
      gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);

      // Create textures
      // Todo: move this in update()
      // context.createTexture("diffuse", 0, gl.ALPHA, gl.ALPHA, 256, 256, 8);
      // context.createTexture("palette", 1, gl.RGBA, gl.SRGB8_ALPHA8, 256, 1, 1);
      // context.createTexture("specular", 2, gl.ALPHA, gl.ALPHA, 256, 256, 8);
      context.createDepthTexture("directional_light_sm", 1024, 1024, 1);
      context.createDepthTexture("spot_light_sm", 256, 256, 1);
      context.createFrameBuffer("DirectionalLightSM");
      context.createFrameBuffer("SpotLightSM");

      // Create buffers
      context.createBuffer("Vertex", gl.ARRAY_BUFFER, gl.STATIC_DRAW, [
        {
          name: "vertices",
          size: 1000,
          attributes: {
            position: { type: gl.FLOAT, size: 3 },
            texCoords: { type: gl.FLOAT, size: 2 },
            normals: { type: gl.FLOAT, size: 3 },
          }
        }
      ]);

      // Note: with UBO we can have a single buffer for all shaders, and remove this one
      context.createBuffer("VertexDepth", gl.ARRAY_BUFFER, gl.STATIC_DRAW, [
        {
          name: "vertices",
          size: 1000,
          attributes: {
            position: { type: gl.FLOAT, size: 3 },
          }
        }
      ]);

      context.createBuffer("Model", gl.ARRAY_BUFFER, gl.DYNAMIC_DRAW, [
        {
          name: "models",
          size: 1000,
          attributes: {
            model: { type: gl.FLOAT, size: 16 },
          }
        }
      ]);

      context.createBuffer("ModelDepth", gl.ARRAY_BUFFER, gl.DYNAMIC_DRAW, [
        {
          name: "models",
          size: 1000,
          attributes: {
            model: { type: gl.FLOAT, size: 16 },
          }
        }
      ]);
    },
    updateTexture(gl, texture, context) {
      switch (texture.role) {
        case 0:
          context.createTexture("diffuse", gl.ALPHA, gl.ALPHA, texture.width, texture.height, texture.depth, texture.pixels);
          break;
        case 1:
          context.createTexture("palette", gl.RGBA, gl.SRGB8_ALPHA8, texture.width, texture.height, texture.depth, texture.pixels);
          break;
        case 2:
          context.createTexture("specular", gl.ALPHA, gl.ALPHA, texture.width, texture.height, texture.depth, texture.pixels);
          break;
      }
    },
    update(gl, scene, context) {
      if (scene.didObjectChange()) {
        const vertices = [];
        // Data need to be arranged per vertex
        for (const object of scene.listObjects()) {
          if (object.vertexCount > 0) {
            for (let i = 0; i < object.vertices.length / 3; ++i) {
              vertices.push({
                position: object.vertices.slice(i * 3, i * 3 + 3),
                texCoords: object.texCoords.slice(i * 2, i * 2 + 2),
                normals: object.normals.slice(i * 3, i * 3 + 3),
              });
            }
          }
        }
        context.updateBuffer("Vertex", "vertices", 0, vertices);
        context.updateBuffer("VertexDepth", "vertices", 0, vertices.map(vt => ({ position: vt.position })));
      }

      // if (scene.didLightChange()) {

      // }

      // todo
      // When parsing shader files, find uniform blocks and generate a single
      // file to get/update it and link it to the context object.

      // context.uniforms.Camera.set({
      //   view: context.scene.camera.view,
      //   projection: context.scene.camera.projection,
      // });

      // Update the whole buffer
      // context.uniforms.DirectionalLight.set({ direction: [0, 0, 0], ambient: [0, 0, 0] });
      // context.uniforms.PointLight.set({ lights: [{ direction: [0, 0, 0], ambient: [0, 0, 0] }], light_count: 1 });

      // For shadows, the shader will override "Camera" block before each draw call with the Light's matrix
      // No: better use a specific UBO "Light" for this, otherwise the Camera UBO will be wrong for next shaders

      // Or create a single UBO for all lights? Heavier to update
      // context.uniforms.Light.set({
      //   directional_light: { direction: [1, 1, 1] },
      //   point_light_count: 1,
      //   point_lights: [{ direction: [1, 1, 1]}],
      // });

      // Update part of the UBO? probably more complicated, less performant and not that interesting
      // context.uniforms.Light.point_light_count.set(0);
      // context.uniforms.Light.point_lights[1].direction.set(0, 0, 0);
    },
  });

  pipeline.addShader(SimpledepthShader, {
    enabled: true,
    setup(gl, program, context) {
      const vao = gl.createVertexArray();
      gl.bindVertexArray(vao);

      // Bind vertex buffer objects
      const position = gl.getAttribLocation(program, "a_position");
      const model = gl.getAttribLocation(program, "a_model");

      gl.bindBuffer(gl.ARRAY_BUFFER, context.getBuffer("VertexDepth"));
      gl.enableVertexAttribArray(position);
      gl.vertexAttribPointer(position, 3, gl.FLOAT, false, 12, 0);

      gl.bindBuffer(gl.ARRAY_BUFFER, context.getBuffer("ModelDepth"));
      gl.enableVertexAttribArray(model);
      gl.vertexAttribPointer(model, 4, gl.FLOAT, false, 64, 0);
      gl.vertexAttribDivisor(model, 1);
      gl.enableVertexAttribArray(model + 1);
      gl.vertexAttribPointer(model + 1, 4, gl.FLOAT, false, 64, 16);
      gl.vertexAttribDivisor(model + 1, 1);
      gl.enableVertexAttribArray(model + 2);
      gl.vertexAttribPointer(model + 2, 4, gl.FLOAT, false, 64, 32);
      gl.vertexAttribDivisor(model + 2, 1);
      gl.enableVertexAttribArray(model + 3);
      gl.vertexAttribPointer(model + 3, 4, gl.FLOAT, false, 64, 48);
      gl.vertexAttribDivisor(model + 3, 1);

      return vao;
    },
    render(gl, program, uniforms, scene, context) {
      gl.cullFace(gl.FRONT);
      gl.enable(gl.DEPTH_TEST);

      const directionalLight = scene.getDirectionalLight();
      if (directionalLight) {
        gl.viewport(0, 0, 1024, 1024);
        context.bindFrameBufferTexture("DirectionalLightSM", "directional_light_sm");
        gl.clear(gl.DEPTH_BUFFER_BIT);

        const lightOrthoSize = 25;
        const lightSpace = m4.mult(
          m4.new(),
          directionalLight.lightSpace,
          m4.ortho(lightOrthoSize, -lightOrthoSize, lightOrthoSize, -lightOrthoSize, -1, -lightOrthoSize * 10),
        );
        uniforms.u_light_space.set(lightSpace);

        for (const object of scene.listObjects()) {
          // Skip non-world objects
          if (object.space !== 0 || object.vertexCount === 0) {
            continue;
          }
          const nodes = scene.listNodes(object);
          if (nodes.length === 0) {
            continue;
          }
          context.updateBuffer("ModelDepth", "models", 0, nodes.map(node => ({ model: node.model })));

          gl.drawArraysInstanced(gl.TRIANGLES, object.vertexOffset, object.vertexCount, nodes.length);
        }
      }

      const spotLights = scene.listSpotLights();
      if (spotLights[0]) {
        const spotLight = spotLights[0];

        gl.viewport(0, 0, 256, 256);
        context.bindFrameBufferTexture("SpotLightSM", "spot_light_sm");
        gl.clear(gl.DEPTH_BUFFER_BIT);

        const lightSpace = m4.mult(
          m4.new(),
          spotLight.lightSpace,
          m4.ortho(10, -10, 10, -10, -0.01, -80),
          // m4.perspective(70 * (Math.PI / 180), context.getAspectRatio(), -0.6, -80)
        );
        uniforms.u_light_space.set(lightSpace);

        for (const object of scene.listObjects()) {
          // Skip non-world objects
          if (object.space !== 0 || object.vertexCount === 0) {
            continue;
          }
          const nodes = scene.listNodes(object);
          if (nodes.length === 0) {
            continue;
          }
          context.updateBuffer("ModelDepth", "models", 0, nodes.map(node => ({ model: node.model })));

          gl.drawArraysInstanced(gl.TRIANGLES, object.vertexOffset, object.vertexCount, nodes.length);
        }
      }

      // Make sure to reset the default frame buffer
      gl.bindFramebuffer(gl.FRAMEBUFFER, null);
    }
  });

  pipeline.addShader(StandardShader, {
    enabled: true,
    setup(gl, program, context) {
      // Note: we can't bind textures here, because setup() is called before any message received
      // It might be best to create textures ahead in pipeline setup(), and then bind them here
      // But it requires to set the correct format & size before.

      // Shadow maps can be bound here because they are initialized on pipeline setup:
      // gl.uniform1i(gl.getUniformLocation(program, "u_dir_light_shadow_map"), context.getTextureIndex("directional_light_sm"));
      // gl.uniform1i(gl.getUniformLocation(program, "u_spot_light_shadow_map"), context.getTextureIndex("spot_light_sm"));

      const vao = gl.createVertexArray();
      gl.bindVertexArray(vao);

      // Bind vertex buffer objects
      const position = gl.getAttribLocation(program, "a_position");
      const texCoords = gl.getAttribLocation(program, "a_uv");
      const normals = gl.getAttribLocation(program, "a_normal");
      const model = gl.getAttribLocation(program, "a_model");

      gl.bindBuffer(gl.ARRAY_BUFFER, context.getBuffer("Vertex"));
      gl.enableVertexAttribArray(position);
      gl.vertexAttribPointer(position, 3, gl.FLOAT, false, 32, 0);
      gl.enableVertexAttribArray(texCoords);
      gl.vertexAttribPointer(texCoords, 2, gl.FLOAT, false, 32, 12);
      gl.enableVertexAttribArray(normals);
      gl.vertexAttribPointer(normals, 3, gl.FLOAT, false, 32, 20);

      gl.bindBuffer(gl.ARRAY_BUFFER, context.getBuffer("Model"));
      gl.enableVertexAttribArray(model);
      gl.vertexAttribPointer(model, 4, gl.FLOAT, false, 64, 0);
      gl.vertexAttribDivisor(model, 1);
      gl.enableVertexAttribArray(model + 1);
      gl.vertexAttribPointer(model + 1, 4, gl.FLOAT, false, 64, 16);
      gl.vertexAttribDivisor(model + 1, 1);
      gl.enableVertexAttribArray(model + 2);
      gl.vertexAttribPointer(model + 2, 4, gl.FLOAT, false, 64, 32);
      gl.vertexAttribDivisor(model + 2, 1);
      gl.enableVertexAttribArray(model + 3);
      gl.vertexAttribPointer(model + 3, 4, gl.FLOAT, false, 64, 48);
      gl.vertexAttribDivisor(model + 3, 1);

      // Bind uniform buffer objects
      // const camera = gl.getUniformBlockIndex(program, "Camera");
      // const directionalLight = gl.getUniformBlockIndex(program, "DirectionalLight");
      // const pointLight = gl.getUniformBlockIndex(program, "PointLight");
      // const spotLight = gl.getUniformBlockIndex(program, "SpotLight");

      // gl.uniformBlockBinding(program, camera, context.getUniformBufferIndex("Camera"));
      // gl.uniformBlockBinding(program, directionalLight, context.getUniformBufferIndex("DirectionalLight"));
      // gl.uniformBlockBinding(program, pointLight, context.getUniformBufferIndex("PointLight"));
      // gl.uniformBlockBinding(program, spotLight, context.getUniformBufferIndex("SpotLight"));

      return vao;
    },
    render(gl, program, uniforms, scene, context) {
      gl.cullFace(gl.BACK);
      gl.viewport(0, 0, gl.canvas.width, gl.canvas.height);
      gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);

      // Bind textures only in first frame, after they are created
      // Note: read other note, ideally it should be in setup()
      if (context.frame === 1) {
        uniforms.u_palette.set(context.getTextureIndex("palette"));
        uniforms.u_material.diffuse.set(context.getTextureIndex("diffuse"));
        uniforms.u_material.specular.set(context.getTextureIndex("specular"));
        uniforms.u_dir_light_shadow_map.set(context.getTextureIndex("directional_light_sm"));
        uniforms.u_spot_light_shadow_map.set(context.getTextureIndex("spot_light_sm"));
      }

      const camera = scene.getCamera();
      if (camera) {
        uniforms.u_view.set(camera.viewMatrix);
        uniforms.u_projection.set(camera.projectionMatrix);
      }

      const directionalLight = scene.getDirectionalLight();
      if (directionalLight) {
        const lightOrthoSize = 25;
        const lightSpace = m4.mult(
          m4.new(),
          directionalLight.lightSpace,
          m4.ortho(lightOrthoSize, -lightOrthoSize, lightOrthoSize, -lightOrthoSize, -1, -lightOrthoSize * 10),
        );
        uniforms.u_dir_light.ambient.set(directionalLight.ambient.x, directionalLight.ambient.y, directionalLight.ambient.z);
        uniforms.u_dir_light.diffuse.set(directionalLight.diffuse.x, directionalLight.diffuse.y, directionalLight.diffuse.z);
        uniforms.u_dir_light.specular.set(directionalLight.specular.x, directionalLight.specular.y, directionalLight.specular.z);
        uniforms.u_dir_light.direction.set(directionalLight.direction.x, directionalLight.direction.y, directionalLight.direction.z);
        uniforms.u_dir_light.intensity.set(.5);
        uniforms.u_directional_light_space.set(lightSpace);
      }

      const spotLights = scene.listSpotLights();
      uniforms.u_spot_light_count.set(spotLights.length);
      for (const [i, spotLight] of spotLights.entries()) {
        const lightSpace = m4.mult(
          m4.new(),
          spotLight.lightSpace,
          m4.ortho(10, -10, 10, -10, -0.01, -80),
          // m4.perspective(70 * (Math.PI / 180), context.getAspectRatio(), -0.6, -80)
        );
        uniforms.u_spot_light_space[i].set(lightSpace);
        uniforms.u_spot_light[i].ambient.set(spotLight.ambient.x, spotLight.ambient.y, spotLight.ambient.z);
        uniforms.u_spot_light[i].diffuse.set(spotLight.diffuse.x, spotLight.diffuse.y, spotLight.diffuse.z);
        uniforms.u_spot_light[i].specular.set(spotLight.specular.x, spotLight.specular.y, spotLight.specular.z);
        uniforms.u_spot_light[i].direction.set(spotLight.direction.x, spotLight.direction.y, spotLight.direction.z);
        uniforms.u_spot_light[i].position.set(spotLight.position.x, spotLight.position.y, spotLight.position.z);
        uniforms.u_spot_light[i].cut_off.set(spotLight.radius);
        uniforms.u_spot_light[i].outer_cut_off.set(spotLight.outerCutOff);
      }

      const pointLights = scene.listPointLights();
      uniforms.u_point_light_count.set(pointLights.length);
      for (const [i, pointLight] of pointLights.entries()) {
        uniforms.u_point_light[i].ambient.set(pointLight.ambient.x, pointLight.ambient.y, pointLight.ambient.z);
        uniforms.u_point_light[i].diffuse.set(pointLight.diffuse.x, pointLight.diffuse.y, pointLight.diffuse.z);
        uniforms.u_point_light[i].specular.set(pointLight.specular.x, pointLight.specular.y, pointLight.specular.z);
        uniforms.u_point_light[i].position.set(pointLight.position.x, pointLight.position.y, pointLight.position.z);
        uniforms.u_point_light[i].radius.set(pointLight.radius);
        uniforms.u_point_light[i].intensity.set(1);
      }

      for (const object of scene.listObjects()) {
        // Skip non-world objects
        if (object.space !== 0 || object.vertexCount === 0) {
          continue;
        }
        const nodes = scene.listNodes(object);
        if (nodes.length === 0) {
          continue;
        }

        uniforms.u_tex_index.set(object.material.diffuse);
        uniforms.u_material.shininess.set(object.material.shininess);

        context.updateBuffer("Model", "models", 0, nodes.map(node => ({ model: node.model })));

        gl.drawArraysInstanced(gl.TRIANGLES, object.vertexOffset, object.vertexCount, nodes.length);
      }
    },
  });

  return pipeline.end();
};