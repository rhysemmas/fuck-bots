package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Server is a struct which contains an http server and logger
type Server struct {
	*http.Server
	logger *zap.SugaredLogger
}

// NewServer returns a server which can be started to run an http server
func NewServer(address string, routes http.Handler, logger *zap.SugaredLogger) *Server {
	srv := &http.Server{
		Addr:         address,
		Handler:      routes,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}

	return &Server{
		Server: srv,
		logger: logger,
	}
}

// Start will start the http server
func (s *Server) Start(errorCh chan error) func(context.Context) {
	go func() {
		err := s.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			errorCh <- err
		}
	}()

	return func(ctx context.Context) {
		s.Shutdown(ctx)
	}
}
