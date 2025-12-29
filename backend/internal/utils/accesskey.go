package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateAccessKey() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return "uk_" + hex.EncodeToString(bytes)
}

func MaskAccessKey(key *string) string {
	if key == nil || *key == "" {
		return ""
	}
	k := *key
	if len(k) < 12 {
		return k
	}
	return k[:7] + "..." + k[len(k)-4:]
}
