package pkce

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// PKCECodes contains the code verifier and challenge for OAuth PKCE flow
type PKCECodes struct {
	CodeVerifier  string
	CodeChallenge string
}

// GeneratePKCECodes generates a new pair of PKCE codes using S256 method.
// It creates a cryptographically random code verifier (96 bytes = 128 chars base64)
// and its corresponding SHA256 code challenge, as specified in RFC 7636.
func GeneratePKCECodes() (*PKCECodes, error) {
	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}

	codeChallenge := generateCodeChallenge(codeVerifier)

	return &PKCECodes{
		CodeVerifier:  codeVerifier,
		CodeChallenge: codeChallenge,
	}, nil
}

// generateCodeVerifier creates a cryptographically secure random string
// using 96 random bytes, resulting in 128 base64 URL-safe characters.
func generateCodeVerifier() (string, error) {
	bytes := make([]byte, 96)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes), nil
}

// generateCodeChallenge creates a code challenge from the verifier
// by taking SHA256 hash and encoding to base64 URL-safe format.
func generateCodeChallenge(codeVerifier string) string {
	hash := sha256.Sum256([]byte(codeVerifier))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])
}
