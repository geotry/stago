package scene

import (
	"testing"

	"github.com/geotry/stago/compute"
)

func TestIsVisible(t *testing.T) {
	s := NewScene(SceneOptions{
		Camera: &CameraSettings{
			Near:       0.01,
			Far:        100.0,
			Projection: Perspective,
		},
	})
	c := s.SpawnCamera()

	c.Transform.Position.Z = 0
	c.Transform.Position.Y = 0
	c.Transform.Position.X = 0

	o := NewObject(SceneObjectArgs{})
	obj := s.Spawn(o, SpawnArgs{Position: compute.Point{X: 2, Y: 1, Z: 3}})

	if !c.Camera.IsVisible(obj) {
		t.Errorf("expected point 0, 0, 0 to be visible")
	}

	obj = s.Spawn(o, SpawnArgs{Position: compute.Point{X: 2, Y: 1, Z: 0}})
	if c.Camera.IsVisible(obj) {
		t.Errorf("expected point 0, 0, 0 to not be visible")
	}

	obj = s.Spawn(o, SpawnArgs{Position: compute.Point{X: 102, Y: 1, Z: 3}})
	if c.Camera.IsVisible(obj) {
		t.Errorf("expected point 0, 0, 0 to not be visible")
	}
}
