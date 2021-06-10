package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"math/big"
	mathRand "math/rand"
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

// Seed math rand using crypto rand. Will exit app if error is encountered.
func SeedRand() {
	n, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		log.Fatalln("could not seed rand:", err)
	}

	mathRand.Seed(n.Int64())
}
