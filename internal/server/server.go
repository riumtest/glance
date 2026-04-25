package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/glanceapp/glance/internal/config"
)

const (
	// DefaultReadTimeout is the maximum duration for reading the entire request.
	DefaultReadTimeout = 10 * time.Second

	// DefaultWriteTimeout is the maximum duration before timing out writes of the response.
	// Increased from 30s to 60s to account for slower widget data fetches on my home server.
	DefaultWriteTimeout = 60 * time.Second

	// DefaultIdleTimeout is the maximum amount of time to wait for the next request.
	DefaultIdleTimeout = 120 * time.Second
)

// Server wraps the HTTP server and holds application state.
type Server struct {
	httpServer *http.Server
	config     *config.Config
}

// New creates a new Server instance with the provided configuration.
func New(cfg *config.Config) *Server {
	mux := http.NewServeMux()

	s := &Server{
		config: cfg,
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
			Handler:      mux,
			ReadTimeout:  DefaultReadTimeout,
			WriteTimeout: DefaultWriteTimeout,
			IdleTimeout:  DefaultIdleTimeout,
		},
	}

	s.registerRoutes(mux)

	return s
}

// registerRoutes sets up the HTTP routes for the server.
func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/health", s.handleHealth)
}

// Start begins listening for incoming HTTP connections.
func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.httpServer.Addr)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// Shutdown gracefully stops the server, waiting for active connections to finish.
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server...")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	log.Println("Server stopped")
	return nil
}

// handleIndex serves the main dashboard page.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	// TODO: render dashboard template
	_, _ = w.Write([]byte("<html><body><h1>Glance Dashboard</h1></body></html>"))
}

// handleHealth returns a simple health check response.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
