package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ratelimiter/service/internal/api"
	"github.com/ratelimiter/service/internal/config"
	"github.com/ratelimiter/service/internal/limiter"
	"github.com/ratelimiter/service/internal/policy"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis ping failed: %v", err)
	}

	lim, err := limiter.New(rdb, cfg.KeyPrefix)
	if err != nil {
		log.Fatalf("limiter init: %v", err)
	}

	store := policy.NewStore(cfg.DefaultPolicy)
	handler := api.New(lim, store)

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      handler.Router(),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	go func() {
		log.Printf("listening on %s", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
	_ = rdb.Close()
	log.Println("stopped")
}
