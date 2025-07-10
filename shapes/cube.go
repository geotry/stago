package shapes

import "github.com/geotry/rass/compute"

func NewCube() Shape {
	geometry := []compute.Point{
		// Front face
		{X: .5, Y: .5, Z: 0},
		{X: -.5, Y: .5, Z: 0},
		{X: -.5, Y: -.5, Z: 0},
		{X: .5, Y: .5, Z: 0},
		{X: -.5, Y: -.5, Z: 0},
		{X: .5, Y: -.5, Z: 0},

		// Left face
		{X: -.5, Y: .5, Z: 0},
		{X: -.5, Y: .5, Z: 1},
		{X: -.5, Y: -.5, Z: 1},
		{X: -.5, Y: .5, Z: 0},
		{X: -.5, Y: -.5, Z: 1},
		{X: -.5, Y: -.5, Z: 0},

		// Right face
		{X: .5, Y: .5, Z: 1},
		{X: .5, Y: .5, Z: 0},
		{X: .5, Y: -.5, Z: 0},
		{X: .5, Y: .5, Z: 1},
		{X: .5, Y: -.5, Z: 0},
		{X: .5, Y: -.5, Z: 1},

		// Back face
		{X: .5, Y: .5, Z: 1},
		{X: -.5, Y: .5, Z: 1},
		{X: -.5, Y: -.5, Z: 1},
		{X: .5, Y: .5, Z: 1},
		{X: -.5, Y: -.5, Z: 1},
		{X: .5, Y: -.5, Z: 1},

		// Top face
		{X: .5, Y: .5, Z: 1},
		{X: -.5, Y: .5, Z: 1},
		{X: -.5, Y: .5, Z: 0},
		{X: .5, Y: .5, Z: 1},
		{X: -.5, Y: .5, Z: 0},
		{X: .5, Y: .5, Z: 0},

		// Bottom face
		{X: .5, Y: -.5, Z: 1},
		{X: -.5, Y: -.5, Z: 1},
		{X: -.5, Y: -.5, Z: 0},
		{X: .5, Y: -.5, Z: 1},
		{X: -.5, Y: -.5, Z: 0},
		{X: .5, Y: -.5, Z: 0},
	}

	textureUv := []compute.Point{
		{X: 1, Y: 0},
		{X: 0, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},

		{X: 1, Y: 0},
		{X: 0, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},

		{X: 1, Y: 0},
		{X: 0, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},

		{X: 1, Y: 0},
		{X: 0, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},

		{X: 1, Y: 0},
		{X: 0, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 0},
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
		{Z: -1},
		{Z: -1},
		{Z: -1},
		{Z: -1},
		{Z: -1},
		{Z: -1},

		// Left
		{X: -1},
		{X: -1},
		{X: -1},
		{X: -1},
		{X: -1},
		{X: -1},

		// Right
		{X: 1},
		{X: 1},
		{X: 1},
		{X: 1},
		{X: 1},
		{X: 1},

		// Back
		{Z: 1},
		{Z: 1},
		{Z: 1},
		{Z: 1},
		{Z: 1},
		{Z: 1},

		// Top
		{Y: 1},
		{Y: 1},
		{Y: 1},
		{Y: 1},
		{Y: 1},
		{Y: 1},

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
