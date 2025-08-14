package compute

type AABB struct {
	Points []Vector3
	Min    Vector3
	Max    Vector3
	Width  float64
	Height float64
	Depth  float64
	Scale  Vector3
}

// Create a new AABB volume from geometry and a transform
func NewAABB(points []Vector3) AABB {
	if len(points) == 0 {
		return AABB{}
	}

	p0 := points[0]

	minX := p0.X
	maxX := p0.X
	minY := p0.Y
	maxY := p0.Y
	minZ := p0.Z
	maxZ := p0.Z

	for _, p := range points {
		if p.X < minX {
			minX = p.X
		} else if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		} else if p.Y > maxY {
			maxY = p.Y
		}
		if p.Z < minZ {
			minZ = p.Z
		} else if p.Z > maxZ {
			maxZ = p.Z
		}
	}

	aabb := AABB{
		Points: []Vector3{
			{X: minX, Y: minY, Z: minZ},
			{X: minX, Y: minY, Z: maxZ},
			{X: maxX, Y: minY, Z: maxZ},
			{X: maxX, Y: minY, Z: minZ},
			{X: minX, Y: maxY, Z: minZ},
			{X: minX, Y: maxY, Z: maxZ},
			{X: maxX, Y: maxY, Z: maxZ},
			{X: maxX, Y: maxY, Z: minZ},
		},
		Width:  maxX - minX,
		Height: maxY - minY,
		Depth:  maxZ - minZ,
		Min:    Vector3{X: minX, Y: minY, Z: minZ},
		Max:    Vector3{X: maxX, Y: maxY, Z: maxZ},
		Scale:  Vector3{X: (maxX - minX) / 2, Y: (maxY - minY) / 2, Z: (maxZ - minZ) / 2},
	}

	return aabb
}

func (aabb AABB) IsEmpty() bool {
	return aabb.Depth == 0 && aabb.Width == 0 && aabb.Height == 0
}

func (aabb AABB) Intersect(t AABB) bool {
	return aabb.Max.X > t.Min.X && aabb.Min.X < t.Max.X &&
		aabb.Max.Y > t.Min.Y && aabb.Min.Y < t.Max.Y &&
		aabb.Max.Z > t.Min.Z && aabb.Min.Z < t.Max.Z
}
