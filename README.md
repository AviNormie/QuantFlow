# StockFlow

Distributed real-time market infrastructure: Go microservices, JWT auth, Redis pub/sub, and TradingView charts.

## Quick start

```bash
cp .env.example .env
# Set FINNHUB_API_KEY and JWT_SECRET in .env

docker compose up --build
```

Frontend (from `stockflow/`):

```bash
cd stockflow && corepack enable && pnpm install && pnpm dev
```

Open http://localhost:3000 — signup/login, then charts at `/charts`.

## Architecture

```text
Finnhub → market-service → Redis → websocket-service → browser
Browser → api-gateway → auth-service | market-service
TradingView chart ← StockFlow datafeed ← gateway market APIs + WebSocket ticks
```

## Docs

- [Deployment (Render / Vercel)](DEPLOYMENT.md)
- [Frontend (stockflow/)](stockflow/README.md)
- [Backend services](backend/README.md)
- [Monitoring](MONITORING.md)

## Verify backend

```bash
curl localhost:8080/ready
curl -X POST localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"yourpassword"}'
curl "localhost:8080/api/market/quotes/AAPL"
```
