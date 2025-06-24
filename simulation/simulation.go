package simulation

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/geotry/rass/scene"
)

type Simulation struct {
	time         time.Time
	currentScene *scene.Scene
	scenes       []*scene.Scene
	sessions     []*Session
	queue        chan *scene.Scene
	dequeue      chan *scene.Scene

	mu sync.Mutex
}

const TICKS_PER_SEC = 60

func NewSimulation() *Simulation {
	r := &Simulation{
		time:     time.Now(),
		scenes:   make([]*scene.Scene, 0),
		sessions: make([]*Session, 0),
		queue:    make(chan *scene.Scene, 10),
		dequeue:  make(chan *scene.Scene, 10),
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
	s.currentScene.Subscribe(camera)

	session := NewSession(sessionId, camera, struct{ UserId string }{UserId: userId})
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
			session.Camera.Scene.Unsubscribe(session.Camera)
			// session.ticker.Stop()
			// session.done <- struct{}{}
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
				s.time = time.Now()

				// if s.ticks%60 == 0 {
				// 	log.Printf("tick=%d scenes=%d viewers=%v objects=%v", s.ticks, len(s.cameras), len(s.objects))
				// }

				for _, scn := range s.scenes {
					scn.Tick()
				}

				// Update session snapshots
				for _, ss := range s.sessions {
					ss.snapshot()
				}
			}
		}
	}()
}
