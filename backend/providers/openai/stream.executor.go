package openai

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"aigateway-backend/providers"
)

// executeHTTPStream performs a streaming HTTP request to OpenAI API
func executeHTTPStream(ctx context.Context, req *HTTPRequest) (*providers.StreamResponse, error) {
	endpoint := BaseURL + EndpointChatCompletions

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(req.Payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", ContentType)
	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)
	httpReq.Header.Set("User-Agent", UserAgent)
	httpReq.Header.Set("Accept", "text/event-stream")

	// Create HTTP client with optional proxy
	client, err := createHTTPClient(req.ProxyURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Execute request
	startTime := time.Now()
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check status code
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		body, _ := io.ReadAll(httpResp.Body)
		httpResp.Body.Close()
		return &providers.StreamResponse{
			StatusCode: httpResp.StatusCode,
		}, fmt.Errorf("upstream error: status %d, body: %s", httpResp.StatusCode, string(body))
	}

	// Create channels
	dataCh := make(chan []byte, 10)
	errCh := make(chan error, 1)
	done := make(chan struct{})

	// Start goroutine to read SSE stream
	go func() {
		defer close(dataCh)
		defer close(errCh)
		defer close(done)
		defer httpResp.Body.Close()

		if err := readOpenAIStream(httpResp.Body, dataCh); err != nil && err != io.EOF {
			errCh <- err
		}

		fmt.Printf("[DEBUG] Stream completed in %dms\n", time.Since(startTime).Milliseconds())
	}()

	return &providers.StreamResponse{
		StatusCode: httpResp.StatusCode,
		Headers:    extractHeaders(httpResp.Header),
		DataCh:     dataCh,
		ErrCh:      errCh,
		Done:       done,
	}, nil
}

// readOpenAIStream reads SSE events from OpenAI stream
func readOpenAIStream(body io.Reader, dataCh chan<- []byte) error {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		// Check for "data: " prefix
		if !bytes.HasPrefix(line, []byte("data: ")) {
			continue
		}

		// Extract data after "data: "
		data := line[6:]

		// Check for [DONE] marker
		if bytes.Equal(data, []byte("[DONE]")) {
			break
		}

		// Send chunk to channel (copy to avoid race)
		chunk := make([]byte, len(data))
		copy(chunk, data)
		dataCh <- chunk
	}

	return scanner.Err()
}

// extractHeaders converts http.Header to map[string]string
func extractHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}
