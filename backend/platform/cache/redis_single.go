package cache

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func newRedisSingleClient(ctx context.Context) (*RedisClient, error) {
	dbNumber, _ := strconv.Atoi(os.Getenv("REDIS_DB_NUMBER"))

	options := &redis.Options{
		Addr:         fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password:     os.Getenv("REDIS_PASSWORD"),
		DB:           dbNumber,
		PoolSize:     40,
		MinIdleConns: 15,
		MaxRetries:   3, // Reduced retries
		DialTimeout:  5 * time.Second, // Fail fast
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	}

	client := redis.NewClient(options)

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to Redis Standalone: %w", err)
	}

	return &RedisClient{
		Client:      client,
		connectType: "standalone",
	}, nil
}
