package compute

import (
	"testing"
)

func TestGJKOnCollidingCubes(t *testing.T) {
	cube := []Vector3{
		{X: -1, Y: -1, Z: -1},
		{X: 1, Y: -1, Z: -1},
		{X: 1, Y: -1, Z: 1},
		{X: -1, Y: -1, Z: 1},
		{X: -1, Y: 1, Z: -1},
		{X: 1, Y: 1, Z: -1},
		{X: 1, Y: 1, Z: 1},
		{X: -1, Y: 1, Z: 1},
	}
	colliderA := make([]Vector3, len(cube))
	colliderB := make([]Vector3, len(cube))

	transform := NewTransform(nil)
	transform.ObjectToWorld(cube, colliderA)

	transform.Position.X = .5
	transform.Position.Z = .5
	transform.Position.Y = .5
	transform.ObjectToWorld(cube, colliderB)

	hit, info := GJK(colliderA, colliderB)

	if hit == false {
		t.Errorf("expected GJK to return true for two cubes")
		return
	}
	if info.Depth < 1.5 || info.Depth >= 1.51 {
		t.Errorf("expected depth to be 1.5, got %v", info.Depth)
		return
	}

	transform.Position.X = 2
	transform.Position.Z = 2
	transform.Position.Y = 2
	transform.ObjectToWorld(cube, colliderB)

	if hit, _ := GJK(colliderA, colliderB); hit == true {
		t.Errorf("expected GJK to return false for two cubes")
		return
	}
}

func BenchmarkGJKOnCollidingCubes(b *testing.B) {
	cube := []Vector3{
		{X: -1, Y: -1, Z: -1},
		{X: 1, Y: -1, Z: -1},
		{X: 1, Y: -1, Z: 1},
		{X: -1, Y: -1, Z: 1},
		{X: -1, Y: 1, Z: -1},
		{X: 1, Y: 1, Z: -1},
		{X: 1, Y: 1, Z: 1},
		{X: -1, Y: 1, Z: 1},
	}
	colliderA := make([]Vector3, len(cube))
	colliderB := make([]Vector3, len(cube))

	transform := NewTransform(nil)
	transform.ObjectToWorld(cube, colliderA)

	transform.Position.X = .5
	transform.Position.Z = .5
	transform.Position.Y = .5
	transform.ObjectToWorld(cube, colliderB)

	for b.Loop() {
		GJK(colliderA, colliderB)
	}
}
