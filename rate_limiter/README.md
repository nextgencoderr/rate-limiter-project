# Rate Limiter Service

Go microservice implementing **Token Bucket** and **Sliding Window** rate limiting with per-client policies, Redis-backed atomic counters (Lua scripts), Prometheus metrics, and Kubernetes deployment manifests.

## Features

- **Algorithms**: `token_bucket` and `sliding_window`, selected per client
- **Per-client throttling**: policies keyed by `client_id`
- **Redis + Lua**: atomic check-and-update across concurrent requests
- **Runtime policy API**: update limits without restart
- **Observability**: `/metrics` (Prometheus), request and throttle counters
- **Health**: `/healthz` (liveness), `/readyz` (readiness, includes Redis ping)
- **Docker**: multi-stage build on Alpine
- **Kubernetes**: Deployment, ConfigMap, Service, HPA, probes

## Quick start (Docker Compose)

```bash
docker compose up --build
```

```bash
# Check a client (default: 100 req / 60s sliding window)
curl -s -X POST http://localhost:8080/v1/check \
  -H "Content-Type: application/json" \
  -d '{"client_id":"user-1"}'

# Set token-bucket policy at runtime
curl -s -X PUT http://localhost:8080/v1/policies/user-1 \
  -H "Content-Type: application/json" \
  -d '{"algorithm":"token_bucket","limit":50,"refill_rate":5}'

curl -s http://localhost:8080/metrics
```

## API

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/check` | Body: `{"client_id":"..."}` — allow or 429 |
| GET | `/v1/policies` | List configured client policies |
| GET | `/v1/policies/{id}` | Get effective policy for client |
| PUT | `/v1/policies/{id}` | Create/update policy (no restart) |
| DELETE | `/v1/policies/{id}` | Remove client override |
| GET | `/metrics` | Prometheus metrics |
| GET | `/healthz` | Liveness |
| GET | `/readyz` | Readiness (Redis) |

### Policy fields

| Field | Description |
|-------|-------------|
| `algorithm` | `token_bucket` or `sliding_window` |
| `limit` | Bucket capacity or max requests per window |
| `window_sec` | Sliding window length (sliding window only) |
| `refill_rate` | Tokens per second (token bucket only) |

## Local development

Requires Go 1.22+ and Redis.

```bash
go mod tidy
go run ./cmd/server
```

Environment variables: `HTTP_ADDR`, `REDIS_ADDR`, `REDIS_PASSWORD`, `DEFAULT_ALGORITHM`, `DEFAULT_LIMIT`, `DEFAULT_WINDOW_SEC`, `DEFAULT_REFILL_RATE`.

## Minikube

```bash
minikube start
# Linux/macOS
./scripts/minikube-deploy.sh
# Windows (PowerShell)
.\scripts\minikube-deploy.ps1
```

## Metrics

- `rate_limit_requests_total{client_id,algorithm,result}` — `result` is `allowed` or `throttled`
- `rate_limit_throttle_events_total{client_id,algorithm}`
- `rate_limit_policy_updates_total`

## Project layout

```
cmd/server/          HTTP entrypoint
internal/api/        REST handlers
internal/limiter/    Redis Lua limiters
internal/policy/     In-memory policy store
k8s/                 Kubernetes manifests
lua/                 Reference Lua scripts (embedded copies in internal/limiter/scripts)
```
