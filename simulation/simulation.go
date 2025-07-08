package simulation

import (
	"context"
	"image/png"
	"log"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/geotry/rass/rendering"
	"github.com/geotry/rass/scene"
)

type Simulation struct {
	currentScene *scene.Scene
	rm           *rendering.ResourceManager
	state        *State
	scenes       []*scene.Scene
	sessions     []*Session
	queue        chan *scene.Scene
	dequeue      chan *scene.Scene
	ticker       *scene.Ticker
	bench        *scene.Ticker
	mu           sync.Mutex
}

const TICKS_PER_SEC = 60

func NewSimulation(rm *rendering.ResourceManager) *Simulation {
	r := &Simulation{
		rm:       rm,
		state:    NewState(),
		scenes:   make([]*scene.Scene, 0),
		sessions: make([]*Session, 0),
		queue:    make(chan *scene.Scene, 10),
		dequeue:  make(chan *scene.Scene, 10),
		ticker:   scene.NewTicker(),
		bench:    scene.NewTicker(),
	}
	return r
}

func (s *Simulation) AddScene(scene *scene.Scene) {
	s.queue <- scene
}

func (s *Simulation) RemoveScene(scene *scene.Scene) {
	s.dequeue <- scene
}

// Create or return existing session. Second value returns true if session was created.
func (s *Simulation) OpenSession(sessionId string, userId string) (*Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sIndex := slices.IndexFunc(s.sessions, func(ss *Session) bool { return ss.Id == sessionId })
	if sIndex != -1 {
		session := s.sessions[sIndex]
		session.Count++
		return session, false
	}

	if s.currentScene == nil {
		return nil, false
	}

	camera := scene.NewCamera(s.currentScene)
	s.currentScene.AddCamera(camera)

	session := NewSession(sessionId, s, camera, struct{ UserId string }{UserId: userId})
	s.sessions = append(s.sessions, session)

	return session, true
}

func (s *Simulation) GetSession(sessionId string) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	sIndex := slices.IndexFunc(s.sessions, func(ss *Session) bool { return ss.Id == sessionId })
	if sIndex != -1 {
		return s.sessions[sIndex]
	}

	return nil
}

func (s *Simulation) CloseSession(sessionId string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	sIndex := slices.IndexFunc(s.sessions, func(ss *Session) bool { return ss.Id == sessionId })
	if sIndex != -1 {
		session := s.sessions[sIndex]
		session.Count--
		if session.Count <= 0 {
			session.Camera.Scene.RemoveCamera(session.Camera)
			s.state.DeleteCamera(session.Camera)
			s.sessions = slices.Delete(s.sessions, sIndex, sIndex+1)
		}
		return true
	}
	return false
}

// Starts the main loop
func (s *Simulation) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Second / time.Duration(TICKS_PER_SEC))

	go func() {
		for {
			select {
			case scn := <-s.queue:
				if s.currentScene == nil {
					s.currentScene = scn
				}
				s.scenes = append(s.scenes, scn)
			case scn := <-s.dequeue:
				if s.currentScene == scn {
					s.currentScene = nil
				}
				s.scenes = slices.DeleteFunc(s.scenes, func(dscn *scene.Scene) bool { return dscn == scn })
			case <-ctx.Done():
				return
			case <-ticker.C:
				tick, _ := s.ticker.Tick()

				s.bench.Reset()
				for _, scn := range s.scenes {
					scn.Update()
				}
				_, updateTime := s.bench.Tick()

				s.bench.Reset()
				s.saveState()
				_, saveTime := s.bench.Tick()

				// To debug textures stored in state
				if tick == 1 {
					palette, _ := s.state.GetTextureRGBA(1)
					diffuse, _ := s.state.GetTexturePaletted(2, palette)
					specular, _ := s.state.GetTextureGrayScale(3)
					f1, _ := os.Create(".out/palette.png")
					png.Encode(f1, palette)
					f2, _ := os.Create(".out/diffuse.png")
					png.Encode(f2, diffuse)
					f3, _ := os.Create(".out/specular.png")
					png.Encode(f3, specular)
				}

				if tick%TICKS_PER_SEC == 0 {
					bufferSize := s.state.buffer.Offset()
					bufferCapacity := s.state.buffer.Capacity()
					bufferUsage := (float64(bufferSize) / float64(bufferCapacity)) * 100.0
					log.Printf("tick=%d buffer_size=%.3fMb usage=%.2f%% blocks=%d update=%vμs save=%vμs", tick, float64(bufferSize)/float64(MiB), bufferUsage, s.state.buffer.BlockCount(), updateTime.Microseconds(), saveTime.Microseconds())
				}
			}
		}
	}()
}

func (s *Simulation) saveState() {
	s.state.WriteTextureOnce(s.rm.Palette)
	s.state.WriteTextureGroupOnce(s.rm.Diffuse)
	s.state.WriteTextureGroupOnce(s.rm.Specular)

	for _, obj := range s.currentScene.NewObjects {
		log.Printf("added object %v", obj)
	}

	for _, obj := range s.currentScene.OldObjects {
		b := s.state.sceneObjectInstances[obj.Id]
		// b.Free()
		delete(s.state.sceneObjectInstances, obj.Id)
		log.Printf("deleted object %v (%v)", obj, b)
	}

	for _, obj := range s.currentScene.Objects() {
		s.state.WriteSceneObjectOnce(obj.SceneObject)
		s.state.WriteSceneObjectInstance(obj)
	}

	for _, session := range s.sessions {
		s.state.WriteCamera(session.Camera)
	}

	// todo: compact buffer to reclaim free space by shifting offsets
}
