package shapes

import "github.com/geotry/rass/compute"

type Shape struct {
	Geometry []compute.Point
	Texture  []compute.Point
	Normals  []compute.Point
}
