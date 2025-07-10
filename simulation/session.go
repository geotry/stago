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

	// The camera attached to this session
	Camera *scene.Camera

	buffer []byte

	Ticker *time.Ticker
	Closed chan struct{}

	readCount int

	objectsSent int
	instances   map[*scene.SceneObjectInstance]bool
}

func NewSession(id string, simulation *Simulation, camera *scene.Camera, context any) *Session {
	return &Session{
		Id:     id,
		sim:    simulation,
		Count:  1,
		buffer: make([]byte, 1024*1024),
		Ticker: time.NewTicker(time.Second / time.Duration(60)),
		Closed: make(chan struct{}),
		Camera: camera,

		objectsSent: 0,

		// objectsSent:   make(map[*scene.SceneObject]bool),
		instances: make(map[*scene.SceneObjectInstance]bool),
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

	offset += state.CopyCamera(s.buffer[offset:], s.Camera)
	offset += state.CopySceneObjectInstances(s.buffer[offset:])

	// offset += state.CopyCamera(s.buffer[offset:], s.Camera)
	// offset += state.CopySceneObjectInstances(s.buffer[offset:])

	s.readCount++

	return s.buffer[0:offset]
}

// func (s *Session) updateBuffer() {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	// offset := 0

// 	// if !s.paletteSent {
// 	// 	offset += s.sim.palette.Copy(s.buffer[offset:])
// 	// }
// 	// if !s.textureSent {
// 	// 	offset += s.sim.texture.Copy(s.buffer[offset:])
// 	// }

// 	// for _, obj := range s.Camera.Scene.Scan(s.Camera) {
// 	// 	if !s.objectsSent[obj.SceneObject] {
// 	// 		offset += s.sim.objects[obj.SceneObject].Copy(s.buffer[offset:])
// 	// 	}
// 	// 	if !s.instancesSent[obj] {
// 	// 		offset += s.sim.instances[obj].Copy(s.buffer[offset:])
// 	// 	}
// 	// }

// 	// if !s.cameraSent {
// 	// 	offset += s.sim.cameras[s.Camera].Copy(s.buffer[offset:])
// 	// }

// 	s.bufferSize = offset
// }

// func (s *Session) checkTexturePalette() bool {
// 	return s.paletteId != s.sim.rm.Palette0Id
// }

// func (s *Session) checkTextures() bool {
// 	return s.textures != len(s.sim.rm.TextureIndex)
// }

func (s *Session) encodeViewMatrices() {
	// check if they have changed since last tick
}

// Take a snapshot of the scene from the camera
// func (s *Session) snapshot() {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	objects := s.Camera.Scene.Scan(s.Camera)

// 	enc := s.encoder
// 	enc.Reset()

// 	enc.PutUint16(uint16(len(objects)))

// 	// Encode view matrices
// 	view3D, viewUI := s.Camera.ViewMatrix()

// 	enc.PutUint8(6)
// 	enc.PutUint8(0)
// 	enc.PutMatrix(viewUI)

// 	enc.PutUint8(6)
// 	enc.PutUint8(1)
// 	enc.PutMatrix(view3D)

// 	// Encode scene objects
// 	sceneObjects := make(map[int32]*scene.SceneObject)
// 	for _, obj := range objects {
// 		sceneObjects[obj.SceneObject.Id] = obj.SceneObject
// 	}
// 	for _, obj := range sceneObjects {
// 		// Encode vertex data (geometry, uv mapping, normals...) in a single block
// 		// It should be sent upfront like textures, then use ObjectID to reference it in other blocks
// 		enc.PutUint8(10)
// 		// Encode Object ID so webgl can index buffer position with ID and update partial buffer
// 		// without rewriting the whole buffer from zero
// 		// ObjectID is the SceneObject.Id, not SceneObjectInstance.Id
// 		enc.PutUint32(uint32(obj.Id))

// 		enc.PutUint16(6) // Number of vertices

// 		// Geometry: 3*4 bytes per vertex + 2 byte for number of vertices
// 		// For now, use a simple 1x1 quad but can support more advanced shapes
// 		enc.NewArray(6 * 3)
// 		enc.PutVector3Float32(1.0, 1.0, 0.0)
// 		enc.PutVector3Float32(-1.0, 1.0, 0.0)
// 		enc.PutVector3Float32(-1.0, -1.0, 0.0)
// 		enc.PutVector3Float32(1.0, 1.0, 0.0)
// 		enc.PutVector3Float32(-1.0, -1.0, 0.0)
// 		enc.PutVector3Float32(1.0, -1.0, 0.0)

// 		// UV Mapping
// 		// Use texture size / dimension to stretch correctly
// 		rx := float32(float64(obj.Size.X) / 128)
// 		ry := float32(float64(obj.Size.Y) / 128)

// 		enc.NewArray(6 * 2)
// 		// enc.PutVector2Float32(0.0, 0.0)
// 		// enc.PutVector2Float32(0.0, 0.0)
// 		// enc.PutVector2Float32(0.0, 0.0)
// 		// enc.PutVector2Float32(0.0, 0.0)
// 		// enc.PutVector2Float32(0.0, 0.0)
// 		// enc.PutVector2Float32(0.0, 0.0)
// 		enc.PutVector2Float32(rx, 0.0)
// 		enc.PutVector2Float32(0.0, 0.0)
// 		enc.PutVector2Float32(0.0, ry)
// 		enc.PutVector2Float32(rx, 0.0)
// 		enc.PutVector2Float32(0.0, ry)
// 		enc.PutVector2Float32(rx, ry)

// 		// TODO: Normals
// 	}

// 	// Encode models of instances
// 	for index, obj := range objects {
// 		// Texture (4 bytes) (deprecated)
// 		enc.PutUint8(encoding.TextureBlock)
// 		enc.PutUint16(uint16(index))
// 		enc.PutUint8(uint8(obj.SceneObject.Texture))

// 		// Matrix (67 bytes)
// 		enc.PutUint8(encoding.MatrixBlock)
// 		enc.PutUint16(uint16(index))
// 		enc.PutUint32(uint32(obj.SceneObject.Id))
// 		// View matrix index
// 		if obj.SceneObject.UIElement {
// 			enc.PutUint8(0)
// 		} else {
// 			enc.PutUint8(1)
// 		}
// 		enc.PutMatrix(obj.ModelMatrix())
// 	}

// 	s.objectCount = uint16(len(objects))
// }
