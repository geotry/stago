package scene

import (
	"fmt"
	"slices"
	"sync"
	"time"

	"maps"

	"github.com/geotry/rass/compute"
	"github.com/geotry/rass/pb"
)

type Scene struct {
	nodes   map[uint32]*Node
	sorted  []*Node
	queue   chan func()
	cameras []*Camera

	// Nodes created and destroyed during last Update()
	NewNodes []*Node
	OldNodes []*Node

	nextId uint32
	ticker *Ticker

	cameraSettings    *CameraSettings // default camera settings applied
	cameraSceneObject *SceneObject

	mu sync.RWMutex
}

type SceneOptions struct {
	Camera           *CameraSettings
	CameraController *SceneObjectController
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

	// Update all nodes
	for _, o := range s.sorted {
		if o.Object.Controller.Update != nil {
			o.Object.Controller.Update(o, deltaTime)
		}
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
		if a.Position.Z > b.Position.Z {
			return -1
		} else if a.Position.Z < b.Position.Z {
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
	Parent   *Node
	Data     map[string]any
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
		Position:  args.Position,
		Rotation:  args.Rotation,
		Scale:     compute.Point{X: 1, Y: 1, Z: 1},
		SpawnTime: time.Now(),
		Hidden:    args.Hidden,

		model: compute.NewMatrix4(),
	}

	if args.Scale.X != 0 && args.Scale.Y != 0 {
		obj.Scale = args.Scale
	}

	if obj.Camera != nil {
		obj.Camera.Parent = obj
		obj.Camera.updateProjectionMatrix()
		obj.Camera.normalizeLookAt()
	}

	if args.Data != nil {
		maps.Copy(obj.Data, args.Data)
	}

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
