package openai

const (
	// ProviderID is the unique identifier for OpenAI provider
	ProviderID = "openai"

	// AuthType defines the authentication method
	AuthType = "api_key"

	// BaseURL is the OpenAI API base URL
	BaseURL = "https://api.openai.com/v1"

	// EndpointChatCompletions is the chat completions endpoint
	EndpointChatCompletions = "/chat/completions"

	// UserAgent is the HTTP User-Agent header value
	UserAgent = "aigateway-backend/1.0"

	// ContentType is the HTTP Content-Type header value
	ContentType = "application/json"
)

// SupportedModels returns the list of models supported by OpenAI
var SupportedModels = []string{
	"gpt-4",
	"gpt-4-turbo",
	"gpt-3.5-turbo",
}
