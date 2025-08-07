package shapes

import "github.com/geotry/rass/compute"

func NewQuad() Shape {
	geometry := []compute.Point{
		{X: 1, Y: 1},
		{X: -1, Y: 1},
		{X: -1, Y: -1},
		{X: 1, Y: 1},
		{X: -1, Y: -1},
		{X: 1, Y: -1},
	}

	textureUv := []compute.Point{
		{X: 1, Y: 0},
		{X: 0, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},
	}

	normals := []compute.Point{
		{X: 0, Y: 0, Z: -1},
		{X: 0, Y: 0, Z: -1},
		{X: 0, Y: 0, Z: -1},
		{X: 0, Y: 0, Z: -1},
		{X: 0, Y: 0, Z: -1},
		{X: 0, Y: 0, Z: -1},
	}

	return Shape{
		Geometry: geometry,
		Texture:  textureUv,
		Normals:  normals,
	}
}
