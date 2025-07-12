/**
 * @typedef {{
 *  x:number,
 *  y:number,
 *  z:number
 * }} Vector3
 */

/**
 * @typedef {{
 *  diffuse: number,
 *  specular: number,
 *  shininess: number,
 * }} Material
 */


/**
 * @typedef {{
 *  id: number,
 *  material: Material,
 *  space: number,
 *  vertices: Float32Array,
 *  texCoords: Float32Array,
 *  normals: Float32Array,
 *  vertexCount: number,
 *  vertexOffset: number,
 * }} SceneObject
 */

/**
 * @typedef {{
 *  id:number,
 *  objectId: number,
 *  position: Vector3,
 *  rotation: Vector3,
 *  scale: Vector3,
 *  model: Float32Array,
 * }} SceneNode
 */

/**
 * @typedef {{
 *  id:number,
 *  type: number,
 *  lightSpace: Float32Array,
 *  ambient: Vector3,
 *  diffuse: Vector3,
 *  specular: Vector3,
 *  position: Vector3,
 *  direction: Vector3,
 *  radius: number,
 *  outerCutOff: number,
 * }} SceneLight
 */

/**
 * @typedef {{
 *  id:number,
 *  viewMatrix: Float32Array,
 *  projectionMatrix: Float32Array,
 * }} Camera
 */


/**
 * TODO: Move the blocks and schema in server and generate it.
 */



/**
 * Represents the local scene data.
 *
 * @returns 
 */
export const createScene = () => {
  /**
   * @type {Map<number, SceneObject>}
   */
  const objects = new Map();
  let objectChanged = false;
  /**
   * @type {Map<number, SceneNode>}
   */
  const nodes = new Map();
  /**
   * @type {Map<number, Set<SceneNode>>}
   */
  const nodesByObject = new Map();
  /**
   * @type {Map<number, Camera>}
   */
  const cameras = new Map();
  /**
   * @type {Map<number, SceneLight>}
   */
  const lights = new Map();

  /**
   * 
   * @param {number} id 
   * @param {Camera} camera 
   */
  const updateCamera = (id, camera) => {
    cameras.set(id, camera);
  };

  /**
   * 
   * @param {number} id 
   * @param {Partial<Omit<SceneObject, "id">>} data 
   */
  const updateObject = (id, data) => {
    objectChanged = true;
    const object = objects.get(id);
    if (!object) {
      // Compute vertex count and vertex offset for draw calls
      let vertexCount = 0;
      let vertexOffset = 0;
      if (data.vertices) {
        vertexCount = data.vertices.length / 3;
        for (const [_, obj] of objects) {
          vertexOffset += obj.vertexCount;
        }
      }
      objects.set(id, { ...data, id, vertexOffset, vertexCount, });
      nodesByObject.set(id, new Set());
    } else {
      Object.entries(data).forEach(([key, value]) => {
        object[key] = value;
        // todo: update vertex offset and count if vertices change
      });
    }
  };

  /**
   * 
   * @param {number} id 
   * @param {Partial<Omit<SceneNode, "id">>} data 
   */
  const updateNode = (id, data) => {
    const node = nodes.get(id);
    if (!node) {
      nodes.set(id, { ...data, id });
      nodesByObject.get(data.objectId)?.add(nodes.get(id));
    } else {
      Object.entries(data).forEach(([key, value]) => {
        node[key] = value;
      });
    }
  };

  /**
   * 
   * @param {number} id 
   * @param {Partial<Omit<SceneLight, "id">>} data 
   */
  const updateLight = (id, data) => {
    const light = lights.get(id);
    if (!light) {
      lights.set(id, { ...data, id });
    } else {
      Object.entries(data).forEach(([key, value]) => {
        light[key] = value;
      })
    }
  };

  const deleteNode = (id) => {
    nodes.delete(id);
    // Delete 
    // for (const [_, nodes] of nodesByObject) {
    // }

  };

  const deleteLight = (id) => {
    lights.delete(id);
  };

  const listObjects = () => {
    return objects.values();
  };

  const listSpotLights = () => {
    const spotLights = Array.from(lights.values()).filter(l => l.type === 2);
    return spotLights;
  };

  const listPointLights = () => {
    const pointLights = Array.from(lights.values()).filter(l => l.type === 1);
    return pointLights;
  };

  /**
   * Returns true if a scene objects have changed since last frame.
   *
   * @returns 
   */
  const didObjectChange = () => {
    return objectChanged;
  };

  /**
   * 
   * @param {SceneObject} object 
   */
  const listNodes = (object) => {
    if (!nodesByObject.has(object.id)) {
      return [];
    }
    return Array.from(nodesByObject.get(object.id));
  };

  const update = () => {
    objectChanged = false;
  };

  const getCamera = () => {
    return Array.from(cameras.values())[0];
  };

  const getDirectionalLight = () => {
    const light = Array.from(lights.values()).filter(l => l.type === 0);
    return light[0];
  };

  return {
    update,
    updateCamera,
    getCamera,
    deleteLight,
    deleteNode,
    getDirectionalLight,
    updateLight,
    updateNode,
    updateObject,
    listObjects,
    didObjectChange,
    listNodes,
    listPointLights,
    listSpotLights,
    clear() {
      objects.clear();
      lights.clear();
      nodes.clear();
      cameras.clear();
    },
  };
};
