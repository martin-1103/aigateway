package codex

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// ParseIDToken parses a JWT id_token without verifying signature
// This is safe because we trust the token came from OpenAI's token endpoint
func ParseIDToken(idToken string) (*IDTokenClaims, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format: expected 3 parts, got %d", len(parts))
	}

	// Decode payload (second part)
	payload, err := decodeSegment(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var claims IDTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse JWT claims: %w", err)
	}

	return &claims, nil
}

// decodeSegment decodes a base64url encoded JWT segment
func decodeSegment(seg string) ([]byte, error) {
	// Add padding if needed
	switch len(seg) % 4 {
	case 2:
		seg += "=="
	case 3:
		seg += "="
	}

	// Use URL encoding (base64url as per JWT spec)
	return base64.URLEncoding.DecodeString(seg)
}

// GetAccountID extracts account ID from id_token
func GetAccountID(idToken string) (string, error) {
	claims, err := ParseIDToken(idToken)
	if err != nil {
		return "", err
	}

	if claims.Sub == "" {
		return "", fmt.Errorf("no subject (account_id) in token")
	}

	return claims.Sub, nil
}

// GetEmail extracts email from id_token
func GetEmail(idToken string) (string, error) {
	claims, err := ParseIDToken(idToken)
	if err != nil {
		return "", err
	}

	return claims.Email, nil
}
