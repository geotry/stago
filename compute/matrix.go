package compute

import (
	"fmt"
	"math"
)

type Matrix = []float64

type Matrix4 struct {
	Out  Matrix
	buf  Matrix
	buf2 Matrix
}

// Note: Matrix4 uses column-major order to follow opengl convention

// Row-major
// 00 01 02 03
// 04 05 06 07
// 08 09 10 11
// 12 13 14 15

// Column-major
// 00 04 08 12
// 01 05 09 13
// 02 06 10 14
// 03 07 11 15

var unitMatrix4 = Matrix{
	1, 0, 0, 0,
	0, 1, 0, 0,
	0, 0, 1, 0,
	0, 0, 0, 1,
}

func NewMatrix4() *Matrix4 {
	var out = make(Matrix, len(unitMatrix4))
	var buf = make(Matrix, len(unitMatrix4))
	var buf2 = make(Matrix, len(unitMatrix4))
	copy(out, unitMatrix4)
	copy(buf, unitMatrix4)
	copy(buf2, unitMatrix4)
	return &Matrix4{
		Out:  out,
		buf:  buf,
		buf2: buf2,
	}
}

func (m *Matrix4) String() string {
	return fmt.Sprintf(
		"\n%.2f %.2f %.2f %.2f\n%.2f %.2f %.2f %.2f\n%.2f %.2f %.2f %.2f\n%.2f %.2f %.2f %.2f",
		m.Out[0],
		m.Out[1],
		m.Out[2],
		m.Out[3],
		m.Out[4],
		m.Out[5],
		m.Out[6],
		m.Out[7],
		m.Out[8],
		m.Out[9],
		m.Out[10],
		m.Out[11],
		m.Out[12],
		m.Out[13],
		m.Out[14],
		m.Out[15],
	)

}

func mult4(out Matrix, a Matrix, b Matrix) {
	// var a00, a01, a02, a03 = a[0], a[1], a[2], a[3]
	var a00, a01, a02, a03 = a[0], a[4], a[8], a[12]
	// var a10, a11, a12, a13 = a[4], a[5], a[6], a[7]
	var a10, a11, a12, a13 = a[1], a[5], a[9], a[13]
	// var a20, a21, a22, a23 = a[8], a[9], a[10], a[11]
	var a20, a21, a22, a23 = a[2], a[6], a[10], a[14]
	// var a30, a31, a32, a33 = a[12], a[13], a[14], a[15]
	var a30, a31, a32, a33 = a[3], a[7], a[11], a[15]
	// Cache only the current line of the second matrix
	// var b0, b1, b2, b3 = b[0], b[1], b[2], b[3]
	var b0, b1, b2, b3 = b[0], b[4], b[8], b[12]
	out[0] = b0*a00 + b1*a10 + b2*a20 + b3*a30
	// out[1] = b0*a01 + b1*a11 + b2*a21 + b3*a31
	out[4] = b0*a01 + b1*a11 + b2*a21 + b3*a31
	// out[2] = b0*a02 + b1*a12 + b2*a22 + b3*a32
	out[8] = b0*a02 + b1*a12 + b2*a22 + b3*a32
	// out[3] = b0*a03 + b1*a13 + b2*a23 + b3*a33
	out[12] = b0*a03 + b1*a13 + b2*a23 + b3*a33
	// b0 = b[4]
	b0 = b[1]
	b1 = b[5]
	// b2 = b[6]
	b2 = b[9]
	// b3 = b[7]
	b3 = b[13]
	// out[4] = b0*a00 + b1*a10 + b2*a20 + b3*a30
	out[1] = b0*a00 + b1*a10 + b2*a20 + b3*a30
	out[5] = b0*a01 + b1*a11 + b2*a21 + b3*a31
	// out[6] = b0*a02 + b1*a12 + b2*a22 + b3*a32
	out[9] = b0*a02 + b1*a12 + b2*a22 + b3*a32
	// out[7] = b0*a03 + b1*a13 + b2*a23 + b3*a33
	out[13] = b0*a03 + b1*a13 + b2*a23 + b3*a33
	// b0 = b[8]
	b0 = b[2]
	// b1 = b[9]
	b1 = b[6]
	b2 = b[10]
	// b3 = b[11]
	b3 = b[14]
	// out[8] = b0*a00 + b1*a10 + b2*a20 + b3*a30
	out[2] = b0*a00 + b1*a10 + b2*a20 + b3*a30
	// out[9] = b0*a01 + b1*a11 + b2*a21 + b3*a31
	out[6] = b0*a01 + b1*a11 + b2*a21 + b3*a31
	out[10] = b0*a02 + b1*a12 + b2*a22 + b3*a32
	// out[11] = b0*a03 + b1*a13 + b2*a23 + b3*a33
	out[14] = b0*a03 + b1*a13 + b2*a23 + b3*a33
	// b0 = b[12]
	b0 = b[3]
	// b1 = b[13]
	b1 = b[7]
	// b2 = b[14]
	b2 = b[11]
	b3 = b[15]
	// out[12] = b0*a00 + b1*a10 + b2*a20 + b3*a30
	out[3] = b0*a00 + b1*a10 + b2*a20 + b3*a30
	// out[13] = b0*a01 + b1*a11 + b2*a21 + b3*a31
	out[7] = b0*a01 + b1*a11 + b2*a21 + b3*a31
	// out[14] = b0*a02 + b1*a12 + b2*a22 + b3*a32
	out[11] = b0*a02 + b1*a12 + b2*a22 + b3*a32
	out[15] = b0*a03 + b1*a13 + b2*a23 + b3*a33
}

func (m *Matrix4) Mult(matrix Matrix) *Matrix4 {
	mult4(m.Out, m.Out, matrix)
	return m
}

func (m *Matrix4) Scale(scale Size) *Matrix4 {
	copy(m.buf, unitMatrix4)
	m.buf[0] = scale.X
	m.buf[5] = scale.Y
	m.buf[10] = scale.Z
	mult4(m.Out, m.Out, m.buf)
	return m
}

func (m *Matrix4) Rotate(q Quaternion) *Matrix4 {
	xx := q.X * q.X
	yy := q.Y * q.Y
	zz := q.Z * q.Z

	copy(m.buf, unitMatrix4)

	m.buf[0] = 1 - 2*yy - 2*zz
	m.buf[1] = 2*q.X*q.Y + 2*q.Z*q.W
	m.buf[2] = 2*q.X*q.Z - 2*q.Y*q.W

	m.buf[4] = 2*q.X*q.Y - 2*q.Z*q.W
	m.buf[5] = 1 - 2*xx - 2*zz
	m.buf[6] = 2*q.Y*q.Z + 2*q.X*q.W

	m.buf[8] = 2*q.X*q.Z + 2*q.Y*q.W
	m.buf[9] = 2*q.Y*q.Z - 2*q.X*q.W
	m.buf[10] = 1 - 2*xx - 2*yy

	mult4(m.Out, m.Out, m.buf)

	return m
}

func (m *Matrix4) Translate(point Point) *Matrix4 {
	copy(m.buf, unitMatrix4)
	// (row-major)
	// m.buf[3] = point.X
	// m.buf[7] = point.Y
	// m.buf[11] = point.Z
	m.buf[12] = point.X
	m.buf[13] = point.Y
	m.buf[14] = point.Z
	mult4(m.Out, m.Out, m.buf)
	return m
}

func (m *Matrix4) LookAt(center Point, eye Point) *Matrix4 {
	zaxis := center.Sub(eye).Normalize()
	xaxis := zaxis.Cross(Point{X: 0, Y: 1, Z: 0}).Normalize()
	yaxis := xaxis.Cross(zaxis)

	// Rotation buffer
	copy(m.buf, unitMatrix4)
	m.buf[0] = xaxis.X
	m.buf[1] = yaxis.X
	m.buf[2] = zaxis.X
	m.buf[4] = xaxis.Y
	m.buf[5] = yaxis.Y
	m.buf[6] = zaxis.Y
	m.buf[8] = xaxis.Z
	m.buf[9] = yaxis.Z
	m.buf[10] = zaxis.Z

	// Translation buffer
	copy(m.buf2, unitMatrix4)
	m.buf2[12] = -eye.X
	m.buf2[13] = -eye.Y
	m.buf2[14] = -eye.Z

	mult4(m.buf, m.buf2, m.buf)
	mult4(m.Out, m.Out, m.buf)

	return m
}

func (m *Matrix4) Orthographic(right, left, top, bottom, near, far float64) *Matrix4 {
	copy(m.buf, unitMatrix4)
	m.buf[0] = 2.0 / (right - left)
	// m.buf[3] = -(right + left) / (right - left)
	m.buf[12] = -(right + left) / (right - left)
	m.buf[5] = 2.0 / (top - bottom)
	// m.buf[7] = -(top + bottom) / (top - bottom)
	m.buf[13] = -(top + bottom) / (top - bottom)
	m.buf[10] = -2.0 / (far - near)
	// m.buf[11] = -(far + near) / (far - near)
	m.buf[14] = -(far + near) / (far - near)
	mult4(m.Out, m.Out, m.buf)
	return m
}

func (m *Matrix4) Perspective(fov float64, aspectRatio float64, near, far float64) *Matrix4 {
	// var f = 1.0 / math.Tan(fov/2.0)
	var f = math.Tan(math.Pi*0.5 - 0.5*fov)
	var rangeInv = 1.0 / (near - far)
	copy(m.buf, unitMatrix4)
	m.buf[0] = f / aspectRatio
	m.buf[5] = f
	// m.buf[10] = (near + far) * rangeInv
	m.buf[10] = far * rangeInv
	m.buf[11] = -1
	m.buf[14] = near * far * rangeInv
	// m.buf[14] = near * far * rangeInv * 2.0
	// m.buf[11] = near * far * rangeInv * 2.0
	mult4(m.Out, m.Out, m.buf)
	return m
}

func (m *Matrix4) Reset() *Matrix4 {
	copy(m.Out, unitMatrix4)
	copy(m.buf, unitMatrix4)
	return m
}

func (m *Matrix4) Flip() *Matrix4 {
	buf := make(Matrix, len(m.Out))
	for i := range m.Out {
		buf[i] = m.Out[int(math.Floor(float64(i)/4))+4*(i%4)]
	}
	copy(m.Out, buf)
	return m
}

func (m *Matrix4) Equals(m2 Matrix) bool {
	if len(m.Out) != len(m2) {
		return false
	}
	for i := range m.Out {
		if m.Out[i] != m2[i] {
			return false
		}
	}
	return true
}
