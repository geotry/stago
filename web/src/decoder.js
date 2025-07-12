
/**
 * @typedef {{[BlockTypeSymbol]: number}} GenericBuffer
 */

/**
 * @typedef {{
 *  id: number,
 *  width: number,
 *  height: number,
 *  depth: number,
 *  format: number,
 *  role: number,
 *  pixels: Uint8Array,
 * }} TextureBuffer
 */

/**
 * @typedef {{
 *  id: number,
 *  view: Float32Array,
 *  projection: Float32Array,
 * }} CameraBuffer
 */

/**
 * @typedef {{
 *  id: number,
 *  textureId: number,
 *  textureIndex: number,
 *  shininess: number,
 *  space: number,
 *  vertices: Float32Array,
 *  uv: Float32Array,
 *  normals: Float32Array,
 * }} SceneObjectBuffer
 */

/**
 * @typedef {{
 *  id: number,
 *  objectId: number,
 *  model: Float32Array,
 * }} SceneNodeBuffer
 */

/**
 * @typedef {{
 *  id: number,
 *  objectId: number,
 * }} SceneNodeDeletedBuffer
 */

/**
 * @typedef {{
 *  id: number,
 *  type: number,
 *  ambientR: number,
 *  ambientG: number,
 *  ambientB: number,
 *  diffuseR: number,
 *  diffuseG: number,
 *  diffuseB: number,
 *  specularR: number,
 *  specularG: number,
 *  specularB: number,
 *  posX: number,
 *  posY: number,
 *  posZ: number,
 *  viewSpace: Float32Array,
 *  directionX: number,
 *  directionY: number,
 *  directionZ: number,
 *  radius: number,
 *  outerCutOff: number,
 * }} SceneLightBuffer
 */

/**
 * @typedef {{
 *  id: number,
 *  type: number,
 * }} SceneLightDeletedBuffer
 */

const Block = Object.freeze({
  TEXTURE: 0,
  CAMERA: 1,
  SCENE_OBJECT: 2,
  SCENE_OBJECT_INSTANCE: 3,
  SCENE_OBJECT_INSTANCE_DELETED: 6,
  LIGHT: 4,
  LIGHT_DELETED: 5,
});

const schema = {
  [Block.TEXTURE]: {
    id: "uint8",
    width: "uint16",
    height: "uint16",
    depth: "uint8",
    format: "uint8",
    role: "uint8",
    pixels: "uint8[]",
  },
  [Block.CAMERA]: {
    id: "uint16",
    viewMatrix: "float32[]",
    projectionMatrix: "float32[]",
  },
  // [Block.CAMERA_PROJECTION]: {
  //   cameraId: "uint16",
  //   projectionMatrix: "float32[]",
  // },
  [Block.SCENE_OBJECT]: {
    id: "uint32",
    textureId: "uint8",
    textureIndex: "uint8",
    shininess: "float32",
    space: "uint8", // 0: World, 1: Screen
    vertices: "float32[]",
    uv: "float32[]",
    normals: "float32[]",
  },
  [Block.SCENE_OBJECT_INSTANCE]: {
    id: "uint16",
    objectId: "uint32",
    model: "float32[]",
  },
  [Block.SCENE_OBJECT_INSTANCE_DELETED]: {
    id: "uint16",
    objectId: "uint32",
  },
  [Block.LIGHT]: {
    id: "uint16",
    // 0: directional
    // 1: point
    // 2: spot
    type: "uint8",
    ambientR: "float32",
    ambientG: "float32",
    ambientB: "float32",
    diffuseR: "float32",
    diffuseG: "float32",
    diffuseB: "float32",
    specularR: "float32",
    specularG: "float32",
    specularB: "float32",
    // model: "float32[]",
    posX: "float32",
    posY: "float32",
    posZ: "float32",
    viewSpace: "float32[]",
    directionX: "float32",
    directionY: "float32",
    directionZ: "float32",
    radius: "float32",
    outerCutOff: "float32",
  },
  [Block.LIGHT_DELETED]: {
    id: "uint16",
    type: "uint8",
  }
};

const BlockTypeSymbol = Symbol();

const sceneObjectBlocksEntries = Object.fromEntries(
  Object.keys(schema).map(type => [type, Object.entries(schema[type])])
);

/**
 * @param {GenericBuffer} buffer 
 * @return {buffer is TextureBuffer}
 */
export const assertTexture = (buffer) => {
  return buffer[BlockTypeSymbol] === Block.TEXTURE;
};

/**
 * @param {GenericBuffer} buffer 
 * @return {buffer is CameraBuffer}
 */
export const assertCamera = (buffer) => {
  return buffer[BlockTypeSymbol] === Block.CAMERA;
};

/**
 * @param {GenericBuffer} buffer 
 * @return {buffer is SceneObjectBuffer}
 */
export const assertSceneObject = (buffer) => {
  return buffer[BlockTypeSymbol] === Block.SCENE_OBJECT;
};

/**
 * @param {GenericBuffer} buffer 
 * @return {buffer is SceneNodeBuffer}
 */
export const assertSceneNode = (buffer) => {
  return buffer[BlockTypeSymbol] === Block.SCENE_OBJECT_INSTANCE;
};

/**
 * @param {GenericBuffer} buffer 
 * @return {buffer is SceneNodeDeletedBuffer}
 */
export const assertSceneNodeDeleted = (buffer) => {
  return buffer[BlockTypeSymbol] === Block.SCENE_OBJECT_INSTANCE_DELETED;
};

/**
 * @param {GenericBuffer} buffer 
 * @return {buffer is SceneLightBuffer}
 */
export const assertSceneLight = (buffer) => {
  return buffer[BlockTypeSymbol] === Block.LIGHT;
};

/**
 * @param {GenericBuffer} buffer 
 * @return {buffer is SceneLightDeletedBuffer}
 */
export const assertSceneLightDeleted = (buffer) => {
  return buffer[BlockTypeSymbol] === Block.LIGHT_DELETED;
};


/**
 * Decode a buffer message from server and return blocks.
 *
 * @param {ArrayBuffer} buffer
 * @returns {Generator<GenericBuffer>}
 */
export function* decodeBuffer(buffer) {
  if (buffer.byteLength === 0) {
    return;
  }

  const view = new DataView(buffer);
  let offset = 0;

  while (offset < buffer.byteLength) {
    const blockType = view.getUint8(offset);
    offset += 1;
    const blockSize = view.getUint32(offset, false);
    offset += 4;

    if (blockSize === 0) {
      console.log(`empty block ${blockType} at offset ${offset} - remaining buffer:`, buffer.slice(offset));
    }

    if (schema[blockType] === undefined) {
      console.error(`Error while decoding buffer: block ${blockType} (size=${blockSize}) invalid`);
      console.log(buffer);
      break;
    }

    const block = Object.fromEntries(
      sceneObjectBlocksEntries[blockType].map(([field, type]) => {
        let value;
        try {
          switch (type) {
            case "uint8":
              value = view.getUint8(offset);
              offset += 1;
              break;
            case "boolean":
              value = view.getUint8(offset) !== 0;
              offset += 1;
              break;
            case "uint8[]": {
              // Read next block to get array size
              const byteSize = view.getUint32(offset, false);
              offset += 4
              value = new Uint8Array(byteSize);
              for (let i = 0; i < byteSize; ++i) {
                value[i] = view.getUint8(offset);
                offset += 1;
              }
              break;
            }
            case "uint16":
              value = view.getUint16(offset, false);
              offset += 2;
              break;
            case "uint32":
              value = view.getUint32(offset, false);
              offset += 4;
              break;
            case "float32":
              value = view.getFloat32(offset, false);
              offset += 4;
              break;
            case "float32[]": {
              // Read next block to get array size
              const byteSize = view.getUint32(offset, false);
              const arraySize = byteSize / 4;
              offset += 4
              value = new Float32Array(arraySize);
              for (let i = 0; i < arraySize; ++i) {
                value[i] = view.getFloat32(offset, false);
                offset += 4;
              }
              break;
            }
          }
        } catch (err) {
          console.log(`Error while decoding ${type} field ${field} in block ${blockType} (offset: ${offset}, buffer length: ${buffer.byteLength})`);
          throw err;
        }
        return [field, value];
      })
    );

    // Add block type to discriminate it with assert*()
    block[BlockTypeSymbol] = blockType;

    yield block;
  }
};
