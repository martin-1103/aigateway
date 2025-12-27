package services

import (
	"github.com/tidwall/gjson"
)

// TokenExtractor extracts token usage from provider responses
type TokenExtractor struct{}

// NewTokenExtractor creates a new token extractor
func NewTokenExtractor() *TokenExtractor {
	return &TokenExtractor{}
}

// ExtractAntigravityTokens extracts token count from Antigravity/Google API response
// Response format:
// {"usageMetadata": {"promptTokenCount": 10, "candidatesTokenCount": 50, "totalTokenCount": 60}}
func (e *TokenExtractor) ExtractAntigravityTokens(payload []byte) int64 {
	// Try totalTokenCount first
	total := gjson.GetBytes(payload, "usageMetadata.totalTokenCount").Int()
	if total > 0 {
		return total
	}

	// Fallback: sum prompt + candidates
	prompt := gjson.GetBytes(payload, "usageMetadata.promptTokenCount").Int()
	candidates := gjson.GetBytes(payload, "usageMetadata.candidatesTokenCount").Int()
	if prompt+candidates > 0 {
		return prompt + candidates
	}

	// Last fallback: estimate from payload size (~4 chars per token)
	if len(payload) > 0 {
		return int64(len(payload) / 4)
	}

	return 0
}

// ExtractOpenAITokens extracts token count from OpenAI API response
// Response format:
// {"usage": {"prompt_tokens": 10, "completion_tokens": 50, "total_tokens": 60}}
func (e *TokenExtractor) ExtractOpenAITokens(payload []byte) int64 {
	total := gjson.GetBytes(payload, "usage.total_tokens").Int()
	if total > 0 {
		return total
	}

	prompt := gjson.GetBytes(payload, "usage.prompt_tokens").Int()
	completion := gjson.GetBytes(payload, "usage.completion_tokens").Int()
	if prompt+completion > 0 {
		return prompt + completion
	}

	return 0
}

// ExtractTokens extracts token count based on provider
func (e *TokenExtractor) ExtractTokens(providerID string, payload []byte) int64 {
	switch providerID {
	case "antigravity":
		return e.ExtractAntigravityTokens(payload)
	case "openai":
		return e.ExtractOpenAITokens(payload)
	default:
		// Generic fallback - try common patterns
		if total := gjson.GetBytes(payload, "usageMetadata.totalTokenCount").Int(); total > 0 {
			return total
		}
		if total := gjson.GetBytes(payload, "usage.total_tokens").Int(); total > 0 {
			return total
		}
		// Estimate from size
		if len(payload) > 0 {
			return int64(len(payload) / 4)
		}
		return 0
	}
}
