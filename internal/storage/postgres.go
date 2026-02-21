package storage

import (
	"context"
	"fmt"
	"os"
	"time"

	"fast-ingest/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func (p *PostgresStore) Ping(ctx context.Context) error                       { return p.pool.Ping(ctx) }
func (p *PostgresStore) InsertEvent(ctx context.Context, e model.Event) error { /* ... */ return nil }
func (p *PostgresStore) InsertEvents(ctx context.Context, events []model.Event) error { /* ... */
	return nil
}

// NewPostgres initializes a new PostgresStore with a connection pool.
func NewPostgres(ctx context.Context) (*PostgresStore, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL not set")
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// Define defaults for our ingestion workload
	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute
	cfg.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Verify connection
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctxTimeout); err != nil {
		pool.Close()
		return nil, err
	}

	return &PostgresStore{pool: pool}, nil
}

func (p *PostgresStore) Close() {
	p.pool.Close()
}
