package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env.dev")
	if err != nil {
		log.Println("Error loading .env file")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, `DROP TABLE IF EXISTS events`)
	if err != nil {
		log.Fatalf("Failed to drop events table: %v", err)
	}
	log.Println("Dropped table: events")

	_, err = conn.Exec(ctx, `DELETE FROM schema_migrations`)
	if err != nil {
		log.Fatalf("Failed to delete schema_migrations rows: %v", err)
	}
	log.Println("Cleared schema_migrations")
}
