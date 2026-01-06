package cache

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
)

// RedisClient holds the Redis client instances optimized for caching
type RedisClient struct {
	Client      redis.UniversalClient
	connectType string
}

// Close closes the Redis client connection
func (rc *RedisClient) Close() error {
	return rc.Client.Close()
}

// NewRedisClient creates a new Redis client based on REDIS_MODE or auto-detection
func NewRedisClient(ctx context.Context) (*RedisClient, error) {
	var (
		client *RedisClient
		err    error
	)

	// Prefer explicit mode if provided, else auto-detect
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("REDIS_MODE")))

	switch mode {
	case "cluster":
		client, err = newRedisClusterClient(ctx)
	case "sentinel":
		client, err = newRedisSentinelClient(ctx)
	case "single", "standalone":
		client, err = newRedisSingleClient(ctx)
	case "", "auto":
		// Auto-detect by presence of envs
		switch {
		case strings.TrimSpace(os.Getenv("REDIS_CLUSTER_ADDRS")) != "":
			client, err = newRedisClusterClient(ctx)
		case strings.TrimSpace(os.Getenv("REDIS_SENTINEL_MASTER_NAME")) != "" &&
			strings.TrimSpace(os.Getenv("REDIS_SENTINEL_ADDRS")) != "":
			client, err = newRedisSentinelClient(ctx)
		case strings.TrimSpace(os.Getenv("REDIS_HOST")) != "" &&
			strings.TrimSpace(os.Getenv("REDIS_PORT")) != "":
			client, err = newRedisSingleClient(ctx)
		default:
			err = fmt.Errorf("redis configuration not found: set REDIS_MODE or required envs")
		}
	default:
		err = fmt.Errorf("unsupported REDIS_MODE: %s", mode)
	}

	if err != nil {
		return nil, err
	}

	// Register Prometheus Metrics Hook
	client.Client.AddHook(NewRedisMetricsHook())

	return client, nil
}

// HealthCheck checks Redis connection health
func (rc *RedisClient) HealthCheck(ctx context.Context) error {
	return rc.Client.Ping(ctx).Err()
}

// IsCluster returns true if the client is a cluster client
func (rc *RedisClient) IsCluster() bool {
	return rc.connectType == "cluster"
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

// GetStats returns detailed connection statistics for monitoring
func (rc *RedisClient) GetStats(ctx context.Context) map[string]interface{} {
	stats := make(map[string]interface{})
	stats["type"] = rc.connectType

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
		info := client.Info(ctx, "server", "memory", "stats")
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
		nodes := client.ClusterNodes(ctx)
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
		slots := client.ClusterSlots(ctx)
		if slots.Err() == nil {
			stats["cluster_slots_ranges"] = len(slots.Val())
		}

		// Get cluster info
		if clusterInfo := client.ClusterInfo(ctx); clusterInfo.Err() == nil {
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
