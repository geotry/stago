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
