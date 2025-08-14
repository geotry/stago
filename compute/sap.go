package compute

import "slices"

type Object interface {
	AABB() AABB
}

type Pair[T Object] struct {
	A, B T
}

// Returns pairs of objects having their AABB volume overlapping
// (not optimized)
func SweepAndPrune[T Object](objects []T) []Pair[T] {
	candidatesX := make([]Pair[T], 0)
	pairs := make([]Pair[T], 0)
	var set []T

	sorted := make([]T, 0)

	for _, o := range objects {
		if !o.AABB().IsEmpty() {
			sorted = append(sorted, o)
		}
	}

	// Sort on x axis
	slices.SortFunc(sorted, func(a T, b T) int {
		if a.AABB().Min.X < b.AABB().Min.X {
			return -1
		}
		if a.AABB().Min.X > b.AABB().Min.X {
			return 1
		}
		return 0
	})

	for _, ox := range sorted {
		set = slices.DeleteFunc(set, func(e T) bool { return e.AABB().Max.X <= ox.AABB().Min.X })
		for _, xo := range set {
			candidatesX = append(candidatesX, Pair[T]{A: ox, B: xo})
		}
		set = append(set, ox)
	}
	set = nil

	// Filter on other axis if objects length is enough
	// Sort on y axis
	// slices.SortFunc(sorted, func(a T, b T) int {
	// 	if a.AABB().Min.Y < b.AABB().Min.Y {
	// 		return -1
	// 	}
	// 	if a.AABB().Min.Y > b.AABB().Min.Y {
	// 		return 1
	// 	}
	// 	return 0
	// })
	// candidatesXY := make([]Pair[T], 0)
	// xyMap := make(map[*T]*T)
	// for _, oy := range sorted {
	// 	set = slices.DeleteFunc(set, func(e T) bool { return e.AABB().Max.Y <= oy.AABB().Min.Y })
	// 	for _, yo := range set {
	// 		// Filter candidates
	// 		// check oy and yo exists in candidates
	// 		if xMap[&oy] == &yo {
	// 			xyMap[&oy] = &yo
	// 			xyMap[&yo] = &oy
	// 		}
	// 	}
	// 	set = append(set, oy)
	// }
	// set = nil

	// Reduce candidates to keep only those colliding on every axis
	for _, pair := range candidatesX {
		A := pair.A
		B := pair.B
		if A.AABB().Intersect(B.AABB()) {
			pairs = append(pairs, Pair[T]{A: A, B: B})
		}
	}

	return pairs
}
