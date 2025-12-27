package antigravity

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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
	baseURLs []string
}

// NewExecutor creates a new Executor instance
func NewExecutor() *Executor {
	return &Executor{
		baseURLs: BaseURLs,
	}
}

// Execute performs a non-streaming request to Antigravity API with fallback URLs
func (e *Executor) Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
	var lastErr error
	var lastResp *ExecuteResponse

	for _, baseURL := range e.baseURLs {
		endpoint := baseURL + EndpointGenerate
		// Claude models always use streaming endpoint with alt=sse
		if req.Stream || IsClaudeModel(req.Model) {
			endpoint = baseURL + EndpointStream + "?alt=sse"
		}

		fmt.Printf("[DEBUG] Model: %s, IsClaudeModel: %v, Endpoint: %s\n", req.Model, IsClaudeModel(req.Model), endpoint)

		resp, err := e.executeRequest(ctx, req, endpoint)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		lastResp = resp
		lastErr = err

		if resp != nil && resp.StatusCode == 401 {
			return resp, err
		}
	}

	if lastResp != nil {
		return lastResp, lastErr
	}
	return nil, lastErr
}

// resolveHost extracts host from URL for Host header
func resolveHost(base string) string {
	parsed, err := url.Parse(base)
	if err != nil {
		return ""
	}
	if parsed.Host != "" {
		return parsed.Host
	}
	return strings.TrimPrefix(strings.TrimPrefix(base, "https://"), "http://")
}

func (e *Executor) executeRequest(ctx context.Context, req *ExecuteRequest, endpoint string) (*ExecuteResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(req.Payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers exactly like reference
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+req.AccessToken)
	httpReq.Header.Set("User-Agent", UserAgent)

	// Stream uses text/event-stream, non-stream uses application/json
	if req.Stream || IsClaudeModel(req.Model) {
		httpReq.Header.Set("Accept", "text/event-stream")
	} else {
		httpReq.Header.Set("Accept", "application/json")
	}

	// Set Host header
	if host := resolveHost(endpoint); host != "" {
		httpReq.Host = host
	}

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

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return &ExecuteResponse{
			StatusCode: httpResp.StatusCode,
			Body:       nil,
			Latency:    latency,
			Error:      fmt.Errorf("failed to read response body: %w", err),
		}, err
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		fmt.Printf("[DEBUG] Error response: %s\n", string(body))
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

// ExecuteStream performs a streaming request to Antigravity API with fallback URLs
func (e *Executor) ExecuteStream(ctx context.Context, req *ExecuteRequest, handler StreamHandler) (*ExecuteResponse, error) {
	var lastErr error
	var lastResp *ExecuteResponse

	for _, baseURL := range e.baseURLs {
		endpoint := baseURL + EndpointStream + "?alt=sse"

		resp, err := e.executeStreamRequest(ctx, req, endpoint, handler)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		lastResp = resp
		lastErr = err

		if resp != nil && resp.StatusCode == 401 {
			return resp, err
		}
	}

	if lastResp != nil {
		return lastResp, lastErr
	}
	return nil, lastErr
}

func (e *Executor) executeStreamRequest(ctx context.Context, req *ExecuteRequest, endpoint string, handler StreamHandler) (*ExecuteResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(req.Payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", ContentType)
	httpReq.Header.Set("Authorization", "Bearer "+req.AccessToken)
	httpReq.Header.Set("User-Agent", UserAgent)
	httpReq.Header.Set("Accept", "text/event-stream")

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

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		body, _ := io.ReadAll(httpResp.Body)
		return &ExecuteResponse{
			StatusCode: httpResp.StatusCode,
			Body:       body,
			Latency:    latency,
			Error:      fmt.Errorf("upstream error: status %d", httpResp.StatusCode),
		}, fmt.Errorf("upstream error: status %d", httpResp.StatusCode)
	}

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
