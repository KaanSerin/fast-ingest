package model

type MetricsDTO struct {
	EventName string `json:"event_name"`
	From      int64  `json:"from"`
	To        int64  `json:"to"`
	GroupBy   string `json:"group_by"`
}
