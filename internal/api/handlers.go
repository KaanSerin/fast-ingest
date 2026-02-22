package api

import (
	"encoding/json"
	"fast-ingest/internal/model"
	"fmt"
	"net/http"
	"strconv"
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
	r.Body = http.MaxBytesReader(w, r.Body, 20*1024*1024) // Limit request body to 20MB

	var events []model.Event
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // Strict decoding to catch unexpected fields

	if err := dec.Decode(&events); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	if len(events) == 0 {
		http.Error(w, "events is required", http.StatusBadRequest)
		return
	}

	if len(events) > 1000 {
		http.Error(w, "too many events (max 1000)", http.StatusBadRequest)
		return
	}

	// Validate each event in the batch
	for i := 0; i < len(events); i++ {
		e := events[i]
		// Validate required fields
		if e.EventName == "" || e.Channel == "" || e.UserID == "" || e.Timestamp == 0 {
			http.Error(w, fmt.Sprintf("invalid event at index %d", i), http.StatusBadRequest)
			return
		}
	}

	// Queue events for processing
	for i := 0; i < len(events); i++ {
		e := events[i]
		select {
		case s.Queue <- e:
		default:
			// queue full
			w.Header().Set("Retry-After", "1")
			http.Error(w, "ingest queue full", http.StatusTooManyRequests)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":   "accepted",
		"accepted": len(events),
	})
}

// HandleGetMetrics handles GET /metrics
// Returns aggregated metric data over a time range.
func (s *Server) HandleGetMetrics(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for from and to timestamps
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		http.Error(w, "from and to query parameters are required", http.StatusBadRequest)
		return
	}

	from, err := strconv.ParseInt(fromStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid from timestamp", http.StatusBadRequest)
		return
	}

	to, err := strconv.ParseInt(toStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid to timestamp", http.StatusBadRequest)
		return
	}

	var metricsDTO model.MetricsDTO

	// Get query parameters
	metricsDTO.EventName = r.URL.Query().Get("event_name")
	metricsDTO.GroupBy = r.URL.Query().Get("group_by")
	metricsDTO.From = from
	metricsDTO.To = to

	// Validate required fields
	if metricsDTO.EventName == "" {
		http.Error(w, "event_name is required", http.StatusBadRequest)
		return
	}

	if metricsDTO.GroupBy != "" && metricsDTO.GroupBy != "day" && metricsDTO.GroupBy != "hour" && metricsDTO.GroupBy != "channel" {
		http.Error(w, "invalid group_by value", http.StatusBadRequest)
		return
	}

	if metricsDTO.From >= metricsDTO.To {
		http.Error(w, "from must be before to", http.StatusBadRequest)
		return
	}

	if time.Unix(metricsDTO.From, 0).Before(time.Now().Add(-30 * 24 * time.Hour)) {
		http.Error(w, "from must be within the last 30 days", http.StatusBadRequest)
		return
	}

	// Retrieve metrics from the store
	metrics, err := s.Store.GetMetrics(r.Context(), metricsDTO)
	if err != nil {
		http.Error(w, "failed to retrieve metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(metrics)
}
