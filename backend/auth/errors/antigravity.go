package errors

import (
	"github.com/tidwall/gjson"
)

// AntigravityParser parses Google Cloud/Antigravity API errors
//
// Error format:
//
//	{"error": {"status": "...", "message": "...", "details": [{"reason": "QUOTA_EXCEEDED"}]}}
//
// Key distinction for 429:
//   - reason="RATE_LIMIT_EXCEEDED" → RateLimit (retry quickly, ~1s)
//   - reason="QUOTA_EXCEEDED" → QuotaExceeded (wait longer, ~1h)
//   - reason="USER_RATE_LIMIT_EXCEEDED" → RateLimit (retry ~10s)
type AntigravityParser struct{}

// Parse implements ErrorParser for Antigravity/Google Cloud API
func (p *AntigravityParser) Parse(statusCode int, body []byte) *ParsedError {
	errorStatus := gjson.GetBytes(body, "error.status").String()
	message := gjson.GetBytes(body, "error.message").String()
	reason := p.extractReason(body)

	if message == "" {
		message = extractMessage(body)
	}

	parsed := &ParsedError{
		StatusCode: statusCode,
		Message:    message,
		RawBody:    body,
		RawType:    errorStatus,
		RawCode:    reason,
	}

	// First, set defaults by status code
	p.setDefaultsByStatus(parsed, statusCode)

	// Then, check for specific reasons on 429
	if statusCode == 429 {
		p.handle429(parsed, reason, message)
	}

	return parsed
}

func (p *AntigravityParser) extractReason(body []byte) string {
	// Check error.details array for reason
	details := gjson.GetBytes(body, "error.details")
	if details.IsArray() {
		for _, detail := range details.Array() {
			if reason := detail.Get("reason").String(); reason != "" {
				return reason
			}
		}
	}

	// Fallback: check error.reason directly
	return gjson.GetBytes(body, "error.reason").String()
}

func (p *AntigravityParser) setDefaultsByStatus(parsed *ParsedError, statusCode int) {
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

func (p *AntigravityParser) handle429(parsed *ParsedError, reason, message string) {
	switch reason {
	case "RATE_LIMIT_EXCEEDED":
		parsed.Type = ErrTypeRateLimit
		parsed.Retryable = true
		parsed.CooldownDur = 1 * CooldownRateLimit / 5 // ~1 second

	case "QUOTA_EXCEEDED", "RESOURCE_EXHAUSTED":
		parsed.Type = ErrTypeQuotaExceeded
		parsed.Retryable = false
		parsed.CooldownDur = CooldownQuotaExceed

	case "USER_RATE_LIMIT_EXCEEDED":
		parsed.Type = ErrTypeRateLimit
		parsed.Retryable = true
		parsed.CooldownDur = 2 * CooldownRateLimit // ~10 seconds

	default:
		// Check message for hints
		if containsIgnoreCase(message, "quota") {
			parsed.Type = ErrTypeQuotaExceeded
			parsed.Retryable = false
			parsed.CooldownDur = CooldownQuotaExceed
		} else {
			parsed.Type = ErrTypeRateLimit
			parsed.Retryable = true
			parsed.CooldownDur = CooldownRateLimit
		}
	}
}
