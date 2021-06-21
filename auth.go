package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// Will return true if API token is valid.
func ValidateApiToken(ctx context.Context, userId string, token string, rdb *redis.Client) (bool, error) {
	val, err := rdb.Get(ctx, userId).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		} else {
			return false, err
		}
	}

	return val == token, nil
}

// Will generate an API token and store it in Redis.
func CreateApiToken(ctx context.Context, rdb *redis.Client, userId uuid.UUID) (string, error) {
	token, err := GenerateToken(API_TOKEN_LEN)
	if err != nil {
		return "", err
	}

	cmd := rdb.SetEX(ctx, fmt.Sprintf("%v", userId.String()), token, API_TOKEN_TTL)
	if cmd.Err() != nil {
		return "", err
	}

	return token, nil
}
