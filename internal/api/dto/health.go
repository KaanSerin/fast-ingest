package api

type HealthCheckResponseDTO struct {
	Status      string `json:"status"`
	QueueLength int    `json:"queue_length"`
}

type HealthCheckErrorResponseDTO struct {
	Error string `json:"error"`
}
