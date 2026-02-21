package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fast-ingest/internal/api"
	"fast-ingest/internal/storage"
	"fast-ingest/internal/worker"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env.dev file
	err := godotenv.Load(".env.dev")
	if err != nil {
		log.Println("Error loading .env file")
	}

	// Create context that listens for the interrupt signal
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize the storage layer
	store, err := storage.NewPostgres(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer store.Close()

	// Set up the router
	server := api.NewServer(store, 20000) // Queue size of 20,000 for event processing
	r := api.NewRouter(*server)

	// Get the port from environment variables, default to 8080 if not set
	port := os.Getenv("PORT")
	if port == "" {
		log.Println("PORT environment variable is not set, defaulting to 8080")
		port = "8080" // Default to 8080 if not set
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: r,
	}

	// Start the server in a separate goroutine
	go func() {
		log.Printf("Starting the server on :%s\n", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on :%s: %v\n", port, err)
		}
	}()

	w := &worker.Writer{
		Store: store,
		In:    server.Queue,
	}

	// Start the writer in a separate goroutine
	go w.Run(ctx)

	// Listen for the interrupt signal
	<-ctx.Done()

	fmt.Println("\nShutting down gracefully, press Ctrl+C again to force")

	// Create shutdown context with 30-second timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Trigger graceful shutdown
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}
