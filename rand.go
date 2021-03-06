package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"math"
	"math/big"
	mathRand "math/rand"
)

// Generate a random string.
func GenerateToken(length int) (string, error) {
	if length < 2 {
		return "", errors.New("cannot generate a token with length less than 2")
	}

	bytes := make([]byte, length/2)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
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

func RandInt(min int, max int) int {
	return mathRand.Int()%max + 1
}
