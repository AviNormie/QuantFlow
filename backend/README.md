# StockFlow Backend

Go microservices for StockFlow. Each service is an independent Gin HTTP server with its own module, port, and Docker image.

## Architecture

```text
Finnhub → market-service → Redis (cache + pub/sub) → websocket-service → clients
Clients → api-gateway → auth-service | market-service
```

## Services

| Service             | Default port | Health / Ready                          |
|---------------------|-------------:|-----------------------------------------|
| `api-gateway`       | 8080         | `/health`, `/ready`                     |
| `auth-service`      | 8084         | `/health`, `/ready`                     |
| `market-service`    | 8082         | `/health`, `/ready`                     |
| `websocket-service` | 8083         | `/health`, `/ready`                     |

**API gateway routes**

- `/api/auth/*` → auth-service (signup, register, login, refresh, logout, me)
- `/api/market/*` → market-service (symbols, quotes, candles)

**WebSocket:** `ws://localhost:8083/ws?token=<access_jwt>` with JSON subscribe/unsubscribe messages.

## Auth endpoints (via gateway)

```text
POST /api/auth/signup
POST /api/auth/register   (alias)
POST /api/auth/login
POST /api/auth/refresh
POST /api/auth/logout
GET  /api/auth/me
```

## Market endpoints (via gateway)

```text
GET /api/market/symbols/search?q=AAPL
GET /api/market/symbols/{symbol}
GET /api/market/quotes/{symbol}
GET /api/market/candles/{symbol}?resolution=D&from=&to=
```

Compatibility query routes: `/api/market/quote?symbol=`, `/api/market/candles?symbol=`.

## Prerequisites

- **Go 1.25+**
- **Docker** (optional)

## Environment variables

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL (auth-service) |
| `REDIS_URL` | Redis for sessions, cache, pub/sub |
| `JWT_SECRET` | JWT signing (auth-service, websocket optional auth) |
| `AUTH_SERVICE_URL` | Gateway → auth-service |
| `MARKET_SERVICE_URL` | Gateway → market-service |
| `FINNHUB_API_KEY` | market-service provider |
| `MARKET_PUBSUB_CHANNEL` | Redis channel (default `market:updates`) |
| `MARKET_DEFAULT_SYMBOLS` | Symbols to ingest (default `AAPL,MSFT,GOOGL`) |
| `WS_REQUIRE_AUTH` | Require JWT on websocket connect (`true`/`false`) |

See root `.env.example` for Sentry, PostHog, and port overrides.

## Project layout

```
<service>/
├── cmd/main.go
├── internal/
│   ├── handler/
│   ├── service/
│   ├── repository/
│   ├── middleware/
│   └── model/
├── Dockerfile
└── go.mod
```

`shared/` provides redis client, JWT, health probes, gateway middleware, and monitoring.

## Run locally

```bash
docker compose up -d postgres redis
cd backend/auth-service && JWT_SECRET=dev-secret DATABASE_URL=... REDIS_URL=... go run ./cmd
# repeat for api-gateway, market-service, websocket-service
```

Or from repo root:

```bash
docker compose up --build
```

## Verify

```bash
curl http://localhost:8080/health
curl http://localhost:8080/ready
curl -X POST http://localhost:8080/api/auth/signup \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@example.com","password":"password123"}'
curl "http://localhost:8080/api/market/symbols/search?q=AAPL"
```

## Tests

```bash
cd backend/auth-service && go test ./...
cd backend/market-service && go test ./...
cd backend/websocket-service && go test ./...
cd backend/shared && go test ./...
```

## Monitoring

See [`../MONITORING.md`](../MONITORING.md) for Sentry and PostHog setup.
