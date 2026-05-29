# StockFlow Backend

Go microservices for StockFlow. Each service is an independent Gin HTTP server with its own module, port, and Docker image.

## Services

| Service             | Default port | Health check                          |
|---------------------|-------------:|---------------------------------------|
| `api-gateway`       | 8080         | http://localhost:8080/health          |
| `auth-service`      | 8084         | http://localhost:8084/health          |
| `market-service`    | 8082         | http://localhost:8082/health          |
| `websocket-service` | 8083         | http://localhost:8083/health          |

Example health response:

```json
{"status":"ok","service":"auth-service"}
```

## Prerequisites

- **Go 1.25+** — required by Gin v1.12 (`go version` to check)
- **Docker** (optional) — for container builds and runs

## Project layout

Each runnable service follows the same structure:

```
<service>/
├── cmd/
│   └── main.go           # entrypoint
├── internal/
│   ├── handler/
│   ├── service/
│   ├── repository/
│   └── middleware/
├── configs/
├── .env.local            # local env (not loaded automatically; see below)
├── Dockerfile
├── go.mod
└── go.sum
```

`shared/` is a separate Go module for shared libraries (no HTTP server).

## Environment variables

| Variable       | Description                          |
|----------------|--------------------------------------|
| `PORT`         | HTTP listen port                     |
| `SERVICE_NAME` | Service identifier (in `.env.local`) |

Per-service defaults are in each `.env.local`:

```bash
# backend/auth-service/.env.local
PORT=8084
SERVICE_NAME=auth-service
```

`main.go` reads `PORT` from the process environment. To use `.env.local` when running locally:

```bash
cd backend/auth-service
set -a && source .env.local && set +a
go run ./cmd
```

Or export inline:

```bash
PORT=8084 go run ./cmd
```

## Run locally (Go)

From the repo root, in **separate terminals** (one per service):

```bash
cd backend/api-gateway      && go run ./cmd
cd backend/auth-service     && go run ./cmd
cd backend/market-service   && go run ./cmd
cd backend/websocket-service && go run ./cmd
```

Verify all services:

```bash
curl http://localhost:8080/health
curl http://localhost:8084/health
curl http://localhost:8082/health
curl http://localhost:8083/health
```

Override port for a single run:

```bash
PORT=9000 go run ./cmd
```

## Build binaries

```bash
cd backend/<service>
go build -o bin/server ./cmd
./bin/server
```

Build all services from `backend/`:

```bash
for svc in api-gateway auth-service market-service websocket-service; do
  (cd "$svc" && go build -o bin/server ./cmd && echo "built $svc")
done
```

## Docker

Build and run a single service (example: `auth-service`):

```bash
cd backend/auth-service
docker build -t stockflow-auth-service .
docker run --rm -p 8084:8084 -e PORT=8084 stockflow-auth-service
```

| Service             | Build tag                    | Run (host port)        |
|---------------------|------------------------------|------------------------|
| api-gateway         | `stockflow-api-gateway`      | `-p 8080:8080`         |
| auth-service        | `stockflow-auth-service`     | `-p 8084:8084`         |
| market-service      | `stockflow-market-service`   | `-p 8082:8082`         |
| websocket-service   | `stockflow-websocket-service`| `-p 8083:8083`         |

Build all images:

```bash
cd backend
for svc in api-gateway auth-service market-service websocket-service; do
  docker build -t "stockflow-$svc" "./$svc"
done
```

## Module & dependencies

Initialize or add deps inside a service directory (not from `backend/` root):

```bash
cd backend/auth-service
go mod init auth-service          # already done
go get github.com/gin-gonic/gin
go mod tidy
```

## Useful commands

```bash
# Format
go fmt ./...

# Vet
go vet ./...

# Test (when tests exist)
go test ./...

# Tidy module
go mod tidy
```

## Docker Compose (full stack)

Everything runs from the **repo root**: `../docker-compose.yml` (from `backend/`).

```bash
# From StockFlow repo root
cp .env.example .env          # optional — defaults work for local dev
docker compose up --build
# or: docker-compose up --build
```

Expected containers: **postgres**, **redis**, **api-gateway**, **auth-service**, **market-service**, **websocket-service**.

```bash
docker compose ps
docker compose logs -f
docker compose down           # stop containers
docker compose down -v        # stop and delete volumes (wipes DB data)
```

| Service             | Host URL                         | Inside Docker network   |
|---------------------|----------------------------------|-------------------------|
| PostgreSQL          | `localhost:5432`                 | `postgres:5432`         |
| Redis               | `localhost:6379`                 | `redis:6379`            |
| api-gateway         | http://localhost:8080/health     | `api-gateway:8080`      |
| auth-service        | http://localhost:8084/health     | `auth-service:8084`     |
| market-service      | http://localhost:8082/health     | `market-service:8082`   |
| websocket-service   | http://localhost:8083/health     | `websocket-service:8083`|

```bash
curl http://localhost:8080/health
curl http://localhost:8084/health
curl http://localhost:8082/health
curl http://localhost:8083/health
```

**Docker networking:** containers on `stockflow-network` resolve each other by **service name**, not `localhost`. Backend containers receive:

```text
postgres://stockflow:stockflow@postgres:5432/stockflow
redis://redis:6379
```

From your host (Go services via `go run`, Next.js, etc.), use `localhost` and the published ports.

**PostgreSQL**

- **Ports:** `5432:5432` (override with `POSTGRES_PORT` in `.env`)
- **Volumes:** `stockflow-postgres-data` persists data under `/var/lib/postgresql/data`
- **Credentials:** `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` (defaults: `stockflow` / `stockflow` / `stockflow`)

**Redis**

- **Ports:** `6379:6379`
- **Networking:** same `stockflow-network` as Postgres
- **Volume:** `stockflow-redis-data` (AOF enabled for persistence)

Verify:

```bash
docker compose exec postgres psql -U stockflow -d stockflow -c "SELECT 1"
docker compose exec redis redis-cli ping
```

## Running everything at once

**Docker (recommended):** from repo root:

```bash
docker compose up --build
```

**Local Go only:** four terminals with `go run ./cmd`, plus `docker compose up -d postgres redis` for infra.

## Troubleshooting

| Issue | Fix |
|-------|-----|
| Port already in use | Change `PORT` or stop the process on that port (`lsof -i :8080`) |
| `go: requires go >= 1.25` | Upgrade Go: https://go.dev/dl/ |
| Health check fails | Confirm the service is running and you are using the correct port |
| `.env.local` ignored | Export vars or `source .env.local` before `go run` (see above) |
