package glm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"aigateway-backend/providers"
)

// executeHTTP performs the actual HTTP request to GLM API
// Handles both streaming and non-streaming requests
func executeHTTP(ctx context.Context, req *providers.ExecuteRequest) (*providers.ExecuteResponse, error) {
	// Extract API key from account auth data
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

	// Create HTTP client with optional proxy
	client := createHTTPClient(req.ProxyURL)

	// Execute request and measure latency
	startTime := time.Now()
	httpResp, err := client.Do(httpReq)
	latencyMs := int(time.Since(startTime).Milliseconds())

	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Return response
	return &providers.ExecuteResponse{
		StatusCode: httpResp.StatusCode,
		Payload:    body,
		LatencyMs:  latencyMs,
	}, nil
}

// extractAPIKey extracts the API key from the account's auth data
// AuthData should be a JSON object with an "api_key" field
func extractAPIKey(req *providers.ExecuteRequest) (string, error) {
	// Try to use Token field first
	if req.Token != "" {
		return req.Token, nil
	}

	// Fall back to parsing AuthData JSON
	if req.Account.AuthData == "" {
		return "", fmt.Errorf("no auth data provided")
	}

	var authData map[string]interface{}
	if err := json.Unmarshal([]byte(req.Account.AuthData), &authData); err != nil {
		return "", fmt.Errorf("failed to parse auth data: %w", err)
	}

	apiKey, ok := authData["api_key"].(string)
	if !ok || apiKey == "" {
		return "", fmt.Errorf("api_key not found in auth data")
	}

	return apiKey, nil
}

// createHTTPClient creates an HTTP client with optional proxy configuration
func createHTTPClient(proxyURL string) *http.Client {
	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	if proxyURL != "" {
		// Note: In production, you'd parse proxyURL and configure transport
		// For now, using default transport
		// transport := &http.Transport{
		// 	Proxy: http.ProxyURL(proxyURLParsed),
		// }
		// client.Transport = transport
	}

	return client
}
