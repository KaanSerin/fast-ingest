package api

import "fast-ingest/internal/model"

type MetricsRequestDTO struct {
	EventName string `json:"event_name"`
	From      int64  `json:"from"`
	To        int64  `json:"to"`
	GroupBy   string `json:"group_by"`
}

type MetricsResponseDTO struct {
	Metrics model.Metrics `json:"metrics"`
}
