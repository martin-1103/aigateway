package antigravity

import "time"

const (
	// ProviderID is the unique identifier for Antigravity provider
	ProviderID = "antigravity"

	// AuthType defines the authentication method
	AuthType = "oauth"

	// OAuth credentials for Antigravity (Google Cloud Code)
	OAuthClientID     = "1071006060591-tmhssin2h21lcre235vtolojh4g403ep.apps.googleusercontent.com"
	OAuthClientSecret = "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"
	OAuthTokenURL     = "https://oauth2.googleapis.com/token"

	// RefreshSkew is the time before expiry to refresh token
	RefreshSkew = 3000 * time.Second

	// Base URLs with fallback order
	BaseURLDaily   = "https://daily-cloudcode-pa.googleapis.com"
	BaseURLSandbox = "https://daily-cloudcode-pa.sandbox.googleapis.com"
	BaseURLProd    = "https://cloudcode-pa.googleapis.com"

	// Default base URL
	BaseURL = BaseURLDaily

	// EndpointGenerate is the non-streaming endpoint
	EndpointGenerate = "/v1internal:generateContent"

	// EndpointStream is the streaming endpoint (without query params)
	EndpointStream = "/v1internal:streamGenerateContent"

	// UserAgent is the HTTP User-Agent header value (same as reference)
	UserAgent = "antigravity/1.104.0 darwin/arm64"

	// ContentType is the HTTP Content-Type header value
	ContentType = "application/json"
)

// BaseURLs returns the list of base URLs in priority order
var BaseURLs = []string{
	BaseURLDaily,
	BaseURLSandbox,
	BaseURLProd,
}

// SupportedModels returns the list of models supported by Antigravity
var SupportedModels = []string{
	// Gemini 2.5 Series
	"gemini-2.5-flash",
	"gemini-2.5-flash-lite",
	"gemini-2.5-pro",
	"gemini-2.5-computer-use-preview-10-2025",

	// Gemini 3 Series
	"gemini-3-pro-preview",
	"gemini-3-pro-image-preview",
	"gemini-3-flash-preview",

	// Claude via Antigravity (using upstream names directly)
	"claude-sonnet-4-5",
	"claude-sonnet-4-5-thinking",
	"claude-opus-4-5",
	"claude-opus-4-5-thinking",
}

// ThinkingSupport defines thinking/reasoning configuration for models
type ThinkingSupport struct {
	Min            int      // Minimum thinking budget
	Max            int      // Maximum thinking budget
	ZeroAllowed    bool     // Whether budget=0 is allowed
	DynamicAllowed bool     // Whether dynamic budget (-1) is allowed
	Levels         []string // Supported thinking levels (e.g., "low", "high")
}

// ModelConfig contains configuration for specific models
type ModelConfig struct {
	Thinking            *ThinkingSupport
	MaxCompletionTokens int
	Name                string
}

// IsClaudeModel checks if the model is a Claude model
func IsClaudeModel(model string) bool {
	return len(model) >= 7 && model[0:7] == "claude-"
}

// GetModelConfig returns static configuration for antigravity models
func GetModelConfig() map[string]*ModelConfig {
	return map[string]*ModelConfig{
		"gemini-2.5-flash": {
			Thinking: &ThinkingSupport{Min: 0, Max: 24576, ZeroAllowed: true, DynamicAllowed: true},
			Name:     "models/gemini-2.5-flash",
		},
		"gemini-2.5-flash-lite": {
			Thinking: &ThinkingSupport{Min: 0, Max: 24576, ZeroAllowed: true, DynamicAllowed: true},
			Name:     "models/gemini-2.5-flash-lite",
		},
		"gemini-2.5-pro": {
			Thinking: &ThinkingSupport{Min: 0, Max: 24576, ZeroAllowed: true, DynamicAllowed: true},
			Name:     "models/gemini-2.5-pro",
		},
		"gemini-2.5-computer-use-preview-10-2025": {
			Name: "models/gemini-2.5-computer-use-preview-10-2025",
		},
		"gemini-3-pro-preview": {
			Thinking: &ThinkingSupport{Min: 128, Max: 32768, ZeroAllowed: false, DynamicAllowed: true, Levels: []string{"low", "high"}},
			Name:     "models/gemini-3-pro-preview",
		},
		"gemini-3-pro-image-preview": {
			Thinking: &ThinkingSupport{Min: 128, Max: 32768, ZeroAllowed: false, DynamicAllowed: true, Levels: []string{"low", "high"}},
			Name:     "models/gemini-3-pro-image-preview",
		},
		"gemini-3-flash-preview": {
			Thinking: &ThinkingSupport{Min: 128, Max: 32768, ZeroAllowed: false, DynamicAllowed: true, Levels: []string{"minimal", "low", "medium", "high"}},
			Name:     "models/gemini-3-flash-preview",
		},
		"claude-sonnet-4-5": {
			Name: "claude-sonnet-4-5",
		},
		"claude-sonnet-4-5-thinking": {
			Thinking:            &ThinkingSupport{Min: 1024, Max: 200000, ZeroAllowed: false, DynamicAllowed: true},
			MaxCompletionTokens: 64000,
			Name:                "claude-sonnet-4-5-thinking",
		},
		"claude-opus-4-5": {
			Name: "claude-opus-4-5",
		},
		"claude-opus-4-5-thinking": {
			Thinking:            &ThinkingSupport{Min: 1024, Max: 200000, ZeroAllowed: false, DynamicAllowed: true},
			MaxCompletionTokens: 64000,
			Name:                "claude-opus-4-5-thinking",
		},
	}
}
