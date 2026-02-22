package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

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
