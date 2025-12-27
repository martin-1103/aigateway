package errors

import (
	"github.com/tidwall/gjson"
)

// CodexParser parses OpenAI/Codex API errors
//
// Error format:
//
//	{"error": {"type": "...", "code": "insufficient_quota", "message": "..."}}
//
// Key distinction for 429:
//   - code="insufficient_quota" → QuotaExceeded (wait long, maybe 24h)
//   - code="rate_limit_exceeded" → RateLimit (retry quickly)
type CodexParser struct{}

// Parse implements ErrorParser for Codex/OpenAI API
func (p *CodexParser) Parse(statusCode int, body []byte) *ParsedError {
	errorType := gjson.GetBytes(body, "error.type").String()
	errorCode := gjson.GetBytes(body, "error.code").String()
	message := gjson.GetBytes(body, "error.message").String()

	if message == "" {
		message = extractMessage(body)
	}

	parsed := &ParsedError{
		StatusCode: statusCode,
		Message:    message,
		RawBody:    body,
		RawType:    errorType,
		RawCode:    errorCode,
	}

	// First, set defaults by status code
	p.setDefaultsByStatus(parsed, statusCode)

	// Then, check for quota vs rate limit on 429
	if statusCode == 429 {
		p.handle429(parsed, errorCode, errorType, message)
	}

	return parsed
}

func (p *CodexParser) setDefaultsByStatus(parsed *ParsedError, statusCode int) {
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
		// Will be overridden by handle429
		parsed.Type = ErrTypeRateLimit
		parsed.Retryable = true
		parsed.CooldownDur = CooldownRateLimit

	case 500, 502, 503, 504:
		parsed.Type = ErrTypeTransient
		parsed.Retryable = true
		parsed.CooldownDur = CooldownTransient

	default:
		parsed.Type = ErrTypeUnknown
		parsed.Retryable = false
	}
}

func (p *CodexParser) handle429(parsed *ParsedError, errorCode, errorType, message string) {
	// Check for quota exceeded indicators
	isQuotaExceeded := errorCode == "insufficient_quota" ||
		errorType == "insufficient_quota" ||
		containsIgnoreCase(message, "exceeded your current quota") ||
		containsIgnoreCase(message, "insufficient_quota")

	if isQuotaExceeded {
		parsed.Type = ErrTypeQuotaExceeded
		parsed.Retryable = false
		parsed.CooldownDur = CooldownQuotaExceed
		return
	}

	// Check for rate limit indicators
	isRateLimit := errorCode == "rate_limit_exceeded" ||
		containsIgnoreCase(message, "rate limit") ||
		containsIgnoreCase(message, "too many requests")

	if isRateLimit {
		parsed.Type = ErrTypeRateLimit
		parsed.Retryable = true
		parsed.CooldownDur = CooldownRateLimit
		return
	}

	// Default 429 to rate limit
	parsed.Type = ErrTypeRateLimit
	parsed.Retryable = true
	parsed.CooldownDur = CooldownRateLimit
}
