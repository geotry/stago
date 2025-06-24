package scene

import (
	"encoding/binary"
	"math"
	"time"

	"github.com/geotry/rass/compute"
	"github.com/geotry/rass/pb"
	"github.com/geotry/rass/rendering"
)

type Texture struct {
	Size struct{ X, Y int }
	Data []uint8
}

type SceneObject struct {
	Texture   int
	Size      compute.Size
	UIElement bool
	Init      func(self *SceneObjectInstance)
	Update    func(self *SceneObjectInstance, deltaTime time.Duration)
	Input     func(self *SceneObjectInstance, event *pb.InputEvent)
}

type SceneObjectArgs struct {
	Texture   *rendering.PalettedTexture
	UIElement bool

	Init   func(self *SceneObjectInstance)
	Update func(self *SceneObjectInstance, deltaTime time.Duration)
	Input  func(self *SceneObjectInstance, event *pb.InputEvent)
}

type SceneObjectInstance struct {
	Id uint32

	SceneObject *SceneObject
	Scene       *Scene

	SpawnTime time.Time

	// Attach this object to a Camera to be visible on this camera only
	// Its position becomes relative to Camera position
	Camera *Camera

	Data map[string]any

	Position compute.Point
	Scale    compute.Size
	Rotation compute.Point

	matrix *compute.Matrix4
}

func NewObject(args SceneObjectArgs) *SceneObject {
	o := &SceneObject{
		Init:      args.Init,
		Update:    args.Update,
		Input:     args.Input,
		UIElement: args.UIElement,
	}

	if args.Texture != nil {
		o.Texture = args.Texture.Id
		o.Size = compute.Point{X: args.Texture.Size.X, Y: args.Texture.Size.Y}
	}

	return o
}

func (o *SceneObjectInstance) Destroy() {
	o.Scene.oldObjects <- o
}

func (o *SceneObjectInstance) Size() compute.Size {
	return compute.Size{
		X: o.SceneObject.Size.X * o.Scale.X,
		Y: o.SceneObject.Size.Y * o.Scale.Y,
		Z: o.SceneObject.Size.Z * o.Scale.Z,
	}
}

func (c *SceneObjectInstance) Move(offx float64, offy float64, offz float64) {
	c.Position.X += offx
	c.Position.Y += offy
	c.Position.Z += offz
}

// Move objet toward a destination
func (c *SceneObjectInstance) MoveToward(pt compute.Point, s float64) {
	d := c.Position.DistanceTo(pt)
	ds := math.Min(d, s)

	if ds >= 0 {
		dx := math.Abs(pt.X - c.Position.X)
		dy := math.Abs(pt.Y - c.Position.Y)
		dvx := (dx / (dx + dy)) * s
		dvy := (dy / (dx + dy)) * s
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

func (o *SceneObjectInstance) Origin() compute.Point {
	size := o.Size()
	return compute.Point{
		X: o.Position.X + size.X/2,
		Y: o.Position.Y + size.Y/2,
		Z: o.Position.Z + size.Z/2,
	}
}

func (o *SceneObjectInstance) Area() compute.Rectangle2D {
	return compute.Rectangle2D{
		Min: o.Position,
		Max: compute.Point{X: o.Position.X + o.Size().X, Y: o.Position.Y + o.Size().Y},
	}
}

type BlockType = uint8

const (
	TextureBlock BlockType = iota
	ScaleBlock
	RotationBlock
	TranslateBlock
	ProjectionBlock
	MatrixBlock
)

const EncodeBufferSize = 4 + 67

// Encode a scene object instance at index in buffer and returns new buffer offset
func (o *SceneObjectInstance) Encode(c *Camera, index uint16, buffer []uint8) int {
	offset := int(index) * EncodeBufferSize

	// Texture (4 bytes)
	offset = putUint8(buffer, offset, TextureBlock)
	offset = putUint16(buffer, offset, index)
	offset = putUint8(buffer, offset, uint8(o.SceneObject.Texture))

	// Matrix (67 bytes)
	offset = putUint8(buffer, offset, MatrixBlock)
	offset = putUint16(buffer, offset, index)
	m := o.ComputeMatrix(c)
	for i := range m {
		offset = putFloat32(buffer, offset, float32(m[i]))
	}

	return offset
}

func (o *SceneObjectInstance) ComputeMatrix(c *Camera) compute.Matrix {
	o.matrix.Reset()

	// Apply object transformation first, then camera
	o.matrix.Scale(o.Scale)
	o.matrix.Rotate(o.Rotation)
	o.matrix.Translate(compute.Point{
		X: (o.Position.X - c.Position.X),
		Y: (o.Position.Y - c.Position.Y),
		Z: (o.Position.Z - c.Position.Z),
	})

	o.matrix.Scale(c.Scale)
	if !o.SceneObject.UIElement {
		o.matrix.Rotate(c.Rotation)
	}

	if o.SceneObject.UIElement {
		o.matrix.Orthographic(c.Width/2.0, -c.Width/2.0, c.Height/2.0, -c.Height/2.0, 0, -1)
	} else if c.Perspective {
		o.matrix.Perspective(c.Fov, c.AspectRatio, -.5-c.Near, -c.Far)
	} else {
		o.matrix.Orthographic(c.Width/2.0, -c.Width/2.0, c.Height/2.0, -c.Height/2.0, 0, -c.Far)
	}

	return o.matrix.Out
}

func putUint8(buffer []uint8, offset int, value uint8) int {
	buffer[offset] = value
	return offset + 1
}

func putUint16(buffer []uint8, offset int, value uint16) int {
	binary.BigEndian.PutUint16(buffer[offset:], value)
	return offset + 2
}

func putFloat32(buffer []uint8, offset int, value float32) int {
	binary.BigEndian.PutUint32(buffer[offset:], math.Float32bits(value))
	return offset + 4
}
