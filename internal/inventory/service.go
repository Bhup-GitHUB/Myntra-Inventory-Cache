package inventory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"

	"golang.org/x/sync/singleflight"

	"inventory-cache-lab/internal/cache"
)

type Service struct {
	repo   *Repository
	l1     *cache.L1Cache
	redis  *cache.RedisCache
	group  singleflight.Group
	logger *slog.Logger
}

func NewService(repo *Repository, l1 *cache.L1Cache, redis *cache.RedisCache, logger *slog.Logger) *Service {
	return &Service{repo: repo, l1: l1, redis: redis, logger: logger}
}

func (s *Service) GetInventory(ctx context.Context, productID int64) (Response, error) {
	key := cache.InventoryKey(productID)
	path := []string{"api"}

	if value, ok := s.l1.Get(ctx, key); ok {
		var inventory Inventory
		if err := json.Unmarshal(value, &inventory); err == nil {
			s.logger.InfoContext(ctx, "inventory_read", "product_id", productID, "source", "l1")
			return buildResponse(inventory, "l1", []string{"api", "l1_cache"}), nil
		}
	}
	path = append(path, "l1_miss")

	if value, ok, err := s.redis.Get(ctx, key); err == nil && ok {
		var inventory Inventory
		if err := json.Unmarshal(value, &inventory); err == nil {
			s.l1.Set(ctx, key, value)
			s.logger.InfoContext(ctx, "inventory_read", "product_id", productID, "source", "l2")
			return buildResponse(inventory, "l2", append(path, "l2_cache", "fill_l1")), nil
		}
	} else if err != nil {
		s.logger.WarnContext(ctx, "inventory_l2_read_failed", "product_id", productID, "error", err)
	}
	path = append(path, "l2_miss")

	// When a hot key expires, singleflight keeps one goroutine on the DB path
	// while the rest wait for the same result.
	result, err, _ := s.group.Do(key, func() (any, error) {
		inventory, err := s.repo.GetByProductID(ctx, productID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return Inventory{}, ErrNotFound
			}
			return Inventory{}, err
		}
		body, err := json.Marshal(inventory)
		if err == nil {
			if err := s.redis.Set(ctx, key, body); err != nil {
				s.logger.WarnContext(ctx, "inventory_l2_write_failed", "product_id", productID, "error", err)
			}
			s.l1.Set(ctx, key, body)
		}
		return inventory, nil
	})
	if err != nil {
		return Response{}, err
	}

	inventory := result.(Inventory)
	s.logger.InfoContext(ctx, "inventory_read", "product_id", productID, "source", "db")
	path = append(path, "mysql_read_replica", "fill_l2", "fill_l1")
	return buildResponse(inventory, "db", path), nil
}

func (s *Service) InvalidateInventory(ctx context.Context, productID int64) error {
	key := cache.InventoryKey(productID)
	s.l1.Delete(ctx, key)
	if err := s.redis.Delete(ctx, key); err != nil {
		return err
	}
	s.logger.InfoContext(ctx, "cache_invalidated", "key", key)
	return nil
}

func buildResponse(inventory Inventory, source string, path []string) Response {
	return Response{
		ProductID:    inventory.ProductID,
		AvailableQty: inventory.AvailableQty,
		ReservedQty:  inventory.ReservedQty,
		CacheSource:  source,
		RequestPath:  path,
	}
}

var ErrNotFound = errors.New("inventory_not_found")
