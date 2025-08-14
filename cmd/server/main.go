package main

import (
	"flag"
	"fmt"
	"log"

	"net/http"
	_ "net/http/pprof"

	"github.com/geotry/stago/server"
	"github.com/gorilla/websocket"
)

var port = flag.Int("port", 9090, "The websocket server port")

var upgrader = websocket.Upgrader{
	Subprotocols: []string{"render", "input"},
	CheckOrigin: func(r *http.Request) bool {
		return r.Host == fmt.Sprintf("localhost:%d", *port)
	},
}

func main() {
	flag.Parse()

	// Websocket server
	http.HandleFunc("/", wsHandler(server.NewWebsocketServer()))
	log.Printf("Websocket server listening at %v", fmt.Sprintf(":%d", *port))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		log.Fatalf("failed to serve websocket server: %v", err)
	}
}

func wsHandler(s *server.WebsocketServer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()

		if err := s.Handle(c); err != nil {
			log.Printf("%v", err)
			return
		}
	}
}
