package api

type EventResponseDTO struct {
	Instant string `json:"instant"`
}

type EventsBulkResponseDTO struct {
	Accepted int `json:"accepted"`
}
