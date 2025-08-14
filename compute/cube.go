package compute

func NewCube() Shape {
	geometry := []Point{

		// Front face
		{X: 1, Y: 1, Z: -1},
		{X: -1, Y: 1, Z: -1},
		{X: -1, Y: -1, Z: -1},
		{X: 1, Y: 1, Z: -1},
		{X: -1, Y: -1, Z: -1},
		{X: 1, Y: -1, Z: -1},

		// Left face
		{X: -1, Y: 1, Z: -1},
		{X: -1, Y: 1, Z: 1},
		{X: -1, Y: -1, Z: 1},
		{X: -1, Y: 1, Z: -1},
		{X: -1, Y: -1, Z: 1},
		{X: -1, Y: -1, Z: -1},

		// Right face
		{X: 1, Y: 1, Z: 1},
		{X: 1, Y: 1, Z: -1},
		{X: 1, Y: -1, Z: -1},
		{X: 1, Y: 1, Z: 1},
		{X: 1, Y: -1, Z: -1},
		{X: 1, Y: -1, Z: 1},

		// Back face
		{X: -1, Y: -1, Z: 1},
		{X: -1, Y: 1, Z: 1},
		{X: 1, Y: 1, Z: 1},
		{X: 1, Y: -1, Z: 1},
		{X: -1, Y: -1, Z: 1},
		{X: 1, Y: 1, Z: 1},

		// Top face
		{X: 1, Y: 1, Z: 1},
		{X: -1, Y: 1, Z: 1},
		{X: -1, Y: 1, Z: -1},
		{X: 1, Y: 1, Z: 1},
		{X: -1, Y: 1, Z: -1},
		{X: 1, Y: 1, Z: -1},

		// Bottom face
		{X: -1, Y: -1, Z: -1},
		{X: -1, Y: -1, Z: 1},
		{X: 1, Y: -1, Z: 1},
		{X: 1, Y: -1, Z: -1},
		{X: -1, Y: -1, Z: -1},
		{X: 1, Y: -1, Z: 1},
	}

	textureUv := []Point{
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

		{X: 0, Y: 1},
		{X: 0, Y: 0},
		{X: 1, Y: 0},
		{X: 1, Y: 1},
		{X: 0, Y: 1},
		{X: 1, Y: 0},

		{X: 1, Y: 0},
		{X: 0, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},

		{X: 0, Y: 1},
		{X: 0, Y: 0},
		{X: 1, Y: 0},
		{X: 1, Y: 1},
		{X: 0, Y: 1},
		{X: 1, Y: 0},
	}

	normals := []Point{
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

	// Convex bounding hull
	collider := []Vector3{
		{X: -1, Y: -1, Z: -1},
		{X: 1, Y: -1, Z: -1},
		{X: 1, Y: -1, Z: 1},
		{X: -1, Y: -1, Z: 1},
		{X: -1, Y: 1, Z: -1},
		{X: 1, Y: 1, Z: -1},
		{X: 1, Y: 1, Z: 1},
		{X: -1, Y: 1, Z: 1},
	}

	return Shape{
		Geometry: geometry,
		Texture:  textureUv,
		Normals:  normals,
		Collider: collider,
	}
}
