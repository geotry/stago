package scene

import (
	"github.com/geotry/rass/compute"
)

type Physics struct {
	// The original mass of the object
	Mass float64
	// The layer in which the object can collide
	CollisionLayer int
}

type Force struct {
	// The intensity of force being applied, in m/s²
	Intensity float64
	// The normalized direction of the force
	Direction compute.Vector3
}
