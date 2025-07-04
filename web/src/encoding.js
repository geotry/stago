const Block = Object.freeze({
  TEXTURE: 0,
  CAMERA: 1,
  SCENE_OBJECT: 2,
  SCENE_OBJECT_INSTANCE: 3,
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
    size: "uint8",
    ortho: "float32[]",
    perspective: "float32[]",
  },
  [Block.SCENE_OBJECT]: {
    id: "uint32",
    textureId: "uint8",
    textureIndex: "uint8",
    isUI: "boolean",
    vertices: "float32[]",
    uv: "float32[]",
    normals: "float32[]",
  },
  [Block.SCENE_OBJECT_INSTANCE]: {
    id: "uint16",
    objectId: "uint32",
    model: "float32[]",
  },
};

const sceneObjectBlocksEntries = Object.fromEntries(
  Object.keys(schema).map(type => [type, Object.entries(schema[type])])
);

/**
 * Read buffer and call handler on each block.
 *
 * @param {DataView} buffer
 * @param {number} frame
 * @param {Object} handler 
 * @param {(texture: {id: number, width: number, height: number, depth: number, format: number, role: number, pixels: Uint8Array}) => void} handler.onTextureUpdated
 * @param {(camera: {ortho: Float32Array, perspective: Float32Array}) => void} handler.onCameraUpdated
 * @param {(sceneObject: {id: number, textureId: number, textureIndex: number, isUI: boolean, vertices: Float32Array, uv: Float32Array, normals: Float32Array}) => void} handler.onSceneObjectUpdated
 * @param {(sceneObjectInstance: {id: number, objectId: number, model: Float32Array}) => void} handler.onSceneObjectInstanceUpdated
 */
export const readSceneObjectBuffer = (buffer, frame, handler) => {
  if (buffer.byteLength === 0) {
    return;
  }

  let offset = 0;

  const blocks = []; // for debugging first frame

  while (offset < buffer.byteLength) {
    const blockType = buffer.getUint8(offset);
    offset += 1;
    const blockSize = buffer.getUint32(offset, false);
    offset += 4;

    if (blockSize === 0) {
      console.log(`empty block ${blockType} at offset ${offset} - remaining buffer:`, buffer.buffer.slice(offset));
    }

    if (schema[blockType] === undefined) {
      console.error(`Error while decoding buffer: block ${blockType} (size=${blockSize}) invalid`);
      console.log(buffer.buffer);
      break;
    }

    const block = Object.fromEntries(
      sceneObjectBlocksEntries[blockType].map(([field, type]) => {
        let value;
        try {
          switch (type) {
            case "uint8":
              value = buffer.getUint8(offset);
              offset += 1;
              break;
            case "boolean":
              value = buffer.getUint8(offset) !== 0;
              offset += 1;
              break;
            case "uint8[]": {
              // Read next block to get array size
              const byteSize = buffer.getUint32(offset, false);
              offset += 4
              value = new Uint8Array(byteSize);
              for (let i = 0; i < byteSize; ++i) {
                value[i] = buffer.getUint8(offset);
                offset += 1;
              }
              break;
            }
            case "uint16":
              value = buffer.getUint16(offset, false);
              offset += 2;
              break;
            case "uint32":
              value = buffer.getUint32(offset, false);
              offset += 4;
              break;
            case "float32":
              value = buffer.getFloat32(offset, false);
              offset += 4;
              break;
            case "float32[]": {
              // Read next block to get array size
              const byteSize = buffer.getUint32(offset, false);
              const arraySize = byteSize / 4;
              offset += 4
              value = new Float32Array(arraySize);
              for (let i = 0; i < arraySize; ++i) {
                value[i] = buffer.getFloat32(offset, false);
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

    if (frame === 1) {
      blocks.push(block);
    }

    switch (blockType) {
      case Block.TEXTURE:
        handler.onTextureUpdated(block);
        break;
      case Block.CAMERA:
        handler.onCameraUpdated(block);
        break;
      case Block.SCENE_OBJECT:
        handler.onSceneObjectUpdated(block);
        break;
      case Block.SCENE_OBJECT_INSTANCE:
        handler.onSceneObjectInstanceUpdated(block);
        break;
    }
  }

  // if (frame === 1) {
  //   console.log(blocks);
  // }
};

