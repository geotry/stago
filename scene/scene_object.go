package scene

import (
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"github.com/geotry/rass/compute"
	"github.com/geotry/rass/pb"
	"github.com/geotry/rass/rendering"
	"github.com/geotry/rass/shapes"
)

type Texture struct {
	Size struct{ X, Y int }
	Data []uint8
}

type SceneSpace uint8

const (
	WorldSpace SceneSpace = iota
	ScreenSpace
)

type SceneObject struct {
	Id int32

	Material   *rendering.Material
	Size       compute.Size
	Shape      shapes.Shape
	Space      SceneSpace
	Controller SceneObjectController
}

type SceneObjectController struct {
	Init   func(self *SceneObjectInstance)
	Update func(self *SceneObjectInstance, deltaTime time.Duration)
	Input  func(self *SceneObjectInstance, event *pb.InputEvent)
}

type SceneObjectArgs struct {
	Material  *rendering.Material
	Shape     shapes.Shape
	UIElement bool
	Init      func(self *SceneObjectInstance)
	Update    func(self *SceneObjectInstance, deltaTime time.Duration)
	Input     func(self *SceneObjectInstance, event *pb.InputEvent)
}

type SceneObjectInstance struct {
	Id uint32

	SceneObject *SceneObject
	Scene       *Scene
	Parent      *SceneObjectInstance
	Hidden      bool

	SpawnTime time.Time

	// This object is attached to a Camera
	Camera *Camera

	Data map[string]any

	Position compute.Point
	Scale    compute.Size
	Rotation compute.Point

	model *compute.Matrix4
}

func NewObject(args SceneObjectArgs) *SceneObject {
	o := &SceneObject{
		Id:       rand.Int32(),
		Material: args.Material,
		Shape:    args.Shape,
		Controller: SceneObjectController{
			Init:   args.Init,
			Update: args.Update,
			Input:  args.Input,
		},
	}

	if args.UIElement {
		o.Space = ScreenSpace
	} else {
		o.Space = WorldSpace
	}

	if o.Material != nil {
		o.Size = compute.Point{
			X: float64(args.Material.Diffuse.Width),
			Y: float64(args.Material.Diffuse.Height),
		}.Normalize()
	}

	return o
}

func (o *SceneObjectInstance) Destroy() {
	o.Scene.scheduleOldObjectInstance(o)
}

func (o *SceneObjectInstance) Size() compute.Size {
	return o.SceneObject.Size
}

func (c *SceneObjectInstance) Move(x, y, z float64) {
	c.Position.X += x
	c.Position.Y += y
	c.Position.Z += z
}

func (c *SceneObjectInstance) MoveAt(pos compute.Point) {
	c.Position.X = pos.X
	c.Position.Y = pos.Y
	c.Position.Z = pos.Z
}

// Move objet toward a destination
func (c *SceneObjectInstance) MoveToward(pt compute.Point, s float64) {
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

func (c *SceneObjectInstance) Rotate(rotation compute.Point) {
	c.Rotation.X += rotation.X
	c.Rotation.Y += rotation.Y
	c.Rotation.Z += rotation.Z
}

func (c *SceneObjectInstance) RotateZ(value float64) {
	c.Rotation.Z += value
}

func (c *SceneObjectInstance) RotateX(value float64) {
	c.Rotation.X += value
}

func (c *SceneObjectInstance) RotateY(value float64) {
	c.Rotation.Y += value
}

func (o *SceneObjectInstance) Resize(x, y, z float64) {
	o.Scale.X += x
	o.Scale.Y += y
	o.Scale.Z += z

	if o.Camera != nil {
		o.Camera.updateProjectionMatrix()
		o.Camera.normalizeLookAt()
	}
}

func (o *SceneObjectInstance) ScaleAt(x, y float64) {
	o.Scale.X = x
	o.Scale.Y = y

	if o.Camera != nil {
		o.Camera.updateProjectionMatrix()
		o.Camera.normalizeLookAt()
	}
}

func (o *SceneObjectInstance) IsDescendant(b *SceneObjectInstance) bool {
	if o == b || o.Parent == b {
		return true
	}
	if o.Parent == nil {
		return false
	}
	return o.Parent.IsDescendant(b)
}

func (o *SceneObjectInstance) Points() []compute.Point {
	size := o.Size().Scale(o.Scale)

	x1 := compute.Point{X: o.Position.X + size.X, Y: o.Position.Y + size.Y, Z: o.Position.Z}.Rotate(o.Rotation)
	x2 := compute.Point{X: o.Position.X - size.X, Y: o.Position.Y + size.Y, Z: o.Position.Z}.Rotate(o.Rotation)
	x3 := compute.Point{X: o.Position.X + size.X, Y: o.Position.Y - size.Y, Z: o.Position.Z}.Rotate(o.Rotation)
	x4 := compute.Point{X: o.Position.X - size.X, Y: o.Position.Y - size.Y, Z: o.Position.Z}.Rotate(o.Rotation)

	return []compute.Point{x1, x2, x3, x4}
}

func (o *SceneObjectInstance) ModelMatrix() compute.Matrix {
	o.model.Reset()
	if o.Parent != nil {
		o.model.Mult(o.Parent.ModelMatrix())
	}

	var xRatio float64
	if o.SceneObject.Size.X > o.SceneObject.Size.Y {
		xRatio = o.SceneObject.Size.Y / o.SceneObject.Size.X
	} else {
		xRatio = o.SceneObject.Size.X / o.SceneObject.Size.Y
	}

	o.model.Scale(compute.Size{
		X: o.Scale.X * xRatio,
		Y: o.Scale.Y,
		Z: o.Scale.Z,
	})
	o.model.Rotate(o.Rotation)
	o.model.Translate(o.Position)

	// Note: ScreenSpace should have a dedicated projection matrix
	// that is fixed (defined on client?) with space contained in
	// (0, 0), (1, 1) and looking forward (Z:-1), in orthographic view.

	// ScreenSpace objects should be bound to the current active session
	// meaning that server should not stream other session objects. This
	// could be defined in a separate field to allow shareable UI elements
	// {
	//    Scope: Global / Local
	// }
	return o.model.Out
}

func (o *SceneObject) String() string {
	return fmt.Sprintf("id=%d space=%v w=%.2f h=%.2f", o.Id, o.Space, o.Size.X, o.Size.Y)
}

func (o *SceneObjectInstance) String() string {
	return fmt.Sprintf("id=%d pos=%v object=%v", o.Id, o.Position, o.SceneObject)
}
