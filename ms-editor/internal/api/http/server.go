package http

import (
	"context"
	"net/http"
	"time"
)

const (
	defaultReadTimeout  = 10 * time.Second
	defaultWriteTimeout = 10 * time.Second
	defaultIdleTimeout  = 60 * time.Second
)

type Server struct {
	httpServer *http.Server
}

func NewServer(address string, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         address,
			Handler:      handler,
			ReadTimeout:  defaultReadTimeout,
			WriteTimeout: defaultWriteTimeout,
			IdleTimeout:  defaultIdleTimeout,
		},
	}
}

func (server *Server) Start() error {
	return server.httpServer.ListenAndServe()
}

func (server *Server) Shutdown(ctx context.Context) error {
	return server.httpServer.Shutdown(ctx)
}
