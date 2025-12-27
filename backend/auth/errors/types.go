package errors

import "time"

// ErrorType represents categorized API error types
type ErrorType string

const (
	ErrTypeRateLimit      ErrorType = "rate_limit"      // Too many requests, retry quickly
	ErrTypeQuotaExceeded  ErrorType = "quota_exceeded"  // Quota exhausted, wait longer
	ErrTypeAuthentication ErrorType = "authentication"  // Auth failed, disable account
	ErrTypePermission     ErrorType = "permission"      // No permission, disable account
	ErrTypeNotFound       ErrorType = "not_found"       // Resource not found
	ErrTypeOverloaded     ErrorType = "overloaded"      // Server overloaded, retry with backoff
	ErrTypeTransient      ErrorType = "transient"       // Temporary error, retry
	ErrTypeInvalidRequest ErrorType = "invalid_request" // Bad request, don't retry
	ErrTypeUnknown        ErrorType = "unknown"         // Unknown error
)

// ParsedError contains parsed error information from API response
type ParsedError struct {
	Type        ErrorType     // Categorized error type
	StatusCode  int           // HTTP status code
	Message     string        // Error message from API
	Retryable   bool          // Whether request can be retried
	CooldownDur time.Duration // Suggested cooldown before retry
	RawBody     []byte        // Original response body
	RawType     string        // Original error type from API (e.g., "rate_limit_error")
	RawCode     string        // Original error code from API (e.g., "insufficient_quota")
}

// Error implements error interface
func (e *ParsedError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

// IsRetryable returns true if the error allows retry
func (e *ParsedError) IsRetryable() bool {
	if e == nil {
		return false
	}
	return e.Retryable
}

// ShouldDisableAccount returns true if account should be disabled
func (e *ParsedError) ShouldDisableAccount() bool {
	if e == nil {
		return false
	}
	return e.Type == ErrTypeAuthentication || e.Type == ErrTypePermission
}

// ShouldDisableForModel returns true if account should be disabled for specific model
func (e *ParsedError) ShouldDisableForModel() bool {
	if e == nil {
		return false
	}
	return e.Type == ErrTypeNotFound
}
