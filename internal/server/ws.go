package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type WSServer struct {
	mux        *http.ServeMux
	srv        *http.Server
	wsUpgrager *websocket.Upgrader
	clients    map[*websocket.Conn]struct{}
}

func NewWSServer() *WSServer {
	cfg := GetWSConfig()
	mux := http.NewServeMux()

	return &WSServer{
		mux: mux,
		srv: &http.Server{
			Addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		},
		wsUpgrager: &websocket.Upgrader{},
		clients:    make(map[*websocket.Conn]struct{}),
	}
}

func (ws *WSServer) Start() error {
	return ws.srv.ListenAndServe()
}
