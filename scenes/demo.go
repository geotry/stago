package scenes

import (
	"math"
	"math/rand/v2"
	"time"

	"github.com/geotry/rass/compute"
	"github.com/geotry/rass/pb"
	"github.com/geotry/rass/rendering"
	"github.com/geotry/rass/scene"
)

const ballBorderColor = 0
const ballFillColor = 10

// https://mycours.es/crc/F0D62941

var Palette = []string{
	"0B0F00", // #0B0F00
	"281703", // #281703
	"3F0F07", // #3F0F07
	"564907", // #564907
	"3D700E", // #3D700E
	"18872E", // #18872E
	"219E8F", // #219E8F
	"2D69B7", // #2D69B7
	"603BCE", // #603BCE
	"CF4ED8", // #CF4ED8
	"E06494", // #E06494
	"E8A081", // #E8A081
	"EDEB9C", // #EDEB9C
	"CAF4B5", // #CAF4B5
	"D6FFE4", // #D6FFE4
	"EFFEFF", // #EFFEFF
}

func NewDemo() (*scene.Scene, *rendering.ResourceManager) {
	rm := rendering.NewResourceManager()

	rm.UseRGBPalette(Palette)

	s := scene.NewScene(scene.SceneOptions{})

	background := scene.NewObject(scene.SceneObjectArgs{
		Texture: rm.NewTexturePalette([]uint8{7, 5, 5}, 1),
		Update: func(self *scene.SceneObjectInstance, deltaTime time.Duration) {
			if self.Camera != nil {

				self.Position.X = self.Camera.Position.X
				self.Position.Y = self.Camera.Position.Y
				self.Position.Z = self.Camera.Position.Z + 100

				// self.Position.Z = self.Camera.Position.Z - 1000

				// self.Camera.Rotation.Y -= step(1, deltaTime)
				// self.Scale.X = self.Camera.Width * (self.Camera.Width * (self.SceneObject.Size.X / self.SceneObject.Size.Y))
				// self.Scale.X = self.Camera.Width + (self.Camera.AspectRatio * self.Camera.Width)
				// self.Scale.Y = self.Camera.Height / self.Camera.AspectRatio
				// self.Scale.X = (self.Camera.Width / self.SceneObject.Size.X) * self.Camera.AspectRatio
				// self.Scale.Y = (self.Camera.Height / self.SceneObject.Size.Y) / self.Camera.AspectRatio

				cameraScreenWidth, cameraScreenHeight := (self.Camera.Width / self.Camera.Scale.X), (self.Camera.Height / self.Camera.Scale.Y)
				self.Scale.X = cameraScreenWidth / 2 / self.SceneObject.Size.X
				self.Scale.Y = cameraScreenHeight / 2 / self.SceneObject.Size.Y

				// self.Position.X = self.Camera.Position.X - (cameraScreenWidth / 2) + cameraScreenWidth*float64(event.X)
				// self.Position.Y = self.Camera.Position.Y + (cameraScreenHeight / 2) - cameraScreenHeight*float64(event.Y*float32(self.Camera.AspectRatio))

				// self.Move(0, 0, step(.1, deltaTime))

				// self.Rotation.Z = .2
				// self.Rotation.X = .2
				// self.Position.Y = 15

				// TODO: Improve this
				// if self.SceneObject.Size.X > self.SceneObject.Size.Y {
				// 	self.Scale.X = (self.Camera.Width * self.Camera.WidthRatio) / self.SceneObject.Size.X
				// 	self.Scale.Y = (self.Camera.Height * self.Camera.HeightRatio) / self.SceneObject.Size.Y
				// } else {
				// 	self.Scale.X = (self.Camera.Width * self.Camera.HeightRatio) / self.SceneObject.Size.X
				// 	self.Scale.Y = (self.Camera.Height * self.Camera.WidthRatio) / self.SceneObject.Size.Y
				// }
			}
		},
	})

	// s.Spawn(scene.NewObject(scene.SceneObjectArgs{
	// 	Texture: rm.NewTexturePalette([]uint8{11, 11, 8, 8}, 2),
	// 	// Update: func(self *scene.SceneObjectInstance, deltaTime time.Duration) {

	// 	// },
	// }), scene.SpawnArgs{
	// 	Position: compute.Point{X: 0, Y: 0, Z: 0},
	// 	Rotation: compute.Rotation{X: math.Pi / 2.0, Y: 0, Z: 0},
	// 	Scale:    compute.Size{X: 20, Y: 20, Z: 1},
	// })

	ground := scene.NewObject(scene.SceneObjectArgs{
		Texture: rm.NewTextureRGBAFromFile("assets/Sprite-0001.png"),
		// Update: func(self *scene.SceneObjectInstance, deltaTime time.Duration) {
		// },
	})
	s.Spawn(ground, scene.SpawnArgs{
		Position: compute.Point{X: 0, Y: -10, Z: 0},
		Rotation: compute.Rotation{X: math.Pi / 2.0, Y: 0, Z: 0},
		Scale:    compute.Size{X: 10, Y: 10, Z: 1},
	})
	s.Spawn(ground, scene.SpawnArgs{
		Position: compute.Point{X: 0, Y: -10, Z: 10},
		Rotation: compute.Rotation{X: math.Pi / 2.0, Y: math.Pi, Z: 0},
		Scale:    compute.Size{X: 10, Y: 10, Z: 1},
	})

	square := scene.NewObject(scene.SceneObjectArgs{
		Texture: rm.NewTexturePalette([]uint8{1, 2, 3, 4}, 2),
		Update: func(self *scene.SceneObjectInstance, deltaTime time.Duration) {
			self.Rotate(compute.Point{X: step(compute.PI, deltaTime), Y: step(compute.PI, deltaTime), Z: step(compute.PI, deltaTime)})
		},
	})

	rock := scene.NewObject(scene.SceneObjectArgs{
		Texture: rm.NewTextureRGBAFromFile("assets/Sprite-0003.png"),
	})

	ball := scene.NewObject(scene.SceneObjectArgs{
		Texture: rm.NewTexturePalette([]uint8{
			255, 255, ballBorderColor, ballBorderColor, 255, 255,
			255, ballBorderColor, ballFillColor, ballFillColor + 1, ballBorderColor, 255,
			ballBorderColor, ballFillColor, ballFillColor, ballFillColor, ballFillColor + 1, ballBorderColor,
			ballBorderColor, ballFillColor, ballFillColor, ballFillColor, ballFillColor + 1, ballBorderColor,
			255, ballBorderColor, ballFillColor, ballFillColor, ballBorderColor, 255,
			255, 255, ballBorderColor, ballBorderColor, 255, 255,
		}, 6),
		Init: func(self *scene.SceneObjectInstance) {
			if self.Camera != nil && self.Position.X == 0 && self.Position.Y == 0 {
				self.Position = compute.Point{
					X: self.Camera.Position.X + (rand.Float64() * 2) - 1,
					Y: self.Camera.Position.Y + (rand.Float64() * 2) - 1,
				}
			}
			scaleFactor := .2 + (rand.Float64() * .8)
			self.Scale = compute.Point{
				X: scaleFactor,
				Y: scaleFactor,
			}
			self.Data["offset"] = compute.Point{
				X: (rand.Float64() * 4) - 2,
				Y: (rand.Float64() * 4) - 2,
			}
			self.Data["velocity"] = .1 + rand.Float64()
		},
		Update: func(self *scene.SceneObjectInstance, deltaTime time.Duration) {
			target := self.Data["target"].(*scene.SceneObjectInstance)
			offset := self.Data["offset"].(compute.Point)
			velocity := self.Data["velocity"].(float64)

			if target != nil {
				p := compute.Point{X: target.Position.X + offset.X, Y: target.Position.Y + offset.Y}
				d := self.Position.DistanceTo(p)
				self.MoveToward(p, step(d*velocity, deltaTime))
			}

			self.RotateZ(step(1, deltaTime))

			// var t = newTime / float64(time.Second)
			// var t = float64(time.Now().UnixNano()) / float64(time.Second)
			// var scale float64 = 1.7 + 0.8*m.Sin(2*float64(t)*2*float64(compute.PI))

			// Scale down over time

			// self.Scale.X = scale
			// self.Scale.Y = scale
			if self.Scale.X >= 0 {
				scale := step(.5*float64(time.Since(self.SpawnTime)/time.Second), deltaTime)
				self.Scale.X -= scale
				self.Scale.Y -= scale
			}

			if time.Since(self.SpawnTime) >= time.Duration(time.Second*5) {
				self.Destroy()
			}
		},
	})

	cursor := scene.NewObject(scene.SceneObjectArgs{
		Texture: rm.NewTexturePalette([]uint8{
			255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
			255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
			255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
			255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
			255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
			255, 255, 255, 255, 255, 0, 0, 255, 255, 255, 255,
			255, 255, 255, 255, 255, 0, 15, 0, 255, 255, 255,
			255, 255, 255, 255, 255, 0, 15, 15, 0, 255, 255,
			255, 255, 255, 255, 255, 0, 15, 15, 15, 0, 255,
			255, 255, 255, 255, 255, 0, 15, 0, 0, 255, 255,
			255, 255, 255, 255, 255, 0, 0, 255, 255, 255, 255,
		}, 11),
		UIElement: true,
		Update: func(self *scene.SceneObjectInstance, deltaTime time.Duration) {
			if self.Camera == nil {
				return
			}
			self.Position.Z = self.Camera.Position.Z
		},
		Input: func(self *scene.SceneObjectInstance, event *pb.InputEvent) {
			if self.Camera == nil {
				return
			}
			if event.Device == pb.InputDevice_MOUSE {
				cameraScreenWidth, cameraScreenHeight := (self.Camera.Width / self.Camera.Scale.X), (self.Camera.Height / self.Camera.Scale.Y)
				self.Position.X = self.Camera.Position.X - (cameraScreenWidth / 2) + cameraScreenWidth*float64(event.X)
				self.Position.Y = self.Camera.Position.Y + (cameraScreenHeight / 2) - cameraScreenHeight*float64(event.Y)

				if event.Pressed {
					self.Scene.Spawn(ball, scene.SpawnArgs{
						Camera:   self.Camera,
						Data:     map[string]any{"target": self},
						Position: compute.Point{X: self.Position.X, Y: self.Position.Y, Z: 1},
					})
				}

				// TODO
				// Implement a better "lookAt" where camera follows the center of the screen
				if self.Camera.Perspective {
					self.Camera.Rotation.Y = (.5 - float64(event.X)) * 2 * math.Pi
					self.Camera.Rotation.X = (.5 - float64(event.Y)) * 2 * math.Pi
				}
			}
		},
	})

	// Spawn objects in scene
	s.Spawn(square, scene.SpawnArgs{
		Position: compute.Point{X: 0, Y: 0, Z: 2},
		Scale:    compute.Size{X: 1, Y: 1, Z: 1},
	})

	point := scene.NewObject(scene.SceneObjectArgs{
		Texture:   rm.NewTexturePalette([]uint8{12}, 1),
		UIElement: true,
		Init: func(self *scene.SceneObjectInstance) {
			self.Scale = compute.Size{X: .02, Y: .02, Z: 1}
		},
		Update: func(self *scene.SceneObjectInstance, deltaTime time.Duration) {
			if self.Camera != nil {
				self.Position.Z = self.Camera.Position.Z + 1
			}
		},
	})

	s.Spawn(rock, scene.SpawnArgs{
		Scale:    compute.Point{X: 1, Y: 1, Z: 1},
		Position: compute.Point{X: 0, Y: -10, Z: 10},
	})

	cameraController := scene.NewObject(scene.SceneObjectArgs{
		Texture: rm.NewTexturePalette([]uint8{255}, 1),
		Update: func(self *scene.SceneObjectInstance, deltaTime time.Duration) {
			if self.Camera == nil {
				return
			}
			speed := 5.0
			if self.Data["boost"] == true {
				speed *= 1.5
			}

			offset := compute.Point{}
			if self.Data["right"] == true {
				offset.X += step(speed, deltaTime)
			}
			if self.Data["left"] == true {
				offset.X -= step(speed, deltaTime)
			}
			if self.Data["up"] == true {
				offset.Y += step(speed, deltaTime)
			}
			if self.Data["down"] == true {
				offset.Y -= step(speed, deltaTime)
			}
			if self.Data["forward"] == true {
				offset.Z += step(speed, deltaTime)
			}
			if self.Data["backward"] == true {
				offset.Z -= step(speed, deltaTime)
			}

			rotate := compute.Point{}
			if self.Data["rotateLeft"] == true {
				rotate.Z -= step(speed/10.0, deltaTime)
			}
			if self.Data["rotateRight"] == true {
				rotate.Z += step(speed/10.0, deltaTime)
			}
			if self.Data["rotateForward"] == true {
				rotate.X += step(speed/10.0, deltaTime)
			}
			if self.Data["rotateBackward"] == true {
				rotate.X -= step(speed/10.0, deltaTime)
			}

			self.Camera.Move(offset)
			self.Camera.Rotate(rotate)
		},
		Input: func(self *scene.SceneObjectInstance, event *pb.InputEvent) {
			if self.Camera == nil {
				return
			}

			if event.Device == pb.InputDevice_KEYBOARD {
				switch event.Code {
				case "ShiftLeft":
					self.Data["boost"] = event.Pressed
				case "ArrowUp":
				case "KeyW": // Z
					if self.Camera.Perspective {
						self.Data["forward"] = event.Pressed
					} else {
						self.Data["up"] = event.Pressed
					}
				case "ArrowDown":
				case "KeyS":
					if self.Camera.Perspective {
						self.Data["backward"] = event.Pressed
					} else {
						self.Data["down"] = event.Pressed
					}
				case "ArrowRight":
				case "KeyD":
					self.Data["right"] = event.Pressed
				case "ArrowLeft":
				case "KeyA": // Q
					self.Data["left"] = event.Pressed
				case "KeyQ": // A
					self.Data["rotateLeft"] = event.Pressed
				case "KeyE":
					self.Data["rotateRight"] = event.Pressed
				case "KeyX": // A
					self.Data["rotateForward"] = event.Pressed
				case "KeyV":
					self.Data["rotateBackward"] = event.Pressed
				case "KeyZ": // W
					self.Data["forward"] = event.Pressed
				case "KeyC":
					if self.Camera.Perspective {
						self.Data["down"] = event.Pressed
					} else {
						self.Data["backward"] = event.Pressed
					}
				case "Space":
					if self.Camera.Perspective {
						self.Data["up"] = event.Pressed
					} else {
						self.Data["forward"] = event.Pressed
					}

				case "Tab":
					if event.Pressed {
						self.Camera.Perspective = !self.Camera.Perspective
						for k := range self.Data {
							delete(self.Data, k)
						}
					}
				}
			}
			if event.Device == pb.InputDevice_MOUSE {
				if event.Scrolled {
					offset := -.2 / float64(event.Delta)
					self.Camera.Scale.X += offset
					self.Camera.Scale.Y += offset
					self.Camera.Scale.Z += offset
				}
			}
		},
	})

	// Spawn objects when Camera is added
	s.WithCamera(func(c *scene.Camera) {
		s.Spawn(cameraController, scene.SpawnArgs{Camera: c})

		for i := range 20 {
			s.Spawn(point, scene.SpawnArgs{
				Camera:   c,
				Position: compute.Point{X: -1 * float64(i), Y: 0},
			})
			s.Spawn(point, scene.SpawnArgs{
				Camera:   c,
				Position: compute.Point{X: 1 * float64(i), Y: 0},
			})
			s.Spawn(point, scene.SpawnArgs{
				Camera:   c,
				Position: compute.Point{X: 0, Y: 1 * float64(i)},
			})
			s.Spawn(point, scene.SpawnArgs{
				Camera:   c,
				Position: compute.Point{X: 0, Y: -1 * float64(i)},
			})
		}

		s.Spawn(background, scene.SpawnArgs{Camera: c})
		s.Spawn(cursor, scene.SpawnArgs{Camera: c})
	})

	return s, rm
}

func step(v float64, deltaTime time.Duration) float64 {
	return v * float64(deltaTime) / float64(time.Second)
}
