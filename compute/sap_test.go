package compute

import (
	"math/rand"
	"testing"
)

type testAABB struct {
	aabb AABB
}

func (t *testAABB) AABB() AABB {
	return t.aabb
}

var cube = []Vector3{
	{X: -1, Y: -1, Z: -1},
	{X: 1, Y: -1, Z: -1},
	{X: 1, Y: -1, Z: 1},
	{X: -1, Y: -1, Z: 1},
	{X: -1, Y: 1, Z: -1},
	{X: 1, Y: 1, Z: -1},
	{X: 1, Y: 1, Z: 1},
	{X: -1, Y: 1, Z: 1},
}

func TestSweepAndPruneOverlap(t *testing.T) {
	transform := NewTransform(nil)
	bounding := make([]Vector3, len(cube))
	copy(bounding, cube)

	objects := []*testAABB{}
	objects = append(objects, &testAABB{aabb: NewAABB(bounding)})

	transform.Position.X += 0.5
	transform.Position.Z += 0.5
	transform.Position.Y += 0.5
	transform.ObjectToWorld(cube, bounding)
	objects = append(objects, &testAABB{aabb: NewAABB(bounding)})

	pairs := SweepAndPrune(objects)

	if len(pairs) != 1 {
		t.Errorf("expected to have one pair, got %v", len(pairs))
		return
	}
}

func TestSweepAndPruneNoOverlap(t *testing.T) {
	transform := NewTransform(nil)
	bounding := make([]Vector3, len(cube))
	copy(bounding, cube)

	objects := []*testAABB{}
	objects = append(objects, &testAABB{aabb: NewAABB(bounding)})

	transform.Position.X += 2
	transform.Position.Z += 2
	transform.Position.Y += 2
	transform.ObjectToWorld(cube, bounding)
	objects = append(objects, &testAABB{aabb: NewAABB(bounding)})

	pairs := SweepAndPrune(objects)

	if len(pairs) != 0 {
		t.Errorf("expected to have 0 pair, got %v", len(pairs))
		return
	}
}

func BenchmarkSweepAndPruneCube1000(b *testing.B) {
	objects := []*testAABB{}
	transform := NewTransform(nil)
	bounding := make([]Vector3, len(cube))
	copy(bounding, cube)
	for range 1000 {
		transform.Position.X = -10 + rand.Float64()*20
		transform.Position.Y = -10 + rand.Float64()*20
		transform.Position.Z = -10 + rand.Float64()*20
		transform.Rotation.X = -1 + rand.Float64()*2
		transform.Rotation.Y = -1 + rand.Float64()*2
		transform.Rotation.Z = -1 + rand.Float64()*2
		transform.Scale.X = .3 + rand.Float64()*5
		transform.Scale.Y = .3 + rand.Float64()*5
		transform.Scale.Z = .3 + rand.Float64()*5
		transform.ObjectToWorld(cube, bounding)
		objects = append(objects, &testAABB{aabb: NewAABB(bounding)})
	}

	for b.Loop() {
		SweepAndPrune(objects)
	}
}
