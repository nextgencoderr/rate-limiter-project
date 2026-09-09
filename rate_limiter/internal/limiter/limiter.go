package limiter

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/ratelimiter/service/internal/policy"
	"github.com/redis/go-redis/v9"
)

//go:embed scripts/token_bucket.lua
var tokenBucketScript string

//go:embed scripts/sliding_window.lua
var slidingWindowScript string

type Service struct {
	rdb       *redis.Client
	tbSHA     string
	swSHA     string
	keyPrefix string
}

func New(rdb *redis.Client, keyPrefix string) (*Service, error) {
	s := &Service{rdb: rdb, keyPrefix: keyPrefix}
	ctx := context.Background()
	tb, err := rdb.ScriptLoad(ctx, tokenBucketScript).Result()
	if err != nil {
		return nil, fmt.Errorf("load token bucket script: %w", err)
	}
	sw, err := rdb.ScriptLoad(ctx, slidingWindowScript).Result()
	if err != nil {
		return nil, fmt.Errorf("load sliding window script: %w", err)
	}
	s.tbSHA = tb
	s.swSHA = sw
	return s, nil
}

type Result struct {
	Allowed   bool    `json:"allowed"`
	Remaining float64 `json:"remaining"`
	Algorithm string  `json:"algorithm"`
}

func (s *Service) Allow(ctx context.Context, p policy.Policy) (Result, error) {
	now := time.Now().UnixMilli()
	switch p.Algorithm {
	case policy.TokenBucket:
		return s.tokenBucket(ctx, p, now)
	case policy.SlidingWindow:
		return s.slidingWindow(ctx, p, now)
	default:
		return Result{}, fmt.Errorf("unknown algorithm: %s", p.Algorithm)
	}
}

func (s *Service) tokenBucket(ctx context.Context, p policy.Policy, nowMs int64) (Result, error) {
	refill := p.RefillRate
	if refill <= 0 {
		refill = float64(p.Limit) / float64(max(p.WindowSec, 1))
	}
	key := s.keyPrefix + ":tb:" + p.ClientID
	raw, err := s.rdb.EvalSha(ctx, s.tbSHA, []string{key},
		p.Limit, refill, nowMs, 1,
	).Result()
	if err != nil {
		return Result{}, err
	}
	vals, ok := raw.([]interface{})
	if !ok || len(vals) < 2 {
		return Result{}, fmt.Errorf("unexpected token bucket response")
	}
	allowed := toInt(vals[0]) == 1
	return Result{
		Allowed:   allowed,
		Remaining: toFloat(vals[1]),
		Algorithm: string(policy.TokenBucket),
	}, nil
}

func (s *Service) slidingWindow(ctx context.Context, p policy.Policy, nowMs int64) (Result, error) {
	windowMs := p.WindowSec * 1000
	if windowMs <= 0 {
		windowMs = 60_000
	}
	key := s.keyPrefix + ":sw:" + p.ClientID
	member := uniqueID(nowMs)
	raw, err := s.rdb.EvalSha(ctx, s.swSHA, []string{key},
		windowMs, p.Limit, nowMs, member,
	).Result()
	if err != nil {
		return Result{}, err
	}
	vals, ok := raw.([]interface{})
	if !ok || len(vals) < 2 {
		return Result{}, fmt.Errorf("unexpected sliding window response")
	}
	allowed := toInt(vals[0]) == 1
	return Result{
		Allowed:   allowed,
		Remaining: toFloat(vals[1]),
		Algorithm: string(policy.SlidingWindow),
	}, nil
}

func (s *Service) Ping(ctx context.Context) error {
	return s.rdb.Ping(ctx).Err()
}

func toInt(v interface{}) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case string:
		i, _ := strconv.ParseInt(n, 10, 64)
		return i
	default:
		return 0
	}
}

func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int64:
		return float64(n)
	case string:
		f, _ := strconv.ParseFloat(n, 64)
		return f
	default:
		return 0
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func uniqueID(nowMs int64) string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("%d-%s", nowMs, hex.EncodeToString(b[:]))
}
