package simplews

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Server struct {
	Handler    func(http.ResponseWriter, *http.Request)
	Connection *websocket.Conn
}

func New() *Server {
	server := &Server{}
	log.Println("make server")
	handler := func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(w, "Not a websocket handshake", 400)
			return
		} else if err != nil {
			log.Println(err)
			return
		}
		// close previous connection
		if server.Connection != nil {
			server.Connection.Close()
		}
		server.Connection = conn
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				log.Println("read error: ", err)
				return
			}
		}
	}
	server.Handler = handler
	log.Println("return handler")
	return server
}

func (s *Server) WriteJSON(data interface{}) {
	if s.Connection != nil {
		s.Connection.WriteJSON(data)
	}
}

func (s *Server) WriteString(data string) {
	if s.Connection != nil {
		s.Connection.WriteMessage(websocket.TextMessage, []byte(data))
	}
}

func Serve() *Server {
	ws := New()
	go ws.Serve()
	return ws
}

func (s *Server) Serve() {
	http.HandleFunc("/ws", s.Handler)
	log.Fatal(http.ListenAndServe(":6060", nil))
}

func reader(ws *websocket.Conn) {
	defer ws.Close()

	pongWait := 60 * 60 * time.Second
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}
