package scene

import (
	"math"

	"github.com/geotry/rass/compute"
)

type Camera struct {
	Scene *Scene

	Position compute.Point
	Scale    compute.Size
	Rotation compute.Rotation

	Width       float64
	Height      float64
	ScreenSizeX int
	ScreenSizeY int
	AspectRatio float64

	Perspective bool
	Fov         float64
	Near, Far   float64

	c chan struct{}
}

type Viewport = compute.Rectangle2D

// Create a new Camera
func NewCamera(s *Scene) *Camera {
	c := &Camera{
		Scene:       s,
		Position:    compute.Point{X: 0, Y: 0, Z: 0},
		Rotation:    compute.Rotation{X: 0, Y: 0, Z: 0},
		Scale:       compute.Scale(.05), // 1 unit is 2% of screen
		Width:       1,
		Height:      1,
		AspectRatio: 1.0,
		Fov:         70 * (math.Pi / 180),
		Near:        0.01,
		Far:         100.0,
		Perspective: true,
	}

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
	}
}

// Return the area visible by the Camera
func (c *Camera) Viewport() Viewport {
	return compute.Rectangle2D{
		Min: compute.Point{X: c.Position.X - c.Width/2, Y: c.Position.Y - c.Height/2},
		Max: compute.Point{X: c.Position.X + c.Width/2, Y: c.Position.Y + c.Height/2},
	}
}

func (c *Camera) Move(offset compute.Point) {
	c.Position.X += offset.X
	c.Position.Y += offset.Y
	c.Position.Z += offset.Z
}

func (c *Camera) Rotate(offset compute.Rotation) {
	c.Rotation.X += offset.X
	c.Rotation.Y += offset.Y
	c.Rotation.Z += offset.Z
}

func (c *Camera) Resize(scale compute.Size) {
	c.Scale.X += scale.X
	c.Scale.Y += scale.Y
	c.Scale.Z += scale.Z
}

func (c *Camera) Front() {
	c.Scale.X = 1
	c.Scale.Y = 1
	c.Scale.Z = 1
	c.Rotation.X = 0
	c.Rotation.Y = 0
	c.Rotation.Z = 0
}

func (c *Camera) Isometric() {
	c.Rotation.X = (2.0 * math.Pi) / 3
	c.Rotation.Y = 0
	c.Rotation.Z = (2.0 * math.Pi) / 3
}

func (c *Camera) Dimetric() {
	c.Scale.X = c.Scale.X * (math.Sqrt(5.0) / 2.0)
	c.Scale.Z = c.Scale.Z * (math.Sqrt(5.0) / 2.0)
	c.Rotation.X = math.Atan(1.0/2.0) + (math.Pi / 2.0)
	c.Rotation.Y = 0
	c.Rotation.Z = 2.0 * math.Atan(2.0)
}

func (c *Camera) Set(p compute.Point) {
	c.Position.X = p.X
	c.Position.Y = p.Y
	c.Position.Z = p.Z
}

// Convert a point to screen pixel
// func (c *Camera) ScreenSize(s compute.Size) (int, int) {
// 	return int(s.X * c.unitPixel), int(s.Y * c.unitPixel)
// }

// Return screen position relative to camera
// func (c *Camera) ScreenPosition(p compute.Point) (int, int) {
// 	rx, ry := p.X-c.Position.X, p.Y-c.Position.Y

// 	cx, cy := float64(c.Width)/2, float64(c.Height)/2

// 	// Make Y-axis bottom-up
// 	return int((cx + rx) * c.unitPixel), int((cy + ry) * c.unitPixel)
// }

// Render a frame from this camera
func (c *Camera) Render(frame []uint8) error {
	// var objCount = 0

	// var viewport = c.Viewport()

	// // Clear frame
	// for i := range frame {
	// 	frame[i] = 0
	// }

	// for layer := range c.Scene.layers {
	// 	// get objects on layer in viewport
	// 	for _, obj := range c.Scene.Scan(layer, viewport) {
	// 		// Don't render object attached on a different camera
	// 		if obj.Camera != nil && obj.Camera != c {
	// 			continue
	// 		}

	// 		objCount = objCount + 1

	// 		// Get screen position and size of object, in pixels
	// 		screenX, screenY := c.ScreenPosition(obj.Position)
	// 		screenW, screenH := c.ScreenSize(obj.Size())
	// 		tex := obj.SceneObject.Texture
	// 		_ = tex

	// 		if screenW == 0 || screenH == 0 {
	// 			continue
	// 		}

	// 		// Compute offsets to skip offscreen pixels
	// 		minOffsetX, minOffsetY, maxOffsetX, maxOffsetY := 0, 0, screenW, screenH
	// 		if screenX < 0 {
	// 			minOffsetX = -screenX
	// 		}
	// 		if screenY < 0 {
	// 			minOffsetY = -screenY
	// 		}
	// 		if screenX+screenW > c.ScreenSizeX {
	// 			maxOffsetX = screenW - ((screenX + screenW) - c.ScreenSizeX)
	// 		}
	// 		if screenY+screenH > c.ScreenSizeY {
	// 			maxOffsetY = screenH - ((screenY + screenH) - c.ScreenSizeY)
	// 		}

	// 		texture := c.Scene.RM.TextureIndex[tex]
	// 		startIndex := (screenY * c.ScreenSizeX) + screenX
	// 		texSizeXf := texture.Size.X
	// 		texSizeYf := texture.Size.Y
	// 		screenHf, screenWf := float64(screenH), float64(screenW)

	// 		for y := minOffsetY; y < maxOffsetY; y++ {
	// 			yOffset := int(m.Floor((float64(y)/screenHf)*texSizeYf)) * int(texture.Size.X)
	// 			yShift := (y * screenW)
	// 			for x := minOffsetX; x < maxOffsetX; x++ {
	// 				xIndex := int(m.Floor((float64(x) / screenWf) * texSizeXf))
	// 				pixel := texture.Pixels[yOffset+xIndex]
	// 				if pixel != ColorTransparent {
	// 					i := yShift + x
	// 					frameIndex := startIndex + (i/screenW)*c.ScreenSizeX + i%screenW
	// 					frame[frameIndex] = pixel
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	return nil
}

// Sub to new objects in viewport, and unsub old objects
// func (c *Camera) Snapshot() {
// 	// c.mu.Lock()
// 	// defer c.mu.Unlock()

// 	v := c.Viewport()

// 	// track new and old objects
// 	new := map[uint32]*SceneObjectInstance{}
// 	old := map[uint32]*SceneObjectInstance{}

// 	c.Scene.mu.RLock()
// 	for _, o := range c.Scene.objects {
// 		size := o.Size()
// 		rect := compute.Rectangle2D{
// 			Min: compute.Point{X: o.Position.X, Y: o.Position.Y},
// 			Max: compute.Point{X: o.Position.X + size.X, Y: o.Position.Y + size.Y},
// 		}

// 		//
// 		// TODO
// 		// Encode in []uint8
// 		// then compare?

// 		if v.Overlaps(rect) {
// 			if c.objects[o.Id] == nil {
// 				new[o.Id] = o
// 				// e := struct {
// 				// 	Id       uint32
// 				// 	Position compute.Point
// 				// 	Size     compute.Size
// 				// 	Scale    compute.Size
// 				// 	Texture  *Texture
// 				// }{
// 				// 	Id:       o.Id,
// 				// 	Position: o.Position,
// 				// 	Size:     o.SceneObject.Size,
// 				// 	Scale:    o.Scale,
// 				// 	Texture:  o.SceneObject.Texture,
// 				// }
// 			}
// 		} else if c.objects[o.Id] != nil {
// 			old[o.Id] = o
// 			delete(c.objects, o.Id)
// 		}
// 	}
// 	c.Scene.mu.RUnlock()

// 	// Create a copy of Position, Scale, etc. of every object + new/old

// 	// objs := make([]struct{}, len(c.objects)+len(old))

// 	// i := 0
// 	// for id, o := range old {
// 	// 	objs[i] = struct{}{}
// 	// }
// 	// for id, o := range new {
// 	// 	objs[i] = struct{}{}
// 	// }
// 	// for id, o := range c.objects {
// 	// 	objs[i] = struct{}{}
// 	// }

// 	// On first server connection
// 	// send all scene objects (id/texture/size)
// 	// then send object instance (id + scene object id + pos + scale)

// 	// 123,23.44,21.2244,22......
// 	// 123,23.44,21.2244,22......
// 	// 123,23.44,21.2244,22......
// 	// 123,23.44,21.2244,22......

// 	//

// }

// When subscribed object updates, it sends an event to Camera channel
// When websocket server ticks, it reads event from channel, dedup them, and encode it
// to send to client
func (c *Camera) Read() {

	// Consume messages
	for range len(c.c) {
		// event := <-c.c

	}
}
