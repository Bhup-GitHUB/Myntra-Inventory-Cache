package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	amqp "github.com/rabbitmq/amqp091-go"

	"inventory-cache-lab/internal/cache"
	"inventory-cache-lab/internal/checkout"
	"inventory-cache-lab/internal/config"
	"inventory-cache-lab/internal/db"
	"inventory-cache-lab/internal/events"
	"inventory-cache-lab/internal/httpx"
	"inventory-cache-lab/internal/inventory"
	"inventory-cache-lab/internal/observability"
	"inventory-cache-lab/internal/product"
)

func main() {
	logger := observability.NewLogger()
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mysqlDB, err := openMySQLWithRetry(ctx, cfg.MySQLDSN, logger)
	if err != nil {
		logger.Error("mysql_connect_failed", "error", err)
		os.Exit(1)
	}
	defer mysqlDB.Close()

	if err := db.RunMigrations(mysqlDB, "migrations"); err != nil {
		logger.Error("migrations_failed", "error", err)
		os.Exit(1)
	}

	l1 := cache.NewL1Cache(cfg.L1TTL)
	redisCache := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, cfg.L2TTL)
	defer redisCache.Close()
	if err := pingRedisWithRetry(ctx, redisCache, logger); err != nil {
		logger.Error("redis_connect_failed", "error", err)
		os.Exit(1)
	}

	rabbitConn, err := dialRabbitWithRetry(ctx, cfg.RabbitMQURL, logger)
	if err != nil {
		logger.Error("rabbitmq_connect_failed", "error", err)
		os.Exit(1)
	}
	defer rabbitConn.Close()

	publisher, err := events.NewPublisher(ctx, rabbitConn)
	if err != nil {
		logger.Error("rabbitmq_publisher_failed", "error", err)
		os.Exit(1)
	}
	defer publisher.Close()

	productService := product.NewService(product.NewRepository(mysqlDB), redisCache, logger)
	inventoryService := inventory.NewService(inventory.NewRepository(mysqlDB), l1, redisCache, logger)
	checkoutService := checkout.NewService(mysqlDB, publisher, logger)

	router := buildRouter(mysqlDB, l1, redisCache, productService, inventoryService, checkoutService, logger)
	server := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("api_started", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api_failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("api_shutdown_failed", "error", err)
	}
}

func buildRouter(
	mysqlDB *sql.DB,
	l1 *cache.L1Cache,
	redisCache *cache.RedisCache,
	productService *product.Service,
	inventoryService *inventory.Service,
	checkoutService *checkout.Service,
	logger *slog.Logger,
) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(httpx.SimulatedInfra(logger))
	router.Use(middleware.Recoverer)

	productHandler := product.NewHandler(productService)
	inventoryHandler := inventory.NewHandler(inventoryService)
	checkoutHandler := checkout.NewHandler(checkoutService)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	router.Get("/products/{productID}", productHandler.GetProduct)
	router.Get("/inventory/{productID}", inventoryHandler.GetInventory)
	router.Post("/inventory/{productID}/invalidate", inventoryHandler.InvalidateInventory)
	router.Post("/checkout", checkoutHandler.PlaceOrder)

	router.Get("/debug/cache/stats", func(w http.ResponseWriter, r *http.Request) {
		l1Stats := l1.Stats()
		l2Stats := redisCache.Stats()
		httpx.JSON(w, http.StatusOK, observability.CacheStats{
			L1Hits:   l1Stats.Hits,
			L1Misses: l1Stats.Misses,
			L2Hits:   l2Stats.Hits,
			L2Misses: l2Stats.Misses,
		})
	})
	router.Post("/debug/reset", func(w http.ResponseWriter, r *http.Request) {
		if err := db.ResetDemoData(r.Context(), mysqlDB); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "reset_failed")
			return
		}
		for _, productID := range []int64{101, 102, 103} {
			l1.Delete(r.Context(), cache.InventoryKey(productID))
			_ = redisCache.Delete(r.Context(), cache.InventoryKey(productID))
			_ = redisCache.Delete(r.Context(), cache.ProductKey(productID))
		}
		httpx.JSON(w, http.StatusOK, map[string]bool{"reset": true})
	})

	return router
}

func openMySQLWithRetry(ctx context.Context, dsn string, logger *slog.Logger) (*sql.DB, error) {
	var lastErr error
	for i := 0; i < 30; i++ {
		conn, err := db.OpenMySQL(ctx, dsn)
		if err == nil {
			return conn, nil
		}
		lastErr = err
		logger.InfoContext(ctx, "mysql_waiting", "attempt", i+1)
		if !sleep(ctx, 2*time.Second) {
			break
		}
	}
	return nil, lastErr
}

func pingRedisWithRetry(ctx context.Context, redisCache *cache.RedisCache, logger *slog.Logger) error {
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

func dialRabbitWithRetry(ctx context.Context, url string, logger *slog.Logger) (*amqp.Connection, error) {
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
