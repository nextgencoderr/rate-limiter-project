package config

import (
	"os"
	"strconv"
	"time"

	"github.com/ratelimiter/service/internal/policy"
)

type Config struct {
	Addr         string
	RedisAddr    string
	RedisPass    string
	RedisDB      int
	KeyPrefix    string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	DefaultPolicy policy.Policy
}

func Load() Config {
	return Config{
		Addr:         env("HTTP_ADDR", ":8080"),
		RedisAddr:    env("REDIS_ADDR", "localhost:6379"),
		RedisPass:    env("REDIS_PASSWORD", ""),
		RedisDB:      envInt("REDIS_DB", 0),
		KeyPrefix:    env("REDIS_KEY_PREFIX", "rl"),
		ReadTimeout:  envDuration("HTTP_READ_TIMEOUT", 5*time.Second),
		WriteTimeout: envDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
		DefaultPolicy: policy.Policy{
			Algorithm:  policy.Algorithm(env("DEFAULT_ALGORITHM", "sliding_window")),
			Limit:      envInt("DEFAULT_LIMIT", 100),
			WindowSec:  envInt("DEFAULT_WINDOW_SEC", 60),
			RefillRate: envFloat("DEFAULT_REFILL_RATE", 10),
		},
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			return n
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
