package api

import (
	"fast-ingest/internal/model"
	"fast-ingest/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Server represents the API server with its dependencies.
type Server struct {
	Store storage.Store
	Queue chan model.Event
}

func NewServer(store storage.Store, queueSize int) *Server {
	return &Server{
		Store: store,
		Queue: make(chan model.Event, queueSize),
	}
}

// NewRouter sets up the API routes and returns a chi.Mux router.
func NewRouter(s *Server) *chi.Mux {
	r := chi.NewRouter()

	// Add middleware for logging and request ID generation
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	// Define API routes
	r.Get("/health", s.HandleHealthCheck)

	r.Post("/events", s.HandleIngestEvent)
	r.Post("/events/bulk", s.HandleBulkIngestEvents)

	r.Get("/metrics", s.HandleGetMetrics)

	return r
}
