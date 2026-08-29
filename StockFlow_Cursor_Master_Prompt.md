# StockFlow — Cursor Master Prompt

## Production-Grade Distributed Real-Time Market Infrastructure Platform

You are working inside an existing repository called **StockFlow**.

The goal is to build StockFlow as a **backend-focused, production-grade microservices system** for real-time market data. The frontend is secondary and may be implemented heavily with AI assistance.

The developer wants to understand the backend and architecture, not merely generate code.

---

# 1. NON-NEGOTIABLE WORKING RULE

**Inspect before changing anything.**

Before implementing a feature:

1. Inspect the repository tree.
2. Inspect existing Go services and their code.
3. Inspect Docker Compose, Dockerfiles, and environment configuration.
4. Inspect existing PostgreSQL and Redis setup.
5. Inspect existing routes, middleware, models, repositories, and configuration.
6. Determine what is already:
   - DONE
   - PARTIALLY DONE
   - MISSING
   - INCORRECT / NEEDS REFACTOR
7. Do not recreate completed work.
8. Do not duplicate functionality.
9. Preserve working code unless it conflicts with the architecture below.
10. Implement only the next required milestone.

If something is already correctly implemented, reuse it and continue from there.

When a significant architectural change is required, explain the reason before making it.

---

# 2. PROJECT OBJECTIVE

Build a distributed real-time market infrastructure platform capable of:

- ingesting real-time market data from an external provider
- normalizing market data
- caching current prices
- serving historical market data
- streaming market updates through WebSockets
- powering TradingView Charting Library
- authenticating users
- managing sessions
- supporting watchlists and paper-trading functionality as the project grows
- eventually processing market events asynchronously
- eventually providing production observability and cloud deployment

The project is intentionally **microservices-first**.

Do NOT build a monolith and split it later.

---

# 3. FINAL TECHNOLOGY DIRECTION

## Backend

- Go
- Gin
- GORM
- PostgreSQL
- Redis
- WebSockets

## Frontend

- Next.js
- TypeScript
- TradingView Charting Library

## Infrastructure

- Docker
- Docker Compose

## Future infrastructure

Kafka, Kubernetes, Prometheus, Grafana, Loki, OpenTelemetry, CI/CD, and advanced resilience/scaling are intentionally deferred. Do not implement them during the current milestone unless explicitly requested.

---

# 4. DATABASE DECISION — FINAL

Use **PostgreSQL** as the relational database.

Supabase PostgreSQL can be used as the hosted PostgreSQL provider.

For Go services, use **GORM** as the ORM.

### Why GORM

- mature Go ORM
- reduces repetitive CRUD/database boilerplate
- supports PostgreSQL well
- supports models, relationships, transactions, migrations, and connection pooling
- lets the project focus on microservices, realtime systems, Redis, and distributed backend concepts
- supports raw SQL when an explicit query is justified

Use GORM behind a repository layer.

Do not introduce Prisma or a separate Node backend just to access the database.

Do not add another ORM.

---

# 5. MICROSERVICE ARCHITECTURE

Initial services:

```text
api-gateway
      |
      +---- auth-service
      |
      +---- market-service
      |
      +---- websocket-service
```

These are independent Go applications/processes and should be independently containerizable.

Do not create unnecessary tiny services.

The initial four services are the correct boundary for the first milestone.

---

# 6. API GATEWAY

The API Gateway is the public entry point.

Responsibilities:

- route requests to internal services
- authentication verification where appropriate
- request ID generation/propagation
- rate limiting structure
- centralized request logging
- timeout handling
- consistent error handling
- public health/readiness endpoints

The gateway must NOT contain business logic belonging to another service.

---

# 7. AUTH SERVICE

The Auth Service owns authentication.

Responsibilities:

- signup
- login
- access-token generation
- refresh-token flow
- logout
- session management
- password hashing
- protected `/me` endpoint

Database:

- PostgreSQL through GORM

Redis:

- refresh/session state
- session invalidation
- TTL-based session data

Use:

- secure password hashing
- short-lived JWT access tokens
- longer-lived refresh tokens
- refresh-token/session rotation where appropriate
- JWT validation middleware

Do not use Supabase Auth or Better Auth for the core authentication system.

The purpose of this project is to learn backend authentication architecture.

Initial endpoints:

```text
POST /api/auth/signup
POST /api/auth/login
POST /api/auth/refresh
POST /api/auth/logout
GET  /api/auth/me
```

Never store plaintext passwords.
Never log passwords, tokens, or secrets.

---

# 8. MARKET DATA SERVICE

This is the core service of StockFlow.

Responsibilities:

- connect to an external market-data provider
- consume real-time market data
- normalize provider-specific payloads
- validate incoming data
- maintain latest prices
- publish normalized updates internally
- provide symbol information
- provide historical market-data APIs
- aggregate data into OHLC candles where necessary
- reconnect safely if the external provider disconnects

Initial provider options:

- Finnhub
- Twelve Data

Keep the provider behind an internal adapter/interface so the rest of StockFlow does not depend on the provider's exact payload format.

The frontend must NEVER receive provider credentials or connect directly to the provider.

Flow:

```text
External Market Provider
          |
          v
   Market Data Service
          |
          +---- Redis latest-price cache
          |
          +---- internal realtime stream
```

---

# 9. WEBSOCKET SERVICE

The WebSocket Service owns client WebSocket connections.

Responsibilities:

- accept client connections
- authenticate connections
- manage subscriptions
- subscribe/unsubscribe symbols
- broadcast market updates
- handle disconnects
- clean up connection state
- handle reconnect-related behavior

Initial realtime architecture:

```text
Market Service
      |
      v
Redis Pub/Sub
      |
      v
WebSocket Service
      |
      v
Frontend Clients
```

Redis Pub/Sub is for realtime fan-out here, not permanent storage.

---

# 10. REDIS

Use Redis for:

- authentication sessions
- refresh-token/session state
- latest market prices
- realtime Pub/Sub
- rate limiting infrastructure

Do not treat Redis as the source of truth for permanent financial data.

Use TTLs wherever temporary data has a natural expiration.

---

# 11. GORM DATABASE ARCHITECTURE

Use a clean repository structure:

```text
handler
   |
service
   |
repository
   |
GORM
   |
PostgreSQL
```

Handlers should handle HTTP concerns.

Services should contain business logic.

Repositories should contain database access.

GORM models should not become the entire business layer.

Use transactions for operations that must be atomic.

Use database constraints and indexes intentionally.

Use migrations rather than manually changing production schemas.

---

# 12. GO SERVICE STRUCTURE

Each service should follow a structure similar to:

```text
service-name/
├── cmd/
│   └── main.go
├── internal/
│   ├── handler/
│   ├── service/
│   ├── repository/
│   ├── model/
│   ├── middleware/
│   ├── config/
│   └── ...
├── migrations/
├── Dockerfile
├── go.mod
└── README.md
```

Do not create abstractions only to make the folder tree look sophisticated.

---

# 13. CONFIGURATION

Use environment variables.

Never commit secrets.

Existing variables may include:

```text
POSTGRES_USER
POSTGRES_PASSWORD
POSTGRES_DB
POSTGRES_PORT

REDIS_PORT

API_GATEWAY_PORT
AUTH_SERVICE_PORT
MARKET_SERVICE_PORT
WEBSOCKET_SERVICE_PORT

DATABASE_URL
REDIS_URL
```

Inside Docker, services must use Docker service names rather than `localhost`.

For example:

```text
postgres:5432
redis:6379
```

Host development can use:

```text
localhost:5432
localhost:6379
```

Keep service configuration explicit.

---

# 14. HEALTH AND READINESS

Every service should expose:

```text
GET /health
GET /ready
```

`/health` should indicate that the process is alive.

`/ready` should indicate that required dependencies are available.

Do not perform expensive operations in health checks.

---

# 15. DOCKER

Every Go service should have its own Dockerfile.

Docker Compose should run the complete local system.

Initial local infrastructure:

```text
postgres
redis
api-gateway
auth-service
market-service
websocket-service
```

Use:

- multi-stage builds
- environment variables
- Docker networking
- health checks
- sensible image sizes
- non-root execution where practical

Understand that containers communicate using service names, not `localhost`.

---

# 16. WEB SOCKET DATA FLOW

The intended initial realtime flow is:

```text
Finnhub/Twelve Data
        |
        v
Market Service
        |
        v
Normalize tick
        |
        v
Redis latest price
        |
        v
Redis Pub/Sub
        |
        v
WebSocket Service
        |
        v
Next.js / TradingView
```

The WebSocket Service should not directly depend on the external provider.

The Market Service should be responsible for provider ingestion.

---

# 17. TRADINGVIEW INTEGRATION

The TradingView Charting Library is already available.

The frontend datafeed should communicate with StockFlow's backend, not directly with the external market provider.

Backend capabilities needed:

```text
symbol search
symbol resolution
historical bars
real-time bar updates
```

Conceptually:

```text
TradingView
     |
     v
Next.js Datafeed
     |
     v
API Gateway
     |
     v
Market Service
```

Realtime updates:

```text
Market Service
     |
     v
Redis Pub/Sub
     |
     v
WebSocket Service
     |
     v
Next.js
     |
     v
TradingView
```

---

# 18. LOGGING AND ERROR HANDLING

Use structured logging.

Logs should eventually include:

- timestamp
- level
- service name
- request ID
- route
- status
- latency
- useful error context

Do not log secrets or tokens.

Use explicit error handling.

Do not panic for normal application errors.

Use context cancellation correctly.

---

# 19. CONCURRENCY AND LIFECYCLE

Use Go concurrency intentionally.

Important rules:

- use `context.Context`
- avoid goroutine leaks
- cancel long-running goroutines
- cleanly close connections
- handle external provider reconnects
- do not create unlimited goroutines per request
- support graceful service shutdown

The market-data connection is a long-running process and must have controlled lifecycle management.

---

# 20. TESTING

Add tests around important backend behavior.

Initial critical tests:

## Auth

- signup
- duplicate email
- invalid password
- login
- refresh
- logout
- protected endpoint

## Market

- provider payload parsing
- normalization
- invalid payload handling
- Redis latest-price updates

## WebSocket

- connect
- authenticate
- subscribe
- unsubscribe
- broadcast
- disconnect cleanup

Tests should verify behavior rather than implementation details.

---

# 21. FIRST IMPLEMENTATION MILESTONES

Implement in this order, unless the repository audit shows that a milestone is already complete.

## Milestone 1 — Foundation

- four Go services
- Gin servers
- health/readiness endpoints
- Dockerfiles
- Docker Compose
- PostgreSQL
- Redis
- environment configuration

## Milestone 2 — Authentication

- GORM setup
- database migrations
- users table
- password hashing
- signup
- login
- JWT access token
- refresh token
- Redis sessions
- logout
- `/me`
- gateway routing

## Milestone 3 — Market Data

- provider adapter
- external WebSocket connection
- symbol subscription
- normalized market model
- latest-price Redis cache
- reconnect handling
- historical data endpoint

## Milestone 4 — WebSocket Service

- client connection management
- authentication
- symbol subscriptions
- Redis Pub/Sub
- market update broadcasting
- unsubscribe
- disconnect cleanup

## Milestone 5 — TradingView

- symbol search
- symbol resolution
- historical bars
- real-time bars
- TradingView datafeed integration

---

# 22. FUTURE WORK — DO NOT IMPLEMENT YET

After the initial microservices + authentication + market-data + WebSocket + TradingView system is stable, the project will later expand with:

- Kafka and durable event-driven processing
- portfolio service
- analytics service
- notification service
- Prometheus/Grafana/Loki/OpenTelemetry
- Kubernetes
- CI/CD
- advanced retries, circuit breakers, DLQs, idempotency, load testing, profiling, and horizontal scaling

Do not implement these now unless explicitly requested.

---

# 23. CURRENT SPRINT RULE

The user is building the project incrementally.

If Day 1/foundation is already complete, do NOT touch it unnecessarily.

For example, if these already exist:

```text
api-gateway
auth-service
market-service
websocket-service
docker-compose.yml
.env.example
PostgreSQL
Redis
health endpoints
```

then move directly to the next missing milestone.

At the beginning of each task, provide:

```text
PROJECT AUDIT

DONE
- ...

PARTIAL
- ...

MISSING
- ...

NEXT
- ...

FILES TO CHANGE
- ...
```

Then implement the next required piece.

---

# 24. DEVELOPMENT STYLE

The user is learning backend engineering while building this project.

For significant components:

1. Explain what is being built.
2. Explain why it exists.
3. Explain the important backend concept.
4. Implement it.
5. List files changed.
6. Explain how to run it.
7. Give verification commands.
8. Mention important failure cases.

Do not dump huge amounts of code without explanation.

For small obvious changes, implement directly.

---

# 25. CODE QUALITY RULES

Prefer:

- small focused functions
- explicit dependencies
- dependency injection
- context-aware operations
- typed domain models
- clear package boundaries
- repository/service/handler separation
- structured logging
- consistent errors
- tests for important behavior

Avoid:

- global mutable state
- giant `main.go`
- giant service objects
- unnecessary interfaces
- unnecessary abstraction layers
- duplicated configuration
- magic constants
- secrets in source code
- swallowed errors
- `panic()` for normal application errors

---

# 26. DEFINITION OF DONE

A feature is not done just because it compiles.

Verify that it:

- builds
- starts correctly
- has correct configuration
- works through the intended service boundary
- handles expected errors
- has useful logs
- has tests where appropriate
- works inside Docker
- does not break existing services

---

# 27. ARCHITECTURAL PRINCIPLE

Every technology must have a reason.

Do not add a tool simply because it looks good on a resume.

The project should demonstrate that the developer understands:

- microservices
- Go backend engineering
- Gin
- GORM
- PostgreSQL
- Redis
- WebSockets
- realtime systems
- authentication
- concurrency
- service boundaries
- Docker
- distributed-system fundamentals

The final system should be something the developer can explain and defend in a backend interview.
