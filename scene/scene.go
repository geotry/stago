package scene

import (
	"fmt"
	"image/color"
	"math"
	"slices"
	"sync"
	"time"

	"maps"

	"github.com/geotry/stago/compute"
	"github.com/geotry/stago/pb"
)

type Scene struct {
	nodes   map[uint32]*Node
	sorted  []*Node
	queue   chan func()
	cameras []*Camera

	// Nodes created and destroyed during last Update()
	NewNodes []*Node
	OldNodes []*Node

	gravity compute.Vector3

	nextId uint32
	ticker *Ticker

	cameraSettings    *CameraSettings // default camera settings applied
	cameraSceneObject *SceneObject

	mu sync.RWMutex
}

type SceneOptions struct {
	Camera           *CameraSettings
	CameraController *SceneObjectController
	Gravity          *Force
}

func NewScene(opts SceneOptions) *Scene {
	scene := &Scene{
		nodes:   map[uint32]*Node{},
		sorted:  make([]*Node, 0),
		nextId:  1,
		ticker:  &Ticker{},
		cameras: make([]*Camera, 0),
		queue:   make(chan func(), 1000),

		NewNodes: make([]*Node, 0),
		OldNodes: make([]*Node, 0),

		gravity: compute.Vector3{Y: -9.8},

		cameraSettings: opts.Camera,
		cameraSceneObject: NewObject(SceneObjectArgs{
			Init:   opts.CameraController.Init,
			Update: opts.CameraController.Update,
			Input:  opts.CameraController.Input,
		}),
	}
	return scene
}

func (s *Scene) Update() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear old, new nodes
	s.OldNodes = nil
	s.NewNodes = nil

	// Process all events in queue
queue:
	for {
		select {
		case fn := <-s.queue:
			fn()
		default:
			break queue
		}
	}

	_, deltaTime := s.ticker.Tick()
	// deltaTime /= 10

	// Save transforms before update to compare with new transforms
	oldTransforms := make(map[*Node]compute.Transform)
	for _, o := range s.sorted {
		oldTransforms[o] = *o.Transform
	}

	// Update all nodes
	for _, o := range s.sorted {
		if o.Object.Controller.Update != nil {
			o.Object.Controller.Update(o, deltaTime)
		}
	}

	// Update motion of physical objects
	for _, o := range s.sorted {
		// Reset collisions targets
		o.CollisionTargets = nil
		if o.Object.Physics != nil && !o.IsKinematic && !o.IsStatic() {
			o.UpdatePhysicsMotion(deltaTime)
		}
		o.UpdateCollider()
	}

	// Compute collisions

	// 1. Broad phase with Sweep and Prune
	pairs := compute.SweepAndPrune(s.sorted)

	// 2. Narrow phase with GJK
	collisions := make([]Collision, 0)
	for _, pair := range pairs {
		a := pair.A
		b := pair.B
		// Always make a the moving object
		if pair.A.IsStatic() {
			a = pair.B
			b = pair.A
		}
		// Ignore collisions of two static objects
		if a.IsStatic() {
			continue
		}

		// Check if colliders A and B really collided with GJK and EPA
		hitAB, hitInfoAB := compute.GJK(a.Collider, b.Collider)
		if hitAB {
			collisions = append(collisions, Collision{
				Source: a,
				Target: b,
				Hit:    hitInfoAB,
			})
		}

		// If b is also moving, compute the opposite collision
		if !b.IsStatic() {
			hitBA, hitInfoBA := compute.GJK(b.Collider, a.Collider)
			if hitBA {
				collisions = append(collisions, Collision{
					Source: b,
					Target: a,
					Hit:    hitInfoBA,
				})
			}
		}
	}

	// 3. Resolution
	for _, collision := range collisions {
		source := collision.Source
		target := collision.Target
		depth := collision.Hit.Depth
		norm := collision.Hit.Normal

		v := collision.Source.TranslationVelocity
		if v.Dot(norm) < 0 {
			norm = norm.Mult(-1)
		}
		newVelocity := v.Sub(norm.Mult(2 * (v.Dot(norm))))

		source.CollisionTargets = append(source.CollisionTargets, target)

		// Update position (move at surface) and velocity (take mirror velocity from normal)
		source.Transform.Position = source.Transform.Position.Sub(norm.Mult(depth))
		source.TranslationVelocity = newVelocity

		// Transfer momentum
		// v1f = ((m1-m2)*v1i + 2 * m2 * v2i) / (m1+m2)
		// v2f = ((m2-m1)*v2i + 2 * m1 * v1i) / (m1+m2)

		m1 := source.Mass
		m2 := target.Mass
		if target.IsKinematic {
			m2 = m1 + 100000
		}
		v1i := source.TranslationVelocity
		v2i := target.TranslationVelocity
		v1f := v1i.Mult(m1 - m2).Add(v2i.Mult(m2 * 2)).Div(m1 + m2)

		source.TranslationVelocity = v1f.Mult(-1)
		source.UpdateMomentum()

		// If object is colliding with an object beneath and velocity is small enough, make it kinematic
		if (norm.IsZero() || norm.Y < -.75) && source.TranslationVelocity.Length() <= 0.1 {
			source.IsKinematic = true
			source.TranslationVelocity = compute.Vector3{}
			source.AngularVelocity = compute.Vector3{}
		}

		// Debug: set objects as kinematic to freeze collision point
		// source.IsKinematic = true
		// source.TranslationVelocity = compute.Vector3{}
		// source.AngularVelocity = compute.Vector3{}
		// collision.B.IsKinematic = true
		// collision.B.TranslationVelocity = compute.Vector3{}
		// collision.B.AngularVelocity = compute.Vector3{}

		// log.Println(collision.A.Transform.Position, collision.B.Transform.Position, collision.Hit)
	}

	s.sortNodes()
}

// Sort nodes by z-index
func (s *Scene) sortNodes() {
	slices.SortFunc(s.sorted, func(a *Node, b *Node) int {
		if a.Object.Space == ScreenSpace && b.Object.Space != ScreenSpace {
			return 1
		}
		if a.Object.Space != ScreenSpace && b.Object.Space == ScreenSpace {
			return -1
		}
		if a.Transform.Position.Z > b.Transform.Position.Z {
			return -1
		} else if a.Transform.Position.Z < b.Transform.Position.Z {
			return 1
		} else {
			return int(b.Id) - int(a.Id)
		}
	})
}

// Queue an input event
func (s *Scene) ReceiveInput(event *pb.InputEvent, source *Node) {
	s.queue <- func() {
		for _, o := range s.nodes {
			if o.Object.Controller.Input != nil && o.IsDescendant(source) {
				o.Object.Controller.Input(o, event)
			}
		}
	}
}

func (s *Scene) SpawnCamera() *Node {
	return s.Spawn(s.cameraSceneObject, SpawnArgs{camera: NewCamera(s.cameraSettings)})
}

type SpawnArgs struct {
	Position compute.Point
	Rotation compute.Rotation
	Scale    compute.Point
	Mass     float64
	Parent   *Node
	Data     map[string]any
	Tint     color.RGBA
	Hidden   bool

	camera *Camera
}

func (s *Scene) Spawn(o *SceneObject, args SpawnArgs) *Node {
	obj := &Node{
		Object:    o,
		Scene:     s,
		Parent:    args.Parent,
		Camera:    args.camera,
		Data:      make(map[string]any),
		SpawnTime: time.Now(),
		Mass:      args.Mass,
		Hidden:    args.Hidden,
		Tint:      color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Transform: compute.NewTransform(nil),
		// TransformOld: compute.NewTransform(nil),
	}

	if args.Tint.A != 0 {
		obj.Tint = args.Tint
	}

	if args.Parent != nil {
		obj.Transform.Parent = args.Parent.Transform
	}

	if args.Mass == 0 && o.Physics != nil {
		obj.Mass = o.Physics.Mass
	}

	if obj.Mass > 0 {
		dragCoef := 1.05 // cube
		pArea := 1.0     // 1m² for a 1x1 cube
		density := 1.2   // air
		obj.TerminalVelocity = math.Sqrt((2 * obj.Mass * s.gravity.Length()) / (density * pArea * dragCoef))
	}

	if args.Scale.X != 0 && args.Scale.Y != 0 {
		obj.Transform.Scale = args.Scale
	}
	if !args.Position.IsZero() {
		obj.Transform.Position = args.Position
	}
	if !args.Rotation.IsZero() {
		obj.Transform.Rotation = compute.NewQuaternionFromEuler(args.Rotation)
	}

	if obj.Camera != nil {
		obj.Camera.Parent = obj
		obj.Camera.updateProjectionMatrix()
	}

	if args.Data != nil {
		maps.Copy(obj.Data, args.Data)
	}

	obj.UpdateCollider()

	s.scheduleNewObjectInstance(obj)

	return obj
}

func (s *Scene) scheduleNewObjectInstance(o *Node) {
	s.queue <- func() {
		o.Id = s.nextId
		s.nextId = s.nextId + 1
		if o.Object.Controller.Init != nil {
			o.Object.Controller.Init(o)
		}
		s.nodes[o.Id] = o
		s.sorted = append(s.sorted, o)
		s.NewNodes = append(s.NewNodes, o)
	}
}

func (s *Scene) scheduleOldObjectInstance(o *Node) {
	s.queue <- func() {
		deleted := make([]*Node, 0)
		for _, obj := range s.nodes {
			if obj == o || obj.IsDescendant(o) {
				deleted = append(deleted, obj)
			}
		}

		for _, obj := range deleted {
			delete(s.nodes, obj.Id)
			s.OldNodes = append(s.OldNodes, obj)
		}

		s.sorted = slices.Collect(maps.Values(s.nodes))
		s.sortNodes()
	}
}

func (s *Scene) Objects() []*Node {
	return s.sorted
}

// Return nodes visible by camera
func (s *Scene) Scan(c *Camera) []*Node {
	objs := make([]*Node, 0)

	for _, obj := range s.sorted {
		if c.IsVisible(obj) {
			objs = append(objs, obj)
		}
	}

	return objs
}

func (s *Scene) String() string {
	str := "\nScene\n----\n"
	str = fmt.Sprintf("%sscene has %d nodes\n", str, len(s.nodes))
	return str
}
