package stream

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockResponseWriter implements http.ResponseWriter and http.Flusher
type mockResponseWriter struct {
	*httptest.ResponseRecorder
}

func (m *mockResponseWriter) Flush() {}

func TestNewForwarder(t *testing.T) {
	w := &mockResponseWriter{httptest.NewRecorder()}

	forwarder, err := NewForwarder(w, nil)
	if err != nil {
		t.Fatalf("NewForwarder() error = %v", err)
	}
	if forwarder == nil {
		t.Error("NewForwarder() returned nil")
	}
}

func TestNewForwarderWithoutFlusher(t *testing.T) {
	// httptest.ResponseRecorder without Flusher wrapper should fail
	w := httptest.NewRecorder()

	// Create a wrapper that doesn't implement Flusher
	nonFlusher := &nonFlusherWriter{ResponseWriter: w}

	_, err := NewForwarder(nonFlusher, nil)
	if err == nil {
		t.Error("NewForwarder() should error when writer doesn't support flushing")
	}
}

type nonFlusherWriter struct {
	http.ResponseWriter
}

func TestForwardData(t *testing.T) {
	w := &mockResponseWriter{httptest.NewRecorder()}
	forwarder, _ := NewForwarder(w, nil)

	dataCh := make(chan []byte, 3)
	errCh := make(chan error, 1)
	done := make(chan struct{})

	// Send test data
	go func() {
		dataCh <- []byte("chunk1")
		dataCh <- []byte("chunk2")
		dataCh <- []byte("chunk3")
		close(dataCh)
	}()

	err := forwarder.Forward(dataCh, errCh, done)
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	body := w.Body.String()
	if !bytes.Contains([]byte(body), []byte("chunk1")) {
		t.Error("Response missing chunk1")
	}
	if !bytes.Contains([]byte(body), []byte("chunk2")) {
		t.Error("Response missing chunk2")
	}
	if !bytes.Contains([]byte(body), []byte("chunk3")) {
		t.Error("Response missing chunk3")
	}
}

func TestForwardWithTranslator(t *testing.T) {
	w := &mockResponseWriter{httptest.NewRecorder()}

	translator := func(data []byte) []byte {
		return []byte("translated:" + string(data))
	}

	forwarder, _ := NewForwarder(w, translator)

	dataCh := make(chan []byte, 1)
	errCh := make(chan error, 1)
	done := make(chan struct{})

	go func() {
		dataCh <- []byte("original")
		close(dataCh)
	}()

	err := forwarder.Forward(dataCh, errCh, done)
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	body := w.Body.String()
	if !bytes.Contains([]byte(body), []byte("translated:original")) {
		t.Errorf("Response not translated, got: %s", body)
	}
}

func TestForwardDoneSignal(t *testing.T) {
	w := &mockResponseWriter{httptest.NewRecorder()}
	forwarder, _ := NewForwarder(w, nil)

	dataCh := make(chan []byte, 10)
	errCh := make(chan error, 1)
	done := make(chan struct{})

	// Close done channel immediately
	close(done)

	err := forwarder.Forward(dataCh, errCh, done)
	if err != nil {
		t.Fatalf("Forward() error = %v, want nil", err)
	}
}

func TestWriteSSEEvent(t *testing.T) {
	w := &mockResponseWriter{httptest.NewRecorder()}
	forwarder, _ := NewForwarder(w, nil)

	err := forwarder.WriteSSEEvent("message", []byte(`{"type":"test"}`))
	if err != nil {
		t.Fatalf("WriteSSEEvent() error = %v", err)
	}

	body := w.Body.String()
	if !bytes.Contains([]byte(body), []byte("event: message\n")) {
		t.Error("Response missing event line")
	}
	if !bytes.Contains([]byte(body), []byte(`data: {"type":"test"}`)) {
		t.Error("Response missing data line")
	}
}

func TestWriteSSEEventWithoutEventName(t *testing.T) {
	w := &mockResponseWriter{httptest.NewRecorder()}
	forwarder, _ := NewForwarder(w, nil)

	err := forwarder.WriteSSEEvent("", []byte(`{"type":"test"}`))
	if err != nil {
		t.Fatalf("WriteSSEEvent() error = %v", err)
	}

	body := w.Body.String()
	if bytes.Contains([]byte(body), []byte("event:")) {
		t.Error("Response should not have event line when event name is empty")
	}
	if !bytes.Contains([]byte(body), []byte(`data: {"type":"test"}`)) {
		t.Error("Response missing data line")
	}
}

func TestForwardConcurrency(t *testing.T) {
	w := &mockResponseWriter{httptest.NewRecorder()}
	forwarder, _ := NewForwarder(w, nil)

	dataCh := make(chan []byte, 100)
	errCh := make(chan error, 1)
	done := make(chan struct{})

	// Produce data concurrently
	go func() {
		for i := 0; i < 100; i++ {
			dataCh <- []byte("data")
		}
		close(dataCh)
	}()

	errDone := make(chan error, 1)
	go func() {
		errDone <- forwarder.Forward(dataCh, errCh, done)
	}()

	select {
	case err := <-errDone:
		if err != nil {
			t.Fatalf("Forward() error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Forward() timed out")
	}
}
