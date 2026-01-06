package cache

import (
	"context"
	"net"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

var (
	redisCommandsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_commands_total",
			Help: "Total number of Redis commands executed",
		},
		[]string{"command", "status"},
	)

	redisCommandDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_command_duration_seconds",
			Help:    "Duration of Redis commands",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"command"},
	)

	redisPoolHits = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "redis_pool_hits_total",
			Help: "Total number of successful connection pool hits",
		},
	)

	redisPoolMisses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "redis_pool_misses_total",
			Help: "Total number of connection pool misses",
		},
	)

	redisPoolTimeouts = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "redis_pool_timeouts_total",
			Help: "Total number of connection pool timeouts",
		},
	)
)

func init() {
	prometheus.MustRegister(redisCommandsTotal)
	prometheus.MustRegister(redisCommandDuration)
	prometheus.MustRegister(redisPoolHits)
	prometheus.MustRegister(redisPoolMisses)
	prometheus.MustRegister(redisPoolTimeouts)
}

// RedisMetricsHook implements redis.Hook
type RedisMetricsHook struct{}

func NewRedisMetricsHook() *RedisMetricsHook {
	return &RedisMetricsHook{}
}

func (h *RedisMetricsHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (h *RedisMetricsHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd)
		duration := time.Since(start).Seconds()

		status := "success"
		if err != nil && err != redis.Nil {
			status = "error"
		}

		redisCommandsTotal.WithLabelValues(cmd.Name(), status).Inc()
		redisCommandDuration.WithLabelValues(cmd.Name()).Observe(duration)

		return err
	}
}

func (h *RedisMetricsHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmds)
		duration := time.Since(start).Seconds()

		status := "success"
		if err != nil && err != redis.Nil {
			status = "error"
		}

		redisCommandsTotal.WithLabelValues("pipeline", status).Inc()
		redisCommandDuration.WithLabelValues("pipeline").Observe(duration)

		return err
	}
}
