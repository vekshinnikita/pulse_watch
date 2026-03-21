package server

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type WebServer struct {
	httpServer *http.Server
}

func (s *WebServer) Run(
	handler http.Handler,
) error {
	config := GetWebConfig()

	s.httpServer = &http.Server{
		Addr:           fmt.Sprintf(":%v", config.Port),
		Handler:        handler,
		MaxHeaderBytes: 1 << 20, // 1 MB
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
	}
	return s.httpServer.ListenAndServe()
}

func (s *WebServer) Shutdown(ctx context.Context) error {

	return s.httpServer.Shutdown(ctx)
}

func NewWebServer() *WebServer {
	return &WebServer{}
}
