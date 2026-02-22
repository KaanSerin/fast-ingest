# fast-ingest

A high-throughput event ingestion API backed by PostgreSQL. Events are queued in-memory and flushed to the database via a background worker.

## API

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/events` | Ingest a single event |
| `POST` | `/events/bulk` | Ingest up to 1000 events |
| `GET` | `/metrics` | Query aggregated metrics |

### POST /events

```json
{
  "event_name": "page_view",
  "channel": "web",
  "user_id": "user_123",
  "timestamp": 1769904000
}
```

### POST /events/bulk

```json
[
  { "event_name": "page_view", "channel": "web", "user_id": "user_123", "timestamp": 1769904000 },
  { "event_name": "click",     "channel": "web", "user_id": "user_456", "timestamp": 1769904060 }
]
```

### GET /metrics

Query parameters:

| Param | Required | Description |
|-------|----------|-------------|
| `event_name` | yes | Event name to query |
| `from` | yes | Unix timestamp (start, inclusive) |
| `to` | yes | Unix timestamp (end, exclusive) |
| `group_by` | no | `day`, `hour`, or `channel` |

```
GET /metrics?event_name=page_view&from=1769904000&to=1771753229&group_by=day
```

---

## Development (local)

### Prerequisites

- Go 1.24+
- PostgreSQL 16+

### Setup

1. Copy the example env file and fill in values:

```bash
cp .env.example .env.dev
```

```bash
# .env.dev
PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/fastingest?sslmode=disable
```

2. Run migrations:

```bash
make migrate-up
```

3. Start the server:

```bash
make run
```

### Other commands

```bash
make migrate-down   # drop events table and clear schema_migrations
```

---

## Docker

### Prerequisites

- Docker
- Docker Compose

### Setup

No additional configuration needed — environment variables are defined in `docker-compose.yml`.

### Run

```bash
docker compose up --build
```

Services start in order: **db** → **migrate** → **server**. The server is available at `http://localhost:8080`.

### Stop

```bash
docker compose down
```

### Logs

```bash
docker compose logs -f server
```