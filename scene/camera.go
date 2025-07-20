package scene

import (
	"math"

	"github.com/geotry/rass/compute"
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

	lookAt           compute.Point
	lookAtNormalized compute.Point

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

		lookAt: compute.Point{X: 0, Y: 0, Z: 1},

		projectionMatrix: compute.NewMatrix4(),
		viewMatrix:       compute.NewMatrix4(),
		matrixTicker:     NewTicker(),
	}

	c.normalizeLookAt()
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
		Min: compute.Point{X: c.Parent.Position.X - c.Width/2, Y: c.Parent.Position.Y - c.Height/2},
		Max: compute.Point{X: c.Parent.Position.X + c.Width/2, Y: c.Parent.Position.Y + c.Height/2},
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

func (c *Camera) SetLookAt(x, y float64) {
	c.lookAt.X = x
	c.lookAt.Y = y
	c.normalizeLookAt()
}

func (c *Camera) MoveLookAt(x, y float64) {
	c.lookAt.X += x
	c.lookAt.Y += y
	c.normalizeLookAt()
}

func (c *Camera) LookAt() compute.Point {
	return c.lookAtNormalized
}

func (c *Camera) normalizeLookAt() {
	xi, xf := math.Modf(c.lookAt.X)
	yi, yf := math.Modf(c.lookAt.Y)

	xmod := math.Mod(math.Abs(xi), 4)
	ymod := math.Mod(math.Abs(yi), 4)

	negativeZ := false

	if xmod == 0 {
		c.lookAtNormalized.X = xf
	}
	if ymod == 0 {
		c.lookAtNormalized.Y = yf
	}

	if xmod == 1 {
		if xf == 0 {
			c.lookAtNormalized.X = 1
		} else {
			negativeZ = !negativeZ
			c.lookAtNormalized.X = 1 - math.Abs(xf)
		}
	}
	if ymod == 1 {
		if yf == 0 {
			c.lookAtNormalized.Y = 1
			negativeZ = !negativeZ
		} else {
			c.lookAtNormalized.Y = 1 - math.Abs(yf)
		}
	}

	if xmod == 2 {
		negativeZ = !negativeZ
		if xf == 0 {
			c.lookAtNormalized.X = 0
		} else {
			c.lookAtNormalized.X = -math.Abs(xf)
		}
	}
	if ymod == 2 {
		negativeZ = !negativeZ
		if yf == 0 {
			c.lookAtNormalized.Y = 0
		} else {
			c.lookAtNormalized.Y = -math.Abs(yf)
		}
	}
	if xmod == 3 {
		if xf > 0 {
			c.lookAtNormalized.X = -1 + xf
		} else {
			c.lookAtNormalized.X = -1 - xf
		}
	}
	if ymod == 3 {
		if yf > 0 {
			c.lookAtNormalized.Y = -1 + yf
		} else {
			c.lookAtNormalized.Y = -1 - yf
		}
	}

	if xi < 0 {
		c.lookAtNormalized.X = -c.lookAtNormalized.X
	}
	if yi < 0 {
		c.lookAtNormalized.Y = -c.lookAtNormalized.Y
	}

	if negativeZ {
		c.lookAtNormalized.Z = -1 + (math.Max(math.Abs(c.lookAtNormalized.X), math.Abs(c.lookAtNormalized.Y)))
	} else {
		c.lookAtNormalized.Z = 1 - (math.Max(math.Abs(c.lookAtNormalized.X), math.Abs(c.lookAtNormalized.Y)))
	}
}

func (c *Camera) updateProjectionMatrix() {
	c.projectionMatrix.Reset()
	switch c.Projection {
	case Perspective:
		c.projectionMatrix.Perspective(c.Fov, c.AspectRatio, -.5-c.Near, -c.Far)
	case Orthographic:
		// c.projectionMatrix.Scale(compute.Size{X: c.Scale, Y: c.Scale, Z: c.Scale})
		c.projectionMatrix.Orthographic(c.Width/c.Scale/2.0, -c.Width/c.Scale/2.0, c.Height/c.Scale/2.0, -c.Height/c.Scale/2.0, -c.Near, -c.Far)
	}
}

func (c *Camera) ProjectionMatrix() compute.Matrix {
	return c.projectionMatrix.Out
}

func (c *Camera) ViewMatrix() compute.Matrix {
	if !c.matrixTicker.IsSynced(c.Parent.Scene.ticker) {
		c.matrixTicker.Sync(c.Parent.Scene.ticker)
		c.viewMatrix.Reset()
		position := c.Parent.WorldPosition()
		c.viewMatrix.LookAt(position, position.Add(c.LookAt()))
		c.viewMatrix.Rotate(c.Parent.WorldRotation().Opposite())
	}
	return c.viewMatrix.Out
}

func (c *Camera) Front() {
	c.Parent.Scale.X = 1
	c.Parent.Scale.Y = 1
	c.Parent.Scale.Z = 1
	c.Parent.Rotation.X = 0
	c.Parent.Rotation.Y = 0
	c.Parent.Rotation.Z = 0
}

func (c *Camera) Isometric() {
	c.Parent.Rotation.X = (2.0 * math.Pi) / 3
	c.Parent.Rotation.Y = 0
	c.Parent.Rotation.Z = (2.0 * math.Pi) / 3
}

func (c *Camera) Dimetric() {
	c.Parent.Scale.X = c.Parent.Scale.X * (math.Sqrt(5.0) / 2.0)
	c.Parent.Scale.Z = c.Parent.Scale.Z * (math.Sqrt(5.0) / 2.0)
	c.Parent.Rotation.X = math.Atan(1.0/2.0) + (math.Pi / 2.0)
	c.Parent.Rotation.Y = 0
	c.Parent.Rotation.Z = 2.0 * math.Atan(2.0)
}
