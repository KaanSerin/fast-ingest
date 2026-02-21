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

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env.dev file
	err := godotenv.Load(".env.dev")
	if err != nil {
		log.Println("Error loading .env file")
	}

	// Set up the router
	r := api.NewRouter()

	port := os.Getenv("PORT")
	if port == "" {
		log.Println("PORT environment variable is not set, defaulting to 8080")
		port = "8080" // Default to 8080 if not set
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: r,
	}

	// Create context that listens for the interrupt signal
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start the server in a separate goroutine
	go func() {
		log.Printf("Starting the server on :%s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on :%s: %v\n", port, err)
		}
	}()

	// Listen for the interrupt signal
	<-ctx.Done()

	// Create shutdown context with 30-second timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Trigger graceful shutdown
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}
