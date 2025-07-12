
const unitMatrix4 = [
  1, 0, 0, 0,
  0, 1, 0, 0,
  0, 0, 1, 0,
  0, 0, 0, 1,
];

const MAT4_DIM = 4;
const MAT4_SIZE = 16;

/**
 * 
 * @param {number} count 
 * @returns 
 */
const createMatrix4 = (count) => {
  if (count === undefined) {
    count = 1;
  }
  const mat = new Float32Array(count * MAT4_SIZE);
  for (let i = 0; i < count; ++i) {
    setMatrix4(mat, i, unitMatrix4);
  }
  return mat;
};

const createRotationXMatrix4 = (angle) => {
  return [
    1, 0, 0, 0,
    0, Math.cos(angle), -Math.sin(angle), 0,
    0, Math.sin(angle), Math.cos(angle), 0,
    0, 0, 0, 1,
  ];
};

const createRotationYMatrix4 = (angle) => {
  return [
    Math.cos(angle), 0, Math.sin(angle), 0,
    0, 1, 0, 0,
    -Math.sin(angle), 0, Math.cos(angle), 0,
    0, 0, 0, 1,
  ];
};

const createRotationZMatrix4 = (angle) => {
  return [
    Math.cos(angle), -Math.sin(angle), 0, 0,
    Math.sin(angle), Math.cos(angle), 0, 0,
    0, 0, 1, 0,
    0, 0, 0, 1,
  ];
};

/**
 * 
 * @param {number[]} matrices 
 * @param {number} index 
 * @returns
 */
const getMatrix4 = (matrices, index) => {
  const mat = new Float32Array(MAT4_SIZE);
  for (let i = 0; i < MAT4_SIZE; ++i) {
    mat[i] = matrices[index * MAT4_SIZE + Math.floor(i / MAT4_DIM) + MAT4_DIM * (i % MAT4_DIM)];
  }
  return mat;
};

const setMatrix4 = (matrices, index, elements) => {
  for (let i = 0; i < MAT4_SIZE; ++i) {
    matrices[index * MAT4_SIZE + i] = elements[Math.floor(i / MAT4_DIM) + MAT4_DIM * (i % MAT4_DIM)];
  }
};

const resetMatrix4 = (matrices, index) => {
  setMatrix4(matrices, index, unitMatrix4);
};

/**
 * Multiplies two mat4s
 *
 * @param {number[]} out the receiving matrix
 * @param {number[]} a the first operand
 * @param {number[]} b the second operand
 * @returns {number[]} out
 */
function multMatrix4(out, a, b) {
  let a00 = a[0],
    a01 = a[4],
    a02 = a[8],
    a03 = a[12];
  let a10 = a[1],
    a11 = a[5],
    a12 = a[9],
    a13 = a[13];
  let a20 = a[2],
    a21 = a[6],
    a22 = a[10],
    a23 = a[14];
  let a30 = a[3],
    a31 = a[7],
    a32 = a[11],
    a33 = a[15];
  // Cache only the current line of the second matrix
  let b0 = b[0],
    b1 = b[4],
    b2 = b[8],
    b3 = b[12];
  out[0] = b0 * a00 + b1 * a10 + b2 * a20 + b3 * a30;
  out[4] = b0 * a01 + b1 * a11 + b2 * a21 + b3 * a31;
  out[8] = b0 * a02 + b1 * a12 + b2 * a22 + b3 * a32;
  out[12] = b0 * a03 + b1 * a13 + b2 * a23 + b3 * a33;
  b0 = b[1];
  b1 = b[5];
  b2 = b[9];
  b3 = b[13];
  out[1] = b0 * a00 + b1 * a10 + b2 * a20 + b3 * a30;
  out[5] = b0 * a01 + b1 * a11 + b2 * a21 + b3 * a31;
  out[9] = b0 * a02 + b1 * a12 + b2 * a22 + b3 * a32;
  out[13] = b0 * a03 + b1 * a13 + b2 * a23 + b3 * a33;
  b0 = b[2];
  b1 = b[6];
  b2 = b[10];
  b3 = b[14];
  out[2] = b0 * a00 + b1 * a10 + b2 * a20 + b3 * a30;
  out[6] = b0 * a01 + b1 * a11 + b2 * a21 + b3 * a31;
  out[10] = b0 * a02 + b1 * a12 + b2 * a22 + b3 * a32;
  out[14] = b0 * a03 + b1 * a13 + b2 * a23 + b3 * a33;
  b0 = b[3];
  b1 = b[7];
  b2 = b[11];
  b3 = b[15];
  out[3] = b0 * a00 + b1 * a10 + b2 * a20 + b3 * a30;
  out[7] = b0 * a01 + b1 * a11 + b2 * a21 + b3 * a31;
  out[11] = b0 * a02 + b1 * a12 + b2 * a22 + b3 * a32;
  out[15] = b0 * a03 + b1 * a13 + b2 * a23 + b3 * a33;
  return out;
}

const multnMatrix4 = (out, ...m) => {
  for (let i = 0; i < m.length - 1; i = i + 1) {
    if (i === 0) {
      multMatrix4(out, m[i], m[i + 1]);
    } else {
      multMatrix4(out, out, m[i + 1]);
    }
  }
  return out;
};

const scaleMatrix4 = (x, y, z) => {
  return [
    x, 0, 0, 0,
    0, y === undefined ? 1 : y, 0, 0,
    0, 0, z === undefined ? 1 : z, 0,
    0, 0, 0, 1,
  ];
};

const rotateMatrix4 = (x, y, z) => {
  return multnMatrix4(
    createMatrix4(),
    x === undefined ? createMatrix4() : createRotationXMatrix4(x),
    y === undefined ? createMatrix4() : createRotationYMatrix4(y),
    z === undefined ? createMatrix4() : createRotationZMatrix4(z),
  );
};

const translateMatrix4 = (x, y, z) => {
  return [
    1, 0, 0, x,
    0, 1, 0, y,
    0, 0, 1, z,
    0, 0, 0, 1,
  ];
};

const perspective = (fieldOfViewInRadians, aspectRatio, near, far) => {
  const f = 1.0 / Math.tan(fieldOfViewInRadians / 2);
  const rangeInv = 1 / (near - far);

  return [
    f / aspectRatio, 0, 0, 0,
    0, f, 0, 0,
    0, 0, (near + far) * rangeInv, near * far * rangeInv * 2,
    0, 0, -1, 0,
  ];
}

const orthoMatrix4 = (right, left, top, bottom, near, far) => {
  return [
    2 / (right - left), 0, 0, 0,
    0, 2 / (top - bottom), 0, 0,
    0, 0, -2 / (far - near), 0,
    -((right + left) / (right - left)), -((top + bottom) / (top - bottom)), -((far + near) / (far - near)), 1,
  ];
};

export const m4 = {
  mult: multMatrix4,
  multn: multnMatrix4,
  reset: resetMatrix4,
  get: getMatrix4,
  set: setMatrix4,
  new: createMatrix4,
  rotateX: createRotationXMatrix4,
  rotateY: createRotationYMatrix4,
  rotateZ: createRotationZMatrix4,
  scale: scaleMatrix4,
  rotate: rotateMatrix4,
  translate: translateMatrix4,
  ortho: orthoMatrix4,
  perspective: perspective,
};