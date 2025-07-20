package scene

import (
	"fmt"
	"math"
	"time"

	"github.com/geotry/rass/compute"
)

type Node struct {
	Id uint32

	Object *SceneObject
	Scene  *Scene
	Parent *Node
	Hidden bool

	SpawnTime time.Time

	// This object is attached to a Camera
	Camera *Camera
	// This object is a light source
	Light Light

	Data map[string]any

	Position compute.Point
	Scale    compute.Size
	Rotation compute.Point

	model *compute.Matrix4
}

func (n *Node) Destroy() {
	n.Scene.scheduleOldObjectInstance(n)
}

func (n *Node) Size() compute.Size {
	return n.Object.Size
}

func (n *Node) WorldPosition() compute.Point {
	pos := n.Position
	if n.Parent != nil {
		p := n.Parent.WorldPosition()
		pos.X += p.X
		pos.Y += p.Y
		pos.Z += p.Z
	}
	return pos
}

func (c *Node) Move(x, y, z float64) {
	c.Position.X += x
	c.Position.Y += y
	c.Position.Z += z
}

func (c *Node) MoveAt(pos compute.Point) {
	c.Position.X = pos.X
	c.Position.Y = pos.Y
	c.Position.Z = pos.Z
}

// Move objet toward a destination
func (c *Node) MoveToward(pt compute.Point, s float64) {
	d := c.Position.DistanceTo(pt)
	ds := math.Min(d, s)

	if ds >= 0 {
		dx := math.Abs(pt.X - c.Position.X)
		dy := math.Abs(pt.Y - c.Position.Y)
		dz := math.Abs(pt.Z - c.Position.Z)

		dvx := (dx / (dx + dy + dz)) * s
		dvy := (dy / (dx + dy + dz)) * s
		dvz := (dz / (dx + dy + dz)) * s

		if c.Position.X < pt.X {
			c.Position.X += dvx
		} else {
			c.Position.X -= dvx
		}
		if c.Position.Y < pt.Y {
			c.Position.Y += dvy
		} else {
			c.Position.Y -= dvy
		}
		if c.Position.Z < pt.Z {
			c.Position.Z += dvz
		} else {
			c.Position.Z -= dvz
		}
	}
}

func (n *Node) WorldRotation() compute.Point {
	rot := n.Rotation
	if n.Parent != nil {
		p := n.Parent.WorldRotation()
		rot.X += p.X
		rot.Y += p.Y
		rot.Z += p.Z
	}
	return rot
}

func (c *Node) Rotate(rotation compute.Point) {
	c.Rotation.X += rotation.X
	c.Rotation.Y += rotation.Y
	c.Rotation.Z += rotation.Z
}

func (c *Node) RotateZ(value float64) {
	c.Rotation.Z += value
}

func (c *Node) RotateX(value float64) {
	c.Rotation.X += value
}

func (c *Node) RotateY(value float64) {
	c.Rotation.Y += value
}

func (n *Node) Resize(x, y, z float64) {
	n.Scale.X += x
	n.Scale.Y += y
	n.Scale.Z += z

	if n.Camera != nil {
		n.Camera.updateProjectionMatrix()
		n.Camera.normalizeLookAt()
	}
}

func (n *Node) ScaleAt(x, y float64) {
	n.Scale.X = x
	n.Scale.Y = y

	if n.Camera != nil {
		n.Camera.updateProjectionMatrix()
		n.Camera.normalizeLookAt()
	}
}

func (n *Node) IsDescendant(p *Node) bool {
	if n == p || n.Parent == p {
		return true
	}
	if n.Parent == nil {
		return false
	}
	return n.Parent.IsDescendant(p)
}

func (n *Node) Points() []compute.Point {
	size := n.Size().Scale(n.Scale)

	x1 := compute.Point{X: n.Position.X + size.X, Y: n.Position.Y + size.Y, Z: n.Position.Z}.Rotate(n.Rotation)
	x2 := compute.Point{X: n.Position.X - size.X, Y: n.Position.Y + size.Y, Z: n.Position.Z}.Rotate(n.Rotation)
	x3 := compute.Point{X: n.Position.X + size.X, Y: n.Position.Y - size.Y, Z: n.Position.Z}.Rotate(n.Rotation)
	x4 := compute.Point{X: n.Position.X - size.X, Y: n.Position.Y - size.Y, Z: n.Position.Z}.Rotate(n.Rotation)

	return []compute.Point{x1, x2, x3, x4}
}

func (n *Node) ModelMatrix() compute.Matrix {

	var xRatio float64
	if n.Object.Size.X > n.Object.Size.Y {
		xRatio = n.Object.Size.Y / n.Object.Size.X
	} else {
		xRatio = n.Object.Size.X / n.Object.Size.Y
	}

	n.model.Reset()
	n.model.Scale(compute.Size{X: n.Scale.X * xRatio, Y: n.Scale.Y, Z: n.Scale.Z})
	n.model.Rotate(n.WorldRotation())
	n.model.Translate(n.WorldPosition())

	// Note: ScreenSpace should have a dedicated projection matrix
	// that is fixed (defined on client?) with space contained in
	// (0, 0), (1, 1) and looking forward (Z:-1), in orthographic view.

	// ScreenSpace objects should be bound to the current active session
	// meaning that server should not stream other session objects. This
	// could be defined in a separate field to allow shareable UI elements
	// {
	//    Scope: Global / Local
	// }
	return n.model.Out
}

func (n *Node) String() string {
	switch {
	case n.Camera != nil:
		return fmt.Sprintf("camera id=%d object_id=%v pos=%v", n.Id, n.Object.Id, n.Position)
	case n.Light != nil:
		return fmt.Sprintf("light id=%d object_id=%v pos=%v", n.Id, n.Object.Id, n.Position)
	default:
		return fmt.Sprintf("object id=%d object_id=%v pos=%v", n.Id, n.Object.Id, n.Position)
	}
}
