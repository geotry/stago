package simulation

import (
	"time"

	"github.com/geotry/rass/scene"
)

type Session struct {
	Id string

	// Number of opened sessions
	Count int

	sim *Simulation

	// The root object attached to this session (the camera)
	Root *scene.SceneObjectInstance

	buffer []byte

	Ticker *time.Ticker
	Closed chan struct{}

	readCount int

	objectsSent int
	instances   map[*scene.SceneObjectInstance]bool
}

func NewSession(id string, simulation *Simulation, root *scene.SceneObjectInstance, context any) *Session {
	return &Session{
		Id:     id,
		sim:    simulation,
		Count:  1,
		buffer: make([]byte, 1024*1024),
		Ticker: time.NewTicker(time.Second / time.Duration(60)),
		Closed: make(chan struct{}),
		Root:   root,

		objectsSent: 0,
		instances:   make(map[*scene.SceneObjectInstance]bool),
	}
}

func (s *Session) RenderCount() int {
	return s.readCount
}

func (s *Session) SetFps(fps int) {
	if fps > 0 {
		s.Ticker.Reset(time.Second / time.Duration(fps))
	} else if fps == -1 {
		s.Ticker.Stop()
	}
}

func (s *Session) Render() []byte {
	state := s.sim.state
	offset := 0

	stateObjectsCount := len(state.sceneObjects)

	if s.readCount == 0 {
		offset += state.CopyTextures(s.buffer[offset:])
	}

	if stateObjectsCount != s.objectsSent {
		offset += state.CopySceneObjects(s.buffer[offset:])
		s.objectsSent = stateObjectsCount
	}

	// Check new and old objects
	offset += state.CopyCamera(s.buffer[offset:], s.Root.Id)
	offset += state.CopyLights(s.buffer[offset:])
	offset += state.CopyLightsDeleted(s.buffer[offset:])
	offset += state.CopySceneObjectInstances(s.buffer[offset:])
	offset += state.CopySceneObjectInstancesDeleted(s.buffer[offset:])

	s.readCount++

	return s.buffer[0:offset]
}
