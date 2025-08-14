package examples

import (
	"image/color"
	"math"
	"math/rand/v2"
	"time"

	"github.com/geotry/stago/compute"
	"github.com/geotry/stago/pb"
	"github.com/geotry/stago/rendering"
	"github.com/geotry/stago/scene"
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

	showAABBs := false
	showCollisions := false

	aabb := scene.NewObject(scene.SceneObjectArgs{
		Material: &rendering.Material{
			Diffuse:   rm.NewTextureFromAtlas("assets/Environment_64x64.png", rendering.Diffuse, 176, 128, 32, 32),
			Specular:  rm.NewTextureFromAtlas("assets/Environment_64x64.png", rendering.Specular, 176, 128, 32, 32),
			Shininess: 0.0,
			Opaque:    false,
		},
		Shape: compute.NewCube(),
		Init: func(self *scene.Node) {
			self.Data["parent"] = self.Parent
			// Detach parent node to not be affected by parent physics
			self.Parent = nil
			self.Transform.Parent = nil
			self.Tint = color.RGBA{R: 0, G: 255, B: 0, A: 255}
		},
		Update: func(self *scene.Node, deltaTime time.Duration) {
			parent := self.Data["parent"].(*scene.Node)
			aabb := parent.AABB()
			self.Transform.Position = parent.Transform.WorldPosition()
			self.Transform.Scale.X = aabb.Scale.X
			self.Transform.Scale.Y = aabb.Scale.Y
			self.Transform.Scale.Z = aabb.Scale.Z

			if showCollisions {
				if len(parent.CollisionTargets) > 0 {
					self.Tint = color.RGBA{R: 255, G: 0, B: 0, A: 255}
				} else {
					self.Tint = color.RGBA{R: 0, G: 255, B: 0, A: 255}
				}
			}
		},
	})

	ground := scene.NewObject(scene.SceneObjectArgs{
		Material: &rendering.Material{
			Diffuse:   rm.NewTextureFromAtlas("assets/Environment_64x64.png", rendering.Diffuse, 96, 96, 64, 64),
			Specular:  rm.NewTextureFromAtlas("assets/Environment_64x64_specular.png", rendering.Specular, 96, 96, 64, 64),
			Shininess: 128.0,
			Opaque:    true,
		},
		Shape:   compute.NewCube(),
		Physics: &scene.Physics{},
		Init: func(self *scene.Node) {
			self.IsKinematic = true
			if showAABBs {
				self.Scene.Spawn(aabb, scene.SpawnArgs{Parent: self})
			}
		},
		Update: func(self *scene.Node, deltaTime time.Duration) {
			if showCollisions {
				if len(self.CollisionTargets) > 0 {
					self.Tint = color.RGBA{R: 255, G: 0, B: 0, A: 255}
				} else {
					self.Tint = color.RGBA{R: 255, G: 255, B: 255, A: 255}
				}
			}
		},
	})

	cube := scene.NewObject(scene.SceneObjectArgs{
		Material: &rendering.Material{
			Diffuse:   rm.NewTextureFromAtlas("assets/Environment_64x64.png", rendering.Diffuse, 176, 96, 32, 32),
			Specular:  rm.NewTextureFromAtlas("assets/Environment_64x64.png", rendering.Specular, 176, 96, 32, 32),
			Shininess: 32.0,
			Opaque:    true,
		},
		Shape: compute.NewCube(),
		Physics: &scene.Physics{
			Mass:           1.0,
			CollisionLayer: 1,
		},
		Init: func(self *scene.Node) {
			if showAABBs {
				self.Scene.Spawn(aabb, scene.SpawnArgs{Parent: self})
			}
		},
		Update: func(self *scene.Node, deltaTime time.Duration) {
			if self.Transform.Position.Y < -100 {
				self.Destroy()
			}
			if showCollisions {
				if len(self.CollisionTargets) > 0 {
					self.Tint = color.RGBA{R: 255, G: 0, B: 0, A: 255}
				} else {
					self.Tint = color.RGBA{R: 255, G: 255, B: 255, A: 255}
				}
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
		Shape: compute.NewCube(),
		Init: func(self *scene.Node) {
			self.IsKinematic = true
			self.Data["velocity"] = 1.0
			self.Data["rotateSpeedX"] = 1 + (rand.Float64() * 2)
		},
		Physics: &scene.Physics{Mass: 1},
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
			if showCollisions {
				if len(self.CollisionTargets) > 0 {
					self.Tint = color.RGBA{R: 255, G: 0, B: 0, A: 255}
				} else {
					self.Tint = color.RGBA{R: 255, G: 255, B: 255, A: 255}
				}
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
			Opaque:    true,
		},
		Shape: compute.NewCube(),
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
							self.Data["spot"] = self.Scene.Spawn(spot, scene.SpawnArgs{Parent: self, Position: compute.Vector3{X: 0, Y: 0, Z: -1}})
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
			Opaque:    true,
		},
		Physics: &scene.Physics{Mass: 1},
		Shape:   compute.NewPyramid(),
		Init: func(self *scene.Node) {
			self.IsKinematic = true
			if showAABBs {
				self.Scene.Spawn(aabb, scene.SpawnArgs{Parent: self})
			}
		},
		Update: func(self *scene.Node, deltaTime time.Duration) {
			if showCollisions {
				if len(self.CollisionTargets) > 0 {
					self.Tint = color.RGBA{R: 255, G: 0, B: 0, A: 255}
				} else {
					self.Tint = color.RGBA{R: 255, G: 255, B: 255, A: 255}
				}
			}
		},
	})

	cameraController := &scene.SceneObjectController{
		Init: func(self *scene.Node) {
			self.Transform.Position = compute.Point{X: 0, Y: 0, Z: 0}
			self.Data["mousemode"] = false
			self.Scene.Spawn(player, scene.SpawnArgs{Parent: self, Position: compute.Point{Y: 0}})
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
				case "KeyT":
					if event.Pressed {
						lookAt := self.Camera.LookAt()
						cb := self.Scene.Spawn(cube, scene.SpawnArgs{
							Position: self.Transform.Position.Add(lookAt.Mult(5)),
							Rotation: self.Camera.PitchYawRoll(),
							Mass:     70,
							Scale:    compute.Vector3{X: .3, Y: .3, Z: .3},
						})
						cb.PushLocal(
							lookAt,
							2500+rand.Float64()*5000,
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
					self.Camera.UpdatePitchYawRoll(-float64(event.DeltaY), float64(event.DeltaX), 0)
				}
			}
		},
	}

	sun := scene.NewObject(scene.SceneObjectArgs{
		Init: func(self *scene.Node) {
			c := color.RGBA{R: 255, G: 255, B: 255, A: 255}
			light := scene.NewDirectionalLight(c, 20, 255, 128)
			light.Direction.Y = -.8
			light.Direction.X = 0
			light.Direction.Z = 1
			self.Light = light
			self.Transform.Position.Y = 20
		},
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
	scn.Spawn(lamp, scene.SpawnArgs{
		Position: compute.Point{
			X: 50, Y: -5, Z: 50,
		},
	})

	for i := range 10 {
		for j := range 10 {
			tint := color.RGBA{R: 255, G: 255, B: 255, A: 255}
			if i%2 == 0 {
				tint.G = 0
			}
			if j%2 == 0 {
				tint.B = 0
			}
			scn.Spawn(ground, scene.SpawnArgs{
				Position: compute.Point{X: float64(i) * 20, Y: -20, Z: float64(j) * 20},
				Rotation: compute.Rotation{X: math.Pi / 2.0, Y: 0, Z: 0},
				Scale:    compute.Size{X: 10, Y: 10, Z: 10},
				Tint:     tint,
			})
		}
	}

	scn.Spawn(rock, scene.SpawnArgs{
		Scale:    compute.Point{X: 1, Y: 1, Z: 1},
		Position: compute.Point{X: 0, Y: -8.9, Z: 10},
	})
	scn.Spawn(rock, scene.SpawnArgs{
		Scale:    compute.Point{X: 5, Y: 5, Z: 5},
		Position: compute.Point{X: 12, Y: -4.9, Z: 32},
		Rotation: compute.Rotation{X: 0, Y: .6, Z: 0},
	})
	scn.Spawn(rock, scene.SpawnArgs{
		Scale:    compute.Point{X: .5, Y: .5, Z: 1},
		Position: compute.Point{X: 4, Y: -9.4, Z: 10},
	})

	wall := NewWall(rm)
	roomGround := NewGround(rm)
	roomScale := compute.Vector3{X: 1, Y: 3, Z: 1}
	roomPos := compute.Vector3{X: 0, Y: 0, Z: 20}
	roomSize := roomScale.X * 4 * 3
	roomHeight := roomScale.Y * 2

	scn.Spawn(roomGround, scene.SpawnArgs{
		Position: roomPos.Add(compute.Vector3{Y: -(roomHeight / 2) - 1}),
		Scale:    compute.Vector3{X: roomSize / 2, Y: 1, Z: roomSize / 2},
	})
	// South wall
	scn.Spawn(wall, scene.SpawnArgs{
		Scale:    roomScale,
		Position: roomPos.Add(compute.Vector3{X: 0, Y: 0, Z: -roomSize / 2}),
		Rotation: compute.Vector3{X: 0, Y: 0, Z: 0},
	})
	// east
	scn.Spawn(wall, scene.SpawnArgs{
		Scale:    roomScale,
		Position: roomPos.Add(compute.Vector3{X: roomSize / 2, Y: 0, Z: 0}),
		Rotation: compute.Vector3{X: 0, Y: math.Pi / 2, Z: 0},
	})
	// north
	scn.Spawn(wall, scene.SpawnArgs{
		Scale:    roomScale,
		Position: roomPos.Add(compute.Vector3{X: 0, Y: 0, Z: roomSize / 2}),
		Rotation: compute.Vector3{X: 0, Y: 0, Z: 0},
	})
	// west
	scn.Spawn(wall, scene.SpawnArgs{
		Scale:    roomScale,
		Position: roomPos.Add(compute.Vector3{X: -roomSize / 2, Y: 0, Z: 0}),
		Rotation: compute.Vector3{X: 0, Y: math.Pi / 2, Z: 0},
	})

	// South east
	scn.Spawn(wall, scene.SpawnArgs{
		Scale:    roomScale.Add(compute.Vector3{X: .5}),
		Position: roomPos.Add(compute.Vector3{X: (2.0 / 3.0) * roomSize / 2, Z: -(2.0 / 3.0) * roomSize / 2}),
		Rotation: compute.Vector3{X: 0, Y: -math.Pi / 4, Z: 0},
	})
	// North east
	scn.Spawn(wall, scene.SpawnArgs{
		Scale:    roomScale.Add(compute.Vector3{X: .5}),
		Position: roomPos.Add(compute.Vector3{X: (2.0 / 3.0) * roomSize / 2, Z: (2.0 / 3.0) * roomSize / 2}),
		Rotation: compute.Vector3{X: 0, Y: math.Pi / 4, Z: 0},
	})
	// North west
	scn.Spawn(wall, scene.SpawnArgs{
		Scale:    roomScale.Add(compute.Vector3{X: .5}),
		Position: roomPos.Add(compute.Vector3{X: -(2.0 / 3.0) * roomSize / 2, Z: (2.0 / 3.0) * roomSize / 2}),
		Rotation: compute.Vector3{X: 0, Y: -math.Pi / 4, Z: 0},
	})
	// South west
	scn.Spawn(wall, scene.SpawnArgs{
		Scale:    roomScale.Add(compute.Vector3{X: .5}),
		Position: roomPos.Add(compute.Vector3{X: -(2.0 / 3.0) * roomSize / 2, Z: -(2.0 / 3.0) * roomSize / 2}),
		Rotation: compute.Vector3{X: 0, Y: math.Pi / 4, Z: 0},
	})

	return scn, rm
}
