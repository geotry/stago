package scene

import (
	"fmt"
	"log"
	"slices"
	"sync"
	"time"

	"maps"

	"github.com/geotry/rass/compute"
	"github.com/geotry/rass/pb"
)

const ColorTransparent = 255

type Scene struct {
	objects map[uint32]*SceneObjectInstance

	inputs chan struct {
		event  *pb.InputEvent
		source *Camera
	}
	newObjects, oldObjects chan *SceneObjectInstance

	newCameras, oldCameras chan *Camera
	cameras                []*Camera

	nextId     uint32
	lastTick   time.Time
	ticks      int
	initCamera func(c *Camera)

	mu sync.RWMutex
}

type SceneOptions struct {
}

func NewScene(opts SceneOptions) *Scene {
	scene := &Scene{
		objects:  map[uint32]*SceneObjectInstance{},
		nextId:   1,
		lastTick: time.Now(),

		cameras: make([]*Camera, 0),

		inputs: make(chan struct {
			event  *pb.InputEvent
			source *Camera
		}, 1024),
		newCameras: make(chan *Camera, 1024),
		oldCameras: make(chan *Camera, 1024),
		newObjects: make(chan *SceneObjectInstance, 1024),
		oldObjects: make(chan *SceneObjectInstance, 1024),
	}
	return scene
}

// Run the main pipeline
func (s *Scene) Tick() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ticks++

	if s.ticks%60 == 0 {
		log.Printf("tick=%d cameras=%v objects=%v", s.ticks, len(s.cameras), len(s.objects))
	}

	// Process all events in queue
queue:
	for {
		select {
		case newCamera := <-s.newCameras:
			vIndex := slices.IndexFunc(s.cameras, func(c *Camera) bool { return c == newCamera })
			if vIndex == -1 {
				s.initCamera(newCamera)
				s.cameras = append(s.cameras, newCamera)
			}
			log.Printf("subscribed camera to scene")
		case oldCamera := <-s.oldCameras:
			vIndex := slices.IndexFunc(s.cameras, func(v *Camera) bool { return v == oldCamera })
			if vIndex != -1 {
				camera := s.cameras[vIndex]
				// camera.Count = camera.Count - 1
				s.cameras = slices.Delete(s.cameras, vIndex, vIndex+1)
				for _, o := range s.objects {
					if o.Camera == camera {
						s.oldObjects <- o
					}
				}
				log.Printf("unsubscribed camera from scene")
			}
		case e := <-s.inputs:
			for _, o := range s.objects {
				if o.SceneObject.Input != nil && o.Camera == e.source {
					o.SceneObject.Input(o, e.event)
				}
			}
		case o := <-s.newObjects:
			o.Id = s.nextId
			s.nextId = s.nextId + 1
			if o.SceneObject.Init != nil {
				o.SceneObject.Init(o)
			}
			s.objects[o.Id] = o
		case o := <-s.oldObjects:
			delete(s.objects, o.Id)
		default:
			break queue
		}
	}

	now := time.Now()
	deltaTime := now.Sub(s.lastTick)
	s.lastTick = now

	// Update all objects
	for _, o := range s.objects {
		if o.SceneObject.Update != nil {
			o.SceneObject.Update(o, deltaTime)
		}
	}
}

// Queue an input event
func (s *Scene) ReceiveInput(event *pb.InputEvent, source *Camera) {
	s.inputs <- struct {
		event  *pb.InputEvent
		source *Camera
	}{event: event, source: source}
}

func (s *Scene) Subscribe(camera *Camera) {
	s.newCameras <- camera
}

func (s *Scene) Unsubscribe(camera *Camera) {
	s.oldCameras <- camera
}

type SpawnArgs struct {
	Camera   *Camera
	Position compute.Point
	Rotation compute.Rotation
	Scale    compute.Point
	Parent   *SceneObjectInstance
	Data     map[string]any
	Hidden   bool
}

func (s *Scene) Spawn(o *SceneObject, args SpawnArgs) *SceneObjectInstance {
	obj := &SceneObjectInstance{
		SceneObject: o,
		Scene:       s,
		Parent:      args.Parent,
		Camera:      args.Camera,
		Data:        make(map[string]any),
		Position:    args.Position,
		Rotation:    args.Rotation,
		Scale:       compute.Point{X: 1, Y: 1, Z: 1},
		SpawnTime:   time.Now(),
		Hidden:      args.Hidden,

		model: compute.NewMatrix4(),

		matrix: compute.NewMatrix4(),
	}

	if args.Scale.X != 0 && args.Scale.Y != 0 {
		obj.Scale = args.Scale
	}

	// Make position relative to camera
	// if obj.Camera != nil {
	// 	obj.Position = compute.Point{
	// 		X: obj.Camera.Position.X + obj.Position.X,
	// 		Y: obj.Camera.Position.Y + obj.Position.Y,
	// 		Z: obj.Position.Z,
	// 	}
	// }

	if args.Data != nil {
		maps.Copy(obj.Data, args.Data)
	}

	s.newObjects <- obj

	return obj
}

func (s *Scene) ScanViewport(v Viewport) []*SceneObjectInstance {
	objs := make([]*SceneObjectInstance, 0)

	for _, obj := range s.objects {
		// size := obj.Size()
		// rect := compute.Rectangle2D{
		// 	Min: compute.Point{X: obj.Position.X, Y: obj.Position.Y},
		// 	Max: compute.Point{X: obj.Position.X + size.X, Y: obj.Position.Y + size.Y},
		// }

		if !obj.Hidden {
			objs = append(objs, obj)
		}

		// if v.Overlaps(rect) {
		// }
	}

	// Sort objects by z-index
	slices.SortFunc(objs, func(a *SceneObjectInstance, b *SceneObjectInstance) int {
		if a.SceneObject.UIElement && !b.SceneObject.UIElement {
			return 1
		}
		if !a.SceneObject.UIElement && b.SceneObject.UIElement {
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

	return objs
}

func (s *Scene) WithCamera(fn func(c *Camera)) {
	s.initCamera = fn
}

func NewBuffer(size int) []uint8 {
	return make([]uint8, 2+size*EncodeBufferSize)
}

func (s *Scene) String() string {
	str := "\nScene\n----\n"
	str = fmt.Sprintf("%sscene has %d objects\n", str, len(s.objects))
	return str
}
