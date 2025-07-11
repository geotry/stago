const { readSceneObjectBuffer } = require("./encoding.js");
const { loadShaderProgram, prepareProgram } = require("./webgl/shader.js");
const { createArrayBuffer, createDepthMapBuffer, createTexture } = require("./webgl/buffers.js");
const { uniforms } = require("./webgl/uniforms.js");

const MAX_OBJECTS = 1024;
const MAX_VERTICES = MAX_OBJECTS * 12;

const createStandardShader = async (gl) => {
  const program = await loadShaderProgram(gl, "standard");

  const vao = gl.createVertexArray();
  gl.bindVertexArray(vao);

  // Create buffers
  const vertexBuffer = createArrayBuffer(gl, program, MAX_VERTICES, 1, gl.STATIC_DRAW, [
    {
      name: "a_position",
      type: gl.FLOAT,
      size: 3,
    },
    {
      name: "a_uv",
      type: gl.FLOAT,
      size: 2,
    },
    {
      name: "a_normal",
      type: gl.FLOAT,
      size: 3,
    },
  ]);

  const modelBuffer = createArrayBuffer(gl, program, MAX_OBJECTS, MAX_OBJECTS, gl.DYNAMIC_DRAW, [
    {
      name: "a_model",
      type: gl.FLOAT,
      size: 4,
      repeat: 4,
      instance: 1,
    },
  ]);

  // Setup uniforms
  const uniformData = {
    viewMatrix: uniforms.createMatrix4fv(gl, program, "u_view"),
    projectionMatrix: uniforms.createMatrix4fv(gl, program, "u_projection"),
    palette: uniforms.create1i(gl, program, "u_palette"),
    texIndex: uniforms.create1i(gl, program, "u_tex_index"),
    material: {
      diffuse: uniforms.create1i(gl, program, "u_material.diffuse"),
      specular: uniforms.create1i(gl, program, "u_material.specular"),
      shininess: uniforms.create1f(gl, program, "u_material.shininess"),
    },
    directionalLight: {
      direction: uniforms.create3f(gl, program, "u_dir_light.direction"),
      ambient: uniforms.create3f(gl, program, "u_dir_light.ambient"),
      diffuse: uniforms.create3f(gl, program, "u_dir_light.diffuse"),
      specular: uniforms.create3f(gl, program, "u_dir_light.specular"),
      intensity: uniforms.create1f(gl, program, "u_dir_light.intensity"),
    },
    pointLightCount: uniforms.create1i(gl, program, "u_point_light_count"),
    pointLights: new Array(10).fill(0).map((_, i) => ({
      position: uniforms.create3f(gl, program, `u_point_light[${i}].position`),
      ambient: uniforms.create3f(gl, program, `u_point_light[${i}].ambient`),
      diffuse: uniforms.create3f(gl, program, `u_point_light[${i}].diffuse`),
      specular: uniforms.create3f(gl, program, `u_point_light[${i}].specular`),
      radius: uniforms.create1f(gl, program, `u_point_light[${i}].radius`),
      intensity: uniforms.create1f(gl, program, `u_point_light[${i}].intensity`),
    })),
    spotLightCount: uniforms.create1i(gl, program, "u_spot_light_count"),
    spotLights: new Array(10).fill(0).map((_, i) => ({
      position: uniforms.create3f(gl, program, `u_spot_light[${i}].position`),
      direction: uniforms.create3f(gl, program, `u_spot_light[${i}].direction`),
      ambient: uniforms.create3f(gl, program, `u_spot_light[${i}].ambient`),
      diffuse: uniforms.create3f(gl, program, `u_spot_light[${i}].diffuse`),
      specular: uniforms.create3f(gl, program, `u_spot_light[${i}].specular`),
      cutOff: uniforms.create1f(gl, program, `u_spot_light[${i}].cut_off`),
      outerCutOff: uniforms.create1f(gl, program, `u_spot_light[${i}].outer_cut_off`),
    })),
  };

  const compiled = prepareProgram(gl, program, vao, [vertexBuffer], [modelBuffer], uniformData);

  return {
    ...compiled,
    buffers: {
      vertex: vertexBuffer,
      model: modelBuffer,
    },
    uniforms: uniformData,
  };
};

const createDepthShader = async (gl) => {
  const program = await loadShaderProgram(gl, "depth");

  const vao = gl.createVertexArray();
  gl.bindVertexArray(vao);

  const vertexBuffer = createArrayBuffer(gl, program, MAX_VERTICES, 1, gl.STATIC_DRAW, [
    {
      name: "a_position",
      type: gl.FLOAT,
      size: 3,
    }
  ]);

  const modelBuffer = createArrayBuffer(gl, program, MAX_OBJECTS, MAX_OBJECTS, gl.DYNAMIC_DRAW, [
    {
      name: "a_model",
      type: gl.FLOAT,
      size: 4,
      repeat: 4,
      instance: 1,
    },
  ]);

  const uniformData = {};

  const compiled = prepareProgram(gl, program, vao, [vertexBuffer], [modelBuffer], uniformData);

  return { ...compiled, buffers: { vertex: vertexBuffer, model: modelBuffer }, uniforms: uniformData };
};

/**
 * 
 * @param {OffscreenCanvas} canvas 
 */
export const createContext = async (canvas) => {
  const gl = canvas.getContext("webgl2", { antialias: false });
  gl.enable(gl.DEPTH_TEST);
  gl.depthFunc(gl.LESS);

  gl.enable(gl.BLEND)
  gl.blendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);

  gl.clearColor(.0, .0, .0, .0);
  gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);

  // Create shaders
  const shaders = {
    std: await createStandardShader(gl),
    depth: await createDepthShader(gl),
  };

  const SHADOW_SIZE = 1024;
  const [depthMapBuffer, depthMap] = createDepthMapBuffer(gl, SHADOW_SIZE);

  let renderCount = 0;

  return {
    reset() { },
    resize(width, height) {
      canvas.width = width;
      canvas.height = height;
      gl.viewport(0, 0, gl.canvas.width, gl.canvas.height);
    },

    /**
     * @param {ArrayBuffer} buffer 
     */
    render(buffer) {
      renderCount++;

      const view = new DataView(buffer);

      readSceneObjectBuffer(view, renderCount, {
        onTextureUpdated(t) {
          createTexture(gl, t.role, t.format, t.width, t.depth, t.pixels);
          switch (t.role) {
            case 0:
              shaders.std.uniforms.material.diffuse.set(t.role);
              break;
            case 1:
              shaders.std.uniforms.palette.set(t.role);
              break;
            case 2:
              shaders.std.uniforms.material.specular.set(t.role);
              break;
            case 3:
              shaders.std.uniforms.material.normal.set(t.role);
              break;
          }
        },
        onSceneObjectUpdated(o) {
          // todo: improve this
          const stdVertexBufferData = [];
          const depthVertexBufferData = [];
          const verticesCount = o.vertices.length / 3;

          // todo: Share buffer data accross buffers (using data view?)
          for (let i = 0; i < verticesCount; i++) {
            stdVertexBufferData.push({
              a_position: o.vertices.slice(i * 3, i * 3 + 3),
              a_uv: o.uv.slice(i * 2, i * 2 + 2),
              a_normal: o.normals.slice(i * 3, i * 3 + 3),
            });
            depthVertexBufferData.push({
              a_position: o.vertices.slice(i * 3, i * 3 + 3),
            });
          }

          shaders.std.buffers.vertex.set(0, o.id, stdVertexBufferData);
          shaders.std.uniforms.texIndex.prepare(o.id, o.textureIndex);
          shaders.std.uniforms.material.shininess.prepare(o.id, o.shininess);

          shaders.depth.buffers.vertex.set(0, o.id, depthVertexBufferData);

          shaders.std.recordObjectVertices(o.id, verticesCount);
          shaders.depth.recordObjectVertices(o.id, verticesCount);
        },
        onCameraUpdated(c) {
          shaders.std.uniforms.viewMatrix.set(c.viewMatrix);
          shaders.std.uniforms.projectionMatrix.set(c.projectionMatrix);
        },
        onSceneObjectInstanceUpdated(i) {
          shaders.std.buffers.model.set(shaders.std.objectIndex.get(i.objectId), i.id, [{ a_model: i.model }]);

          shaders.depth.buffers.model.set(shaders.depth.objectIndex.get(i.objectId), i.id, [{ a_model: i.model }]);
          // Record object and instance in program to draw it
          shaders.std.recordObject(i.objectId, i.id);
          shaders.depth.recordObject(i.objectId, i.id);
        },
        onLightUpdated(l) {
          switch (l.type) {
            // directional light
            case 0: {
              shaders.std.uniforms.directionalLight.ambient.set(l.ambientR, l.ambientG, l.ambientB);
              shaders.std.uniforms.directionalLight.diffuse.set(l.diffuseR, l.diffuseG, l.diffuseB);
              shaders.std.uniforms.directionalLight.specular.set(l.specularR, l.specularG, l.specularB);
              shaders.std.uniforms.directionalLight.direction.set(l.directionX, l.directionY, l.directionZ);
              shaders.std.uniforms.directionalLight.intensity.set(1);
              break;
            }
            // point light
            case 1: {
              const index = shaders.std.pointLightIndex.get(l.id);
              shaders.std.uniforms.pointLightCount.set(shaders.std.pointLightIndex.size());
              shaders.std.uniforms.pointLights[index].position.set(l.posX, l.posY, l.posZ);
              shaders.std.uniforms.pointLights[index].ambient.set(l.ambientR, l.ambientG, l.ambientB);
              shaders.std.uniforms.pointLights[index].diffuse.set(l.diffuseR, l.diffuseG, l.diffuseB);
              shaders.std.uniforms.pointLights[index].specular.set(l.specularR, l.specularG, l.specularB);
              shaders.std.uniforms.pointLights[index].radius.set(l.radius);
              shaders.std.uniforms.pointLights[index].intensity.set(1);
              break;
            }
            // spot light
            case 2: {
              const index = shaders.std.spotLightIndex.get(l.id);
              shaders.std.uniforms.spotLightCount.set(shaders.std.spotLightIndex.size());
              shaders.std.uniforms.spotLights[index].position.set(l.posX, l.posY, l.posZ);
              shaders.std.uniforms.spotLights[index].direction.set(l.directionX, l.directionY, l.directionZ);
              shaders.std.uniforms.spotLights[index].ambient.set(l.ambientR, l.ambientG, l.ambientB);
              shaders.std.uniforms.spotLights[index].diffuse.set(l.diffuseR, l.diffuseG, l.diffuseB);
              shaders.std.uniforms.spotLights[index].specular.set(l.specularR, l.specularG, l.specularB);
              shaders.std.uniforms.spotLights[index].cutOff.set(l.radius)
              shaders.std.uniforms.spotLights[index].outerCutOff.set(l.outerCutoff)
              break;
            }
          }
        }
      });

      // Render depth map texture for shadows
      gl.viewport(0, 0, SHADOW_SIZE, SHADOW_SIZE);
      gl.bindFramebuffer(gl.FRAMEBUFFER, depthMapBuffer);
      gl.clear(gl.DEPTH_BUFFER_BIT);
      // todo: use light view matrix
      shaders.depth.render();
      gl.bindFramebuffer(gl.FRAMEBUFFER, null);

      // Render normal scene
      gl.viewport(0, 0, gl.canvas.width, gl.canvas.height);
      shaders.std.render(renderCount);
    },
  };
};