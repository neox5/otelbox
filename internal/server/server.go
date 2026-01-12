package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server provides HTTP server for Prometheus metrics.
type Server struct {
	addr   string
	path   string
	server *http.Server
	mux    *http.ServeMux
}

// New creates a new HTTP server.
func New(port int, path string, registry *prometheus.Registry) *Server {
	mux := http.NewServeMux()

	// Register Prometheus handler
	mux.Handle(path, promhttp.HandlerFor(
		registry,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))

	addr := fmt.Sprintf(":%d", port)

	return &Server{
		addr: addr,
		path: path,
		mux:  mux,
		server: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
}

// Start begins serving HTTP requests.
func (s *Server) Start(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		slog.Info("starting server", "addr", s.addr, "path", s.path)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return s.shutdown()
	}
}

// shutdown gracefully stops the server.
func (s *Server) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	slog.Info("shutting down server")
	return s.server.Shutdown(ctx)
}
