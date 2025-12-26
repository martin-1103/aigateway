package antigravity

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ExecuteRequest represents a request to execute against Antigravity API
type ExecuteRequest struct {
	Model       string
	Payload     []byte
	Stream      bool
	AccessToken string
	HTTPClient  *http.Client
}

// ExecuteResponse represents the response from Antigravity API
type ExecuteResponse struct {
	StatusCode int
	Body       []byte
	Latency    int64
	Error      error
}

// StreamHandler is a callback function for handling streaming responses
type StreamHandler func(chunk []byte) error

// Executor handles HTTP communication with Antigravity API
type Executor struct {
	baseURL string
}

// NewExecutor creates a new Executor instance
func NewExecutor() *Executor {
	return &Executor{
		baseURL: BaseURL,
	}
}

// Execute performs a non-streaming request to Antigravity API
func (e *Executor) Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
	endpoint := e.baseURL + EndpointGenerate
	if req.Stream {
		endpoint = e.baseURL + EndpointStream
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(req.Payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", ContentType)
	httpReq.Header.Set("Authorization", "Bearer "+req.AccessToken)
	httpReq.Header.Set("User-Agent", UserAgent)

	// Execute request
	startTime := time.Now()
	httpResp, err := req.HTTPClient.Do(httpReq)
	latency := time.Since(startTime).Milliseconds()

	if err != nil {
		return &ExecuteResponse{
			StatusCode: 0,
			Body:       nil,
			Latency:    latency,
			Error:      fmt.Errorf("HTTP request failed: %w", err),
		}, err
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return &ExecuteResponse{
			StatusCode: httpResp.StatusCode,
			Body:       nil,
			Latency:    latency,
			Error:      fmt.Errorf("failed to read response body: %w", err),
		}, err
	}

	// Check for HTTP errors
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return &ExecuteResponse{
			StatusCode: httpResp.StatusCode,
			Body:       body,
			Latency:    latency,
			Error:      fmt.Errorf("upstream error: status %d", httpResp.StatusCode),
		}, fmt.Errorf("upstream error: status %d", httpResp.StatusCode)
	}

	return &ExecuteResponse{
		StatusCode: httpResp.StatusCode,
		Body:       body,
		Latency:    latency,
		Error:      nil,
	}, nil
}

// ExecuteStream performs a streaming request to Antigravity API
func (e *Executor) ExecuteStream(ctx context.Context, req *ExecuteRequest, handler StreamHandler) (*ExecuteResponse, error) {
	endpoint := e.baseURL + EndpointStream

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(req.Payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", ContentType)
	httpReq.Header.Set("Authorization", "Bearer "+req.AccessToken)
	httpReq.Header.Set("User-Agent", UserAgent)
	httpReq.Header.Set("Accept", "text/event-stream")

	// Execute request
	startTime := time.Now()
	httpResp, err := req.HTTPClient.Do(httpReq)
	latency := time.Since(startTime).Milliseconds()

	if err != nil {
		return &ExecuteResponse{
			StatusCode: 0,
			Body:       nil,
			Latency:    latency,
			Error:      fmt.Errorf("HTTP request failed: %w", err),
		}, err
	}
	defer httpResp.Body.Close()

	// Check for HTTP errors before streaming
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		body, _ := io.ReadAll(httpResp.Body)
		return &ExecuteResponse{
			StatusCode: httpResp.StatusCode,
			Body:       body,
			Latency:    latency,
			Error:      fmt.Errorf("upstream error: status %d", httpResp.StatusCode),
		}, fmt.Errorf("upstream error: status %d", httpResp.StatusCode)
	}

	// Process streaming response
	reader := NewSSEReader(httpResp.Body)
	for {
		event, err := reader.ReadEvent()
		if err == io.EOF {
			break
		}
		if err != nil {
			return &ExecuteResponse{
				StatusCode: httpResp.StatusCode,
				Body:       nil,
				Latency:    time.Since(startTime).Milliseconds(),
				Error:      fmt.Errorf("stream read error: %w", err),
			}, err
		}

		// Call handler for each event
		if err := handler(event.Data); err != nil {
			return &ExecuteResponse{
				StatusCode: httpResp.StatusCode,
				Body:       nil,
				Latency:    time.Since(startTime).Milliseconds(),
				Error:      fmt.Errorf("handler error: %w", err),
			}, err
		}
	}

	return &ExecuteResponse{
		StatusCode: httpResp.StatusCode,
		Body:       nil,
		Latency:    time.Since(startTime).Milliseconds(),
		Error:      nil,
	}, nil
}

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	Event string
	Data  []byte
}

// SSEReader reads Server-Sent Events from a stream
type SSEReader struct {
	reader io.Reader
}

// NewSSEReader creates a new SSE reader
func NewSSEReader(reader io.Reader) *SSEReader {
	return &SSEReader{reader: reader}
}

// ReadEvent reads the next SSE event from the stream
func (r *SSEReader) ReadEvent() (*SSEEvent, error) {
	var event SSEEvent
	var dataLines [][]byte

	buf := make([]byte, 4096)
	var line []byte

	for {
		n, err := r.reader.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}

		line = append(line, buf[:n]...)

		// Process complete lines
		for {
			idx := bytes.IndexByte(line, '\n')
			if idx == -1 {
				break
			}

			currentLine := line[:idx]
			line = line[idx+1:]

			// Remove trailing \r if present
			if len(currentLine) > 0 && currentLine[len(currentLine)-1] == '\r' {
				currentLine = currentLine[:len(currentLine)-1]
			}

			// Empty line means end of event
			if len(currentLine) == 0 {
				if len(dataLines) > 0 {
					event.Data = bytes.Join(dataLines, []byte("\n"))
					return &event, nil
				}
				continue
			}

			// Parse field
			if bytes.HasPrefix(currentLine, []byte("event:")) {
				event.Event = string(bytes.TrimSpace(currentLine[6:]))
			} else if bytes.HasPrefix(currentLine, []byte("data:")) {
				data := bytes.TrimSpace(currentLine[5:])
				dataLines = append(dataLines, data)
			}
		}

		if err == io.EOF {
			if len(dataLines) > 0 {
				event.Data = bytes.Join(dataLines, []byte("\n"))
				return &event, nil
			}
			return nil, io.EOF
		}
	}
}
