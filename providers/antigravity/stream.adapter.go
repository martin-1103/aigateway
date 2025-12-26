package antigravity

import (
	"context"
	"fmt"
	"io"
	"time"

	"aigateway/providers"
)

// executeStreamAdapter adapts the existing ExecuteStream to the new interface
func executeStreamAdapter(ctx context.Context, executor *Executor, req *ExecuteRequest) (*providers.StreamResponse, error) {
	// Create channels
	dataCh := make(chan []byte, 10)
	errCh := make(chan error, 1)
	done := make(chan struct{})

	// Start time for latency tracking
	startTime := time.Now()

	// Create handler that forwards to dataCh
	handler := func(chunk []byte) error {
		// Copy chunk to avoid race conditions
		data := make([]byte, len(chunk))
		copy(data, chunk)

		select {
		case dataCh <- data:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Execute stream in goroutine
	go func() {
		defer close(dataCh)
		defer close(errCh)
		defer close(done)

		resp, err := executor.ExecuteStream(ctx, req, handler)
		if err != nil {
			errCh <- err
			return
		}

		if resp.Error != nil {
			errCh <- resp.Error
		}

		fmt.Printf("[DEBUG] Antigravity stream completed in %dms\n", time.Since(startTime).Milliseconds())
	}()

	return &providers.StreamResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "text/event-stream"},
		DataCh:     dataCh,
		ErrCh:      errCh,
		Done:       done,
	}, nil
}

// readAntigravitySSE reads SSE events and converts them to Claude format
func readAntigravitySSE(reader io.Reader, dataCh chan<- []byte) error {
	sseReader := NewSSEReader(reader)

	for {
		event, err := sseReader.ReadEvent()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Translate to Claude format
		translated := TranslateAntigravityStreamToClaude(event.Data, event.Event)

		// Send to channel
		chunk := make([]byte, len(translated))
		copy(chunk, translated)
		dataCh <- chunk
	}

	return nil
}
