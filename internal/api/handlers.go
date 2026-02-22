package api

import (
	"encoding/json"
	api "fast-ingest/internal/api/dto"
	"fast-ingest/internal/model"
	"fast-ingest/internal/storage"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Details any    `json:"details,omitempty"`
}

type SuccessResponse struct {
	Status string `json:"status"`
	Data   any    `json:"data,omitempty"`
}

func WriteError(w http.ResponseWriter, status int, msg string, details any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error:   msg,
		Details: details,
	})
}

func WriteSuccess(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(SuccessResponse{
		Status: "success",
		Data:   data,
	})
}

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

// HandleHealthCheck handles GET /health
// Checks database connectivity and returns queue length.
func (s *Server) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	// Testing DB connection
	err := s.Store.Ping(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "database connection failed", api.HealthCheckErrorResponseDTO{
			Error: err.Error(),
		})
		return
	}

	WriteSuccess(w, http.StatusOK, api.HealthCheckResponseDTO{
		Status:      "ok",
		QueueLength: len(s.Queue),
	})
}

// HandleIngestEvent handles POST /events
// Expects a single event payload in the request body.
func (s *Server) HandleIngestEvent(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 20*1024*1024) // Limit request body to 20MB

	var e model.Event
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // Strict decoding to catch unexpected fields

	if err := dec.Decode(&e); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid JSON payload", nil)
		return
	}

	// Validate required fields
	if e.EventName == "" || e.Channel == "" || e.UserID == "" || e.Timestamp == 0 {
		WriteError(w, http.StatusBadRequest, "missing required fields", nil)
		return
	}

	select {
	case s.Queue <- e:
		WriteSuccess(w, http.StatusAccepted, api.EventResponseDTO{
			Instant: time.Now().UTC().Format(time.RFC3339),
		})
	default:
		// queue full
		w.Header().Set("Retry-After", "1")
		WriteError(w, http.StatusTooManyRequests, "ingest queue full", nil)
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
		WriteError(w, http.StatusBadRequest, "invalid JSON payload", nil)
		return
	}

	if len(events) == 0 {
		WriteError(w, http.StatusBadRequest, "events is required", nil)
		return
	}

	if len(events) > 1000 {
		WriteError(w, http.StatusBadRequest, "too many events (max 1000)", nil)
		return
	}

	// Validate each event in the batch
	for i := 0; i < len(events); i++ {
		e := events[i]
		// Validate required fields
		if e.EventName == "" || e.Channel == "" || e.UserID == "" || e.Timestamp == 0 {
			WriteError(w, http.StatusBadRequest, fmt.Sprintf("invalid event at index %d", i), nil)
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
			WriteError(w, http.StatusTooManyRequests, "ingest queue full", nil)
			return
		}
	}

	WriteSuccess(w, http.StatusAccepted, api.EventsBulkResponseDTO{
		Accepted: len(events),
	})
}

// HandleGetMetrics handles GET /metrics
// Returns aggregated metric data over a time range.
func (s *Server) HandleGetMetrics(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for from and to timestamps
	fromStr := r.URL.Query().Get("from")
	if fromStr == "" {
		WriteError(w, http.StatusBadRequest, "from query parameter is required", nil)
		return
	}

	from, err := strconv.ParseInt(fromStr, 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid from timestamp", nil)
		return
	}

	toStr := r.URL.Query().Get("to")
	var to int64
	// Default 'to' to current time if not provided
	if toStr == "" {
		to = time.Now().Unix()
	} else {
		var err error
		to, err = strconv.ParseInt(toStr, 10, 64)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid to timestamp", nil)
			return
		}
	}

	var metricsDTO api.MetricsRequestDTO

	// Get query parameters
	metricsDTO.EventName = r.URL.Query().Get("event_name")
	metricsDTO.GroupBy = r.URL.Query().Get("group_by")
	metricsDTO.From = from
	metricsDTO.To = to

	// Validate required fields
	// A validation library could be used here for more complex validation rules
	if metricsDTO.EventName == "" {
		WriteError(w, http.StatusBadRequest, "event_name is required", nil)
		return
	}

	if metricsDTO.GroupBy != "" && metricsDTO.GroupBy != "day" && metricsDTO.GroupBy != "hour" && metricsDTO.GroupBy != "channel" {
		WriteError(w, http.StatusBadRequest, "invalid group_by value", nil)
		return
	}

	if metricsDTO.From >= metricsDTO.To {
		WriteError(w, http.StatusBadRequest, "from must be before to", nil)
		return
	}

	if time.Unix(metricsDTO.From, 0).Before(time.Now().Add(-30 * 24 * time.Hour)) {
		WriteError(w, http.StatusBadRequest, "from must be within the last 30 days", nil)
		return
	}

	// Retrieve metrics from the store
	metrics, err := s.Store.GetMetrics(r.Context(), metricsDTO)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to retrieve metrics", nil)
		return
	}

	WriteSuccess(w, http.StatusOK, api.MetricsResponseDTO{
		Metrics: metrics,
	})
}
