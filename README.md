# StockFlow

Full-stack stock charting app with JWT auth, API gateway, and real-time WebSocket prices.

## Quick start

```bash
# 1. Copy env and start infra + backend
cp .env.example .env
docker compose up -d postgres redis
docker compose exec -T postgres psql -U stockflow -d stockflow < backend/auth-service/configs/migrations/001_create_users.sql

# 2. Start backend services (or use docker compose up --build)
# 3. Start frontend — see stockflow/README.md
```

## Docs

- [Frontend (stockflow/)](stockflow/README.md) — signup, login, charts, env vars
- [Backend services](backend/README.md) — Go microservices, Docker, migrations
- [API docs](http://localhost:3000/api/docs) — OpenAPI (when frontend is running)
