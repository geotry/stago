/* Generated file, DO NOT EDIT! */

/**
 * The Camera uniform block.
 */
export const Camera = {
  const view = {
      /**
       * Set the value of uniform `view`.
       *
       * @param {Float32Array} matrix
       * @param {boolean} transpose
       */
      set(matrix, transpose = false) {
        gl.uniformMatrix4fv(locs[`view`], transpose, matrix);
      },
      /**
       * Returns the value of uniform `view`.
       *
       * @returns {Float32Array}
       */
      get() {
        return gl.getUniform(program, locs[`view`]);
      },

  };


  const projection = {
      /**
       * Set the value of uniform `projection`.
       *
       * @param {Float32Array} matrix
       * @param {boolean} transpose
       */
      set(matrix, transpose = false) {
        gl.uniformMatrix4fv(locs[`projection`], transpose, matrix);
      },
      /**
       * Returns the value of uniform `projection`.
       *
       * @returns {Float32Array}
       */
      get() {
        return gl.getUniform(program, locs[`projection`]);
      },

  };


};
