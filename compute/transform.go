package compute

type Transform struct {
	Position      Vector3
	Rotation      Quaternion
	RotationPivot Vector3
	Scale         Vector3

	Parent *Transform

	matrix *Matrix4
}

func NewTransform(parent *Transform) *Transform {
	return &Transform{
		Position:      Vector3{},
		Rotation:      NewQuaternion(Vector3{}),
		RotationPivot: Vector3{X: 0, Y: 0, Z: 0},
		Scale:         Vector3{X: 1, Y: 1, Z: 1},
		Parent:        parent,
		matrix:        NewMatrix4(),
	}
}

// Return the Transform model matrix (object to world space)
func (t *Transform) Model() Matrix {
	t.matrix.Reset()
	t.matrix.Scale(t.Scale)
	// Move to center of rotation before applying rotation
	t.matrix.Translate(t.RotationPivot.Opposite())
	t.matrix.Rotate(t.WorldRotation())
	t.matrix.Translate(t.RotationPivot)
	t.matrix.Translate(t.WorldPosition())

	return t.matrix.Out
}

func (t *Transform) WorldRotation() Quaternion {
	r := t.Rotation
	if t.Parent != nil {
		r = r.Mult(t.Parent.WorldRotation())
	}
	return r
}

func (t *Transform) WorldPosition() Point {
	pos := t.Position
	if t.Parent != nil {
		p := t.Parent.WorldPosition()
		pos.X += p.X
		pos.Y += p.Y
		pos.Z += p.Z
	}
	return pos
}

// Transform points from object space to world space
func (t Transform) ObjectToWorld(src []Vector3, dst []Vector3) {
	model := t.Model()
	for i, p := range src {
		dst[i], _ = p.MultMatrix(model)
	}
}
