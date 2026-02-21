package worker

import (
	"context"
	"log"
	"time"

	"fast-ingest/internal/model"
	"fast-ingest/internal/storage"
)

type Writer struct {
	Store         storage.Store
	In            <-chan model.Event
	BatchSize     int
	FlushInterval time.Duration
}

// Run starts the writer loop that listens for incoming events and flushes them to the storage layer in batches.
func (w *Writer) Run(ctx context.Context) {
	ticker := time.NewTicker(w.FlushInterval)
	defer ticker.Stop()

	// Storing the batch in memory until we flush to the database
	batch := make([]model.Event, 0, w.BatchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		err := w.Store.InsertEvents(ctx, batch)
		if err != nil {
			log.Printf("Error inserting event(s): %v", err)
		}

		// Clear the batch after flushing
		batch = batch[:0]
	}

	for {
		select {
		// Flush on context cancellation to ensure we don't lose any events in the buffer
		case <-ctx.Done():
			flush()
			return

		// Listen for incoming events and add them to the batch
		case e := <-w.In:
			batch = append(batch, e)
			if len(batch) >= w.BatchSize {
				flush()
			}
		// Flush at regular intervals to ensure timely persistence of events even if the batch size isn't reached
		case <-ticker.C:
			flush()
		}
	}
}
