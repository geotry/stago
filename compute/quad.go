package compute

func NewQuad() Shape {
	geometry := []Point{
		{X: 1, Y: 1},
		{X: -1, Y: 1},
		{X: -1, Y: -1},
		{X: 1, Y: 1},
		{X: -1, Y: -1},
		{X: 1, Y: -1},
	}

	textureUv := []Point{
		{X: 1, Y: 0},
		{X: 0, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},
	}

	normals := []Point{
		{X: 0, Y: 0, Z: -1},
		{X: 0, Y: 0, Z: -1},
		{X: 0, Y: 0, Z: -1},
		{X: 0, Y: 0, Z: -1},
		{X: 0, Y: 0, Z: -1},
		{X: 0, Y: 0, Z: -1},
	}

	// Convex bounding hull
	collider := []Vector3{
		{X: 1, Y: 1, Z: 0},
		{X: -1, Y: 1, Z: 0},
		{X: -1, Y: -1, Z: 0},
		{X: 1, Y: -1, Z: 0},
		{X: 1, Y: 1, Z: 1},
		{X: -1, Y: 1, Z: 1},
		{X: -1, Y: -1, Z: 1},
		{X: 1, Y: -1, Z: 1},
	}

	return Shape{
		Geometry: geometry,
		Texture:  textureUv,
		Normals:  normals,
		Collider: collider,
	}
}
