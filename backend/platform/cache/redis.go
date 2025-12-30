/*
Redis Cluster Integration for High-Performance Caching

This package provides Redis Cluster support for caching in large-scale systems.
Redis Cluster offers:
- Automatic data partitioning across multiple nodes
- High availability with master-slave replication
- Automatic failover and recovery
- Horizontal scaling

Note: Message queuing is handled by Kafka in production systems.
*/
package cache

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/create-go-app/fiber-go-template/pkg/utils"
	"github.com/redis/go-redis/v9"
)

// RedisClient holds the Redis client instances optimized for caching
type RedisClient struct {
	Client      redis.Cmdable   // Redis client interface supporting caching commands (SET, GET, DEL, etc.)
	connectType string          // Connection type: "standalone" or "cluster"
	ctx         context.Context // Context for managing Redis connection lifecycle
	maxRetries  int             // Maximum retry attempts for failed operations
	retryDelay  time.Duration   // Delay between retry attempts
}

// NewRedisClient creates a new Redis client with cluster or standalone support
func NewRedisClient() (*RedisClient, error) {
	ctx := context.Background()

	// Check if cluster mode is enabled
	clusterMode := os.Getenv("REDIS_CLUSTER_MODE")
	if clusterMode == "true" {
		log.Println("Attempting to initialize Redis Cluster for high availability caching")

		// Try cluster first, fallback to standalone if cluster fails
		clusterClient, err := newRedisClusterClient(ctx)
		if err != nil {
			log.Printf("❌ Failed to connect to Redis Cluster: %v", err)
			log.Println("🔄 Falling back to Redis Standalone mode")
			return newRedisSingleClient(ctx)
		}

		return clusterClient, nil
	}

	log.Println("Initializing Redis Standalone client")
	return newRedisSingleClient(ctx)
}

// newRedisSingleClient creates a standalone Redis client (for development/testing)
func newRedisSingleClient(ctx context.Context) (*RedisClient, error) {
	dbNumber, _ := strconv.Atoi(os.Getenv("REDIS_DB_NUMBER"))

	redisConnURL, err := utils.ConnectionURLBuilder("redis")
	if err != nil {
		return nil, fmt.Errorf("failed to build Redis connection URL: %w", err)
	}

	options := &redis.Options{
		Addr:         redisConnURL,
		Username:     os.Getenv("REDIS_USERNAME"),
		Password:     os.Getenv("REDIS_PASSWORD"),
		DB:           dbNumber,
		PoolSize:     20, // Increased for better concurrency
		MinIdleConns: 10, // Increased for better performance
		MaxRetries:   5,  // Increased retry attempts
		DialTimeout:  10 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		PoolTimeout:  6 * time.Second,
	}

	client := redis.NewClient(options)

	// Test connection with timeout
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to Redis standalone: %w", err)
	}

	log.Printf("Redis standalone client connected successfully to %s", redisConnURL)

	return &RedisClient{
		Client:      client,
		connectType: "standalone",
		ctx:         ctx,
		maxRetries:  5,
		retryDelay:  500 * time.Millisecond,
	}, nil
}

// translateDockerAddr translates internal Docker addresses to external localhost ports
func translateDockerAddr(addr string) string {
	// Map internal Docker IPs to external localhost ports
	dockerToLocal := map[string]string{
		"172.18.0.3:6379": "localhost:7001", // redis-node-1
		"172.18.0.7:6379": "localhost:7002", // redis-node-2
		"172.18.0.5:6379": "localhost:7003", // redis-node-3
		"172.18.0.6:6379": "localhost:7004", // redis-node-4
		"172.18.0.2:6379": "localhost:7005", // redis-node-5
		"172.18.0.4:6379": "localhost:7006", // redis-node-6
	}

	if localAddr, exists := dockerToLocal[addr]; exists {
		return localAddr
	}
	return addr // Return original if not found in mapping
}

// newRedisClusterClient creates a Redis cluster client for high availability caching
func newRedisClusterClient(ctx context.Context) (*RedisClient, error) {
	clusterAddrsStr := os.Getenv("REDIS_CLUSTER_ADDRS")
	if clusterAddrsStr == "" {
		return nil, fmt.Errorf("REDIS_CLUSTER_ADDRS environment variable is required for cluster mode")
	}

	clusterAddrs := strings.Split(clusterAddrsStr, ",")
	// Trim whitespace from addresses
	for i, addr := range clusterAddrs {
		clusterAddrs[i] = strings.TrimSpace(addr)
	}

	if len(clusterAddrs) < 3 {
		log.Printf("Warning: Redis cluster should have at least 3 nodes for proper failover. Current nodes: %d", len(clusterAddrs))
	}

	options := &redis.ClusterOptions{
		Addrs:    clusterAddrs,
		Username: os.Getenv("REDIS_USERNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),

		// Connection pool settings for high throughput
		PoolSize:        50, // Higher pool size for cluster
		MinIdleConns:    20, // More idle connections for cluster nodes
		MaxRetries:      10, // Increased retries for cluster failover
		MinRetryBackoff: 200 * time.Millisecond,
		MaxRetryBackoff: 5 * time.Second,

		// Cluster-specific settings for reliability - increased timeouts for Docker networking
		DialTimeout:  30 * time.Second, // Increased for Docker network delays
		ReadTimeout:  15 * time.Second, // Increased for network latency
		WriteTimeout: 15 * time.Second, // Increased for network latency
		PoolTimeout:  20 * time.Second, // Increased pool timeout

		// Enable automatic failover
		RouteByLatency: false, // Disable for Docker networking issues
		RouteRandomly:  false, // Disable for more predictable routing
		ReadOnly:       false, // Allow writes to cluster

		// Network optimization for Docker environment
		NewClient: func(opt *redis.Options) *redis.Client {
			// Override node addresses to use localhost ports for external access
			// This helps with Docker network translation issues
			opt.Addr = translateDockerAddr(opt.Addr)
			return redis.NewClient(opt)
		},
	}

	client := redis.NewClusterClient(options)

	// Test connection with comprehensive health check
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to Redis cluster: %w", err)
	}

	// Verify cluster nodes
	nodes, err := client.ClusterNodes(pingCtx).Result()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to retrieve cluster nodes: %w", err)
	}

	nodeCount := len(strings.Split(strings.TrimSpace(nodes), "\n"))
	log.Printf("Redis cluster connected successfully with %d nodes", nodeCount)

	// Log cluster slots distribution for monitoring
	slots, err := client.ClusterSlots(pingCtx).Result()
	if err != nil {
		log.Printf("Warning: could not retrieve cluster slots: %v", err)
	} else {
		log.Printf("Redis cluster initialized with %d slot ranges", len(slots))
	}

	return &RedisClient{
		Client:      client,
		connectType: "cluster",
		ctx:         ctx,
		maxRetries:  5,
		retryDelay:  500 * time.Millisecond,
	}, nil
}

// RedisConnection func for connect to Redis server (backward compatibility)
// Deprecated: Use NewRedisClient() instead for better abstraction
func RedisConnection() (redis.Cmdable, error) {
	redisClient, err := NewRedisClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	return redisClient.Client, nil
}

// GetClusterClient returns the underlying cluster client if available
func (rc *RedisClient) GetClusterClient() (*redis.ClusterClient, bool) {
	if clusterClient, ok := rc.Client.(*redis.ClusterClient); ok {
		return clusterClient, true
	}
	return nil, false
}

// GetStandaloneClient returns the underlying standalone client if available
func (rc *RedisClient) GetStandaloneClient() (*redis.Client, bool) {
	if standaloneClient, ok := rc.Client.(*redis.Client); ok {
		return standaloneClient, true
	}
	return nil, false
}

// IsCluster returns true if the client is a cluster client
func (rc *RedisClient) IsCluster() bool {
	return rc.connectType == "cluster"
}

// HealthCheck checks Redis connection health
func (rc *RedisClient) HealthCheck() error {
	return rc.Client.Ping(rc.ctx).Err()
}

// Close closes the Redis connection
func (rc *RedisClient) Close() error {
	switch client := rc.Client.(type) {
	case *redis.Client:
		return client.Close()
	case *redis.ClusterClient:
		return client.Close()
	default:
		return nil
	}
}

// GetStats returns detailed connection statistics for monitoring
func (rc *RedisClient) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["type"] = rc.connectType
	stats["max_retries"] = rc.maxRetries
	stats["retry_delay_ms"] = rc.retryDelay.Milliseconds()

	switch client := rc.Client.(type) {
	case *redis.Client:
		poolStats := client.PoolStats()
		stats["pool"] = map[string]interface{}{
			"hits":        poolStats.Hits,
			"misses":      poolStats.Misses,
			"timeouts":    poolStats.Timeouts,
			"total_conns": poolStats.TotalConns,
			"idle_conns":  poolStats.IdleConns,
			"stale_conns": poolStats.StaleConns,
		}

		// Get server info
		info := client.Info(rc.ctx, "server", "memory", "stats")
		if info.Err() == nil {
			stats["server_info"] = parseRedisInfo(info.Val())
		}

	case *redis.ClusterClient:
		poolStats := client.PoolStats()
		stats["pool"] = map[string]interface{}{
			"hits":        poolStats.Hits,
			"misses":      poolStats.Misses,
			"timeouts":    poolStats.Timeouts,
			"total_conns": poolStats.TotalConns,
			"idle_conns":  poolStats.IdleConns,
			"stale_conns": poolStats.StaleConns,
		}

		// Get cluster-specific information
		nodes := client.ClusterNodes(rc.ctx)
		if nodes.Err() == nil {
			nodesInfo := nodes.Val()
			nodeLines := strings.Split(strings.TrimSpace(nodesInfo), "\n")
			stats["cluster_nodes_total"] = len(nodeLines)

			masterNodes := 0
			slaveNodes := 0
			for _, line := range nodeLines {
				if strings.Contains(line, "master") {
					masterNodes++
				} else if strings.Contains(line, "slave") {
					slaveNodes++
				}
			}
			stats["cluster_master_nodes"] = masterNodes
			stats["cluster_slave_nodes"] = slaveNodes
		}

		// Get cluster slots information
		slots := client.ClusterSlots(rc.ctx)
		if slots.Err() == nil {
			stats["cluster_slots_ranges"] = len(slots.Val())
		}

		// Get cluster info
		if clusterInfo := client.ClusterInfo(rc.ctx); clusterInfo.Err() == nil {
			info := parseClusterInfo(clusterInfo.Val())
			stats["cluster_info"] = info
		}
	}

	return stats
}

// parseRedisInfo parses Redis INFO command output
func parseRedisInfo(info string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(info, "\r\n")

	for _, line := range lines {
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "#") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				result[parts[0]] = parts[1]
			}
		}
	}

	return result
}

// parseClusterInfo parses Redis CLUSTER INFO command output
func parseClusterInfo(info string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(info), "\r\n")

	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				result[parts[0]] = parts[1]
			}
		}
	}

	return result
}

// CacheSet sets a key-value pair with TTL
func (rc *RedisClient) CacheSet(key string, value interface{}, ttl time.Duration) error {
	return rc.Client.Set(rc.ctx, key, value, ttl).Err()
}

// CacheGet retrieves a value by key
func (rc *RedisClient) CacheGet(key string) (string, error) {
	return rc.Client.Get(rc.ctx, key).Result()
}

// CacheDelete deletes a key
func (rc *RedisClient) CacheDelete(keys ...string) error {
	return rc.Client.Del(rc.ctx, keys...).Err()
}

// CacheExists checks if keys exist
func (rc *RedisClient) CacheExists(keys ...string) (int64, error) {
	return rc.Client.Exists(rc.ctx, keys...).Result()
}

// CacheExpire sets TTL for a key
func (rc *RedisClient) CacheExpire(key string, ttl time.Duration) error {
	return rc.Client.Expire(rc.ctx, key, ttl).Err()
}

// CacheTTL gets remaining TTL for a key
func (rc *RedisClient) CacheTTL(key string) (time.Duration, error) {
	return rc.Client.TTL(rc.ctx, key).Result()
}
