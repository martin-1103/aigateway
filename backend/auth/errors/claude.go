package errors

import (
	"time"

	"github.com/tidwall/gjson"
)

// ClaudeParser parses Anthropic Claude API errors
//
// Error format:
//
//	{"type": "error", "error": {"type": "rate_limit_error", "message": "..."}}
//
// Error types: rate_limit_error, overloaded_error, authentication_error,
// permission_error, not_found_error, invalid_request_error, api_error
type ClaudeParser struct{}

// Parse implements ErrorParser for Claude/Anthropic API
func (p *ClaudeParser) Parse(statusCode int, body []byte) *ParsedError {
	errorType := gjson.GetBytes(body, "error.type").String()
	message := gjson.GetBytes(body, "error.message").String()

	if message == "" {
		message = extractMessage(body)
	}

	parsed := &ParsedError{
		StatusCode: statusCode,
		Message:    message,
		RawBody:    body,
		RawType:    errorType,
	}

	// First, set defaults by status code
	p.setDefaultsByStatus(parsed, statusCode)

	// Then, override by error.type if available
	if errorType != "" {
		p.overrideByErrorType(parsed, errorType)
	}

	return parsed
}

func (p *ClaudeParser) setDefaultsByStatus(parsed *ParsedError, statusCode int) {
	switch statusCode {
	case 400:
		parsed.Type = ErrTypeInvalidRequest
		parsed.Retryable = false

	case 401:
		parsed.Type = ErrTypeAuthentication
		parsed.Retryable = false
		parsed.CooldownDur = CooldownAuthFailure

	case 403:
		parsed.Type = ErrTypePermission
		parsed.Retryable = false
		parsed.CooldownDur = CooldownAuthFailure

	case 404:
		parsed.Type = ErrTypeNotFound
		parsed.Retryable = false
		parsed.CooldownDur = CooldownNotFound

	case 429:
		parsed.Type = ErrTypeRateLimit
		parsed.Retryable = true
		parsed.CooldownDur = CooldownRateLimit

	case 500:
		parsed.Type = ErrTypeTransient
		parsed.Retryable = true
		parsed.CooldownDur = CooldownTransient

	case 529:
		parsed.Type = ErrTypeOverloaded
		parsed.Retryable = true
		parsed.CooldownDur = CooldownOverloaded

	default:
		if statusCode >= 500 {
			parsed.Type = ErrTypeTransient
			parsed.Retryable = true
			parsed.CooldownDur = CooldownTransient
		} else {
			parsed.Type = ErrTypeUnknown
			parsed.Retryable = false
		}
	}
}

func (p *ClaudeParser) overrideByErrorType(parsed *ParsedError, errorType string) {
	switch errorType {
	case "rate_limit_error":
		parsed.Type = ErrTypeRateLimit
		parsed.Retryable = true
		if parsed.CooldownDur == 0 {
			parsed.CooldownDur = CooldownRateLimit
		}

	case "overloaded_error":
		parsed.Type = ErrTypeOverloaded
		parsed.Retryable = true
		parsed.CooldownDur = CooldownOverloaded

	case "authentication_error":
		parsed.Type = ErrTypeAuthentication
		parsed.Retryable = false
		parsed.CooldownDur = CooldownAuthFailure

	case "permission_error":
		parsed.Type = ErrTypePermission
		parsed.Retryable = false
		parsed.CooldownDur = CooldownAuthFailure

	case "not_found_error":
		parsed.Type = ErrTypeNotFound
		parsed.Retryable = false
		parsed.CooldownDur = CooldownNotFound

	case "invalid_request_error":
		parsed.Type = ErrTypeInvalidRequest
		parsed.Retryable = false

	case "api_error":
		parsed.Type = ErrTypeTransient
		parsed.Retryable = true
		parsed.CooldownDur = CooldownTransient
	}
}

// parseRateLimitCooldown extracts cooldown from rate limit response
func (p *ClaudeParser) parseRateLimitCooldown(body []byte) time.Duration {
	// Check for retry-after in message or dedicated field
	message := gjson.GetBytes(body, "error.message").String()

	// Claude sometimes includes "try again in X seconds" in message
	if containsIgnoreCase(message, "try again") {
		// Default to standard rate limit cooldown
		return CooldownRateLimit
	}

	return CooldownRateLimit
}
