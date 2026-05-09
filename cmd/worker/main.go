package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"inventory-cache-lab/internal/cache"
	"inventory-cache-lab/internal/config"
	"inventory-cache-lab/internal/events"
	"inventory-cache-lab/internal/observability"
)

func main() {
	logger := observability.NewLogger()
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	redisCache := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, cfg.L2TTL)
	defer redisCache.Close()
	if err := waitForRedis(ctx, redisCache, logger); err != nil {
		logger.Error("redis_connect_failed", "error", err)
		os.Exit(1)
	}

	rabbitConn, err := dialRabbit(ctx, cfg.RabbitMQURL, logger)
	if err != nil {
		logger.Error("rabbitmq_connect_failed", "error", err)
		os.Exit(1)
	}
	defer rabbitConn.Close()

	consumer, err := events.NewConsumer(ctx, rabbitConn)
	if err != nil {
		logger.Error("rabbitmq_consumer_failed", "error", err)
		os.Exit(1)
	}
	defer consumer.Close()

	deliveries, err := consumer.Consume(ctx)
	if err != nil {
		logger.Error("rabbitmq_consume_failed", "error", err)
		os.Exit(1)
	}

	client := &http.Client{Timeout: 3 * time.Second}
	logger.Info("worker_started")
	for {
		select {
		case <-ctx.Done():
			return
		case delivery, ok := <-deliveries:
			if !ok {
				return
			}
			if err := handleDelivery(ctx, delivery, redisCache, cfg.APIInternalURL, client, logger); err != nil {
				logger.ErrorContext(ctx, "inventory_event_failed", "error", err)
				_ = delivery.Nack(false, true)
				continue
			}
			_ = delivery.Ack(false)
		}
	}
}

func handleDelivery(ctx context.Context, delivery amqp.Delivery, redisCache *cache.RedisCache, apiBaseURL string, client *http.Client, logger *slog.Logger) error {
	event, err := events.DecodeInventoryUpdated(delivery)
	if err != nil {
		return err
	}
	logger.InfoContext(ctx, "inventory_event_received", "product_id", event.ProductID, "version", event.Version)

	key := cache.InventoryKey(event.ProductID)
	if err := redisCache.Delete(ctx, key); err != nil {
		return err
	}
	logger.InfoContext(ctx, "cache_invalidated", "key", key)

	url := fmt.Sprintf("%s/inventory/%d/invalidate", apiBaseURL, event.ProductID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(nil))
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("api_l1_invalidation_failed status=%d", resp.StatusCode)
	}
	return nil
}

func waitForRedis(ctx context.Context, redisCache *cache.RedisCache, logger *slog.Logger) error {
	var lastErr error
	for i := 0; i < 30; i++ {
		if err := redisCache.Ping(ctx); err == nil {
			return nil
		} else {
			lastErr = err
		}
		logger.InfoContext(ctx, "redis_waiting", "attempt", i+1)
		if !sleep(ctx, 2*time.Second) {
			break
		}
	}
	return lastErr
}

func dialRabbit(ctx context.Context, url string, logger *slog.Logger) (*amqp.Connection, error) {
	var lastErr error
	for i := 0; i < 30; i++ {
		conn, err := events.Dial(url)
		if err == nil {
			return conn, nil
		}
		lastErr = err
		logger.InfoContext(ctx, "rabbitmq_waiting", "attempt", i+1)
		if !sleep(ctx, 2*time.Second) {
			break
		}
	}
	return nil, lastErr
}

func sleep(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
