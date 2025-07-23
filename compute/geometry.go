package compute

import (
	"fmt"
	"math"
)

type Point struct {
	X, Y, Z float64
}

type Point2d struct {
	X, Y float64
}

type Rotation = Point
type Size = Point
type Vector3 = Point
type Vector2 = Point2d
type Vector4 struct {
	X, Y, Z, W float64
}
type Quaternion = Vector4

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

func (p Point) Length() float64 {
	return math.Sqrt(math.Pow(p.X, 2) + math.Pow(p.Y, 2) + math.Pow(p.Z, 2))
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

func (p Point) Div(o float64) Point {
	return Point{
		X: p.X / o,
		Y: p.Y / o,
		Z: p.Z / o,
	}
}

func (p Point) Scale(o Point) Point {
	return Point{
		X: p.X * o.X,
		Y: p.Y * o.Y,
		Z: p.Z * o.Z,
	}
}

func (p Point) Pow(v float64) Point {
	return Point{
		X: math.Pow(p.X, v),
		Y: math.Pow(p.Y, v),
		Z: math.Pow(p.Z, v),
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

func (p Point) AngleTo(o Point) float64 {
	return math.Atan2(p.Cross(o).Length(), p.Dot(o))
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

func (v Vector4) Length() float64 {
	return math.Sqrt(math.Pow(v.X, 2) + math.Pow(v.Y, 2) + math.Pow(v.Z, 2) + math.Pow(v.W, 2))
}

func (v Vector4) Normalize() Vector4 {
	m := v.Length()
	if m == 0 {
		return Vector4{}
	}
	return Vector4{X: v.X / m, Y: v.Y / m, Z: v.Z / m, W: v.W / m}
}

func (v Vector4) Scale(f float64) Vector4 {
	return Vector4{X: v.X * f, Y: v.Y * f, Z: v.Z * f, W: v.W * f}
}

func NewQuaternion() Quaternion {
	return Quaternion{X: 0, Y: 0, Z: 0, W: 1}
}

func NewQuaternionFromEuler(e Vector3) Quaternion {
	cx := math.Cos(e.X / 2)
	sx := math.Sin(e.X / 2)
	cy := math.Cos(e.Y / 2)
	sy := math.Sin(e.Y / 2)
	cz := math.Cos(e.Z / 2)
	sz := math.Sin(e.Z / 2)
	return Quaternion{
		X: sx*cy*cz - cx*sy*sz,
		Y: cx*sy*cz + sx*cy*sz,
		Z: cx*cy*sz - sx*sy*cz,
		W: cx*cy*cz + sx*sy*sz,
	}
}

func NewQuaternionFromAngle(axis Vector3, angle float64) Quaternion {
	s := math.Sin(angle / 2)
	return Quaternion{
		axis.X * s,
		axis.Y * s,
		axis.Z * s,
		math.Cos(angle / 2),
	}
}

func (q Quaternion) IsPure() bool {
	return q.W == 0
}

func (q Quaternion) IsUnit() bool {
	return q.Length() == 1
}

func (q Quaternion) IsIdentity() bool {
	return q.X == 0 && q.Y == 0 && q.Z == 0 && math.Abs(q.W) == 1
}

func (q Quaternion) Inverse() Quaternion {
	m := q.Length()
	if m == 0 {
		return Quaternion{}
	}
	m *= m
	return Quaternion{-q.X / m, -q.Y / m, -q.Z / m, q.W / m}
}

// Apply a quaternion rotation a after b
func (b Quaternion) Mult(a Quaternion) Quaternion {
	return Quaternion{
		X: a.W*b.X + a.X*b.W + a.Y*b.Z - a.Z*b.Y,
		Y: a.W*b.Y - a.X*b.Z + a.Y*b.W + a.Z*b.X,
		Z: a.W*b.Z + a.X*b.Y - a.Y*b.X + a.Z*b.W,
		W: a.W*b.W - a.X*b.X - a.Y*b.Y - a.Z*b.Z,
	}
}
