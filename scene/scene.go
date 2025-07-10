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
	objects map[uint32]*SceneObjectInstance
	sorted  []*SceneObjectInstance
	queue   chan func()
	cameras []*Camera

	// Object created and destroyed during last Update()
	NewObjects []*SceneObjectInstance
	OldObjects []*SceneObjectInstance

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
		objects: map[uint32]*SceneObjectInstance{},
		sorted:  make([]*SceneObjectInstance, 0),
		nextId:  1,
		ticker:  &Ticker{},
		cameras: make([]*Camera, 0),
		queue:   make(chan func(), 1000),

		NewObjects: make([]*SceneObjectInstance, 0),
		OldObjects: make([]*SceneObjectInstance, 0),

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

	// Clear old, new objects
	s.OldObjects = nil
	s.NewObjects = nil

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

	// Update all objects
	for _, o := range s.sorted {
		if o.SceneObject.Controller.Update != nil {
			o.SceneObject.Controller.Update(o, deltaTime)
		}
	}

	s.sortObjects()
}

// Sort objects by z-index
func (s *Scene) sortObjects() {
	slices.SortFunc(s.sorted, func(a *SceneObjectInstance, b *SceneObjectInstance) int {
		if a.SceneObject.Space == ScreenSpace && b.SceneObject.Space != ScreenSpace {
			return 1
		}
		if a.SceneObject.Space != ScreenSpace && b.SceneObject.Space == ScreenSpace {
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
func (s *Scene) ReceiveInput(event *pb.InputEvent, source *SceneObjectInstance) {
	s.queue <- func() {
		for _, o := range s.objects {
			if o.SceneObject.Controller.Input != nil && o.IsDescendant(source) {
				o.SceneObject.Controller.Input(o, event)
			}
		}
	}
}

func (s *Scene) SpawnCamera() *SceneObjectInstance {
	return s.Spawn(s.cameraSceneObject, SpawnArgs{camera: NewCamera(s.cameraSettings)})
}

type SpawnArgs struct {
	Position compute.Point
	Rotation compute.Rotation
	Scale    compute.Point
	Parent   *SceneObjectInstance
	Data     map[string]any
	Hidden   bool

	camera *Camera
}

func (s *Scene) Spawn(o *SceneObject, args SpawnArgs) *SceneObjectInstance {
	obj := &SceneObjectInstance{
		SceneObject: o,
		Scene:       s,
		Parent:      args.Parent,
		Camera:      args.camera,
		Data:        make(map[string]any),
		Position:    args.Position,
		Rotation:    args.Rotation,
		Scale:       compute.Point{X: 1, Y: 1, Z: 1},
		SpawnTime:   time.Now(),
		Hidden:      args.Hidden,

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

func (s *Scene) scheduleNewObjectInstance(o *SceneObjectInstance) {
	s.queue <- func() {
		o.Id = s.nextId
		s.nextId = s.nextId + 1
		if o.SceneObject.Controller.Init != nil {
			o.SceneObject.Controller.Init(o)
		}
		s.objects[o.Id] = o
		s.sorted = append(s.sorted, o)
		s.NewObjects = append(s.NewObjects, o)
	}
}

func (s *Scene) scheduleOldObjectInstance(o *SceneObjectInstance) {
	s.queue <- func() {
		delete(s.objects, o.Id)
		s.sorted = slices.DeleteFunc(s.sorted, func(d *SceneObjectInstance) bool { return d == o })
		s.OldObjects = append(s.OldObjects, o)
	}
}

func (s *Scene) Objects() []*SceneObjectInstance {
	return s.sorted
}

// Return objects visible by camera
func (s *Scene) Scan(c *Camera) []*SceneObjectInstance {
	objs := make([]*SceneObjectInstance, 0)

	for _, obj := range s.sorted {
		if c.IsVisible(obj) {
			objs = append(objs, obj)
		}
	}

	return objs
}

func (s *Scene) String() string {
	str := "\nScene\n----\n"
	str = fmt.Sprintf("%sscene has %d objects\n", str, len(s.objects))
	return str
}
