package errors

// StatusError is an error that includes HTTP status code and response body
// Providers should return this error type for proper error parsing
type StatusError interface {
	error
	StatusCode() int
	Body() []byte
}

// ProviderError implements StatusError for provider execution errors
type ProviderError struct {
	Code    int
	RawBody []byte
	Msg     string
}

// NewProviderError creates a new ProviderError
func NewProviderError(code int, body []byte, msg string) *ProviderError {
	return &ProviderError{
		Code:    code,
		RawBody: body,
		Msg:     msg,
	}
}

// Error implements error interface
func (e *ProviderError) Error() string {
	return e.Msg
}

// StatusCode returns the HTTP status code
func (e *ProviderError) StatusCode() int {
	return e.Code
}

// Body returns the raw response body
func (e *ProviderError) Body() []byte {
	return e.RawBody
}

// IsRetryableStatus checks if the status code indicates a retryable error
func IsRetryableStatus(code int) bool {
	switch code {
	case 429, 500, 502, 503, 504, 529:
		return true
	}
	return false
}

// IsAuthError checks if the status code indicates an authentication error
func IsAuthError(code int) bool {
	return code == 401 || code == 403
}
