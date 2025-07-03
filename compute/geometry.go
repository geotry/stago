package compute

import (
	"fmt"
	"math"
)

type Point struct {
	X, Y, Z float64
}

type Rotation = Point
type Size = Point

type Plane struct {
	Min, Max Point
}

func Scale(value float64) Size {
	return Size{X: value, Y: value, Z: value}
}

func (r Plane) Empty() bool {
	return r.Min.X >= r.Max.X || r.Min.Y >= r.Max.Y
}

// Returns true if t and r have a non-empty intersection.
func (r Plane) Overlaps(t Plane) bool {
	return !r.Empty() && !t.Empty() &&
		r.Min.X < t.Max.X && t.Min.X < r.Max.X &&
		r.Min.Y < t.Max.Y && t.Min.Y < r.Max.Y
}

// In reports whether every point in r is in s.
func (r Plane) In(s Plane) bool {
	if r.Empty() {
		return true
	}
	// Note that r.Max is an exclusive bound for r, so that r.In(s)
	// does not require that r.Max.In(s).
	return s.Min.X <= r.Min.X && r.Max.X <= s.Max.X &&
		s.Min.Y <= r.Min.Y && r.Max.Y <= s.Max.Y
}

// Intersect returns the largest rectangle contained by both r and s. If the
// two rectangles do not overlap then the zero rectangle will be returned.
func (r Plane) Intersect(s Plane) Plane {
	if r.Min.X < s.Min.X {
		r.Min.X = s.Min.X
	}
	if r.Min.Y < s.Min.Y {
		r.Min.Y = s.Min.Y
	}
	if r.Max.X > s.Max.X {
		r.Max.X = s.Max.X
	}
	if r.Max.Y > s.Max.Y {
		r.Max.Y = s.Max.Y
	}
	// Letting r0 and s0 be the values of r and s at the time that the method
	// is called, this next line is equivalent to:
	//
	// if max(r0.Min.X, s0.Min.X) >= min(r0.Max.X, s0.Max.X) || likewiseForY { etc }
	if r.Empty() {
		return Plane{}
	}
	return r
}

func (r Plane) Width() float64 {
	return r.Max.X - r.Min.X
}

func (r Plane) Height() float64 {
	return r.Max.Y - r.Min.Y
}

func (p Point) DistanceTo(o Point) float64 {
	dx := math.Pow((p.X - o.X), 2)
	dy := math.Pow((p.Y - o.Y), 2)
	dz := math.Pow((p.Z - o.Z), 2)
	return math.Abs(math.Sqrt(dx + dy + dz))
}

func (p Point) IsZero() bool {
	return p.X == 0 && p.Y == 0 && p.Z == 0
}

func (p Point) Sub(o Point) Point {
	return Point{
		X: p.X - o.X,
		Y: p.Y - o.Y,
		Z: p.Z - o.Z,
	}
}

func (p Point) Mult(o float64) Point {
	return Point{
		X: p.X * o,
		Y: p.Y * o,
		Z: p.Z * o,
	}
}

func (p Point) Scale(o Point) Point {
	return Point{
		X: p.X * o.X,
		Y: p.Y * o.Y,
		Z: p.Z * o.Z,
	}
}

func (p Point) Add(o Point) Point {
	return Point{
		X: p.X + o.X,
		Y: p.Y + o.Y,
		Z: p.Z + o.Z,
	}
}

func (p Point) Opposite() Point {
	return Point{
		X: -p.X,
		Y: -p.Y,
		Z: -p.Z,
	}
}

func (p Point) Inv() Point {
	return Point{
		X: 1 / p.X,
		Y: 1 / p.Y,
		Z: 1 / p.Z,
	}
}

func (a Point) Cross(b Point) Point {
	return Point{
		X: a.Y*b.Z - a.Z*b.Y,
		Y: a.Z*b.X - a.X*b.Z,
		Z: a.X*b.Y - a.Y*b.X,
	}
}

func (a Point) Dot(b Point) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

func (p Point) Rotate(axis Point) Point {
	m := NewMatrix4().Rotate(axis).Out
	return Point{
		X: m[0]*p.X + m[4]*p.Y + m[8]*p.Z,
		Y: m[1]*p.X + m[5]*p.Y + m[9]*p.Z,
		Z: m[2]*p.X + m[6]*p.Y + m[10]*p.Z,
	}
}

func (p Point) MultMatrix(m Matrix) (Point, float64) {
	w := m[3]*p.X + m[7]*p.Y + m[11]*p.Z + m[15]
	return Point{
		X: m[0]*p.X + m[4]*p.Y + m[8]*p.Z + m[12],
		Y: m[1]*p.X + m[5]*p.Y + m[9]*p.Z + m[13],
		Z: m[2]*p.X + m[6]*p.Y + m[10]*p.Z + m[14],
	}, w
}

func (p Point) Normalize() Point {
	f := math.Sqrt(p.X*p.X + p.Y*p.Y + p.Z*p.Z)
	if f == 0 {
		return Point{X: 0, Y: 0, Z: 0}
	}
	return Point{X: p.X / f, Y: p.Y / f, Z: p.Z / f}
}

func (p Point) Equals(o Point) bool {
	return p.String() == o.String()
}

func (p Point) String() string {
	return fmt.Sprintf("(%.2f, %.2f, %.2f)", p.X, p.Y, p.Z)
}

func (r Plane) String() string {
	return fmt.Sprintf("Plane{ (%.2f, %.2f), %.2fx%.2f }", r.Min.X, r.Min.Y, r.Width(), r.Height())
}

func NewQuad() []Point {
	return []Point{
		{X: 1, Y: 1, Z: 0},
		{X: -1, Y: 1, Z: 0},
		{X: -1, Y: -1, Z: 0},
		{X: 1, Y: 1, Z: 0},
		{X: -1, Y: -1, Z: 0},
		{X: 1, Y: -1, Z: 0},
	}
}

func NewQuadUV() []Point {
	return []Point{
		{X: 1, Y: 0},
		{X: 0, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},
	}
}

func NewQuadNormal() []Point {
	return []Point{
		{X: 0, Y: 0, Z: -1},
		{X: 0, Y: 0, Z: -1},
		{X: 0, Y: 0, Z: -1},
		{X: 0, Y: 0, Z: -1},
		{X: 0, Y: 0, Z: -1},
		{X: 0, Y: 0, Z: -1},
	}
}

func NewCube() []Point {
	return []Point{
		// Front top-left
		{X: .5, Y: .5, Z: 0},
		{X: -.5, Y: .5, Z: 0},
		{X: -.5, Y: -.5, Z: 0},

		// Front bot-right
		{X: .5, Y: .5, Z: 0},
		{X: -.5, Y: -.5, Z: 0},
		{X: .5, Y: -.5, Z: 0},

		// Left top-left
		{X: -.5, Y: .5, Z: 0},
		{X: -.5, Y: .5, Z: 1},
		{X: -.5, Y: -.5, Z: 1},

		// Left bot-right
		{X: -.5, Y: .5, Z: 0},
		{X: -.5, Y: -.5, Z: 1},
		{X: -.5, Y: -.5, Z: 0},

		// Right top-left
		{X: .5, Y: .5, Z: 1},
		{X: .5, Y: .5, Z: 0},
		{X: .5, Y: -.5, Z: 0},

		// Right bot-right
		{X: .5, Y: .5, Z: 1},
		{X: .5, Y: -.5, Z: 0},
		{X: .5, Y: -.5, Z: 1},

		// Back top-left
		{X: .5, Y: .5, Z: 1},
		{X: -.5, Y: .5, Z: 1},
		{X: -.5, Y: -.5, Z: 1},

		// Back bot-right
		{X: .5, Y: .5, Z: 1},
		{X: -.5, Y: -.5, Z: 1},
		{X: .5, Y: -.5, Z: 1},

		// Top top-left
		{X: .5, Y: .5, Z: 1},
		{X: -.5, Y: .5, Z: 1},
		{X: -.5, Y: .5, Z: 0},

		// Top bot-right
		{X: .5, Y: .5, Z: 1},
		{X: -.5, Y: .5, Z: 0},
		{X: .5, Y: .5, Z: 0},

		// Bottom top-left
		{X: .5, Y: -.5, Z: 1},
		{X: -.5, Y: -.5, Z: 1},
		{X: -.5, Y: -.5, Z: 0},

		// Bottom bot-right
		{X: .5, Y: -.5, Z: 1},
		{X: -.5, Y: -.5, Z: 0},
		{X: .5, Y: -.5, Z: 0},
	}
}

func NewCubeUV() []Point {
	return []Point{
		{X: 1, Y: 0},
		{X: 0, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},

		{X: 1, Y: 0},
		{X: 0, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},

		{X: 1, Y: 0},
		{X: 0, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},

		{X: 1, Y: 0},
		{X: 0, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},

		{X: 1, Y: 0},
		{X: 0, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},

		{X: 1, Y: 0},
		{X: 0, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},
	}
}

func NewCubeNormal() []Point {
	return []Point{
		// Front
		{Z: -1},
		{Z: -1},
		{Z: -1},
		{Z: -1},
		{Z: -1},
		{Z: -1},

		// Left
		{X: -1},
		{X: -1},
		{X: -1},
		{X: -1},
		{X: -1},
		{X: -1},

		// Right
		{X: 1},
		{X: 1},
		{X: 1},
		{X: 1},
		{X: 1},
		{X: 1},

		// Back
		{Z: 1},
		{Z: 1},
		{Z: 1},
		{Z: 1},
		{Z: 1},
		{Z: 1},

		// Top
		{Y: 1},
		{Y: 1},
		{Y: 1},
		{Y: 1},
		{Y: 1},
		{Y: 1},

		// Bottom
		{Y: -1},
		{Y: -1},
		{Y: -1},
		{Y: -1},
		{Y: -1},
		{Y: -1},
	}
}

func NewPyramid() []Point {
	return []Point{
		// Front
		{X: .5, Y: 1, Z: .5},
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},

		// Left
		{X: .5, Y: 1, Z: .5},
		{X: 0, Y: 0, Z: 1},
		{X: 0, Y: 0, Z: 0},

		// Right
		{X: .5, Y: 1, Z: .5},
		{X: 1, Y: 0, Z: 1},
		{X: 1, Y: 0, Z: 0},

		// Back
		{X: .5, Y: 1, Z: .5},
		{X: 1, Y: 0, Z: 1},
		{X: 0, Y: 0, Z: 1},

		// Bottom
		{X: 1, Y: 0, Z: 1},
		{X: 0, Y: 0, Z: 1},
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 1},
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
	}
}

func NewPyramidUV() []Point {
	return []Point{
		{X: .5, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},

		{X: .5, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},

		{X: .5, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},

		{X: .5, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},

		{X: 1, Y: 0},
		{X: 0, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},
	}
}

func NewPyramidNormal() []Point {
	return []Point{
		// Front
		{Z: -.5, Y: .5},
		{Z: -.5, Y: .5},
		{Z: -.5, Y: .5},

		// Left
		{X: -.5, Y: .5},
		{X: -.5, Y: .5},
		{X: -.5, Y: .5},

		// Right
		{X: .5, Y: .5},
		{X: .5, Y: .5},
		{X: .5, Y: .5},

		// Back
		{Z: .5, Y: .5},
		{Z: .5, Y: .5},
		{Z: .5, Y: .5},

		// Bottom
		{Y: -1},
		{Y: -1},
		{Y: -1},
		{Y: -1},
		{Y: -1},
		{Y: -1},
	}
}
