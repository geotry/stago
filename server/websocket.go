package server

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/geotry/rass/pb"
	"github.com/geotry/rass/scenes"
	"github.com/geotry/rass/simulation"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/encoding/protojson"
)

// Create scene and renderer
var scene, rm = scenes.NewDemo()
var simu = simulation.NewSimulation(rm)

type WebsocketServer struct{}

func NewWebsocketServer() *WebsocketServer {
	simu.AddScene(scene)
	simu.Start(context.Background())

	return &WebsocketServer{}
}

func (s *WebsocketServer) Handle(c *websocket.Conn) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	protocol := c.Subprotocol()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			return err
		}

		switch protocol {
		case "render":
			go s.HandleRender(ctx, c, message)
		case "input":
			go s.HandleInput(ctx, c, message)
		}
	}
}

func (s *WebsocketServer) HandleRender(ctx context.Context, c *websocket.Conn, in []byte) error {
	var req pb.RenderRequest

	if err := protojson.Unmarshal(in, &req); err != nil {
		log.Println("[render] json unmarshal error:", err)
		return err
	}

	if req.SessionId == "" {
		return nil
	}

	// Get session
	session, newSession := simu.OpenSession(req.SessionId, req.UserId)
	defer simu.CloseSession(session.Id)

	// Set frame rate
	session.SetFps(int(req.Fps))
	// session.SetFps(int(1))

	// Update camera settings
	if req.Width > 0 && req.Height > 0 {
		session.Camera.SetSize(int(req.Width), int(req.Height))
	}
	if req.Fov > 0 {
		session.Camera.Fov = float64(req.Fov) * (math.Pi / 180)
	}
	if req.Near > 0 {
		session.Camera.Near = float64(req.Near)
	}
	if req.Far > 0 {
		session.Camera.Far = float64(req.Far)
	}

	log.Printf("[render] session_id=%s[%d] near=%.2f far=%.2f fov=%.2f",
		session.Id,
		session.Count,
		session.Camera.Near,
		session.Camera.Far,
		session.Camera.Fov*180.0/math.Pi,
	)

	// Stop here for existing session
	if !newSession {
		return nil
	}

	go func() {
		for {
			select {
			case <-session.Closed:
				return
			case <-session.Ticker.C:
				if err := c.WriteMessage(websocket.BinaryMessage, session.Render()); err != nil {
					log.Printf("error=%v", err)
				}

				if err := ctx.Err(); err != nil {
					if err.Error() != "context canceled" {
						log.Printf("error=%v", err)
					} else {
						log.Print("[ws] client disconnected")
					}
					session.Ticker.Stop()
					session.Closed <- struct{}{}
				}
			}
		}
	}()

	<-session.Closed

	return nil
}

func (s *WebsocketServer) HandleInput(ctx context.Context, c *websocket.Conn, in []byte) error {
	var req pb.InputEvent

	if err := protojson.Unmarshal(in, &req); err != nil {
		log.Println("[input] json unmarshal error:", err)
		return err
	}

	session := simu.GetSession(req.SessionId)
	if session == nil {
		log.Println("[input] invalid session ", req.SessionId)
		return fmt.Errorf("session does not exist")
	}

	scene.ReceiveInput(&req, session.Camera)

	return nil
}
