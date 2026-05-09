package cache

import (
	"context"
	"sync"
	"time"
)

type L1Cache struct {
	mu     sync.RWMutex
	items  map[string]Item
	ttl    time.Duration
	hits   uint64
	misses uint64
}

type Item struct {
	Value     []byte
	ExpiresAt time.Time
}

type L1Stats struct {
	Hits   uint64
	Misses uint64
}

func NewL1Cache(ttl time.Duration) *L1Cache {
	return &L1Cache{
		items: make(map[string]Item),
		ttl:   ttl,
	}
}

// L1 stays inside one API process. It is the fastest read path, but every API
// replica owns its own copy, so invalidation needs an explicit strategy.
func (c *L1Cache) Get(ctx context.Context, key string) ([]byte, bool) {
	_ = ctx
	now := time.Now()

	c.mu.RLock()
	item, ok := c.items[key]
	if ok && now.Before(item.ExpiresAt) {
		value := append([]byte(nil), item.Value...)
		c.mu.RUnlock()
		c.mu.Lock()
		c.hits++
		c.mu.Unlock()
		return value, true
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if item, ok := c.items[key]; ok && now.After(item.ExpiresAt) {
		delete(c.items, key)
	}
	c.misses++
	return nil, false
}

func (c *L1Cache) Set(ctx context.Context, key string, value []byte) {
	_ = ctx
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = Item{
		Value:     append([]byte(nil), value...),
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

func (c *L1Cache) Delete(ctx context.Context, key string) {
	_ = ctx
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *L1Cache) Stats() L1Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return L1Stats{Hits: c.hits, Misses: c.misses}
}
