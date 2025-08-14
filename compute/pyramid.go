package compute

func NewPyramid() Shape {
	geometry := []Point{
		// Front
		{X: 0, Y: 1, Z: 0},
		{X: -1, Y: -1, Z: -1},
		{X: 1, Y: -1, Z: -1},

		// Left
		{X: 0, Y: 1, Z: 0},
		{X: -1, Y: -1, Z: 1},
		{X: -1, Y: -1, Z: -1},

		// Right
		{X: 1, Y: -1, Z: -1},
		{X: 1, Y: -1, Z: 1},
		{X: 0, Y: 1, Z: 0},

		// Back
		{X: 0, Y: 1, Z: 0},
		{X: 1, Y: -1, Z: 1},
		{X: -1, Y: -1, Z: 1},

		// Bottom
		{X: 1, Y: -1, Z: 1},
		{X: -1, Y: -1, Z: 1},
		{X: -1, Y: -1, Z: -1},
		{X: 1, Y: -1, Z: 1},
		{X: -1, Y: -1, Z: -1},
		{X: 1, Y: -1, Z: -1},
	}

	textureUv := []Point{
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

	normals := []Point{
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

	// Convex bounding hull
	collider := []Vector3{
		{X: -1, Y: -1, Z: -1},
		{X: 1, Y: -1, Z: -1},
		{X: 1, Y: -1, Z: 1},
		{X: -1, Y: -1, Z: 1},
		{X: 0, Y: 1, Z: 0},
	}

	return Shape{
		Geometry: geometry,
		Texture:  textureUv,
		Normals:  normals,
		Collider: collider,
	}
}
