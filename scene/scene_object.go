package scene

import (
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"github.com/geotry/rass/compute"
	"github.com/geotry/rass/pb"
	"github.com/geotry/rass/rendering"
)

type Texture struct {
	Size struct{ X, Y int }
	Data []uint8
}

type SceneEntity interface {
	ModelMatrix() compute.Matrix
}

type SceneObject struct {
	Id        int32
	Texture   *rendering.Texture
	Size      compute.Size
	UIElement bool

	Geometry []compute.Point
	UV       []compute.Point

	Init   func(self *SceneObjectInstance)
	Update func(self *SceneObjectInstance, deltaTime time.Duration)
	Input  func(self *SceneObjectInstance, event *pb.InputEvent)
}

type SceneObjectArgs struct {
	Texture   *rendering.Texture
	Geometry  []compute.Point
	UV        []compute.Point
	UIElement bool

	Init   func(self *SceneObjectInstance)
	Update func(self *SceneObjectInstance, deltaTime time.Duration)
	Input  func(self *SceneObjectInstance, event *pb.InputEvent)
}

type SceneObjectInstance struct {
	Id uint32

	SceneObject *SceneObject
	Scene       *Scene
	Parent      *SceneObjectInstance
	Hidden      bool

	SpawnTime time.Time

	// Attach this object to a Camera to be visible on this camera only
	// Its position becomes relative to Camera position
	Camera *Camera

	Data map[string]any

	Position compute.Point
	Scale    compute.Size
	Rotation compute.Point

	model *compute.Matrix4
}

func NewObject(args SceneObjectArgs) *SceneObject {
	o := &SceneObject{
		Id:        rand.Int32(),
		Init:      args.Init,
		Update:    args.Update,
		Input:     args.Input,
		UIElement: args.UIElement,
		Texture:   args.Texture,
		Geometry:  args.Geometry,
		UV:        args.UV,
	}

	if o.Geometry == nil {
		o.Geometry = compute.NewQuad()
	}
	if o.UV == nil {
		o.UV = compute.NewQuadUV()
	}

	if o.Texture != nil {
		o.Size = compute.Point{
			X: float64(args.Texture.Width),
			Y: float64(args.Texture.Height),
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

func (o *SceneObjectInstance) Grow(x, y, z float64) {
	o.Scale.X += x
	o.Scale.Y += y
	o.Scale.Z += z
}

func (o *SceneObjectInstance) ScaleAt(x, y float64) {
	o.Scale.X = x
	o.Scale.Y = y
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
	return o.model.Out
}

func (o *SceneObject) String() string {
	return fmt.Sprintf("id=%d ui=%v w=%.2f h=%.2f texture=%d (%d)", o.Id, o.UIElement, o.Size.X, o.Size.Y, o.Texture.Id, o.Texture.Index)
}

func (o *SceneObjectInstance) String() string {
	return fmt.Sprintf("id=%d object=%v", o.Id, o.SceneObject)
}
