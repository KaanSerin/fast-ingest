package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()

	// Add middleware for logging and request ID generation
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	// Define API routes
	r.Post("/events", HandleIngestEvent)
	r.Post("/events/bulk", HandleBulkIngestEvents)
	r.Get("/metrics", HandleGetMetrics)

	return r
}
