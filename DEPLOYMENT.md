# StockFlow Deployment Guide

This guide covers deploying StockFlow to **Render** (backend) and **Vercel or Render** (frontend).

## Recommendation: deploy microservices individually (not one mega-Dockerfile)

| Approach | Use when | On Render |
|----------|----------|------------|
| **Individual services** (recommended) | Production, learning microservice boundaries | One Web/Private Service per Go app + managed Postgres + Redis |
| **docker-compose** | Local development only | **Not supported** as a single deploy unit |
| **Single Dockerfile (all services)** | Avoid | One container running 4 binaries + DB is fragile and not true microservices |

Render runs **one process per Web Service**. Use [`render.yaml`](render.yaml) (Blueprint) to provision everything together.

---

## Architecture on Render

```text
Browser
  ├─ HTTPS → stockflow-api-gateway (public) → auth / market (private)
  └─ WSS   → stockflow-websocket (public) ← Redis pub/sub ← market (private)

Managed: PostgreSQL (auth) + Redis (sessions, cache, pub/sub)
```

**Public URLs (browser):**

- API: `https://stockflow-api-gateway.onrender.com`
- WebSocket: `wss://stockflow-websocket.onrender.com/ws`

**Private (internal network only):**

- `stockflow-auth` — auth-service
- `stockflow-market` — market-service

---

## Step 1 — Deploy backend with Render Blueprint

StockFlow is **not on Render yet** until you launch the blueprint once.

1. Open [Render Blueprints](https://dashboard.render.com/blueprints)
2. **New Blueprint Instance** → connect `AviNormie/QuantFlow` → branch `main`
3. When prompted (`sync: false` vars), enter:
   - `FINNHUB_API_KEY` — your Finnhub key
   - `ALLOWED_ORIGINS` — your Vercel URL (e.g. `https://quantflow.vercel.app`)
4. Wait for all services to become **Live**

Validate locally before launching:

```bash
render workspace set tea-ct3m35jtq21c738tajq0
render blueprints validate ./render.yaml
```

### Push env vars via CLI (after blueprint exists)

```bash
cp deploy/render.manual.env.example deploy/render.manual.env
# edit FINNHUB_API_KEY and ALLOWED_ORIGINS
./scripts/render-sync-env.sh
```

The script sets manual env vars on Render, triggers deploys, and prints the Vercel URLs to use.


---

## Step 2 — Deploy frontend

### Option A — Vercel (common)

1. Import `stockflow/` as the project root (or monorepo subfolder).
2. Set environment variables:

| Variable | Value |
|----------|--------|
| `NEXT_PUBLIC_API_URL` | `https://stockflow-api-gateway.onrender.com` |
| `NEXT_PUBLIC_WEBSOCKET_URL` | `https://stockflow-websocket.onrender.com` |

`NEXT_PUBLIC_*` are baked in at **build time** — redeploy after changing them.

3. Build command: `pnpm install && pnpm run build`  
   Install command: `corepack enable && pnpm install`

4. Update `ALLOWED_ORIGINS` on api-gateway to include your Vercel URL.

### Option B — Render (included in blueprint)

The blueprint deploys `stockflow-frontend` with `RENDER_EXTERNAL_URL` wired to API and WebSocket hosts.  
Ensure `ALLOWED_ORIGINS` on the gateway includes the frontend Render URL.

---

## How the frontend calls the correct URLs

All browser traffic uses **public env vars** (see [`stockflow/src/lib/env.ts`](stockflow/src/lib/env.ts)):

| Variable | Used for | Local default |
|----------|----------|---------------|
| `NEXT_PUBLIC_API_URL` | REST: auth, market quotes, candles, symbol search | `http://localhost:8080` |
| `NEXT_PUBLIC_WEBSOCKET_URL` | Live ticks + TradingView real-time bars | `ws://localhost:8083` |

- `https://…` WebSocket URLs are automatically converted to `wss://`.
- Auth tokens are sent as `Authorization: Bearer …` on API calls and `?token=` on WebSocket.

**No Finnhub key in the frontend** — market-service holds `FINNHUB_API_KEY` server-side.

---

## Local development (no Docker required for Go)

If Docker is not installed:

```bash
# Infra (optional if you have local Postgres/Redis)
# Or use cloud Postgres/Redis URLs in .env

# Backend — four terminals, from repo root:
export JWT_SECRET=dev-secret
export DATABASE_URL=postgres://stockflow:stockflow@localhost:5432/stockflow?sslmode=disable
export REDIS_URL=redis://localhost:6379
export FINNHUB_API_KEY=your_key

cd backend/auth-service && go run ./cmd
cd backend/market-service && go run ./cmd
cd backend/websocket-service && go run ./cmd
cd backend/api-gateway && go run ./cmd

# Frontend
cd stockflow && pnpm install && pnpm dev
```

With Docker:

```bash
cp .env.example .env   # set FINNHUB_API_KEY, JWT_SECRET
docker compose up --build
cd stockflow && pnpm install && pnpm dev
```

---

## Build commands

| Component | Command |
|-----------|---------|
| Backend services | `cd backend/<service> && go build ./cmd` |
| Frontend | `cd stockflow && pnpm install && pnpm run build` |

Use **pnpm** for the frontend (`packageManager` is set in `package.json`). If `npm install` fails locally, run `corepack enable && pnpm install`.

---

## Render plans & credits

| Resource | Suggested plan | Notes |
|----------|----------------|-------|
| Postgres | Starter or free | Auth users table |
| Redis | Free / Starter | Sessions + pub/sub |
| auth, market | Starter **private** | No public URL |
| api-gateway, websocket | Starter **web** | Public entry points |
| frontend | Starter or Vercel free | Static/SSR Next.js |

Free web services **spin down** after inactivity (~50s cold start). WebSocket feeds may disconnect; clients reconnect automatically.

---

## Troubleshooting

| Issue | Fix |
|-------|-----|
| CORS errors | Set `ALLOWED_ORIGINS` on api-gateway to exact frontend origin (`https://…`, no trailing slash) |
| WebSocket fails | Use `wss://` in production; confirm `stockflow-websocket` is Live |
| Charts empty | Set `FINNHUB_API_KEY` on `stockflow-market`; check market logs |
| 401 on `/me` | Login again; check `JWT_SECRET` matches auth + websocket |
| Gateway `/ready` fails | Ensure auth + market private services are Live |
| Frontend still hits localhost | Rebuild after setting `NEXT_PUBLIC_*` |

---

## Security checklist

- [ ] Strong `JWT_SECRET` (Render generates one)
- [ ] `FINNHUB_API_KEY` only on market-service
- [ ] `ALLOWED_ORIGINS` restricted to your frontend domain(s)
- [ ] Optional: `WS_REQUIRE_AUTH=true` on websocket-service (requires login for live feed)
