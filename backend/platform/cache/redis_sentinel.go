package cache

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func newRedisSentinelClient(ctx context.Context) (*RedisClient, error) {
	dbNumber, _ := strconv.Atoi(os.Getenv("REDIS_DB_NUMBER"))
	masterName := os.Getenv("REDIS_SENTINEL_MASTER_NAME")
	addrsStr := os.Getenv("REDIS_SENTINEL_ADDRS")
	
	if masterName == "" || addrsStr == "" {
		return nil, fmt.Errorf("REDIS_SENTINEL_MASTER_NAME and REDIS_SENTINEL_ADDRS must be set for sentinel mode")
	}

	addrs := strings.Split(addrsStr, ",")
	for i := range addrs {
		addrs[i] = strings.TrimSpace(addrs[i])
	}

	options := &redis.FailoverOptions{
		MasterName:       masterName,
		SentinelAddrs:    addrs,
		DB:               dbNumber,
		Username:         os.Getenv("REDIS_USERNAME"),
		Password:         os.Getenv("REDIS_PASSWORD"),
		SentinelUsername: os.Getenv("REDIS_SENTINEL_USERNAME"),
		SentinelPassword: os.Getenv("REDIS_SENTINEL_PASSWORD"),
		PoolSize:         40,
		MinIdleConns:     15,
		MaxRetries:       3, // Reduced retries
		DialTimeout:      5 * time.Second, // Fail fast
		ReadTimeout:      3 * time.Second,
		WriteTimeout:     3 * time.Second,
		PoolTimeout:      4 * time.Second,
		// No Dialer hack here!
	}

	client := redis.NewFailoverClient(options)

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect via Redis Sentinel: %w", err)
	}

	return &RedisClient{
		Client:      client,
		connectType: "sentinel",
	}, nil
}
