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

type Rectangle2D struct {
	Min, Max Point
}

func Scale(value float64) Size {
	return Size{X: value, Y: value, Z: value}
}

func (r Rectangle2D) Empty() bool {
	return r.Min.X >= r.Max.X || r.Min.Y >= r.Max.Y
}

// Returns true if t and r have a non-empty intersection.
func (r Rectangle2D) Overlaps(t Rectangle2D) bool {
	return !r.Empty() && !t.Empty() &&
		r.Min.X < t.Max.X && t.Min.X < r.Max.X &&
		r.Min.Y < t.Max.Y && t.Min.Y < r.Max.Y
}

// In reports whether every point in r is in s.
func (r Rectangle2D) In(s Rectangle2D) bool {
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
func (r Rectangle2D) Intersect(s Rectangle2D) Rectangle2D {
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
		return Rectangle2D{}
	}
	return r
}

func (r Rectangle2D) Width() float64 {
	return r.Max.X - r.Min.X
}

func (r Rectangle2D) Height() float64 {
	return r.Max.Y - r.Min.Y
}

func (p Point) DistanceTo(o Point) float64 {
	dx := math.Pow((p.X - o.X), 2)
	dy := math.Pow((p.Y - o.Y), 2)
	return math.Abs(math.Sqrt(dx + dy))
}

func (p Point) Sub(o Point) Point {
	return Point{
		X: p.X - o.X,
		Y: p.Y - o.Y,
		Z: p.Z - o.Z,
	}
}

func (r Rectangle2D) String() string {
	return fmt.Sprintf("Rectangle2D{ (%.2f, %.2f), %.2fx%.2f }", r.Min.X, r.Min.Y, r.Width(), r.Height())
}
