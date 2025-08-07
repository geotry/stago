package scenes

import (
	"image/color"
	"log"
	"math"
	"math/rand/v2"
	"time"

	"github.com/geotry/rass/compute"
	"github.com/geotry/rass/pb"
	"github.com/geotry/rass/rendering"
	"github.com/geotry/rass/scene"
	"github.com/geotry/rass/shapes"
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

	ground := scene.NewObject(scene.SceneObjectArgs{
		Material: &rendering.Material{
			Diffuse:   rm.NewTextureFromAtlas("assets/Environment_64x64.png", rendering.Diffuse, 96, 96, 64, 64),
			Specular:  rm.NewTextureFromAtlas("assets/Environment_64x64_specular.png", rendering.Specular, 96, 96, 64, 64),
			Shininess: 128.0,
		},
		Shape: shapes.NewQuad(),
		Physics: &scene.Physics{
			Mass:   50,
			Static: true,
		},
	})

	cube := scene.NewObject(scene.SceneObjectArgs{
		Material: &rendering.Material{
			Diffuse:   rm.NewTextureFromAtlas("assets/Environment_64x64.png", rendering.Diffuse, 176, 96, 32, 32),
			Specular:  rm.NewTextureFromAtlas("assets/Environment_64x64.png", rendering.Specular, 176, 96, 32, 32),
			Shininess: 32.0,
		},
		Shape: shapes.NewCube(),
		Physics: &scene.Physics{
			Mass:           1.0,
			CollisionLayer: 1,
		},
		Update: func(self *scene.Node, deltaTime time.Duration) {
			if self.Transform.Position.Y < -10 {
				self.Destroy()
			}
		},
	})

	ball := scene.NewObject(scene.SceneObjectArgs{
		Material: rm.NewMaterialPalette(6, []uint8{
			ballBorderColor, ballBorderColor, ballBorderColor, ballBorderColor, ballBorderColor, ballBorderColor,
			ballBorderColor, ballBorderColor, ballFillColor, ballFillColor + 1, ballBorderColor, ballBorderColor,
			ballBorderColor, ballFillColor, ballFillColor, ballFillColor, ballFillColor + 1, ballBorderColor,
			ballBorderColor, ballFillColor, ballFillColor, ballFillColor, ballFillColor + 1, ballBorderColor,
			ballBorderColor, ballBorderColor, ballFillColor, ballFillColor, ballBorderColor, ballBorderColor,
			ballBorderColor, ballBorderColor, ballBorderColor, ballBorderColor, ballBorderColor, ballBorderColor,
		}, []uint8{
			ballBorderColor, ballBorderColor, ballBorderColor, ballBorderColor, ballBorderColor, ballBorderColor,
			ballBorderColor, ballBorderColor, ballFillColor, ballFillColor + 1, ballBorderColor, ballBorderColor,
			ballBorderColor, ballFillColor, ballFillColor, ballFillColor, ballFillColor + 1, ballBorderColor,
			ballBorderColor, ballFillColor, ballFillColor, ballFillColor, ballFillColor + 1, ballBorderColor,
			ballBorderColor, ballBorderColor, ballFillColor, ballFillColor, ballBorderColor, ballBorderColor,
			ballBorderColor, ballBorderColor, ballBorderColor, ballBorderColor, ballBorderColor, ballBorderColor,
		}, 128.0),
		Shape: shapes.NewCube(),
		Init: func(self *scene.Node) {
			self.Data["velocity"] = 1.0
			self.Data["rotateSpeedX"] = 1 + (rand.Float64() * 2)
		},
		Update: func(self *scene.Node, deltaTime time.Duration) {
			if self.Data["Target"] == nil {
				return
			}
			targetPoint := self.Data["Target"].(compute.Point)
			velocity := self.Data["velocity"].(float64)
			rotateSpeedX := self.Data["rotateSpeedX"].(float64)

			d := self.Transform.Position.DistanceTo(targetPoint)
			self.MoveToward(targetPoint, compute.Step(d*velocity, deltaTime))
			self.RotateX(compute.Step(rotateSpeedX, deltaTime))

			if self.Transform.Scale.X >= 0 {
				scale := compute.Step(.5*float64(time.Since(self.SpawnTime)/time.Second), deltaTime)
				self.Resize(-scale, -scale, -scale)
			}

			if time.Since(self.SpawnTime) >= time.Duration(time.Second*5) {
				self.Destroy()
			}
		},
	})

	spot := scene.NewObject(scene.SceneObjectArgs{
		Init: func(self *scene.Node) {
			light := scene.NewSpotLight(color.RGBA{R: 128, G: 255, B: 153}, 5, 128, 255)
			light.Ambient.A = 5
			light.Diffuse.A = 128
			light.Specular.A = 255
			self.Light = light
		},
		Update: func(self *scene.Node, deltaTime time.Duration) {
			light := self.Light.(*scene.SpotLight)
			light.Direction = self.Parent.Parent.Camera.LookAt()
		},
	})

	player := scene.NewObject(scene.SceneObjectArgs{
		Material: &rendering.Material{
			Diffuse:   rm.NewTextureFromAtlas("assets/Environment_64x64.png", rendering.Diffuse, 176, 96, 32, 32),
			Specular:  rm.NewTextureFromAtlas("assets/Environment_64x64.png", rendering.Specular, 176, 96, 32, 32),
			Shininess: 32.0,
		},
		Shape: shapes.NewCube(),
		Init: func(self *scene.Node) {
			self.Data["fireRate"] = time.Second / 5.0
			self.Data["lastFired"] = time.Now()
		},
		Input: func(self *scene.Node, event *pb.InputEvent) {
			if self.Parent == nil {
				return
			}
			camera := self.Parent.Camera
			if camera == nil {
				return
			}
			if event.Device == pb.InputDevice_MOUSE {
				if event.Pressed {
					lastFired := time.Since(self.Data["lastFired"].(time.Time))
					fireRate := self.Data["fireRate"].(time.Duration)
					if lastFired > fireRate {
						self.Data["lastFired"] = time.Now()
						self.Scene.Spawn(ball, scene.SpawnArgs{
							Data:     map[string]any{"Target": self.Parent.Transform.WorldPosition().Add(camera.LookAt().Mult(100.0))},
							Position: self.Parent.Transform.WorldPosition().Sub(compute.Point{X: -2}),
						})
					}
				}
				if self.Parent.Data["mousemode"] != true {
					self.SetRotation(camera.PitchYawRoll())
				}
			}
			if event.Device == pb.InputDevice_KEYBOARD {
				switch event.Code {
				case "KeyF":
					if event.Pressed {
						if self.Data["spot"] == nil {
							self.Data["spot"] = self.Scene.Spawn(spot, scene.SpawnArgs{Parent: self, Position: compute.Vector3{X: 3, Y: 0, Z: 1}})
						} else {
							self.Data["spot"].(*scene.Node).Destroy()
							self.Data["spot"] = nil
						}
					}
				}
			}
		},
	})

	rock := scene.NewObject(scene.SceneObjectArgs{
		Material: &rendering.Material{
			Diffuse:   rm.NewMaterialRGBAFromFile("assets/Sprite-0003.png", rendering.Diffuse),
			Specular:  rm.NewMaterialRGBAFromFile("assets/Sprite-0003-specular.png", rendering.Specular),
			Shininess: 256.0,
		},
		Shape: shapes.NewPyramid(),
	})

	// cursor := scene.NewObject(scene.SceneObjectArgs{
	// 	Material: rm.NewMaterialPalette(11, []uint8{
	// 		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	// 		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	// 		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	// 		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	// 		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	// 		255, 255, 255, 255, 255, 0, 0, 255, 255, 255, 255,
	// 		255, 255, 255, 255, 255, 0, 15, 0, 255, 255, 255,
	// 		255, 255, 255, 255, 255, 0, 15, 15, 0, 255, 255,
	// 		255, 255, 255, 255, 255, 0, 15, 15, 15, 0, 255,
	// 		255, 255, 255, 255, 255, 0, 15, 0, 0, 255, 255,
	// 		255, 255, 255, 255, 255, 0, 0, 255, 255, 255, 255,
	// 	}, nil, 0.0),
	// 	UIElement: true,
	// 	Shape:     shapes.NewQuad(),
	// 	Input: func(self *scene.Node, event *pb.InputEvent) {
	// 		if event.Device == pb.InputDevice_MOUSE {
	// 			// self.Move(float64(event.DeltaX), float64(event.DeltaY), 0)
	// 			self.MoveAt(compute.Point{X: float64(event.X), Y: float64(event.Y), Z: self.Transform.Position.Z})
	// 			// self.MoveAt(compute.Point{
	// 			// 	X: self.Camera.Position.X - (cameraScreenWidth / 2) + cameraScreenWidth*float64(event.X),
	// 			// 	Y: self.Camera.Position.Y + (cameraScreenHeight / 2) - cameraScreenHeight*float64(event.Y),
	// 			// 	Z: self.Transform.Position.Z,
	// 			// })
	// 		}
	// 	},
	// })

	// point := scene.NewObject(scene.SceneObjectArgs{
	// 	Material:  rm.NewMaterialPalette(1, []uint8{12}, []uint8{12}, 0.0),
	// 	UIElement: true,
	// 	Shape:     shapes.NewQuad(),
	// 	Init: func(self *scene.Node) {
	// 		self.Transform.Scale = compute.Size{X: .02, Y: .02, Z: 1}
	// 	},
	// })

	cameraController := &scene.SceneObjectController{
		Init: func(self *scene.Node) {
			self.Transform.Position = compute.Point{X: 0, Y: -5, Z: -5}
			self.Data["mousemode"] = false
			self.Scene.Spawn(player, scene.SpawnArgs{Parent: self, Position: compute.Point{Y: 0}})
			// self.Scene.Spawn(point, scene.SpawnArgs{})
			// self.Scene.Spawn(cursor, scene.SpawnArgs{})
		},
		Update: func(self *scene.Node, deltaTime time.Duration) {
			speed := 5.0
			if self.Data["boost"] == true {
				speed *= 10
			}

			offset := compute.Point{}
			lookAt := self.Camera.LookAt()
			if self.Data["right"] == true {
				r := lookAt.Cross(compute.Vector3{Y: -1})
				offset.X += compute.Step(speed, deltaTime) * r.X
				offset.Y += compute.Step(speed, deltaTime) * r.Y
				offset.Z += compute.Step(speed, deltaTime) * r.Z
			}
			if self.Data["left"] == true {
				l := lookAt.Cross(compute.Vector3{Y: 1})
				offset.X += compute.Step(speed, deltaTime) * l.X
				offset.Y += compute.Step(speed, deltaTime) * l.Y
				offset.Z += compute.Step(speed, deltaTime) * l.Z
			}
			if self.Data["up"] == true {
				offset.Y += compute.Step(speed, deltaTime)
			}
			if self.Data["down"] == true {
				offset.Y -= compute.Step(speed, deltaTime)
			}
			if self.Data["forward"] == true {
				offset.X += compute.Step(speed, deltaTime) * lookAt.X
				offset.Y += compute.Step(speed, deltaTime) * lookAt.Y
				offset.Z += compute.Step(speed, deltaTime) * lookAt.Z
			}
			if self.Data["backward"] == true {
				offset.X -= compute.Step(speed, deltaTime) * lookAt.X
				offset.Y -= compute.Step(speed, deltaTime) * lookAt.Y
				offset.Z -= compute.Step(speed, deltaTime) * lookAt.Z
			}
			self.Move(offset.X, offset.Y, offset.Z)

			rotate := compute.Point{}
			if self.Data["rotateLeft"] == true {
				rotate.Y -= compute.Step(speed/10.0, deltaTime)
			}
			if self.Data["rotateRight"] == true {
				rotate.Y += compute.Step(speed/10.0, deltaTime)
			}
			if self.Data["rotateForward"] == true {
				rotate.X += compute.Step(speed/10.0, deltaTime)
			}
			if self.Data["rotateBackward"] == true {
				rotate.X -= compute.Step(speed/10.0, deltaTime)
			}
			self.Rotate(rotate)
		},
		Input: func(self *scene.Node, event *pb.InputEvent) {
			if event.Device == pb.InputDevice_KEYBOARD {
				switch event.Code {
				case "ShiftLeft":
					self.Data["boost"] = event.Pressed
				case "ArrowUp":
				case "KeyW": // Z
					if self.Camera.Projection == scene.Perspective {
						self.Data["forward"] = event.Pressed
					} else {
						self.Data["up"] = event.Pressed
					}
				case "ArrowDown":
				case "KeyS":
					if self.Camera.Projection == scene.Perspective {
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
					if self.Camera.Projection == scene.Perspective {
						self.Data["down"] = event.Pressed
					} else {
						self.Data["backward"] = event.Pressed
					}
				case "Space":
					if self.Camera.Projection == scene.Perspective {
						self.Data["up"] = event.Pressed
					} else {
						self.Data["forward"] = event.Pressed
					}

				case "Escape":
					if event.Pressed {
						self.Data["mousemode"] = true
					}
				case "Enter":
					if event.Pressed {
						self.Data["mousemode"] = false
					}

				case "Digit1":
					if event.Pressed {
						self.Camera.SetProjection(scene.Perspective)
					}
				case "Digit2":
					if event.Pressed {
						self.Camera.SetProjection(scene.Orthographic)
					}
				// case "KeyF":
				// 	if event.Pressed {
				// 		if self.Data["spot"] == nil {
				// 			// self.Data["spot"] = self.Scene.Spawn(spot, scene.SpawnArgs{Parent: self, Position: compute.Vector3{Z: 2, Y: -2}})
				// 			self.Data["spot"] = self.Scene.Spawn(spot, scene.SpawnArgs{Parent: self, Position: compute.Vector3{Y: -2}})
				// 		} else {
				// 			self.Data["spot"].(*scene.Node).Destroy()
				// 			self.Data["spot"] = nil
				// 		}
				// 	}
				case "KeyT":
					if event.Pressed {
						lookAt := self.Camera.LookAt()
						cb := self.Scene.Spawn(cube, scene.SpawnArgs{
							Position: self.Transform.Position.Add(lookAt.Mult(5)),
							Rotation: self.Camera.PitchYawRoll(),
							Mass:     70,
							Scale:    compute.Vector3{X: .3, Y: .3, Z: .3},
						})
						// cb.Transform.Rotation = self.Transform.Rotation
						log.Println(cb)
						cb.PushLocal(
							// compute.Vector3{X: 0, Y: 0, Z: 1},
							// compute.Vector3{Y: 1.5, Z: 2, X: -1 + rand.Float64()*2},
							// compute.Vector3{Y: 1, Z: 2, X: -1 + rand.Float64()*2},
							lookAt,
							2500+rand.Float64()*5000,
							// 2500,
							// compute.Vector3{X: 0.5, Y: .5, Z: 0},
							// compute.Vector3{X: 0.8, Y: .9, Z: 0},
							compute.Vector3{
								X: compute.Clamp(-1+rand.Float64()*2, -.2, .2),
								Y: compute.Clamp(-1+rand.Float64()*2, -.2, .2),
								Z: 0,
							},
						)
					}
				}
			}

			if event.Device == pb.InputDevice_MOUSE {
				if event.Scrolled {
					offset := -.2 / float64(event.Delta)
					self.Resize(offset, offset, offset)
				}
				if self.Data["mousemode"] != true {
					self.Camera.MoveLookAt(float64(event.DeltaX), -float64(event.DeltaY))
				}
			}
		},
	}

	sun := scene.NewObject(scene.SceneObjectArgs{
		Init: func(self *scene.Node) {
			light := scene.NewDirectionalLight(color.RGBA{R: 233, G: 233, B: 233, A: 255}, 10, 255, 120)
			self.Light = light
			self.Transform.Position.Y = 20
			self.Data["obj"] = self.Scene.Spawn(ball, scene.SpawnArgs{Parent: self, Scale: compute.Vector3{X: 1, Y: .5, Z: 1}})
		},
		Update: func(self *scene.Node, deltaTime time.Duration) {
			light := self.Light.(*scene.DirectionalLight)

			light.Direction.Y = -.8
			light.Direction.X = 0
			light.Direction.Z = 1
			light.Direction.X = compute.Clamp(math.Sin(float64(time.Since(self.SpawnTime))/float64(time.Second)), -1, .5)
			light.Direction.Y = compute.Clamp(math.Sin(float64(time.Since(self.SpawnTime))/float64(time.Second)), -.5, -.2)

			obj := self.Data["obj"].(*scene.Node)
			obj.SetRotation(compute.Vector3{
				Y: light.Direction.X * math.Pi / 2,
				X: light.Direction.Y * math.Pi / 2,
				Z: light.Direction.Z * math.Pi / 2,
			})
			// light.Direction.Y = compute.Clamp(math.Sin(float64(time.Since(self.SpawnTime))/float64(time.Second)), -1, 0)
			// light.Direction = compute.Vector3{X: .6, Y: -.8, Z: 0}
		},
		// 	light := self.Light.(*scene.DirectionalLight)
		// 	light.DiffuseIntensity = compute.Clamp((1.0+(math.Sin(float64(time.Since(self.SpawnTime))/float64(time.Second)*.8)))/2.0, .2, .8)
		// 	light.Diffuse.R = uint8(compute.Clamp((255.0+(math.Sin(float64(time.Since(self.SpawnTime))/float64(time.Second)))*255.0)/2, 0, 255))
		// },
	})

	lamp := scene.NewObject(scene.SceneObjectArgs{
		Init: func(self *scene.Node) {
			light := scene.NewPointLight(color.RGBA{R: 233, G: 64, B: 64, A: 255}, 0, 250, 120)
			light.Radius = 20.0
			light.Diffuse = color.RGBA{B: 200, R: 100, G: 20, A: 255}
			light.Specular = color.RGBA{B: 200, R: 100, G: 20, A: 48}
			self.Light = light
			self.Scene.Spawn(ball, scene.SpawnArgs{Parent: self})
		},
		Update: func(self *scene.Node, deltaTime time.Duration) {
			self.Transform.Position.Y = 1 + 5*math.Sin(float64(time.Since(self.SpawnTime))/float64(time.Second))
			// self.Transform.Position.X = compute.Clamp(math.Sin(float64(time.Since(self.SpawnTime))/float64(time.Second)), -10, 10)
			// self.Light.DiffuseColor()
			// light := self.Light.(*scene.PointLight)
			// light.Ambient.B = 255
			// light.DiffuseIntensity = compute.Clamp((1.0+(math.Sin(float64(time.Since(self.SpawnTime))/float64(time.Second)*.8)))/2.0, 0, 1)
			// light.SpecularIntensity = compute.Clamp((1.0+(math.Sin(float64(time.Since(self.SpawnTime))/float64(time.Second)*.8)))/2.0, 0, 1)
			// light.Diffuse.R = uint8(compute.Clamp((255.0+(math.Sin(float64(time.Since(self.SpawnTime))/float64(time.Second)))*255.0)/2, 0, 255))
		},
	})

	scn := scene.NewScene(scene.SceneOptions{
		// Default camera settings when new camera is added to the scene
		Camera: &scene.CameraSettings{
			Projection: scene.Perspective,
			Fov:        70.0 * (math.Pi / 180),
			Near:       0.1,
			Far:        100.0,
			Scale:      0.05, // For orthographic view
		},
		CameraController: cameraController,
	})

	// Spawn some objects in scene
	scn.Spawn(sun, scene.SpawnArgs{})
	// log.Println(sun)
	scn.Spawn(lamp, scene.SpawnArgs{
		Position: compute.Point{
			X: 50, Y: -5, Z: 50,
		},
	})

	for i := range 10 {
		for j := range 10 {
			scn.Spawn(ground, scene.SpawnArgs{
				Position: compute.Point{X: float64(i) * 20, Y: -10, Z: float64(j) * 20},
				Rotation: compute.Rotation{X: math.Pi / 2.0, Y: 0, Z: 0},
				Scale:    compute.Size{X: 10, Y: 10, Z: 1},
			})
		}
	}

	scn.Spawn(rock, scene.SpawnArgs{
		Scale:    compute.Point{X: 1, Y: 1, Z: 1},
		Position: compute.Point{X: 0, Y: -10, Z: 10},
	})
	scn.Spawn(rock, scene.SpawnArgs{
		Scale:    compute.Point{X: 5, Y: 5, Z: 5},
		Position: compute.Point{X: 4, Y: -10, Z: 12},
		Rotation: compute.Rotation{X: 0, Y: .2, Z: 0},
	})
	scn.Spawn(rock, scene.SpawnArgs{
		Scale:    compute.Point{X: .5, Y: .5, Z: 1},
		Position: compute.Point{X: 4, Y: -10, Z: 10},
	})

	return scn, rm
}
