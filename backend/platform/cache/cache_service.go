package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheService provides caching operations
type CacheService struct {
	client *RedisClient
	ctx    context.Context
}

// NewCacheService creates a new cache service instance
func NewCacheService(ctx context.Context) (*CacheService, error) {
	client, err := NewRedisClient(ctx)
	if err != nil {
		return nil, err
	}

	return &CacheService{
		client: client,
		ctx:    ctx,
	}, nil
}

// Set stores a value in cache with expiration
func (cs *CacheService) Set(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return cs.client.Client.Set(cs.ctx, key, data, expiration).Err()
}

// Get retrieves a value from cache
func (cs *CacheService) Get(key string) (string, error) {
	return cs.client.Client.Get(cs.ctx, key).Result()
}

// GetStruct retrieves a value from cache and unmarshals into struct
func (cs *CacheService) GetStruct(key string, dest interface{}) error {
	data, err := cs.client.Client.Get(cs.ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), dest)
}

// Delete removes a key from cache
func (cs *CacheService) Delete(key string) error {
	return cs.client.Client.Del(cs.ctx, key).Err()
}

// DeletePattern removes all keys matching pattern
func (cs *CacheService) DeletePattern(pattern string) error {
	keys, err := cs.client.Client.Keys(cs.ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return cs.client.Client.Del(cs.ctx, keys...).Err()
	}

	return nil
}

// Exists checks if a key exists in cache
func (cs *CacheService) Exists(key string) (bool, error) {
	count, err := cs.client.Client.Exists(cs.ctx, key).Result()
	return count > 0, err
}

// SetExpire sets expiration for an existing key
func (cs *CacheService) SetExpire(key string, expiration time.Duration) error {
	return cs.client.Client.Expire(cs.ctx, key, expiration).Err()
}

// GetTTL gets the remaining time to live for a key
func (cs *CacheService) GetTTL(key string) (time.Duration, error) {
	return cs.client.Client.TTL(cs.ctx, key).Result()
}

// Increment increments a key's value by 1
func (cs *CacheService) Increment(key string) (int64, error) {
	return cs.client.Client.Incr(cs.ctx, key).Result()
}

// IncrementBy increments a key's value by specified amount
func (cs *CacheService) IncrementBy(key string, value int64) (int64, error) {
	return cs.client.Client.IncrBy(cs.ctx, key, value).Result()
}

// Decrement decrements a key's value by 1
func (cs *CacheService) Decrement(key string) (int64, error) {
	return cs.client.Client.Decr(cs.ctx, key).Result()
}

// DecrementBy decrements a key's value by specified amount
func (cs *CacheService) DecrementBy(key string, value int64) (int64, error) {
	return cs.client.Client.DecrBy(cs.ctx, key, value).Result()
}

// SetNX sets a key only if it doesn't exist (distributed lock)
func (cs *CacheService) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, err
	}

	return cs.client.Client.SetNX(cs.ctx, key, data, expiration).Result()
}

// Hash operations
func (cs *CacheService) HSet(key, field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return cs.client.Client.HSet(cs.ctx, key, field, data).Err()
}

func (cs *CacheService) HGet(key, field string) (string, error) {
	return cs.client.Client.HGet(cs.ctx, key, field).Result()
}

func (cs *CacheService) HGetStruct(key, field string, dest interface{}) error {
	data, err := cs.client.Client.HGet(cs.ctx, key, field).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), dest)
}

func (cs *CacheService) HGetAll(key string) (map[string]string, error) {
	return cs.client.Client.HGetAll(cs.ctx, key).Result()
}

func (cs *CacheService) HDel(key string, fields ...string) error {
	return cs.client.Client.HDel(cs.ctx, key, fields...).Err()
}

// List operations
func (cs *CacheService) LPush(key string, values ...interface{}) error {
	data := make([]interface{}, len(values))
	for i, v := range values {
		jsonData, err := json.Marshal(v)
		if err != nil {
			return err
		}
		data[i] = jsonData
	}

	return cs.client.Client.LPush(cs.ctx, key, data...).Err()
}

func (cs *CacheService) RPush(key string, values ...interface{}) error {
	data := make([]interface{}, len(values))
	for i, v := range values {
		jsonData, err := json.Marshal(v)
		if err != nil {
			return err
		}
		data[i] = jsonData
	}

	return cs.client.Client.RPush(cs.ctx, key, data...).Err()
}

func (cs *CacheService) LPop(key string) (string, error) {
	return cs.client.Client.LPop(cs.ctx, key).Result()
}

func (cs *CacheService) RPop(key string) (string, error) {
	return cs.client.Client.RPop(cs.ctx, key).Result()
}

func (cs *CacheService) LLen(key string) (int64, error) {
	return cs.client.Client.LLen(cs.ctx, key).Result()
}

func (cs *CacheService) LRange(key string, start, stop int64) ([]string, error) {
	return cs.client.Client.LRange(cs.ctx, key, start, stop).Result()
}

// Set operations
func (cs *CacheService) SAdd(key string, members ...interface{}) error {
	data := make([]interface{}, len(members))
	for i, v := range members {
		jsonData, err := json.Marshal(v)
		if err != nil {
			return err
		}
		data[i] = jsonData
	}

	return cs.client.Client.SAdd(cs.ctx, key, data...).Err()
}

func (cs *CacheService) SMembers(key string) ([]string, error) {
	return cs.client.Client.SMembers(cs.ctx, key).Result()
}

func (cs *CacheService) SIsMember(key string, member interface{}) (bool, error) {
	data, err := json.Marshal(member)
	if err != nil {
		return false, err
	}

	return cs.client.Client.SIsMember(cs.ctx, key, data).Result()
}

func (cs *CacheService) SRem(key string, members ...interface{}) error {
	data := make([]interface{}, len(members))
	for i, v := range members {
		jsonData, err := json.Marshal(v)
		if err != nil {
			return err
		}
		data[i] = jsonData
	}

	return cs.client.Client.SRem(cs.ctx, key, data...).Err()
}

// Pipeline operations for batch processing
func (cs *CacheService) Pipeline() redis.Pipeliner {
	return cs.client.Client.Pipeline()
}

// Transaction operations (only available for standalone Redis)
func (cs *CacheService) Transaction(fn func(*redis.Tx) error, keys ...string) error {
	if client, ok := cs.client.Client.(*redis.Client); ok {
		return client.Watch(cs.ctx, func(tx *redis.Tx) error {
			return fn(tx)
		}, keys...)
	}
	return fmt.Errorf("transactions not supported in cluster mode")
}

// GetClient returns the underlying Redis client for advanced operations
func (cs *CacheService) GetClient() *RedisClient {
	return cs.client
}

// Close closes the cache service
func (cs *CacheService) Close() error {
	return cs.client.Close()
}

// CacheBuilder for building cache keys with prefixes
type CacheBuilder struct {
	prefix string
}

// NewCacheBuilder creates a new cache builder
func NewCacheBuilder(prefix string) *CacheBuilder {
	return &CacheBuilder{prefix: prefix}
}

// Key builds a cache key with prefix
func (cb *CacheBuilder) Key(parts ...string) string {
	key := cb.prefix
	for _, part := range parts {
		key += ":" + part
	}
	return key
}

// UserKey builds a user-specific cache key
func (cb *CacheBuilder) UserKey(userID string, parts ...string) string {
	allParts := append([]string{"user", userID}, parts...)
	return cb.Key(allParts...)
}

// TaskKey builds a task-specific cache key
func (cb *CacheBuilder) TaskKey(taskID string, parts ...string) string {
	allParts := append([]string{"task", taskID}, parts...)
	return cb.Key(allParts...)
}

// Helper functions for common cache patterns

// CacheOrFetch tries to get from cache, if not found executes fetchFn and caches the result
func CacheOrFetch[T any](cs *CacheService, key string, expiration time.Duration, fetchFn func() (T, error)) (T, error) {
	var result T

	// Try to get from cache first
	err := cs.GetStruct(key, &result)
	if err == nil {
		return result, nil
	}

	// If not in cache or error, fetch fresh data
	result, err = fetchFn()
	if err != nil {
		return result, err
	}

	// Cache the result
	_ = cs.Set(key, result, expiration)

	return result, nil
}

// InvalidatePattern invalidates cache entries matching a pattern
func InvalidatePattern(cs *CacheService, pattern string) error {
	return cs.DeletePattern(pattern)
}

// WarmupCache pre-loads data into cache
func WarmupCache[T any](cs *CacheService, key string, expiration time.Duration, data T) error {
	return cs.Set(key, data, expiration)
}
