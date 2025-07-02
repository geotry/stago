package scene

import (
	"testing"

	"github.com/geotry/rass/compute"
)

func TestNormalizeLookAt(t *testing.T) {

	c := &Camera{
		lookAt: compute.Point{},
	}

	c.SetLookAt(0, 0)

	if !c.LookAt().Equals(compute.Point{X: 0, Y: 0, Z: 1}) {
		t.Errorf("lookat is incorrect: %v", c.LookAt())
	}

	c.SetLookAt(1, 0)

	if !c.LookAt().Equals(compute.Point{X: 1, Y: 0, Z: 0}) {
		t.Errorf("lookat is incorrect: %v", c.LookAt())
	}

	c.SetLookAt(1.2, 0)

	if !c.LookAt().Equals(compute.Point{X: 0.8, Y: 0, Z: -0.2}) {
		t.Errorf("lookat is incorrect: %v", c.LookAt())
	}

	// c.SetLookAt(1.2, 1.3)

	// if !c.LookAt().Equals(compute.Point{X: 0.8, Y: 0, Z: -0.2}) {
	// 	t.Errorf("lookat is incorrect: %v", c.LookAt())
	// }

	c.SetLookAt(2, 0)

	if !c.LookAt().Equals(compute.Point{X: 0, Y: 0, Z: -1}) {
		t.Errorf("lookat is incorrect: %v", c.LookAt())
	}

	c.SetLookAt(2.4, 0)

	if !c.LookAt().Equals(compute.Point{X: -0.4, Y: 0, Z: -0.6}) {
		t.Errorf("lookat is incorrect: %v", c.LookAt())
	}

	c.SetLookAt(3, 0)

	if !c.LookAt().Equals(compute.Point{X: -1, Y: 0, Z: 0}) {
		t.Errorf("lookat is incorrect: %v", c.LookAt())
	}

	c.SetLookAt(3.2, 0)

	if !c.LookAt().Equals(compute.Point{X: -0.8, Y: 0, Z: 0.2}) {
		t.Errorf("lookat is incorrect: %v", c.LookAt())
	}

	c.SetLookAt(4, 0)

	if !c.LookAt().Equals(compute.Point{X: 0, Y: 0, Z: 1}) {
		t.Errorf("lookat is incorrect: %v", c.LookAt())
	}

	c.SetLookAt(4.2, 0)

	if !c.LookAt().Equals(compute.Point{X: 0.2, Y: 0, Z: 0.8}) {
		t.Errorf("lookat is incorrect: %v", c.LookAt())
	}

	c.SetLookAt(-3.2, 0)

	if !c.LookAt().Equals(compute.Point{X: 0.8, Y: 0, Z: 0.2}) {
		t.Errorf("lookat is incorrect: %v", c.LookAt())
	}
}

func TestIsVisible(t *testing.T) {
	s := NewScene(SceneOptions{})
	c := NewCamera(s)

	c.Position.Z = 0
	c.Position.Y = 0
	c.Position.X = 0

	c.SetSize(1, 1)
	c.Near = 0.01
	c.Far = 100

	o := NewObject(SceneObjectArgs{})
	obj := s.Spawn(o, SpawnArgs{Position: compute.Point{X: 2, Y: 1, Z: 3}})

	if !c.IsVisible(obj) {
		t.Errorf("expected point 0, 0, 0 to be visible")
	}

	obj = s.Spawn(o, SpawnArgs{Position: compute.Point{X: 2, Y: 1, Z: 0}})
	if c.IsVisible(obj) {
		t.Errorf("expected point 0, 0, 0 to not be visible")
	}

	obj = s.Spawn(o, SpawnArgs{Position: compute.Point{X: 102, Y: 1, Z: 3}})
	if c.IsVisible(obj) {
		t.Errorf("expected point 0, 0, 0 to not be visible")
	}
}
