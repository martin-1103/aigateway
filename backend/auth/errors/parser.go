package errors

// ErrorParser parses API error responses into structured ParsedError
type ErrorParser interface {
	// Parse extracts error information from HTTP response
	Parse(statusCode int, body []byte) *ParsedError
}

// GetParser returns appropriate error parser for provider
func GetParser(providerID string) ErrorParser {
	switch providerID {
	case "claude", "anthropic":
		return &ClaudeParser{}
	case "codex", "openai":
		return &CodexParser{}
	case "antigravity":
		return &AntigravityParser{}
	default:
		return &DefaultParser{}
	}
}

// DefaultParser handles unknown providers with status-code-only parsing
type DefaultParser struct{}

// Parse implements ErrorParser for unknown providers
func (p *DefaultParser) Parse(statusCode int, body []byte) *ParsedError {
	return parseByStatusCode(statusCode, body)
}
