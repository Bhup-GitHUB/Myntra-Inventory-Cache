package cache

import (
	"context"
	"testing"
	"time"
)

func TestL1CacheGetSetDeleteAndStats(t *testing.T) {
	ctx := context.Background()
	cache := NewL1Cache(50 * time.Millisecond)

	if _, ok := cache.Get(ctx, "inventory:101"); ok {
		t.Fatal("expected cold cache miss")
	}

	cache.Set(ctx, "inventory:101", []byte(`{"product_id":101}`))
	value, ok := cache.Get(ctx, "inventory:101")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if string(value) != `{"product_id":101}` {
		t.Fatalf("unexpected cached value: %s", value)
	}

	cache.Delete(ctx, "inventory:101")
	if _, ok := cache.Get(ctx, "inventory:101"); ok {
		t.Fatal("expected miss after delete")
	}

	stats := cache.Stats()
	if stats.Hits != 1 || stats.Misses != 2 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

func TestL1CacheExpires(t *testing.T) {
	ctx := context.Background()
	cache := NewL1Cache(10 * time.Millisecond)

	cache.Set(ctx, "inventory:101", []byte("value"))
	time.Sleep(20 * time.Millisecond)

	if _, ok := cache.Get(ctx, "inventory:101"); ok {
		t.Fatal("expected expired value to miss")
	}
}
