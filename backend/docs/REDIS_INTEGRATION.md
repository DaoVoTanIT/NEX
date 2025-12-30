# Redis Cluster Integration Guide for High-Performance Caching

Redis Cluster integration for distributed caching in large-scale Golang Fiber applications.

## Overview

This Redis Cluster integration provides:
- **High-Performance Caching**: Distributed caching with automatic sharding
- **High Availability**: Automatic failover and replica management
- **Horizontal Scaling**: Easy cluster expansion with rebalancing
- **Data Partitioning**: Automatic key distribution across cluster nodes
- **Session Management**: Distributed session storage
- **Rate Limiting**: Cluster-aware rate limiting

**Note**: Message queuing is handled by Apache Kafka for better performance and reliability in large systems.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Fiber App     │    │  Redis Cluster  │    │ Kafka Cluster   │
│                 │    │  (Caching)      │    │ (Messaging)     │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ Cache       │◄┼────┤ │   Master 1  │ │    │ │  Broker 1   │ │
│ │ Middleware  │ │    │ │   Slave 1   │ │    │ │             │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ Controllers │◄┼────┤ │   Master 2  │ │    │ │  Broker 2   │ │
│ │ (Cache)     │ │    │ │   Slave 2   │ │    │ │             │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ Producers   │ │    │ │   Master 3  │ │    │ │  Broker 3   │ │
│ │ (Kafka)     │◄┼────┼─┼─────────────┼─┼────┤ │             │ │
│ └─────────────┘ │    │ │   Slave 3   │ │    │ └─────────────┘ │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Installation and Configuration

### 1. Environment Configuration

Copy and configure the Redis cluster settings:
```bash
cp .env.redis.example .env
```

Development configuration (minimal cluster):
```bash
# Redis Cluster
REDIS_CLUSTER_MODE=true
REDIS_CLUSTER_ADDRS=localhost:7001,localhost:7002,localhost:7003,localhost:7004,localhost:7005,localhost:7006
REDIS_USERNAME=
REDIS_PASSWORD=
```

Production configuration (recommended):
```bash
# Redis Cluster
REDIS_CLUSTER_MODE=true
REDIS_CLUSTER_ADDRS=redis-node1:6379,redis-node2:6379,redis-node3:6379,redis-node4:6379,redis-node5:6379,redis-node6:6379
REDIS_USERNAME=admin
REDIS_PASSWORD=your_strong_cluster_password

# Cache optimization
CACHE_DEFAULT_TTL=1800
CACHE_MAX_MEMORY=2gb
CACHE_EVICTION_POLICY=allkeys-lru

# Security
REDIS_TLS_ENABLED=true
REDIS_REQUIRE_AUTH=true
```

### 2. Redis Cluster Deployment

Development with Docker Compose:
```yaml
version: '3.8'
services:
  redis-node1:
    image: redis:7-alpine
    command: redis-server --port 6379 --cluster-enabled yes --cluster-config-file nodes.conf --cluster-node-timeout 5000 --appendonly yes
    ports: ["7001:6379"]
    volumes: ["redis1_data:/data"]
  
  redis-node2:
    image: redis:7-alpine
    command: redis-server --port 6379 --cluster-enabled yes --cluster-config-file nodes.conf --cluster-node-timeout 5000 --appendonly yes
    ports: ["7002:6379"]
    volumes: ["redis2_data:/data"]
    
  # ... nodes 3-6 with similar config
```

Production with Kubernetes:
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis-cluster
  namespace: production
spec:
  serviceName: redis-cluster
  replicas: 6
  selector:
    matchLabels:
      app: redis-cluster
  template:
    metadata:
      labels:
        app: redis-cluster
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        command: ["redis-server", "/conf/redis.conf"]
        ports:
        - containerPort: 6379
          name: redis
        - containerPort: 16379
          name: redis-cluster
        volumeMounts:
        - name: conf
          mountPath: /conf
        - name: data
          mountPath: /data
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 10Gi
```

### 3. Cluster Initialization

```bash
# Create cluster after all nodes are running
redis-cli --cluster create \
  redis-node1:6379 redis-node2:6379 redis-node3:6379 \
  redis-node4:6379 redis-node5:6379 redis-node6:6379 \
  --cluster-replicas 1
```

## Usage

### 1. HTTP Caching Middleware

```go
import "github.com/create-go-app/fiber-go-template/platform/cache"

app := fiber.New()

// Apply cache middleware for better performance
cacheMiddleware := cache.NewMiddleware(cache.CacheConfig{
    Expiration:   5 * time.Minute,
    CacheControl: true,
    Methods:      []string{fiber.MethodGet, fiber.MethodHead},
})

app.Use("/api/v1/tasks", cacheMiddleware)
```

### 2. Data Caching in Controllers

```go
// Use Redis cluster cache service
cacheService, err := cache.NewCacheService()
if err != nil {
    return err
}
defer cacheService.Close()

// Cache-or-fetch pattern with cluster distribution
tasks, err := cache.CacheOrFetch(cacheService, "tasks:all", 5*time.Minute, func() ([]models.Task, error) {
    return db.GetTasks()
})

// Direct cluster cache operations
client, err := cache.NewRedisClient()
if err != nil {
    return err
}

// Set cache with automatic cluster sharding
err = client.CacheSet("user:123:profile", userProfile, 30*time.Minute)

// Get from cluster
profile, err := client.CacheGet("user:123:profile")
```

### 3. Session Management with Cluster

```go
// Apply session middleware with cluster support
sessionMiddleware := cache.SessionMiddleware(cache.SessionConfig{
    CookieName:      "session_id",
    Expiration:      24 * time.Hour,
    HTTPOnly:        true,
    Secure:          true,  // HTTPS only in production
    SameSite:        "Lax",
    ClusterEnabled:  true,   // Enable cluster session storage
})

app.Use(sessionMiddleware)

// Session data is automatically distributed across cluster
app.Get("/profile", func(c *fiber.Ctx) error {
    session := c.Locals("session").(*cache.Session)
    
    userID := session.Get("user_id")
    if userID == nil {
        return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
    }
    
    // Session data is available from any cluster node
    return c.JSON(fiber.Map{"user_id": userID})
})
```

### 4. Cluster-Aware Rate Limiting

```go
// Apply distributed rate limiting across cluster
rateLimitMiddleware := cache.RateLimitMiddleware(cache.RateLimitConfig{
    Max:      1000,                    // 1000 requests
    Duration: 1 * time.Hour,           // per hour
    KeyGenerator: func(c *fiber.Ctx) string {
        return "api:" + c.IP()
    },
    ClusterEnabled: true,              // Use cluster for accurate counting
})

app.Use("/api/*", rateLimitMiddleware)
```

### 5. Message Queuing with Kafka

**Note**: For message queuing, use Kafka instead of Redis for better performance and reliability:

```go
// For message queuing, use Kafka producers/consumers
// This provides better throughput and durability than Redis

import "github.com/your-org/kafka-client"

// Kafka producer for high-throughput messaging
producer, err := kafka.NewProducer(kafka.Config{
    Brokers: []string{"kafka-1:9092", "kafka-2:9092", "kafka-3:9092"},
    Topic:   "task-events",
})

// Send message to Kafka
err = producer.Send(kafka.Message{
    Key:   "task:created",
    Value: taskPayload,
})
```

### 7. Advanced Patterns

#### Distributed Locking

```go
cacheService, _ := cache.NewCacheService()
lockKey := "lock:critical-operation"

acquired, err := cacheService.SetNX(lockKey, "locked", 30*time.Second)
if acquired {
    // Thực hiện critical section
    defer cacheService.Delete(lockKey)
}
```

#### Counters

```go
// Increment counters
views, err := cacheService.IncrementBy("page_views", 1)

// Set expiration for daily reset
if views == 1 {
    cacheService.SetExpire("page_views", 24*time.Hour)
}
```

#### Hash Operations

```go
// Store user preferences
cacheService.HSet("user:123:prefs", "theme", "dark")
cacheService.HSet("user:123:prefs", "language", "vi")

// Get all preferences
prefs, err := cacheService.HGetAll("user:123:prefs")
```

## Performance Tuning

### 1. Connection Pool

```bash
# Optimize connection pool
REDIS_POOL_SIZE=20
REDIS_MIN_IDLE_CONNS=10
REDIS_POOL_TIMEOUT=4s
```

### 2. Memory Optimization

```bash
# Configure Redis memory policies
redis-cli CONFIG SET maxmemory 1gb
redis-cli CONFIG SET maxmemory-policy allkeys-lru
```

### 3. Persistence Settings

```bash
# Production persistence
REDIS_PERSISTENCE_MODE=AOF
REDIS_AOF_ENABLED=true
REDIS_AOF_SYNC_MODE=everysec
```

## Monitoring và Troubleshooting

### 1. Health Checks

```go
app.Get("/health", func(c *fiber.Ctx) error {
    cacheService, err := cache.NewCacheService()
    if err != nil {
        return c.Status(503).JSON(fiber.Map{
            "status": "unhealthy",
            "redis":  "connection failed",
        })
    }
    
    err = cacheService.GetClient().HealthCheck()
    if err != nil {
        return c.Status(503).JSON(fiber.Map{
            "status": "unhealthy", 
            "redis":  "health check failed",
        })
    }
    
    return c.JSON(fiber.Map{"status": "healthy"})
})
```

### 2. Redis Statistics

```go
app.Get("/redis/stats", func(c *fiber.Ctx) error {
    cacheService, _ := cache.NewCacheService()
    stats := cacheService.GetClient().GetStats()
    
    return c.JSON(stats)
})
```

### 3. Debug Commands

```bash
# Redis CLI commands for debugging
redis-cli INFO memory
redis-cli INFO stats
redis-cli MONITOR
redis-cli SLOWLOG GET 10
redis-cli CLIENT LIST
```

### 4. Performance Monitoring

```go
// Add metrics to your handlers
start := time.Now()
// ... cache operation
duration := time.Since(start)

log.Printf("Cache operation took %v", duration)
```

## Production Deployment

### 1. Redis Cluster Setup

```yaml
# docker-compose.yml cho Redis Cluster
version: '3.8'
services:
  redis-node-1:
    image: redis:7-alpine
    command: redis-server --cluster-enabled yes --cluster-config-file nodes.conf
    ports:
      - "7001:6379"
    volumes:
      - redis1_data:/data

  redis-node-2:
    image: redis:7-alpine  
    command: redis-server --cluster-enabled yes --cluster-config-file nodes.conf
    ports:
      - "7002:6379"
    volumes:
      - redis2_data:/data

  # ... thêm nodes
```

### 2. Security

```bash
# Enable authentication
REDIS_PASSWORD=your_strong_password
REDIS_TLS_ENABLED=true

# Network security
# Bind Redis to private network only
# Use firewall rules  
# Enable Redis AUTH
```

### 3. Backup Strategy

```bash
# Automated backups
*/30 * * * * redis-cli --rdb /backup/dump-$(date +%Y%m%d%H%M).rdb

# Point-in-time recovery với AOF
redis-cli BGREWRITEAOF
```

### 4. High Availability

- **Redis Cluster**: 6 nodes (3 masters, 3 replicas)
- **Redis Sentinel**: 3 sentinel instances
- **Load Balancer**: HAProxy/Nginx cho Redis endpoints
- **Monitoring**: Grafana + Prometheus metrics

## Best Practices

1. **Key Naming**: Sử dụng consistent naming convention
   ```
   app:user:123:profile
   app:cache:tasks:user:456
   app:queue:email:high
   ```

2. **TTL Management**: Luôn set expiration cho cache keys
3. **Error Handling**: Graceful degradation khi Redis unavailable  
4. **Memory Management**: Monitor memory usage và set limits
5. **Security**: Enable AUTH và TLS trong production
6. **Backup**: Regular backups với automated restore testing
7. **Monitoring**: Set up alerts cho Redis health và performance

## API Endpoints

Sau khi tích hợp, các endpoints mới có sẵn:

- `GET /api/v1/tasks/cached` - Cached tasks list
- `GET /api/v1/task/cached/:id` - Cached single task
- `POST /api/v1/task/queued` - Create task with queue
- `PUT /api/v1/task/cached` - Update task with cache  
- `DELETE /api/v1/task/cached` - Delete task with cache
- `GET /api/v1/tasks/stats` - Task statistics with cache
- `GET /api/v1/cache/stats` - Redis cache statistics
- `POST /api/v1/cache/clear` - Clear user cache
- `GET /health` - Health check with Redis status

Tích hợp Redis hoàn tất! 🚀