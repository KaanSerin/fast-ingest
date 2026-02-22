package storage

import (
	"context"

	api "fast-ingest/internal/api/dto"
	"fast-ingest/internal/model"
)

// Store defines the interface for our storage layer, abstracting away the underlying implementation (e.g., Postgres, file system, etc.).
type Store interface {
	// Ping verifies the backing store is reachable.
	Ping(ctx context.Context) error

	// InsertEvent persists a single raw event.
	InsertEvent(ctx context.Context, e model.Event) error

	// InsertEvents persists a batch of raw events (preferred path for ingestion).
	InsertEvents(ctx context.Context, events []model.Event) error

	// GetMetrics retrieves aggregated metrics based on the provided filters and grouping.
	GetMetrics(ctx context.Context, metricsDTO api.MetricsRequestDTO) (model.Metrics, error)

	// Close releases resources (db connections, file handles, etc.).
	Close()
}
