package shapes

import "github.com/geotry/rass/compute"

func NewPyramid() Shape {
	geometry := []compute.Point{
		// Front
		{X: .5, Y: 1, Z: .5},
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},

		// Left
		{X: .5, Y: 1, Z: .5},
		{X: 0, Y: 0, Z: 1},
		{X: 0, Y: 0, Z: 0},

		// Right
		{X: 1, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 1},
		{X: .5, Y: 1, Z: .5},

		// Back
		{X: .5, Y: 1, Z: .5},
		{X: 1, Y: 0, Z: 1},
		{X: 0, Y: 0, Z: 1},

		// Bottom
		{X: 1, Y: 0, Z: 1},
		{X: 0, Y: 0, Z: 1},
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 1},
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
	}

	textureUv := []compute.Point{
		{X: .5, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},

		{X: .5, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},

		{X: 1, Y: 1},
		{X: 0, Y: 1},
		{X: .5, Y: 0},

		{X: .5, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},

		{X: 1, Y: 0},
		{X: 0, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},
	}

	normals := []compute.Point{
		// Front
		{Z: -.5, Y: .5},
		{Z: -.5, Y: .5},
		{Z: -.5, Y: .5},

		// Left
		{X: -.5, Y: .5},
		{X: -.5, Y: .5},
		{X: -.5, Y: .5},

		// Right
		{X: .5, Y: .5},
		{X: .5, Y: .5},
		{X: .5, Y: .5},

		// Back
		{Z: .5, Y: .5},
		{Z: .5, Y: .5},
		{Z: .5, Y: .5},

		// Bottom
		{Y: -1},
		{Y: -1},
		{Y: -1},
		{Y: -1},
		{Y: -1},
		{Y: -1},
	}

	return Shape{
		Geometry: geometry,
		Texture:  textureUv,
		Normals:  normals,
	}
}
