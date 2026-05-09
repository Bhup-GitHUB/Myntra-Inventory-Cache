package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr       string
	MySQLDSN       string
	RedisAddr      string
	RedisPassword  string
	RabbitMQURL    string
	APIInternalURL string
	L1TTL          time.Duration
	L2TTL          time.Duration
	ProductTTL     time.Duration
}

func Load() Config {
	return Config{
		HTTPAddr:       env("HTTP_ADDR", ":8080"),
		MySQLDSN:       env("MYSQL_DSN", "inventory:inventory@tcp(localhost:3306)/inventory_cache?parseTime=true&multiStatements=true"),
		RedisAddr:      env("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  os.Getenv("REDIS_PASSWORD"),
		RabbitMQURL:    env("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		APIInternalURL: env("API_INTERNAL_BASE_URL", "http://localhost:8080"),
		L1TTL:          envDuration("L1_TTL_SECONDS", 5*time.Second),
		L2TTL:          envDuration("L2_TTL_SECONDS", 60*time.Second),
		ProductTTL:     envDuration("PRODUCT_TTL_SECONDS", 5*time.Minute),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}
