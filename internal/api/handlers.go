package api

import (
	"net/http"
)

// HandleIngestEvent handles POST /events
// Expects a single event payload in the request body.
func (s *Server) HandleIngestEvent(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement event ingestion logic here
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// HandleBulkIngestEvents handles POST /events/bulk
// Supports bulk ingestion of multiple event payloads.
func (s *Server) HandleBulkIngestEvents(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement bulk event ingestion logic here
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// HandleGetMetrics handles GET /metrics
// Returns aggregated metric data over a time range.
func (s *Server) HandleGetMetrics(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement metrics retrieval logic here
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
