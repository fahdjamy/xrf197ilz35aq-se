package random

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func TimeBasedString(seconds int64, length int) (string, error) {
	if length < 21 {
		return "", fmt.Errorf("for improved uniqueness, length must be greater than: 21")
	}

	// Generate a cryptographically secure random byte slice
	// Uses the cryptographically secure random number generator
	// from Go's standard library to add further randomness and unpredictability
	randomBytes := make([]byte, length)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}

	// Combine the timestamp and random bytes
	data := []byte(fmt.Sprintf("%d%s", seconds, randomBytes))

	// Encodes the combined data using Base64 for a compact, URL-safe string representation.
	uniqueStr := base64.URLEncoding.EncodeToString(data)

	return uniqueStr, nil
}
