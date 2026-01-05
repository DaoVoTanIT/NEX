package cache

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func newRedisClusterClient(ctx context.Context) (*RedisClient, error) {
	addrsStr := os.Getenv("REDIS_CLUSTER_ADDRS")
	if addrsStr == "" {
		return nil, fmt.Errorf("REDIS_CLUSTER_ADDRS must be set for cluster mode")
	}

	addrs := strings.Split(addrsStr, ",")
	for i := range addrs {
		addrs[i] = strings.TrimSpace(addrs[i])
	}

	options := &redis.ClusterOptions{
		Addrs:        addrs,
		Username:     os.Getenv("REDIS_USERNAME"),
		Password:     os.Getenv("REDIS_PASSWORD"),
		PoolSize:     40,
		MinIdleConns: 15,
		MaxRetries:   3, // Reduced retries
		DialTimeout:  5 * time.Second, // Fail fast
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
		
		// Routing for Cluster
		ReadOnly:       true, // Enable reading from slaves
		RouteByLatency: true, // Route to the closest node
	}

	client := redis.NewClusterClient(options)

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect via Redis Cluster: %w", err)
	}

	return &RedisClient{
		Client:      client,
		connectType: "cluster",
	}, nil
}
