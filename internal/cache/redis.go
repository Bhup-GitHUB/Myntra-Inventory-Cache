package cache

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
	hits   atomic.Uint64
	misses atomic.Uint64
}

type RedisStats struct {
	Hits   uint64
	Misses uint64
}

func NewRedisCache(addr, password string, ttl time.Duration) *RedisCache {
	return &RedisCache{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
		}),
		ttl: ttl,
	}
}

// Redis is the shared cache layer. It costs a network call, but it keeps
// repeated misses away from MySQL across all API instances.
func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	value, err := c.client.Get(ctx, key).Bytes()
	if err == nil {
		c.hits.Add(1)
		return value, true, nil
	}
	if errors.Is(err, redis.Nil) {
		c.misses.Add(1)
		return nil, false, nil
	}
	return nil, false, err
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte) error {
	return c.client.Set(ctx, key, value, c.ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) Stats() RedisStats {
	return RedisStats{Hits: c.hits.Load(), Misses: c.misses.Load()}
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}
