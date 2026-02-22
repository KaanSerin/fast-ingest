package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"fast-ingest/internal/helpers"
	"fast-ingest/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func (p *PostgresStore) Ping(ctx context.Context) error { return p.pool.Ping(ctx) }

func (p *PostgresStore) InsertEvents(ctx context.Context, events []model.Event) error {
	log.Printf("Inserting batch of %d events", len(events))

	// Using a transaction with pgx.Batch to bulk insert events
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	for _, e := range events {
		t := time.Unix(e.Timestamp, 0).UTC()
		tagsJSON, _ := json.Marshal(e.Tags)
		metaJSON, _ := json.Marshal(e.Metadata)

		batch.Queue(`
			INSERT INTO events (dedupe_key, event_name, channel, campaign_id, user_id, ts, tags, metadata)
			VALUES ($1,$2,$3,$4,$5,$6,$7::jsonb,$8::jsonb)
			ON CONFLICT (dedupe_key) DO NOTHING;
		`, helpers.DedupeKey(e), e.EventName, e.Channel, helpers.NullIfEmpty(e.CampaignID), e.UserID, t, tagsJSON, metaJSON)
	}

	start := time.Now()

	br := tx.SendBatch(ctx, batch)
	if err := br.Close(); err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	log.Printf("Batch insert took %d ms", time.Since(start).Milliseconds())

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

func (p *PostgresStore) GetMetrics(ctx context.Context, metricsDTO model.MetricsDTO) (model.Metrics, error) {
	from := time.Unix(metricsDTO.From, 0).UTC()
	to := time.Unix(metricsDTO.To, 0).UTC()

	var metrics model.Metrics = model.Metrics{
		EventName: metricsDTO.EventName,
		From:      from.Format(time.RFC3339),
		To:        to.Format(time.RFC3339),
		GroupBy:   metricsDTO.GroupBy,
	}

	totalsQueryResult, err := p.getTotalsQuery(metricsDTO)
	if err != nil {
		return model.Metrics{}, err
	}

	metrics.TotalEvents = totalsQueryResult.TotalEvents
	metrics.TotalUniqueEventsForUser = totalsQueryResult.TotalUniqueEventsForUser

	// If group_by is specified, we need to run a separate query to get the breakdown by group.
	if metricsDTO.GroupBy == "day" || metricsDTO.GroupBy == "hour" {
		groupQueryResults, err := p.getTimeGroupQuery(metricsDTO)
		if err == nil {
			metrics.GroupBreakdown = groupQueryResults
		}
	} else if metricsDTO.GroupBy == "channel" {
		channelGroupQueryResults, err := p.getChannelGroupQuery(metricsDTO)
		if err == nil {
			metrics.GroupBreakdown = channelGroupQueryResults
		}
	}

	return metrics, nil
}

func (p *PostgresStore) Close() {
	p.pool.Close()
}

func (p *PostgresStore) getTotalsQuery(metricsDTO model.MetricsDTO) (model.MetricsTotalsQueryResult, error) {
	from := time.Unix(metricsDTO.From, 0).UTC()
	to := time.Unix(metricsDTO.To, 0).UTC()

	var totalsQueryResult model.MetricsTotalsQueryResult
	totalsQuery := `SELECT
COUNT(*) AS total_events,
COUNT(DISTINCT user_id) AS total_unique_events_for_user
FROM events
WHERE event_name = $1
AND ts >= $2 AND ts < $3;`
	row := p.pool.QueryRow(context.Background(), totalsQuery, metricsDTO.EventName, from, to)
	if err := row.Scan(&totalsQueryResult.TotalEvents, &totalsQueryResult.TotalUniqueEventsForUser); err != nil {
		return model.MetricsTotalsQueryResult{}, err
	}

	return totalsQueryResult, nil
}

func (p *PostgresStore) getTimeGroupQuery(metricsDTO model.MetricsDTO) ([]model.MetricsTimeGroupQueryResult, error) {
	from := time.Unix(metricsDTO.From, 0).UTC()
	to := time.Unix(metricsDTO.To, 0).UTC()

	groupQuery := `SELECT
DATE_TRUNC($1, ts) AS bucket,
COUNT(*) AS total_count,
COUNT(DISTINCT user_id) AS total_unique_event_for_user_count
FROM events
WHERE event_name = $2
AND ts >= $3 AND ts < $4
GROUP BY bucket
ORDER BY bucket;`
	rows, err := p.pool.Query(context.Background(), groupQuery, metricsDTO.GroupBy, metricsDTO.EventName, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.MetricsTimeGroupQueryResult
	for rows.Next() {
		var r model.MetricsTimeGroupQueryResult
		if err := rows.Scan(&r.Bucket, &r.TotalEvents, &r.TotalUniqueEventsForUser); err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (p *PostgresStore) getChannelGroupQuery(metricsDTO model.MetricsDTO) ([]model.MetricsChannelGroupQueryResult, error) {
	from := time.Unix(metricsDTO.From, 0).UTC()
	to := time.Unix(metricsDTO.To, 0).UTC()

	groupQuery := `SELECT
channel,
COUNT(*) AS total_count,
COUNT(DISTINCT user_id) AS total_unique_event_for_user_count
FROM events
WHERE event_name = $1
AND ts >= $2 AND ts < $3
GROUP BY channel
ORDER BY channel;`
	rows, err := p.pool.Query(context.Background(), groupQuery, metricsDTO.EventName, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.MetricsChannelGroupQueryResult
	for rows.Next() {
		var r model.MetricsChannelGroupQueryResult
		if err := rows.Scan(&r.Channel, &r.TotalEvents, &r.TotalUniqueEventsForUser); err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
