package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

func CreateRedisClient() *redis.Client {
	addr := fmt.Sprintf("%v:6379", os.Getenv("REDIS_HOST"))

	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	cmd := rdb.Ping(context.Background())
	if cmd.Err() != nil {
		log.Fatalln("redis error:", cmd.Err())
	}

	return rdb
}
