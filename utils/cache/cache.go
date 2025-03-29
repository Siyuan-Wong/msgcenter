package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

// CacheEntry 定义缓存条目结构
type CacheEntry struct {
	Key        string
	Value      interface{}
	Expiration time.Time
}

// CacheStats 缓存统计信息
type CacheStats struct {
	LocalHits   int64
	LocalMisses int64
	RedisHits   int64
	RedisMisses int64
	DBFetches   int64
}

// Cache 定义缓存接口
type Cache interface {
	Get(ctx context.Context, key string) (*CacheEntry, error)
	Set(ctx context.Context, entry *CacheEntry) error
	Delete(ctx context.Context, key string) error
	ClearLocal()
	GetOrSet(ctx context.Context, key string, fn func() (interface{}, error), ttl time.Duration) (interface{}, error)
	GetOrSetWithStruct(ctx context.Context, req *CacheRequest) (interface{}, error)
	GetOrSetWithHash(ctx context.Context, hashKey, field string, fn func() (interface{}, error), ttl time.Duration) (interface{}, error)
	GetOrSetWithSet(ctx context.Context, setKey string, fn func() ([]interface{}, error), ttl time.Duration) ([]interface{}, error)
	Stats() *CacheStats
	StartInvalidationListener(ctx context.Context)
	PublishInvalidation(ctx context.Context, key string) error
}

// CacheRequest 缓存请求结构
type CacheRequest struct {
	Key       string
	Fn        func() (interface{}, error)
	TTL       time.Duration
	UseHash   bool
	HashField string
}

// CacheConfig 缓存配置
type CacheConfig struct {
	LocalTTL              time.Duration
	RedisTTL              time.Duration
	StaleWhileRevalidate  time.Duration
	MaxParallelGenerators int
}

type tieredCache struct {
	localCache sync.Map
	redis      *redis.Client
	stats      CacheStats
	statsMu    sync.Mutex
	keyLocks   *keyLock
	config     CacheConfig
}

type keyLock struct {
	locks sync.Map
}

func (kl *keyLock) Lock(key string) func() {
	val, _ := kl.locks.LoadOrStore(key, &sync.Mutex{})
	mu := val.(*sync.Mutex)
	mu.Lock()
	return func() { mu.Unlock() }
}

func NewTieredCache(redisClient *redis.Client, config CacheConfig) Cache {
	if config.LocalTTL == 0 {
		config.LocalTTL = 5 * time.Minute
	}
	if config.RedisTTL == 0 {
		config.RedisTTL = 10 * time.Minute
	}
	if config.StaleWhileRevalidate == 0 {
		config.StaleWhileRevalidate = 1 * time.Minute
	}
	if config.MaxParallelGenerators == 0 {
		config.MaxParallelGenerators = 100
	}

	return &tieredCache{
		redis:    redisClient,
		keyLocks: &keyLock{},
		config:   config,
	}
}

func (c *tieredCache) Get(ctx context.Context, key string) (*CacheEntry, error) {
	if val, ok := c.localCache.Load(key); ok {
		entry := val.(*CacheEntry)
		if entry.Expiration.After(time.Now()) {
			c.statsMu.Lock()
			c.stats.LocalHits++
			c.statsMu.Unlock()
			return entry, nil
		}
		c.localCache.Delete(key)
	}

	c.statsMu.Lock()
	c.stats.LocalMisses++
	c.statsMu.Unlock()

	val, err := c.redis.Get(ctx, key).Result()
	if err == nil {
		var result interface{}
		if err := sonic.Unmarshal([]byte(val), &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal redis value: %w", err)
		}

		entry := &CacheEntry{
			Key:        key,
			Value:      result,
			Expiration: time.Now().Add(c.config.LocalTTL),
		}

		c.localCache.Store(key, entry)

		c.statsMu.Lock()
		c.stats.RedisHits++
		c.statsMu.Unlock()

		return entry, nil
	}

	if !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	c.statsMu.Lock()
	c.stats.RedisMisses++
	c.statsMu.Unlock()

	return nil, nil
}

func (c *tieredCache) Set(ctx context.Context, entry *CacheEntry) error {
	localEntry := &CacheEntry{
		Key:        entry.Key,
		Value:      entry.Value,
		Expiration: time.Now().Add(c.config.LocalTTL),
	}
	c.localCache.Store(entry.Key, localEntry)

	val, err := sonic.Marshal(entry.Value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.redis.Set(ctx, entry.Key, val, c.config.RedisTTL).Err()
}

func (c *tieredCache) Delete(ctx context.Context, key string) error {
	c.localCache.Delete(key)
	return c.redis.Del(ctx, key).Err()
}

func (c *tieredCache) ClearLocal() {
	c.localCache = sync.Map{}
}

func (c *tieredCache) GetOrSet(ctx context.Context, key string, fn func() (interface{}, error), ttl time.Duration) (interface{}, error) {
	if entry, err := c.Get(ctx, key); err != nil {
		return nil, err
	} else if entry != nil {
		return entry.Value, nil
	}

	unlock := c.keyLocks.Lock(key)
	defer unlock()

	if entry, err := c.Get(ctx, key); err != nil {
		return nil, err
	} else if entry != nil {
		return entry.Value, nil
	}

	c.statsMu.Lock()
	c.stats.DBFetches++
	c.statsMu.Unlock()

	result, err := fn()
	if err != nil {
		return nil, err
	}

	entry := &CacheEntry{
		Key:   key,
		Value: result,
	}
	if err := c.Set(ctx, entry); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *tieredCache) GetOrSetWithStruct(ctx context.Context, req *CacheRequest) (interface{}, error) {
	if req.UseHash {
		return c.GetOrSetWithHash(ctx, req.Key, req.HashField, req.Fn, req.TTL)
	}
	return c.GetOrSet(ctx, req.Key, req.Fn, req.TTL)
}

func (c *tieredCache) GetOrSetWithHash(ctx context.Context, hashKey, field string, fn func() (interface{}, error), ttl time.Duration) (interface{}, error) {
	if val, ok := c.localCache.Load(hashKey); ok {
		if hash, ok := val.(map[string]interface{}); ok {
			if value, exists := hash[field]; exists {
				c.statsMu.Lock()
				c.stats.LocalHits++
				c.statsMu.Unlock()
				return value, nil
			}
		}
	}

	c.statsMu.Lock()
	c.stats.LocalMisses++
	c.statsMu.Unlock()

	value, err := c.redis.HGet(ctx, hashKey, field).Result()
	if err == nil {
		var result interface{}
		if err := sonic.Unmarshal([]byte(value), &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal redis hash value: %w", err)
		}

		if val, ok := c.localCache.Load(hashKey); ok {
			if hash, ok := val.(map[string]interface{}); ok {
				hash[field] = result
			}
		} else {
			hash := map[string]interface{}{field: result}
			c.localCache.Store(hashKey, hash)
		}

		c.statsMu.Lock()
		c.stats.RedisHits++
		c.statsMu.Unlock()

		return result, nil
	}

	if !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("redis hget error: %w", err)
	}

	c.statsMu.Lock()
	c.stats.RedisMisses++
	c.statsMu.Unlock()

	unlock := c.keyLocks.Lock(hashKey + ":" + field)
	defer unlock()

	if value, err := c.redis.HGet(ctx, hashKey, field).Result(); err == nil {
		var result interface{}
		if err := sonic.Unmarshal([]byte(value), &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal redis hash value: %w", err)
		}
		return result, nil
	} else if !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("redis hget error: %w", err)
	}

	c.statsMu.Lock()
	c.stats.DBFetches++
	c.statsMu.Unlock()

	result, err := fn()
	if err != nil {
		return nil, err
	}

	val, err := sonic.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.redis.HSet(ctx, hashKey, field, val).Err(); err != nil {
		return nil, fmt.Errorf("redis hset error: %w", err)
	}

	if err := c.redis.Expire(ctx, hashKey, ttl).Err(); err != nil {
		return nil, fmt.Errorf("redis expire error: %w", err)
	}

	if val, ok := c.localCache.Load(hashKey); ok {
		if hash, ok := val.(map[string]interface{}); ok {
			hash[field] = result
		}
	} else {
		hash := map[string]interface{}{field: result}
		c.localCache.Store(hashKey, hash)
	}

	return result, nil
}

func (c *tieredCache) GetOrSetWithSet(ctx context.Context, setKey string, fn func() ([]interface{}, error), ttl time.Duration) ([]interface{}, error) {
	if val, ok := c.localCache.Load(setKey); ok {
		c.statsMu.Lock()
		c.stats.LocalHits++
		c.statsMu.Unlock()
		return val.([]interface{}), nil
	}

	c.statsMu.Lock()
	c.stats.LocalMisses++
	c.statsMu.Unlock()

	members, err := c.redis.SMembers(ctx, setKey).Result()
	if err != nil {
		return nil, fmt.Errorf("redis smembers error: %w", err)
	}

	if len(members) > 0 {
		var result []interface{}
		for _, member := range members {
			var val interface{}
			if err := sonic.Unmarshal([]byte(member), &val); err != nil {
				return nil, fmt.Errorf("failed to unmarshal redis set value: %w", err)
			}
			result = append(result, val)
		}

		c.localCache.Store(setKey, result)

		c.statsMu.Lock()
		c.stats.RedisHits++
		c.statsMu.Unlock()

		return result, nil
	}

	c.statsMu.Lock()
	c.stats.RedisMisses++
	c.statsMu.Unlock()

	unlock := c.keyLocks.Lock(setKey)
	defer unlock()

	if members, err := c.redis.SMembers(ctx, setKey).Result(); err != nil {
		return nil, fmt.Errorf("redis smembers error: %w", err)
	} else if len(members) > 0 {
		var result []interface{}
		for _, member := range members {
			var val interface{}
			if err := sonic.Unmarshal([]byte(member), &val); err != nil {
				return nil, fmt.Errorf("failed to unmarshal redis set value: %w", err)
			}
			result = append(result, val)
		}
		return result, nil
	}

	c.statsMu.Lock()
	c.stats.DBFetches++
	c.statsMu.Unlock()

	result, err := fn()
	if err != nil {
		return nil, err
	}

	pipe := c.redis.Pipeline()
	for _, item := range result {
		val, err := sonic.Marshal(item)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal value: %w", err)
		}
		pipe.SAdd(ctx, setKey, val)
	}
	pipe.Expire(ctx, setKey, ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("redis pipeline error: %w", err)
	}

	c.localCache.Store(setKey, result)

	return result, nil
}

func (c *tieredCache) Stats() *CacheStats {
	c.statsMu.Lock()
	defer c.statsMu.Unlock()
	return &CacheStats{
		LocalHits:   c.stats.LocalHits,
		LocalMisses: c.stats.LocalMisses,
		RedisHits:   c.stats.RedisHits,
		RedisMisses: c.stats.RedisMisses,
		DBFetches:   c.stats.DBFetches,
	}
}

func (c *tieredCache) StartInvalidationListener(ctx context.Context) {
	pubsub := c.redis.Subscribe(ctx, "cache_invalidation")
	ch := pubsub.Channel()

	go func() {
		for msg := range ch {
			c.localCache.Delete(msg.Payload)
		}
	}()
}

func (c *tieredCache) PublishInvalidation(ctx context.Context, key string) error {
	return c.redis.Publish(ctx, "cache_invalidation", key).Err()
}
