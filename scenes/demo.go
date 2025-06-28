package scenes

import (
	"math"
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

				cameraScreenWidth, cameraScreenHeight := (self.Camera.Width / self.Camera.Scale.X), (self.Camera.Height / self.Camera.Scale.Y)
				self.Scale.X = cameraScreenWidth / 2 / self.SceneObject.Size.X
				self.Scale.Y = cameraScreenHeight / 2 / self.SceneObject.Size.Y
			}
		},
	})

	ground := scene.NewObject(scene.SceneObjectArgs{
		Texture: rm.NewTextureRGBAFromFile("assets/Sprite-0001.png"),
	})

	for i := range 10 {
		for j := range 10 {
			s.Spawn(ground, scene.SpawnArgs{
				Position: compute.Point{X: float64(i) * 20, Y: -10, Z: float64(j) * 20},
				Rotation: compute.Rotation{X: math.Pi / 2.0, Y: 0, Z: 0},
				Scale:    compute.Size{X: 10, Y: 10, Z: 1},
			})
		}
	}

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
			// if self.Camera != nil && self.Position.X == 0 && self.Position.Y == 0 {
			// 	self.Position = compute.Point{
			// 		X: self.Camera.Position.X + (rand.Float64() * 2) - 1,
			// 		Y: self.Camera.Position.Y + (rand.Float64() * 2) - 1,
			// 	}
			// }
			// scaleFactor := .2 + (rand.Float64() * .8)
			// self.Scale = compute.Point{
			// 	X: scaleFactor,
			// 	Y: scaleFactor,
			// }
			// self.Data["offset"] = compute.Point{
			// 	X: (rand.Float64() * 4) - 2,
			// 	Y: (rand.Float64() * 4) - 2,
			// }
			// self.Data["velocity"] = .1 + rand.Float64()
			self.Data["offset"] = compute.Point{}
			self.Data["velocity"] = 1.0
		},
		Update: func(self *scene.SceneObjectInstance, deltaTime time.Duration) {
			targetPoint := self.Data["targetPoint"].(compute.Point)
			offset := self.Data["offset"].(compute.Point)
			velocity := self.Data["velocity"].(float64)

			if self.Data["target"] != nil {
				target := self.Data["target"].(*scene.SceneObjectInstance)
				p := compute.Point{X: target.Position.X + offset.X, Y: target.Position.Y + offset.Y}
				d := self.Position.DistanceTo(p)
				self.MoveToward(p, step(d*velocity, deltaTime))
			} else if !targetPoint.IsZero() {
				d := self.Position.DistanceTo(targetPoint)
				self.MoveToward(targetPoint, step(d*velocity, deltaTime))
			}

			self.RotateX(step(1, deltaTime))

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
			self.Hidden = self.Camera.Data["mousemode"] != true
		},
		Input: func(self *scene.SceneObjectInstance, event *pb.InputEvent) {
			if self.Camera == nil {
				return
			}

			if event.Device == pb.InputDevice_MOUSE {
				cameraScreenWidth, cameraScreenHeight := (self.Camera.Width / self.Camera.Scale.X), (self.Camera.Height / self.Camera.Scale.Y)

				// self.Position.X += cameraScreenWidth*float64(event.DeltaX)
				// self.Position.Y -= cameraScreenHeight*float64(event.DeltaY)
				self.Position.X = self.Camera.Position.X - (cameraScreenWidth / 2) + cameraScreenWidth*float64(event.X)
				self.Position.Y = self.Camera.Position.Y + (cameraScreenHeight / 2) - cameraScreenHeight*float64(event.Y)

				// if event.Pressed {
				// 	self.Scene.Spawn(ball, scene.SpawnArgs{
				// 		Camera:   self.Camera,
				// 		Data:     map[string]any{"target": self},
				// 		Position: compute.Point{X: self.Position.X, Y: self.Position.Y, Z: rand.Float64() * 10},
				// 	})
				// }

				// Update LookAt vector
				if self.Camera.Data["mousemode"] != true {
					self.Camera.MoveLookAt(float64(event.DeltaX), -float64(event.DeltaY))
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
	})

	s.Spawn(rock, scene.SpawnArgs{
		Scale:    compute.Point{X: 1, Y: 1, Z: 1},
		Position: compute.Point{X: 0, Y: -9, Z: 10},
	})
	s.Spawn(rock, scene.SpawnArgs{
		Scale:    compute.Point{X: 2, Y: 2, Z: 1},
		Position: compute.Point{X: 4, Y: -8.5, Z: 12},
		Rotation: compute.Rotation{X: 0, Y: .2, Z: 0},
	})
	s.Spawn(rock, scene.SpawnArgs{
		Scale:    compute.Point{X: .5, Y: .5, Z: 1},
		Position: compute.Point{X: 4, Y: -9.5, Z: 10},
	})

	cameraController := scene.NewObject(scene.SceneObjectArgs{
		Texture: rm.NewTexturePalette([]uint8{6}, 1),
		Update: func(self *scene.SceneObjectInstance, deltaTime time.Duration) {
			if self.Camera == nil {
				return
			}

			speed := 20.0
			if self.Data["boost"] == true {
				speed *= 1.5
			}

			offset := compute.Point{}
			lookAt := self.Camera.LookAt()
			if self.Data["right"] == true {
				right := lookAt.Rotate(compute.Point{Y: math.Pi / 2})
				offset.X += step(speed, deltaTime) * right.X
				offset.Y += step(speed, deltaTime) * right.Y
				offset.Z += step(speed, deltaTime) * right.Z
			}
			if self.Data["left"] == true {
				left := lookAt.Rotate(compute.Point{Y: -math.Pi / 2})
				offset.X += step(speed, deltaTime) * left.X
				offset.Y += step(speed, deltaTime) * left.Y
				offset.Z += step(speed, deltaTime) * left.Z
			}
			if self.Data["up"] == true {
				offset.Y += step(speed, deltaTime)
			}
			if self.Data["down"] == true {
				offset.Y -= step(speed, deltaTime)
			}
			if self.Data["forward"] == true {
				offset.X += step(speed, deltaTime) * lookAt.X
				offset.Y += step(speed, deltaTime) * lookAt.Y
				offset.Z += step(speed, deltaTime) * lookAt.Z
			}
			if self.Data["backward"] == true {
				offset.X -= step(speed, deltaTime) * lookAt.X
				offset.Y -= step(speed, deltaTime) * lookAt.Y
				offset.Z -= step(speed, deltaTime) * lookAt.Z
			}
			self.Camera.Move(offset)

			rotate := compute.Point{}
			if self.Data["rotateLeft"] == true {
				rotate.Y -= step(speed/10.0, deltaTime)
			}
			if self.Data["rotateRight"] == true {
				rotate.Y += step(speed/10.0, deltaTime)
			}
			if self.Data["rotateForward"] == true {
				rotate.X += step(speed/10.0, deltaTime)
			}
			if self.Data["rotateBackward"] == true {
				rotate.X -= step(speed/10.0, deltaTime)
			}
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

				case "Escape":
					if event.Pressed {
						if self.Camera.Data["mousemode"] == true {
							self.Camera.Data["mousemode"] = false
						} else {
							self.Camera.Data["mousemode"] = true
						}
					}

				case "Digit1":
					if event.Pressed {
						self.Camera.Perspective = true
						for k := range self.Data {
							delete(self.Data, k)
						}
					}
				case "Digit2":
					if event.Pressed {
						self.Camera.Perspective = false
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

			if event.Device == pb.InputDevice_MOUSE {
				if event.Pressed {
					self.Scene.Spawn(ball, scene.SpawnArgs{
						Camera:   self.Camera,
						Data:     map[string]any{"targetPoint": self.Camera.Position.Add(self.Camera.LookAt().Mult(20.0))},
						Rotation: self.Camera.LookAt(),
						Position: self.Camera.Position.Add(self.Camera.LookAt()),
					})
					// self.Scene.Spawn(ball, scene.SpawnArgs{
					// 	Camera:   self.Camera,
					// 	Data:     map[string]any{"targetPoint": self.Position.Add(self.Camera.LookAt())},
					// 	// Scale: compute.Scale(1),
					// 	Rotation: compute.Point{},
					// 	Position: compute.Point{X: self.Camera.Position.X, Y: self.Camera.Position.Y, Z: self.Camera.Position.Z},
					// })
				}
			}
		},
	})

	// Spawn objects when Camera is added
	s.WithCamera(func(c *scene.Camera) {
		s.Spawn(cameraController, scene.SpawnArgs{Camera: c, Hidden: true})
		s.Spawn(background, scene.SpawnArgs{Camera: c})
		s.Spawn(point, scene.SpawnArgs{Camera: c, Position: compute.Point{X: 0, Y: 0}})
		s.Spawn(cursor, scene.SpawnArgs{Camera: c, Scale: compute.Scale(.2)})
	})

	return s, rm
}

func step(v float64, deltaTime time.Duration) float64 {
	return v * float64(deltaTime) / float64(time.Second)
}
