package scene

import (
	"math"

	"github.com/geotry/stago/compute"
)

type CameraProjection uint8

const (
	Perspective CameraProjection = iota
	Orthographic
)

type Camera struct {
	Width       float64
	Height      float64
	AspectRatio float64

	Projection CameraProjection
	Fov        float64
	Near, Far  float64
	Scale      float64

	Parent *Node

	pitchYawRoll compute.Vector3

	projectionMatrix *compute.Matrix4
	viewMatrix       *compute.Matrix4
	matrixTicker     *Ticker
}

type CameraSettings struct {
	Projection CameraProjection
	Fov        float64
	Near       float64
	Far        float64
	Scale      float64
}

type Viewport = compute.Plane

// Create a new Camera
func NewCamera(settings *CameraSettings) *Camera {
	c := &Camera{
		Width:       1,
		Height:      1,
		AspectRatio: 1.0,
		Fov:         settings.Fov,
		Near:        settings.Near,
		Far:         settings.Far,
		Scale:       settings.Scale,

		Projection: settings.Projection,

		pitchYawRoll:     compute.Vector3{X: 0, Y: -math.Pi / 2, Z: 0},
		projectionMatrix: compute.NewMatrix4(),
		viewMatrix:       compute.NewMatrix4(),
		matrixTicker:     NewTicker(),
	}

	c.updateProjectionMatrix()

	return c
}

func (c *Camera) SetSize(width, height int) {
	if width > 0 && height > 0 {
		c.AspectRatio = float64(width) / float64(height)
		if c.AspectRatio > 1 {
			c.Width = 1
			c.Height = 1 / c.AspectRatio
		} else {
			c.Width = 1 * c.AspectRatio
			c.Height = 1
		}
		c.updateProjectionMatrix()
	}
}

func (c *Camera) SetNear(near float64) {
	c.Near = near
	c.updateProjectionMatrix()
}

func (c *Camera) SetFar(far float64) {
	c.Far = far
	c.updateProjectionMatrix()
}

func (c *Camera) SetFov(fov float64) {
	c.Fov = fov
	c.updateProjectionMatrix()
}

func (c *Camera) SetProjection(projection CameraProjection) {
	if c.Projection != projection {
		c.Projection = projection
		c.updateProjectionMatrix()
	}
}

// Return the area visible by the Camera
func (c *Camera) Viewport() Viewport {
	return compute.Plane{
		Min: compute.Point{X: c.Parent.Transform.Position.X - c.Width/2, Y: c.Parent.Transform.Position.Y - c.Height/2},
		Max: compute.Point{X: c.Parent.Transform.Position.X + c.Width/2, Y: c.Parent.Transform.Position.Y + c.Height/2},
	}
}

// Returns true if one of the point is in camera projection
func (c *Camera) IsVisible(o *Node) bool {
	if o.Hidden {
		return false
	}

	// Check object scope
	// if o.LocalToCamera != nil && o.LocalToCamera != c {
	// 	return false
	// }

	// if o.SceneObject.Space == ScreenSpace {
	// 	return true
	// }

	// m := c.ModelViewMatrix(o)
	// for _, p := range o.Points() {
	// 	vp, w := p.MultMatrix(m)
	// 	if -w < vp.X && vp.X < w && -w < vp.Y && vp.Y < w && -w < vp.Z && vp.Z < w {
	// 		return true
	// 	}
	// }

	return true
}

func (c *Camera) LookAt() compute.Point {
	pitch := c.pitchYawRoll.X
	yaw := c.pitchYawRoll.Y
	xzLen := math.Cos(pitch)
	return compute.Vector3{
		X: xzLen * math.Cos(yaw),
		Y: math.Sin(pitch),
		Z: xzLen * math.Sin(-yaw),
	}
}

func (c *Camera) UpdatePitchYawRoll(pitch float64, yaw float64, roll float64) {
	c.pitchYawRoll.X += pitch
	c.pitchYawRoll.X = math.Max(math.Min(c.pitchYawRoll.X, 1.2), -1.2)
	c.pitchYawRoll.Y += yaw
	if math.Abs(c.pitchYawRoll.Y) >= math.Pi*2 {
		c.pitchYawRoll.Y = 0
	}
	c.pitchYawRoll.Z += roll
}

func (c *Camera) PitchYawRoll() compute.Vector3 {
	return c.pitchYawRoll
}

func (c *Camera) updateProjectionMatrix() {
	c.projectionMatrix.Reset()
	switch c.Projection {
	case Perspective:
		c.projectionMatrix.Perspective(c.Fov, c.AspectRatio, c.Near, c.Far)
	case Orthographic:
		c.projectionMatrix.Orthographic(c.Width/c.Scale/2.0, -c.Width/c.Scale/2.0, c.Height/c.Scale/2.0, -c.Height/c.Scale/2.0, c.Near, c.Far)
	}
}

func (c *Camera) ProjectionMatrix() compute.Matrix {
	return c.projectionMatrix.Out
}

func (c *Camera) ViewMatrix() compute.Matrix {
	if !c.matrixTicker.IsSynced(c.Parent.Scene.ticker) {
		c.matrixTicker.Sync(c.Parent.Scene.ticker)
		c.viewMatrix.Reset()
		position := c.Parent.Transform.WorldPosition()
		c.viewMatrix.LookAt(position, position.Add(c.LookAt()))
		c.viewMatrix.Rotate(c.Parent.Transform.WorldRotation().Inverse())
	}
	return c.viewMatrix.Out
}

func (c *Camera) Front() {
	c.Parent.Transform.Scale.X = 1
	c.Parent.Transform.Scale.Y = 1
	c.Parent.Transform.Scale.Z = 1
	c.Parent.Transform.Rotation = compute.NewQuaternion(compute.Vector3{})
}

func (c *Camera) Isometric() {
	c.Parent.Transform.Rotation = compute.NewQuaternionFromEuler(compute.Vector3{X: (2.0 * math.Pi) / 3, Z: (2.0 * math.Pi) / 3})
}

func (c *Camera) Dimetric() {
	c.Parent.Transform.Scale.X = c.Parent.Transform.Scale.X * (math.Sqrt(5.0) / 2.0)
	c.Parent.Transform.Scale.Z = c.Parent.Transform.Scale.Z * (math.Sqrt(5.0) / 2.0)
	c.Parent.Transform.Rotation = compute.NewQuaternionFromEuler(compute.Vector3{X: math.Atan(1.0/2.0) + (math.Pi / 2.0), Z: 2.0 * math.Atan(2.0)})
}
