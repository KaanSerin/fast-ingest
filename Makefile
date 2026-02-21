include .env.dev
export

run:
	go run ./cmd/server/main.go

migrate-up:
	go run ./cmd/migrate/main.go