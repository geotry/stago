export const SceneObjectBlock = Object.freeze({
  TEXTURE: 0,
  MATRIX: 5,
});

const sceneObjectBlocks = {
  [SceneObjectBlock.TEXTURE]: {
    index: "uint16",
    textureIndex: "uint8",
  },
  [SceneObjectBlock.MATRIX]: {
    index: "uint16",
    matrix: "float32[16]",
  },
};

const sceneObjectBlocksEntries = Object.fromEntries(
  Object.keys(sceneObjectBlocks).map(type => [type, Object.entries(sceneObjectBlocks[type])])
);

/**
 * Read buffer and call cb on each block.
 *
 * @param {DataView} buffer
 * @param {number} bufferOffset
 * @param {Function} cb 
 */
export const readSceneObjectBuffer = (buffer, bufferOffset, cb) => {
  let offset = bufferOffset;

  while (offset < buffer.byteLength) {
    const blockType = buffer.getUint8(offset);
    offset += 1;

    if (offset >= buffer.byteLength) {
      console.log(`Error while decoding buffer: max size reached (block: ${blockType}, offset: ${offset}, buffer length: ${buffer.byteLength})`);
      break;
    }

    if (sceneObjectBlocks[blockType] === undefined) {
      console.log(`Error while decoding buffer: block ${blockType} invalid`);
      continue;
    }

    const block = Object.fromEntries(
      sceneObjectBlocksEntries[blockType].map(([field, type]) => {
        let value;
        let size;

        // Handle value as an array
        if (type.indexOf("[") !== -1) {
          size = Number(type.slice(type.indexOf("[") + 1, type.indexOf("]")));
          type = type.slice(0, type.indexOf("[")) + "[]";
        }

        try {
          switch (type) {
            case "uint8":
              value = buffer.getUint8(offset);
              offset += 1;
              break;
            case "uint16":
              value = buffer.getUint16(offset);
              offset += 2;
              break;
            case "float32":
              value = buffer.getFloat32(offset, false);
              offset += 4;
              break;
            case "float32[]":
              value = new Float32Array(size);
              for (let i = 0; i < size; ++i) {
                value[i] = buffer.getFloat32(offset, false);
                offset += 4;
              }
              break;
          }
        } catch (err) {
          console.log(`Error while decoding ${type} field ${field} in block ${blockType} (offset: ${offset}, buffer length: ${buffer.byteLength})`);
          throw err;
        }
        return [field, value];
      })
    );

    cb(blockType, block);
  }
};

