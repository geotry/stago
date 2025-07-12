/**
 * 
 * @param {WebGL2RenderingContext} gl 
 * @param {number} type 
 * @returns 
 */
export const typeByteSize = (gl, type) => {
  switch (type) {
    case gl.BYTE:
    case gl.UNSIGNED_BYTE:
      return 1;
    case gl.UNSIGNED_SHORT:
    case gl.SHORT:
    case gl.HALF_FLOAT:
      return 2;
    case gl.FLOAT:
    case gl.INT:
    case gl.UNSIGNED_INT:
      return 4;
  }
  throw new Error(`Unknown type ${type}`);
};

/**
 * 
 * @param {WebGL2RenderingContext} gl 
 * @param {number} type 
 * @returns 
 */
export const typeIsInt = (gl, type) => {
  switch (type) {
    case gl.BYTE:
    case gl.UNSIGNED_BYTE:
    case gl.UNSIGNED_SHORT:
    case gl.SHORT:
    case gl.INT:
    case gl.UNSIGNED_INT:
      return true;
  }
  return false;
};

/**
 * Map any ID to an incremental index (0, 1, ...).
 *
 * @returns 
 */
export const createIndexMap = () => {
  /** @type {Map<number, number>} */
  const indices = new Map();
  let lastObjectIndex = undefined;

  const has = (key) => indices.has(key);

  /**
   * 
   * @param {string|number} id 
   * @returns {number}
   */
  const get = (id) => {
    let objectIndex = indices.get(id);
    if (objectIndex === undefined) {
      const lastIndex = lastObjectIndex !== undefined ? indices.get(lastObjectIndex) : -1;
      objectIndex = lastIndex + 1;
      indices.set(id, objectIndex);
      lastObjectIndex = id;
    }
    return objectIndex;
  };

  const del = (key) => {
    const index = indices.get(key);
    if (index !== undefined) {
      // shift all indices after this one
      for (const [iKey, iIndex] of indices.entries()) {
        if (iIndex > index) {
          indices.set(iKey, iIndex - 1);
        }
      }
      indices.delete(key);
      lastObjectIndex = undefined;
      let maxIndex = 0;
      indices.forEach((iIndex, iKey) => {
        if (iIndex > maxIndex) {
          maxIndex = iIndex;
          lastObjectIndex = iKey;
        }
      });
    }
  };

  const size = () => indices.size;

  const clear = () => {
    indices.clear();
    lastObjectIndex = undefined;
  };

  return { has, get, size, del, clear };
};