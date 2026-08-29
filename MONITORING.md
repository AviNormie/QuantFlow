# StockFlow — Monitoring Setup (Sentry + PostHog)

This guide covers every environment variable and external configuration needed to enable **Sentry** (error tracking) and **PostHog** (product analytics) across:

- **Next.js frontend** — `stockflow/` (`@sentry/nextjs`, org: `quantflow`, project: `quantflow`)
- **Go microservices** — `backend/shared/monitoring/`

If monitoring env vars are **not set**, apps start normally with monitoring disabled.

---

## Quick start

### Next.js (`stockflow/`)

```bash
cd stockflow
cp .env.example .env.local
yarn dev
```

PostHog is initialized in `src/instrumentation-client.ts` (client) and `src/lib/posthog/server.ts` (server).

```bash
npx -y @posthog/wizard@latest   # run in interactive terminal for full wizard setup
```

### Go backend + Docker Compose

```bash
# 1. Copy the env template
cp .env.example .env

# 2. Fill in Sentry + PostHog values (see sections below)

# 3. Start the stack
docker compose up --build
```

For local `go run` (without Docker):

```bash
cd backend/auth-service
set -a && source .env.local && set +a
go run ./cmd
```

---

## Environment variables

### Required for monitoring (at least one)

| Variable | Required | Description |
|----------|----------|-------------|
| `SENTRY_DSN` | No* | Sentry Data Source Name. Enables error tracking when set. |
| `POSTHOG_API_KEY` | No* | PostHog project API key. Enables analytics when set. |

\*At least one should be set to enable monitoring. Both can be used together.

### Sentry variables — Go microservices

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `SENTRY_DSN` | No | *(empty)* | DSN from Sentry → **Settings → Client Keys (DSN)** |
| `SENTRY_ENVIRONMENT` | No | `development` | Environment label: `development`, `staging`, `production` |
| `SENTRY_RELEASE` | No | *(empty)* | Release/version tag, e.g. `stockflow@1.0.0` or git SHA |
| `SENTRY_TRACES_SAMPLE_RATE` | No | `0.2` | Performance trace sampling rate between `0.0` and `1.0` |

### Sentry variables — Next.js (`stockflow/`)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `NEXT_PUBLIC_SENTRY_DSN` | Yes* | *(empty)* | Client-side DSN (browser). Same value as `SENTRY_DSN`. |
| `SENTRY_DSN` | No | *(empty)* | Server-side DSN. Usually same as `NEXT_PUBLIC_SENTRY_DSN`. |
| `SENTRY_ORG` | No | `quantflow` | Sentry org slug (used for source map uploads) |
| `SENTRY_PROJECT` | No | `quantflow` | Sentry project slug |
| `SENTRY_AUTH_TOKEN` | No | *(empty)* | Auth token for source map uploads (build/CI only — **secret**) |
| `SENTRY_ENVIRONMENT` | No | `development` | Environment label |
| `SENTRY_TRACES_SAMPLE_RATE` | No | `1.0` dev / `0.1` prod | Trace sampling (set in `.env.local`) |

\*Required to enable Sentry in the browser.

**Get your DSN:** [sentry.io](https://sentry.io) → org **quantflow** → project **quantflow** → **Settings → Client Keys (DSN)**

**Source maps (optional):** Sentry → **Settings → Auth Tokens** → create token with `project:releases` → set `SENTRY_AUTH_TOKEN`

**Tunnel route:** Browser events route through `/monitoring` to avoid ad blockers (configured in `next.config.ts`).

### PostHog variables — Go microservices

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `POSTHOG_API_KEY` | No | *(empty)* | **Project API Key** (`phc_...`) from PostHog → **Project Settings** |
| `POSTHOG_HOST` | No | `https://us.i.posthog.com` | Ingest URL for your PostHog region |

### PostHog variables — Next.js (`stockflow/`)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `NEXT_PUBLIC_POSTHOG_PROJECT_TOKEN` | Yes* | *(empty)* | Same **Project API Key** (`phc_...`) — exposed to browser |
| `NEXT_PUBLIC_POSTHOG_HOST` | No | `https://us.i.posthog.com` | PostHog ingest host |
| `POSTHOG_API_KEY` | No | *(empty)* | Server-side events via `posthog-node` (defaults to project token) |
| `POSTHOG_HOST` | No | `https://us.i.posthog.com` | Server-side ingest host |

**PostHog host by region**

| Region | `POSTHOG_HOST` |
|--------|----------------|
| US (default) | `https://us.i.posthog.com` |
| EU | `https://eu.i.posthog.com` |

### Service identity (already used by all services)

| Variable | Required | Description |
|----------|----------|-------------|
| `SERVICE_NAME` | Yes | Identifies the service in logs and monitoring (`api-gateway`, `auth-service`, etc.) |

---

## Example `.env` (repo root — Docker Compose)

Create this file at the project root:

```bash
cp .env.example .env
```

```env
# Sentry
SENTRY_DSN=https://YOUR_KEY@oYOUR_ORG.ingest.sentry.io/YOUR_PROJECT
SENTRY_ENVIRONMENT=development
SENTRY_RELEASE=stockflow@local
SENTRY_TRACES_SAMPLE_RATE=0.2

# PostHog
POSTHOG_API_KEY=phc_your_project_api_key_here
POSTHOG_HOST=https://us.i.posthog.com
```

Docker Compose reads the root `.env` and passes these values to every Go service via `docker-compose.yml`.

---

## Example `.env.local` (per service — local `go run`)

For each service under `backend/<service>/.env.local`:

```env
PORT=8084
SERVICE_NAME=auth-service

SENTRY_DSN=https://YOUR_KEY@oYOUR_ORG.ingest.sentry.io/YOUR_PROJECT
SENTRY_ENVIRONMENT=development
SENTRY_RELEASE=stockflow@local
SENTRY_TRACES_SAMPLE_RATE=0.2

POSTHOG_API_KEY=phc_your_project_api_key_here
POSTHOG_HOST=https://us.i.posthog.com
```

Load before running:

```bash
set -a && source .env.local && set +a
go run ./cmd
```

---

## External setup

### Sentry (one-time)

1. Go to [https://sentry.io](https://sentry.io) and create an account.
2. Create a new **Project** → choose **Go** as the platform.
3. Open **Settings → Client Keys (DSN)** and copy the DSN.
4. Paste it into `SENTRY_DSN` in your `.env` or `.env.local`.

**Recommended settings**

| Setting | Recommendation |
|---------|----------------|
| Environment | Use `development` locally, `production` in prod |
| Release | Set `SENTRY_RELEASE` to your git tag or commit SHA in CI |
| Traces | Start with `0.2` (20% sampling); increase in prod if needed |

**Optional:** Use one Sentry project for all services (differentiated by `SERVICE_NAME`) or create a separate Sentry project per microservice with different DSNs.

### PostHog (one-time)

1. Go to [https://posthog.com](https://posthog.com) and create an account.
2. Create a **Project**.
3. Open **Project Settings** and copy the **Project API Key** (`phc_...`).
   - Use the **Project API Key**, not a personal API key.
4. Set `POSTHOG_API_KEY` in your `.env` or `.env.local`.
5. Set `POSTHOG_HOST` to match your cloud region (US or EU).

---

## What is captured automatically

| Tool | Event | When |
|------|-------|------|
| **Sentry** | Errors & panics | Unhandled exceptions via Gin middleware |
| **Sentry** | Performance traces | Sampled HTTP requests (`SENTRY_TRACES_SAMPLE_RATE`) |
| **PostHog** | `api_request` | Every HTTP request except `/health` (method, path, status, duration) |

---

## Manual capture in code

Inject or pass the monitoring client from `main` into handlers/services:

```go
// Report an error
mon.CaptureError(err, map[string]string{
    "handler": "Register",
    "email":   email,
})

// Track a custom event
mon.CaptureEvent(userID, "user_registered", map[string]any{
    "plan": "free",
})
```

---

## Verify monitoring is working

### 1. Start the stack

```bash
docker compose up --build
```

### 2. Trigger a request (PostHog)

```bash
curl http://localhost:8080/health   # ignored by PostHog
curl http://localhost:8084/health   # auth-service health
```

Check **PostHog → Activity** for `api_request` events tagged with `service`.

### 3. Trigger a test error (Sentry)

Temporarily add a test route or cause a panic in dev, then check **Sentry → Issues**.

---

## Docker Compose reference

These variables are defined in `docker-compose.yml` under `x-backend-env` and shared by all Go services:

```yaml
SENTRY_DSN: ${SENTRY_DSN:-}
SENTRY_ENVIRONMENT: ${SENTRY_ENVIRONMENT:-development}
SENTRY_RELEASE: ${SENTRY_RELEASE:-}
SENTRY_TRACES_SAMPLE_RATE: ${SENTRY_TRACES_SAMPLE_RATE:-0.2}
POSTHOG_API_KEY: ${POSTHOG_API_KEY:-}
POSTHOG_HOST: ${POSTHOG_HOST:-https://us.i.posthog.com}
```

Each service also receives `SERVICE_NAME` individually.

---

## Troubleshooting

| Problem | Fix |
|---------|-----|
| No events in PostHog | Confirm `POSTHOG_API_KEY` is the **project** key; check `POSTHOG_HOST` region |
| No errors in Sentry | Confirm `SENTRY_DSN` is correct; trigger a non-health request that errors |
| Monitoring works in Docker but not locally | Export vars: `set -a && source .env.local && set +a` before `go run` |
| Services fail to start after adding keys | Check for typos; invalid `SENTRY_TRACES_SAMPLE_RATE` must be `0.0`–`1.0` |
| Too many PostHog events | `/health` is already excluded; adjust middleware in `shared/monitoring/` if needed |

---

## Prometheus + Grafana (local / Docker Compose)

StockFlow exposes Prometheus metrics on every Go service at **`GET /metrics`**.

### Start the monitoring stack

```bash
docker compose up -d prometheus grafana
# Or full stack: docker compose up --build
```

| URL | Purpose |
|-----|---------|
| http://localhost:9090 | Prometheus UI |
| http://localhost:3001 | Grafana UI (default login below) |

### Environment variables (root `.env`)

| Variable | Default | Description |
|----------|---------|-------------|
| `PROMETHEUS_PORT` | `9090` | Host port for Prometheus |
| `GRAFANA_PORT` | `3001` | Host port for Grafana (avoids clash with Next.js on 3000) |
| `GRAFANA_ADMIN_USER` | `admin` | Grafana admin username |
| `GRAFANA_ADMIN_PASSWORD` | `admin` | Grafana admin password — **change in production** |
| `GRAFANA_ROOT_URL` | `http://localhost:3001` | Grafana public URL |

No extra env vars are required on Go services — metrics are always enabled when `/metrics` is scraped.

### Metrics exposed

| Metric | Labels | Description |
|--------|--------|-------------|
| `http_requests_total` | `service`, `method`, `path`, `status` | Request counter |
| `http_request_duration_seconds` | `service`, `method`, `path` | Request latency histogram |

`/health` and `/metrics` are excluded from latency counters to avoid scrape noise.

### Provisioned Grafana dashboard

On first start, Grafana loads:

- Datasource: Prometheus at `http://prometheus:9090`
- Dashboard: **StockFlow Services** — request rate, p95 latency, status breakdown

Config lives in `deploy/grafana/provisioning/` and `deploy/prometheus/prometheus.yml`.

### Render / production note

Prometheus and Grafana are included in **docker-compose for local dev**. Render does not run this stack by default. For production metrics you would either:

- Run Grafana Cloud / hosted Prometheus and scrape public `/metrics` (not recommended without auth), or
- Keep monitoring local while production uses Sentry + PostHog.

---

## File reference

| File | Purpose |
|------|---------|
| `.env.example` | Template with all env vars |
| `.env` | Your local secrets (git-ignored) — used by Docker Compose |
| `backend/<service>/.env.local` | Per-service local overrides (git-ignored) |
| `backend/shared/monitoring/monitoring.go` | Shared Sentry + PostHog initialization |
| `docker-compose.yml` | Passes monitoring env to all Go containers |
| `backend/README.md` | General backend run guide |
