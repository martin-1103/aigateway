package antigravity

const (
	// ProviderID is the unique identifier for Antigravity provider
	ProviderID = "antigravity"

	// AuthType defines the authentication method
	AuthType = "oauth"

	// BaseURL is the Antigravity API base URL
	BaseURL = "https://cloudcode-pa.googleapis.com"

	// EndpointGenerate is the non-streaming endpoint
	EndpointGenerate = "/v1internal:generateContent"

	// EndpointStream is the streaming endpoint
	EndpointStream = "/v1internal:streamGenerateContent?alt=sse"

	// UserAgent is the HTTP User-Agent header value
	UserAgent = "antigravity/1.0"

	// ContentType is the HTTP Content-Type header value
	ContentType = "application/json"
)

// SupportedModels returns the list of models supported by Antigravity
var SupportedModels = []string{
	"gemini-claude-sonnet-4-5",
	"gemini-3-pro-preview",
	"gemini-2.0-flash-exp",
	"gemini-exp-1206",
	"gemini-2.0-flash-thinking-exp-1219",
}
