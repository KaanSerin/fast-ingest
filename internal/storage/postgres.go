package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"fast-ingest/internal/helpers"
	"fast-ingest/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func (p *PostgresStore) Ping(ctx context.Context) error { return p.pool.Ping(ctx) }

func (p *PostgresStore) InsertEvents(ctx context.Context, events []model.Event) error { /* ... */
	return nil
}

func (p *PostgresStore) InsertEvent(ctx context.Context, e model.Event) error {

	// Postgres expects timestamps in UTC, so we convert to UTC before inserting.
	t := time.Unix(e.Timestamp, 0).UTC()

	// Marshal tags and metadata to JSON for storage in jsonb columns.
	tagsJSON, _ := json.Marshal(e.Tags)
	metaJSON, _ := json.Marshal(e.Metadata)

	_, err := p.pool.Exec(ctx, `
		INSERT INTO events (dedupe_key, event_name, channel, campaign_id, user_id, ts, tags, metadata)
VALUES ($1,$2,$3,$4,$5,$6,$7::jsonb,$8::jsonb)
ON CONFLICT (dedupe_key) DO NOTHING;
	`, helpers.DedupeKey(e), e.EventName, e.Channel, helpers.NullIfEmpty(e.CampaignID), e.UserID, t, tagsJSON, metaJSON)

	return err
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
