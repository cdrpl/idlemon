package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// Will return true if API token is valid.
func ValidateApiToken(id string, token string, rdb *redis.Client) (bool, error) {
	val, err := rdb.Get(context.Background(), id).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		} else {
			return false, fmt.Errorf("validate api token error: %v", err)
		}
	}

	return val == token, nil
}

// Will generate an API token and store it in Redis.
func CreateApiToken(rdb *redis.Client, userId int) (string, error) {
	token, err := GenerateToken(API_TOKEN_LEN)
	if err != nil {
		return "", fmt.Errorf("create api token error: %v", err)
	}

	cmd := rdb.SetEX(context.Background(), fmt.Sprintf("%d", userId), token, API_TOKEN_TTL)
	if cmd.Err() != nil {
		return "", fmt.Errorf("create api token error: %v", err)
	}

	return token, nil
}
