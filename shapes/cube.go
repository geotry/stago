package shapes

import "github.com/geotry/rass/compute"

func NewCube() Shape {
	geometry := []compute.Point{
		// Front face
		{X: 1, Y: 1, Z: 0},
		{X: 0, Y: 1, Z: 0},
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 1, Z: 0},
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},

		// Left face
		{X: 0, Y: 1, Z: 0},
		{X: 0, Y: 1, Z: 1},
		{X: 0, Y: 0, Z: 1},
		{X: 0, Y: 1, Z: 0},
		{X: 0, Y: 0, Z: 1},
		{X: 0, Y: 0, Z: 0},

		// Right face
		{X: 1, Y: 1, Z: 1},
		{X: 1, Y: 1, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 1, Y: 1, Z: 1},
		{X: 1, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 1},

		// Back face
		{X: 1, Y: 1, Z: 1},
		{X: 0, Y: 1, Z: 1},
		{X: 0, Y: 0, Z: 1},
		{X: 1, Y: 1, Z: 1},
		{X: 0, Y: 0, Z: 1},
		{X: 1, Y: 0, Z: 1},

		// Top face
		{X: 1, Y: 1, Z: 1},
		{X: 0, Y: 1, Z: 1},
		{X: 0, Y: 1, Z: 0},
		{X: 1, Y: 1, Z: 1},
		{X: 0, Y: 1, Z: 0},
		{X: 1, Y: 1, Z: 0},

		// Bottom face
		{X: 1, Y: 0, Z: 1},
		{X: 0, Y: 0, Z: 1},
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 1},
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
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

	// Physic body shape (bounding box)
	physicsShape := []compute.Point{
		{X: 0, Y: 0, Z: 0}, // front bottom-left
		{X: 0, Y: 1, Z: 0}, // front top-left
		{X: 1, Y: 1, Z: 0}, // front top-right
		{X: 1, Y: 0, Z: 0}, // front bottom-right
		{X: 0, Y: 0, Z: 1}, // back bottom-left
		{X: 0, Y: 1, Z: 1}, // back top-left
		{X: 1, Y: 1, Z: 1}, // back top-right
		{X: 1, Y: 0, Z: 1}, // back bottom-right
	}

	return Shape{
		Geometry:     geometry,
		Texture:      textureUv,
		Normals:      normals,
		PhysicsShape: physicsShape,
	}
}
