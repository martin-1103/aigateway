package glm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"aigateway-backend/providers"
)

// executeHTTPStream performs a streaming HTTP request to GLM API
func executeHTTPStream(ctx context.Context, req *providers.ExecuteRequest) (*providers.StreamResponse, error) {
	// Extract API key
	apiKey, err := extractAPIKey(req)
	if err != nil {
		return nil, fmt.Errorf("failed to extract API key: %w", err)
	}

	// Construct full URL
	url := BaseURL + EndpointChatCompletions

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(req.Payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", ContentType)
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	// Create HTTP client with optional proxy
	client := createHTTPClient(req.ProxyURL)

	// Execute request
	startTime := time.Now()
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
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

		if err := readGLMStream(httpResp.Body, dataCh); err != nil && err != io.EOF {
			errCh <- err
		}

		fmt.Printf("[DEBUG] GLM stream completed in %dms\n", time.Since(startTime).Milliseconds())
	}()

	return &providers.StreamResponse{
		StatusCode: httpResp.StatusCode,
		Headers:    extractHeaders(httpResp.Header),
		DataCh:     dataCh,
		ErrCh:      errCh,
		Done:       done,
	}, nil
}

// readGLMStream reads SSE events from GLM stream (OpenAI-compatible)
func readGLMStream(body io.Reader, dataCh chan<- []byte) error {
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

		// Validate JSON
		if !json.Valid(data) {
			continue
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
