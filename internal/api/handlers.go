package api

import (
	"encoding/json"
	"fast-ingest/internal/model"
	"net/http"
	"time"
)

// HandleIngestEvent handles POST /events
// Expects a single event payload in the request body.
func (s *Server) HandleIngestEvent(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 20*1024*1024) // Limit request body to 20MB

	var e model.Event
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // Strict decoding to catch unexpected fields

	if err := dec.Decode(&e); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if e.EventName == "" || e.Channel == "" || e.UserID == "" || e.Timestamp == 0 {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	select {
	case s.Queue <- e:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "accepted",
			"at":     time.Now().UTC().Format(time.RFC3339),
		})
	default:
		// queue full
		w.Header().Set("Retry-After", "1")
		http.Error(w, "ingest queue full", http.StatusTooManyRequests)
	}
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
