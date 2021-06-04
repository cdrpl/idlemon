package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// Generate a random string.
func GenerateToken(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("cannot generate a token with length of: %v", length)
	}

	bytes := make([]byte, length/2)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("generate token error: %v", err)
	}

	return hex.EncodeToString(bytes), nil
}
