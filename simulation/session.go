package simulation

import (
	"encoding/binary"
	"log"
	"sync"
	"time"

	"github.com/geotry/rass/scene"
)

type Session struct {
	Id string

	// Number of opened sessions
	Count int

	// The camera attached to this session
	Camera *scene.Camera

	// The buffer to store camera data at every tick
	buffer            []uint8
	bufferObjectCount uint16
	bufferOffset      int

	Ticker *time.Ticker
	Closed chan struct{}

	mu sync.RWMutex
}

func NewSession(id string, camera *scene.Camera, context any) *Session {
	return &Session{
		Id:           id,
		Count:        1,
		buffer:       scene.NewBuffer(1024),
		bufferOffset: 0,
		Ticker:       time.NewTicker(time.Second / time.Duration(60)),
		Closed:       make(chan struct{}),
		Camera:       camera,
	}
}

func (s *Session) SetFps(fps int) {
	if fps > 0 {
		s.Ticker.Reset(time.Second / time.Duration(fps))
	} else if fps == -1 {
		s.Ticker.Stop()
	}
}

// How many frames to buffer per subscriber
// Buffer should be around the same size of the render loop fps (60)
const BufferSize = 2

// Copy last frame in buffer
func (s *Session) Copy(buffer []uint8) (int, uint16) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	binary.BigEndian.PutUint16(buffer, s.bufferObjectCount)
	copy(buffer[2:], s.buffer[0:s.bufferOffset])
	return 2 + s.bufferOffset, s.bufferObjectCount
}

// Take a snapshot of the scene from the camera
func (s *Session) snapshot() {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := s.Camera

	bufOffset := 0
	objCount := 0

	for index, obj := range c.Scene.ScanViewport(c.Viewport()) {
		if obj.Camera != nil && obj.Camera != c {
			continue
		}
		if bufOffset < len(s.buffer) {
			bufOffset = obj.Encode(c, uint16(index), s.buffer)
			objCount++
		} else {
			log.Println("buffer is full, max objects in scene reached")
		}
	}

	s.bufferObjectCount = uint16(objCount)
	s.bufferOffset = bufOffset
}
