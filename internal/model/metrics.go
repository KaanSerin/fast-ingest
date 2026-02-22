package model

import "time"

type Metrics struct {
	EventName                string `json:"event_name"`
	From                     string `json:"from"`
	To                       string `json:"to"`
	TotalEvents              int64  `json:"total_events"`
	TotalUniqueEventsForUser int64  `json:"total_unique_events_for_user"`
	GroupBy                  string `json:"group_by,omitempty"`
	GroupBreakdown           any    `json:"group_breakdown,omitempty"`
}

type MetricsTotalsQueryResult struct {
	TotalEvents              int64
	TotalUniqueEventsForUser int64
}

type MetricsTimeGroupQueryResult struct {
	Bucket                   time.Time `json:"bucket"`
	TotalEvents              int64     `json:"total_events"`
	TotalUniqueEventsForUser int64     `json:"total_unique_events_for_user"`
}

type MetricsChannelGroupQueryResult struct {
	Channel                  string `json:"channel"`
	TotalEvents              int64  `json:"total_events"`
	TotalUniqueEventsForUser int64  `json:"total_unique_events_for_user"`
}
