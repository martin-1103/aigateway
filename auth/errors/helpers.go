package errors

import (
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

// Default cooldown durations
const (
	CooldownAuthFailure  = 30 * time.Minute
	CooldownRateLimit    = 5 * time.Second
	CooldownQuotaExceed  = 1 * time.Hour
	CooldownTransient    = 1 * time.Minute
	CooldownOverloaded   = 30 * time.Second
	CooldownNotFound     = 12 * time.Hour
)

// parseByStatusCode creates ParsedError based on status code only (fallback)
func parseByStatusCode(statusCode int, body []byte) *ParsedError {
	parsed := &ParsedError{
		StatusCode: statusCode,
		RawBody:    body,
	}

	switch {
	case statusCode == 401:
		parsed.Type = ErrTypeAuthentication
		parsed.Message = "Authentication failed"
		parsed.Retryable = false
		parsed.CooldownDur = CooldownAuthFailure

	case statusCode == 403:
		parsed.Type = ErrTypePermission
		parsed.Message = "Permission denied"
		parsed.Retryable = false
		parsed.CooldownDur = CooldownAuthFailure

	case statusCode == 404:
		parsed.Type = ErrTypeNotFound
		parsed.Message = "Resource not found"
		parsed.Retryable = false
		parsed.CooldownDur = CooldownNotFound

	case statusCode == 429:
		parsed.Type = ErrTypeRateLimit
		parsed.Message = "Rate limit exceeded"
		parsed.Retryable = true
		parsed.CooldownDur = CooldownRateLimit

	case statusCode >= 500 && statusCode < 600:
		parsed.Type = ErrTypeTransient
		parsed.Message = "Server error"
		parsed.Retryable = true
		parsed.CooldownDur = CooldownTransient

	default:
		parsed.Type = ErrTypeUnknown
		parsed.Message = "Unknown error"
		parsed.Retryable = false
	}

	return parsed
}

// extractMessage extracts error message from common JSON paths
func extractMessage(body []byte) string {
	paths := []string{
		"error.message",
		"error.error_message",
		"message",
		"error",
	}

	for _, path := range paths {
		if msg := gjson.GetBytes(body, path).String(); msg != "" {
			return msg
		}
	}

	return ""
}

// parseRetryAfterHeader parses Retry-After header value to duration
func parseRetryAfterHeader(value string) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}

	// Try parsing as seconds
	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Try parsing as HTTP date (not commonly used, skip for simplicity)
	return 0
}

// containsIgnoreCase checks if s contains substr (case insensitive)
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
