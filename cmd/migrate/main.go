package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env.dev file
	err := godotenv.Load(".env.dev")
	if err != nil {
		log.Println("Error loading .env file")
	}

	// Get the database connection URL from environment variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Connect to the database using pgx
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	// Ensure the migrations tracking table exists.
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename   TEXT        PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create schema_migrations table: %v", err)
	}

	// Collect already-applied migrations.
	rows, err := conn.Query(ctx, `SELECT filename FROM schema_migrations`)
	if err != nil {
		log.Fatalf("Failed to query schema_migrations: %v", err)
	}
	applied := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatalf("Failed to scan migration row: %v", err)
		}
		applied[name] = true
	}
	rows.Close()

	// Discover .sql files in migrations/ and sort them.
	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		log.Fatalf("Failed to glob migration files: %v", err)
	}
	sort.Strings(files)

	if len(files) == 0 {
		log.Println("No migration files found in migrations/")
		return
	}

	ran := 0
	for _, path := range files {
		filename := filepath.Base(path)
		if applied[filename] {
			log.Printf("  skip  %s (already applied)", filename)
			continue
		}

		sql, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("Failed to read %s: %v", path, err)
		}

		// Run each migration inside a transaction so failures roll back cleanly.
		tx, err := conn.Begin(ctx)
		if err != nil {
			log.Fatalf("Failed to begin transaction for %s: %v", filename, err)
		}

		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			_ = tx.Rollback(ctx)
			log.Fatalf("Failed to apply %s: %v", filename, err)
		}

		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (filename) VALUES ($1)`, filename,
		); err != nil {
			_ = tx.Rollback(ctx)
			log.Fatalf("Failed to record migration %s: %v", filename, err)
		}

		if err := tx.Commit(ctx); err != nil {
			log.Fatalf("Failed to commit migration %s: %v", filename, err)
		}

		log.Printf("  apply  %s", filename)
		ran++
	}

	fmt.Printf("\nDone. %d migration(s) applied, %d skipped.\n",
		ran, len(files)-ran)
}
