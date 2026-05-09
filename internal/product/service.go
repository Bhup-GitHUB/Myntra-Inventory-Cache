package product

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"

	"inventory-cache-lab/internal/cache"
)

type Service struct {
	repo   *Repository
	redis  *cache.RedisCache
	logger *slog.Logger
}

func NewService(repo *Repository, redis *cache.RedisCache, logger *slog.Logger) *Service {
	return &Service{repo: repo, redis: redis, logger: logger}
}

func (s *Service) GetProduct(ctx context.Context, productID int64) (Response, error) {
	key := cache.ProductKey(productID)
	if value, ok, err := s.redis.Get(ctx, key); err == nil && ok {
		var product Product
		if err := json.Unmarshal(value, &product); err == nil {
			return response(product, "cache"), nil
		}
	} else if err != nil {
		s.logger.WarnContext(ctx, "product_cache_read_failed", "product_id", productID, "error", err)
	}

	product, err := s.repo.GetByID(ctx, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Response{}, ErrNotFound
		}
		return Response{}, err
	}
	body, err := json.Marshal(product)
	if err == nil {
		if err := s.redis.Set(ctx, key, body); err != nil {
			s.logger.WarnContext(ctx, "product_cache_write_failed", "product_id", productID, "error", err)
		}
	}
	return response(product, "db"), nil
}

func response(product Product, source string) Response {
	return Response{
		ID:         product.ID,
		Name:       product.Name,
		Brand:      product.Brand,
		PriceCents: product.PriceCents,
		Cache:      source,
	}
}

var ErrNotFound = errors.New("product_not_found")
