package worker

import (
	"context"
	"log"
	"time"

	"fast-ingest/internal/model"
	"fast-ingest/internal/storage"
)

type Writer struct {
	Store storage.Store
	In    <-chan model.Event
}

func (w *Writer) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-w.In:
			// Use a short timeout for each insert to prevent blocking the worker indefinitely.
			ctxTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()

			err := w.Store.InsertEvent(ctxTimeout, e)
			if err != nil {
				log.Printf("insert event failed: %v", err)
			}
		}
	}
}
