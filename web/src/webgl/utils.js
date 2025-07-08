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
  const indices = new Map();
  let lastObjectIndex = undefined;

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

  return get;
};