# fast-ingest

A high-throughput event ingestion API backed by PostgreSQL. Events are queued in-memory and flushed to the database via a background worker.

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

## Postman
The postman collection can be found <a href="https://documenter.getpostman.com/view/8807961/2sBXcEkg4g">here</a>

## TODO

Ultimately, the technology used and not considered were decided with the time limit and my experience with them in mind. Given enough resources, those technologies can also be utilized. The project has a lot of room for improvement with a solid base already built.

### Next Improvements
* Add proper request rate limiting per client and/or channel.
* Implement proper graceful shutdown draining the queue before completely exiting the program.
* Add pagination for large metrics requests.
* Add Swagger documentation.

### Trade-offs

* Used an in-memory queue instead of an external streaming system (Kafka) to simplify implementation
* Calculated metrics by querying the events table rather than maintaining a pre-aggregated table.
* Used manual SQL queries (pgx) instead of a more sophisticated data-access abstraction layer.

These trade offs were made considering the time constraint of the project.

### Alternative Approaches Considered

* Using Kafka between ingestion and persistence for better durability and replay capability.
* Using ClickHouse for faster aggregations at high scale.
* Maintaining a separate metrics table updated asynchronously for faster metric queries.
* Using Redis as a short-term write buffer or rate limiter.
* Using an ORM (GORM), however used pgx for better control over performance and batching.

### What I Would Do Differently in Production

1. Introduce a streaming layer like Kafka between ingestion and storage.
2. Use ClickHouse for metrics queries.
3. Separate read and write database workloads.
5. Implement a Dead-letter queue for failed events
6. Consider pre-aggregated tables for heavy metrics queries.