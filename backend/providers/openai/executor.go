package openai

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"aigateway-backend/providers"
)

// HTTPRequest contains parameters for OpenAI HTTP request
type HTTPRequest struct {
	Model    string
	Payload  []byte
	Stream   bool
	APIKey   string
	ProxyURL string
}

// executeHTTP performs the HTTP request to OpenAI API
func executeHTTP(ctx context.Context, req *HTTPRequest) (*providers.ExecuteResponse, error) {
	// Build endpoint URL
	endpoint := BaseURL + EndpointChatCompletions
	if req.Stream {
		endpoint = BaseURL + EndpointChatCompletions // OpenAI uses same endpoint, streaming controlled by request body
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(req.Payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", ContentType)
	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)
	httpReq.Header.Set("User-Agent", UserAgent)

	// Create HTTP client with optional proxy
	client, err := createHTTPClient(req.ProxyURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Execute request and measure latency
	startTime := time.Now()
	httpResp, err := client.Do(httpReq)
	latencyMs := int(time.Since(startTime).Milliseconds())

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Return response
	return &providers.ExecuteResponse{
		StatusCode: httpResp.StatusCode,
		Payload:    body,
		LatencyMs:  latencyMs,
	}, nil
}

// createHTTPClient creates an HTTP client with optional proxy configuration
func createHTTPClient(proxyURL string) (*http.Client, error) {
	transport := &http.Transport{
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	// Configure proxy if provided
	if proxyURL != "" {
		parsedURL, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(parsedURL)
	}

	return &http.Client{
		Transport: transport,
		Timeout:   120 * time.Second, // 2 minute timeout
	}, nil
}
