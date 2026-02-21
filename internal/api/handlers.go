package api

import (
	"net/http"
)

// HandleIngestEvent handles POST /events
// Accepts and processes a single event payload.
func HandleIngestEvent(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement event ingestion logic here
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// HandleBulkIngestEvents handles POST /events/bulk
// Supports bulk ingestion of multiple event payloads.
func HandleBulkIngestEvents(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement bulk event ingestion logic here
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// HandleGetMetrics handles GET /metrics
// Returns aggregated metric data over a time range.
func HandleGetMetrics(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement metrics retrieval logic here
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
