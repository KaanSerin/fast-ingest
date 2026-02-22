include .env.dev
export

run:
	go run ./cmd/server/main.go

migrate-up:
	go run ./cmd/migrate/main.go

migrate-down:
	go run ./cmd/migrate-down/main.go

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f server