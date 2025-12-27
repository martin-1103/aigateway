package pkce

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestGeneratePKCECodes(t *testing.T) {
	codes, err := GeneratePKCECodes()
	if err != nil {
		t.Fatalf("GeneratePKCECodes() error = %v", err)
	}

	// Verify code verifier is 128 characters (96 bytes base64 encoded)
	if len(codes.CodeVerifier) != 128 {
		t.Errorf("CodeVerifier length = %d, want 128", len(codes.CodeVerifier))
	}

	// Verify code challenge is 43 characters (SHA256 hash base64 encoded)
	if len(codes.CodeChallenge) != 43 {
		t.Errorf("CodeChallenge length = %d, want 43", len(codes.CodeChallenge))
	}

	// Verify code challenge is valid base64
	_, err = base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(codes.CodeChallenge)
	if err != nil {
		t.Errorf("CodeChallenge is not valid base64: %v", err)
	}
}

func TestPKCECodesAreRandom(t *testing.T) {
	codes1, _ := GeneratePKCECodes()
	codes2, _ := GeneratePKCECodes()

	if codes1.CodeVerifier == codes2.CodeVerifier {
		t.Error("Expected different code verifiers, got identical")
	}

	if codes1.CodeChallenge == codes2.CodeChallenge {
		t.Error("Expected different code challenges, got identical")
	}
}

func TestCodeChallengeMatchesVerifier(t *testing.T) {
	codes, err := GeneratePKCECodes()
	if err != nil {
		t.Fatalf("GeneratePKCECodes() error = %v", err)
	}

	// Manually compute code challenge from verifier
	hash := sha256.Sum256([]byte(codes.CodeVerifier))
	expectedChallenge := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])

	if codes.CodeChallenge != expectedChallenge {
		t.Errorf("CodeChallenge = %s, want %s", codes.CodeChallenge, expectedChallenge)
	}
}

func TestCodeVerifierBase64URLSafe(t *testing.T) {
	codes, _ := GeneratePKCECodes()

	// Verify it only contains URL-safe characters
	for _, c := range codes.CodeVerifier {
		if !isBase64URLSafe(c) {
			t.Errorf("CodeVerifier contains non-URL-safe character: %c", c)
		}
	}
}

func isBase64URLSafe(c rune) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '_'
}
