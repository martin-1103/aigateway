package stream

import (
	"fmt"
	"net/http"
)

// Forwarder handles forwarding streaming data chunks to HTTP response
type Forwarder struct {
	writer     http.ResponseWriter
	flusher    http.Flusher
	translator func([]byte) []byte
}

// NewForwarder creates a new stream forwarder
func NewForwarder(w http.ResponseWriter, translator func([]byte) []byte) (*Forwarder, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("response writer does not support flushing")
	}

	return &Forwarder{
		writer:     w,
		flusher:    flusher,
		translator: translator,
	}, nil
}

// Forward reads from data and error channels and writes SSE events to the client
func (f *Forwarder) Forward(dataCh <-chan []byte, errCh <-chan error, done <-chan struct{}) error {
	for {
		select {
		case data, ok := <-dataCh:
			if !ok {
				return nil
			}

			// Translate chunk if translator is provided
			if f.translator != nil {
				data = f.translator(data)
			}

			// Write data chunk
			if _, err := f.writer.Write(data); err != nil {
				return fmt.Errorf("failed to write chunk: %w", err)
			}

			// Flush immediately for streaming
			f.flusher.Flush()

		case err := <-errCh:
			if err != nil {
				return fmt.Errorf("stream error: %w", err)
			}

		case <-done:
			return nil
		}
	}
}

// WriteSSEEvent writes a single Server-Sent Event
func (f *Forwarder) WriteSSEEvent(event string, data []byte) error {
	if event != "" {
		if _, err := fmt.Fprintf(f.writer, "event: %s\n", event); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(f.writer, "data: %s\n\n", data); err != nil {
		return err
	}

	f.flusher.Flush()
	return nil
}
