# StockFlow Frontend

Next.js app for live market charts, signup/signin, and API documentation.

## Prerequisites

- Node.js 20+
- Yarn (or npm)
- Running backend: API gateway (`8080`), auth-service (`8084`), websocket-service (`8083`)
- PostgreSQL with the `users` table migrated (see [backend README](../backend/README.md))

## Setup

```bash
cd stockflow
cp .env.example .env.local
# Edit .env.local — set FINNHUB_API_KEY for market quotes (Next.js API routes)
yarn install
```

## Environment variables

| Variable | Description |
|----------|-------------|
| `NEXT_PUBLIC_API_URL` | API gateway base URL (default `http://localhost:8080`) |
| `NEXT_PUBLIC_WEBSOCKET_URL` | WebSocket service URL (default `ws://localhost:8083`) |
| `FINNHUB_API_KEY` | Finnhub API key for quote/candle routes (server-side only) |

## Run locally

**Terminal 1 — backend (from repo root):**

```bash
docker compose up -d postgres redis
# Run users migration once (see backend/README.md)
cd backend/auth-service && set -a && source .env.local && set +a && go run ./cmd
cd backend/api-gateway && go run ./cmd
cd backend/websocket-service && set -a && source .env.local && set +a && go run ./cmd
```

**Terminal 2 — frontend:**

```bash
cd stockflow
yarn dev
```

Open [http://localhost:3000](http://localhost:3000).

## Signup / login flow

1. Go to [http://localhost:3000/signup](http://localhost:3000/signup) and create an account (email + password, min 8 chars).
2. The frontend calls `POST http://localhost:8080/api/auth/register` via the API gateway.
3. JWT access and refresh tokens are stored in `localStorage`.
4. Sign in at [http://localhost:3000/login](http://localhost:3000/login) — calls `POST /api/auth/login`.
5. `/charts` requires authentication; unauthenticated users are redirected to `/login`.

## API gateway URLs

| Endpoint | Method | Description |
|----------|--------|-------------|
| `http://localhost:8080/health` | GET | Gateway health |
| `http://localhost:8080/api/auth/register` | POST | Create account |
| `http://localhost:8080/api/auth/login` | POST | Sign in |
| `http://localhost:8080/api/market/quote?symbol=AAPL` | GET | Stock quote (proxied to Next.js) |
| `http://localhost:8080/api/market/candles?symbol=AAPL` | GET | OHLCV candles (proxied to Next.js) |
| `ws://localhost:8083/ws?symbol=AAPL` | WebSocket | Live prices (direct to websocket-service) |

Interactive API docs: [http://localhost:3000/api/docs](http://localhost:3000/api/docs)

## Test signup/login with curl

```bash
# Register
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

## Build

```bash
yarn build
yarn start
```

## Deploy on Vercel

Set `NEXT_PUBLIC_API_URL` to your deployed API gateway URL and `NEXT_PUBLIC_WEBSOCKET_URL` to your websocket-service URL. Add `https://*.vercel.app` to gateway `ALLOWED_ORIGINS`.
