package model

type Event struct {
	EventName  string         `json:"event_name"`
	Channel    string         `json:"channel"`
	CampaignID string         `json:"campaign_id"`
	UserID     string         `json:"user_id"`
	Timestamp  int64          `json:"timestamp"`
	Tags       []string       `json:"tags"`
	Metadata   map[string]any `json:"metadata"`
}
