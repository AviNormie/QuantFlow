# StockFlow — Architecture Deep Dive

> All answers below are grounded in the actual source code. File paths are relative to `/backend/`.

---

## 1. What happens from the moment a price tick arrives to when a subscribed client sees it update?

There are **7 distinct hops** in the pipeline.

### Hop 1 — Finnhub WebSocket → `trades` channel
`market-service` connects to `wss://ws.finnhub.io?token=<API_KEY>` inside `IngestionService.Run()` (`market-service/internal/service/ingestion.go`). Once connected, it calls `provider.Subscribe(cfg.DefaultSymbols)` which sends JSON frames like `{"type":"subscribe","symbol":"AAPL"}` over the Finnhub WebSocket.

A dedicated goroutine `Provider.readLoop(conn)` (`market-service/internal/provider/finnhub/finnhub.go`) sits in a tight loop calling `conn.ReadMessage()`. Each incoming Finnhub frame can contain a batch:
```json
{"type":"trade","data":[{"s":"AAPL","p":182.50,"v":100,"t":1713000000}]}
```
Every trade in the `data` array is pushed individually into a buffered `trades chan provider.RawTrade` (capacity **256**).

### Hop 2 — `trades` channel → Normalization
`IngestionService.consume()` selects from the `trades` channel. Each `RawTrade{Symbol, Price, Volume, Timestamp}` passes through `Normalizer.Normalize()`, which validates non-empty symbol, price > 0, and timestamp > 0, then produces a `model.NormalizedTick{Symbol, Price, Volume, Timestamp, Source: "finnhub"}`.

### Hop 3 — Redis cache write
`PriceCache.SetLatest()` (`market-service/internal/repository/redis_repository.go`) serializes the tick to JSON and writes it to Redis under the key `market:price:<SYMBOL>` with a **24-hour TTL**. This is the snapshot used by REST quote requests.

### Hop 4 — Redis Pub/Sub publish
`Publisher.Publish()` calls `redis.Publish(ctx, "market:updates", jsonPayload)` on the same normalized tick. The channel name `market:updates` is configurable via `MARKET_PUBSUB_CHANNEL`.

### Hop 5 — `websocket-service` receives from Redis
`RedisSubscriber.RunWithReconnect()` (`websocket-service/internal/subscriber/redis_subscriber.go`) is running in its own goroutine. It has called `redis.Subscribe(ctx, "market:updates")` and is ranging over `pubsub.Channel()`. On receiving a message it:
1. Unmarshals `{symbol, price, volume, timestamp}`.
2. Re-marshals into the client wire format: `{"type":"trade","data":[{"s":"AAPL","p":182.50,"v":100,"t":...}]}`.
3. Calls `hub.Broadcast(tick.Symbol, payload)`.

### Hop 6 — Hub iterates clients → `send` channel
`Hub.Broadcast()` (`websocket-service/internal/hub/hub.go`) acquires a **read lock**, iterates every registered `*Client`, checks `client.IsSubscribed(symbol)`, and calls `client.TrySend(payload)` for matching clients. `TrySend` does a **non-blocking send** into the per-client `send chan []byte` (capacity **64**).

### Hop 7 — `WritePump` → browser
A goroutine running `client.WritePump()` is ranging over the `send` channel and calling `conn.WriteMessage(websocket.TextMessage, payload)` for each message. The frame lands in the browser.

### Frontend render
`useStockWebSocket` (`stockflow/src/hooks/use-stock-websocket.ts`) receives the frame in `ws.onmessage`, parses it, takes `payload.data[payload.data.length - 1]`, filters by `trade.s === symbolRef.current`, and calls `setLastTrade(...)`, triggering a React re-render.

**End-to-end latency is dominated by:** Finnhub's own feed delay + Redis round-trip + WebSocket write. In a healthy local setup this is typically sub-100ms. There is no batching or debouncing at any layer.

---

## 2. Why Redis Pub/Sub instead of in-memory broadcast or a message queue like Kafka/NATS?

### Why not in-memory broadcast?

If `market-service` and `websocket-service` were the same process, a simple Go channel or `sync.Map` of listeners would work fine. They're deliberately separate containers — `market-service` owns the Finnhub connection and Redis cache; `websocket-service` owns client state. Coupling them would mean a single binary can't scale each concern independently and would break the isolation of concerns in the Docker network.

### Why Redis over Kafka/NATS?

| Concern | Redis Pub/Sub | Kafka | NATS |
|---|---|---|---|
| **Already in the stack** | ✅ Redis is already running for session storage and price cache | ❌ New dependency | ❌ New dependency |
| **Setup complexity** | Zero — one `redis.Subscribe()` call | Significant — brokers, partitions, consumer groups, offsets | Moderate |
| **Message persistence** | ❌ Fire-and-forget; missed messages are gone | ✅ Retained by offset | Optional |
| **Fan-out to N consumers** | ✅ Native — every subscriber gets every message | Requires consumer groups per service type | ✅ |
| **Throughput for market data** | Sufficient — thousands of ticks/sec well within Redis limits | Overkill for this scale | Good |
| **Operational cost** | 0 extra ops — same Redis instance | High | Moderate |

**The real answer:** At this scale (a few symbols, hundreds to low thousands of concurrent users), Redis pub/sub is completely appropriate. The deciding factor is that Redis was already required for session management and price caching, so adding pub/sub costs nothing operationally. Kafka would be the right choice if you needed guaranteed delivery, replay of historical ticks, or fan-out to many heterogeneous consumers (analytics pipeline, alerting service, etc.).

**The known tradeoff:** Redis pub/sub is **at-most-once**. If `websocket-service` is restarting or the Redis connection drops for even a second, those ticks are gone — no replay. For a live trading price feed this is acceptable (stale price is worse than a momentary gap). For order execution or audit logs it would not be.

---

## 3. How do you handle a client that disconnects abruptly mid-stream — what cleans up its subscription?

The entire cleanup is driven by the **read loop returning** in `ws_handler.go`:

```go
// ws_handler.go — Handle()
client := hub.NewClient(conn, h.hub)
h.hub.Register(client)

go client.WritePump()    // goroutine: drains send channel

defer func() {
    h.hub.Unregister(client)  // step 1: remove from hub map
    client.Close()            // step 2: close send channel + conn
}()

for {
    _, msg, err := conn.ReadMessage()
    if err != nil {
        return  // ← this is the trigger
    }
    // ... handle subscribe/unsubscribe
}
```

**What fires the `return`:** Any network error — TCP RST, abrupt disconnect, browser tab close, `close()` frame — causes `conn.ReadMessage()` to return a non-nil error. There's no separate heartbeat or ping/pong timeout; disconnection detection is purely reactive.

**Step 1 — `hub.Unregister(client)`:** Acquires a write lock on `hub.mu` and deletes the client from `hub.clients`. From this point, `Hub.Broadcast()` will never call `TrySend` on this client again.

**Step 2 — `client.Close()`:**
```go
func (c *Client) Close() {
    c.mu.Lock()
    if c.closed { c.mu.Unlock(); return }
    c.closed = true
    c.mu.Unlock()

    close(c.send)    // signals WritePump to exit its range loop
    c.conn.Close()   // closes the underlying TCP connection
}
```
Closing the `send` channel causes `WritePump`'s `for payload := range c.send` loop to exit, and that goroutine returns. The `closed` flag + idempotency guard prevent double-close panics.

**Memory cleanup:** After `WritePump` returns, there are no more references to the `*Client` struct (it's been removed from the hub map and the handler goroutine is done). The GC reclaims the client, its `symbols` map, the `send` channel backing array, and the gorilla WebSocket buffers.

**Symbol subscriptions:** The `client.symbols map[string]bool` is part of the `Client` struct itself. When the struct is GC'd, so are the subscriptions. There's no external subscription registry to clean up.

---

## 4. What happens if one WebSocket client is much slower to consume than the others — do you drop messages, buffer, or disconnect it?

**The answer is: silent drop. No disconnect.**

```go
// hub/hub.go — TrySend
func (c *Client) TrySend(payload []byte) {
    c.mu.RLock()
    if c.closed { c.mu.RUnlock(); return }
    c.mu.RUnlock()

    select {
    case c.send <- payload:   // succeeds if buffer has room
    default:                  // buffer full → message silently dropped
    }
}
```

Each client has a `send chan []byte` with capacity **64**. `TrySend` is called from `Hub.Broadcast()` which holds a read lock, so it **cannot block** — if the channel is full it must return immediately. The `default` branch discards the tick.

**Consequences for a slow client:**
- It stays connected — no forced disconnect.
- It misses ticks without any notification.
- The buffer can absorb bursts of up to 64 messages, so a briefly slow client on a good connection is fine.
- A client on a consistently slow connection (e.g. 56kbps mobile) will continuously drop ticks during market hours.

**What this means in practice:** The UI will show updates that jump in price (gaps in the stream) but won't crash or show an error. The client won't know it missed messages.
w
**Known gap:** There is no dropped-message counter, no slow-consumer alert, and no automatic disconnect after N consecutive drops. A production hardening would add either a disconnect threshold (if the buffer has been full for X seconds, close the connection) or a counter metric exposed to Prometheus.

---

## 5. How are WebSocket connections load-balanced across multiple server instances? Do you use sticky sessions?

**Currently: there is only one instance. No load balancing is configured.**

`docker-compose.yml` has `container_name: stockflow-websocket-service` — a fixed name implies a single container. There is no Swarm, Kubernetes, nginx upstream block, or HAProxy config in the repository.

**If you were to scale horizontally**, here's what the architecture supports and where it breaks:

**What works out of the box:** Because `websocket-service` subscribes to Redis pub/sub channel `market:updates`, and Redis delivers the same message to every subscriber, scaling to N instances is mostly correct for market data. Every instance gets every tick and fans it out to its local clients.

**What breaks:** The Hub is entirely in-process (`sync.RWMutex` + `map[string]*Client`). Client subscriptions are local to one instance. There is no cross-instance client registry. If an HTTP load balancer round-robins WebSocket connections, a client that reconnects to a different instance starts with no subscriptions — it must re-send its `subscribe` action (which the frontend hook does on `ws.onopen`). This is actually fine if the client reconnects cleanly.

**Sticky sessions:** Not strictly required because clients resubscribe on `onopen`. But without sticky sessions, a rolling restart or instance crash forces all clients to reconnect and resubscribe simultaneously — a thundering herd on the remaining instances. Sticky sessions (via `ip_hash` in nginx or cookie-based affinity) would reduce this.

**True horizontal scaling would require:**
1. An L4/L7 load balancer (nginx, Traefik, AWS ALB).
2. Sticky sessions or connection-aware reconnect logic on the client.
3. No change to the Redis pub/sub fan-out (it already works per-instance).

---

## 6. What's stored in Redis vs Postgres, and why split it that way?

| Data | Store | Key Pattern | TTL | Reason |
|---|---|---|---|---|
| User accounts | **Postgres** | `users` table | Permanent | Durable identity data; needs ACID guarantees and relational queries |
| Refresh token sessions | **Redis** | `session:<userID>:<tokenID>` | 7 days | Ephemeral, high-read, needs fast existence checks; no relational joins needed |
| Latest price per symbol | **Redis** | `market:price:<SYMBOL>` | 24 hours | Millisecond reads for REST quotes; naturally expires; no durability requirement |
| Live price stream | **Redis Pub/Sub** | channel: `market:updates` | N/A (fire-and-forget) | Fan-out to N consumers with zero persistence; at-most-once is acceptable |

**The design rationale:**

*Postgres* holds the only data that must survive a Redis flush: user identities and credentials. It's written to at registration and read at login — low frequency, durability required.

*Redis* holds everything time-sensitive. Session tokens are intentionally ephemeral — they should auto-expire, and checking existence must be sub-millisecond (it happens on every authenticated request). Putting sessions in Postgres would mean a DB round-trip on every API call.

Price data is inherently ephemeral. Storing AAPL's last trade in Postgres would be wasteful — you'd never query historical ticks from this table (that belongs in a time-series DB like InfluxDB or TimescaleDB if you needed it). Redis with a 24h TTL gives you the latest snapshot cheaply and naturally handles the "price is stale after market close" case.

**What's notably absent:** There is no time-series storage. Historical OHLCV data, trade history, and portfolio positions are not persisted anywhere in the current stack. This is appropriate for an MVP focused on live streaming, but would be the next thing to add.

---

## 7. How do you prevent stale data from being shown as live during a Redis or network partition?

**Honestly — you don't, not fully. Here's what you have and what's missing:**

### What exists

**Price cache TTL:** `market:price:<SYMBOL>` expires after **24 hours**. If the Finnhub connection drops and no new ticks arrive, the cached price will eventually expire and `GetLatest()` will return "price not found", falling back to a live Finnhub REST call. So there's a bounded staleness window on the REST quote path.

**Ingestion reconnect:** `IngestionService.Run()` has a reconnect loop:
```go
for {
    if err := s.provider.Connect(ctx); err != nil {
        sleepOrDone(ctx, 3*time.Second)
        continue
    }
    // ... consume loop
}
```
If the Finnhub WebSocket drops, it retries every 3 seconds. During that window, no new ticks are published to Redis, and WebSocket clients see frozen prices.

**Redis subscriber reconnect:** `RedisSubscriber.RunWithReconnect()` retries immediately (no backoff) when the channel closes.

### What's missing

- **No timestamp staleness check on the frontend.** `useStockWebSocket` renders `lastTrade` indefinitely — if you received a tick at 9:30am and it's now 4:00pm, the UI still shows the old price without any "data may be stale" warning.
- **No `lastUpdated` age indicator in the UI.**
- **No Redis health check gating the live indicator.** The frontend connection status is `connected | disconnected | error` based on the WebSocket connection, not on whether fresh data is flowing through it.
- **Access tokens during Redis partition:** If Redis is down, `SessionRepository.SessionExists()` errors, and all token refreshes fail with 500. Users can still use unexpired access tokens (15 min window) but can't get new ones until Redis recovers.

**Production hardening would add:** A timestamp on each tick frame; frontend logic to mark data stale if `Date.now() - lastTrade.timestamp > threshold`; a heartbeat message from the server every N seconds so the client can distinguish "no trades right now" from "connection is dead but open."

---

## 8. What's your reconnection/backoff strategy on the client side if the WebSocket drops?

**There is no automatic reconnection or backoff in the current frontend code.**

```typescript
// use-stock-websocket.ts
ws.onerror = () => setStatus("error");
ws.onclose = () => setStatus("disconnected");
```

When the connection closes, state flips to `"disconnected"` or `"error"` and the UI reflects that. No retry timer fires. The exported `reconnect` function (which is just the `connect` callback) must be called manually, or the component reconnects when the `symbol` prop changes (which triggers `useEffect([connect])`).

**There is no:**
- Exponential backoff (e.g. 1s → 2s → 4s → ... → 30s cap)
- Jitter to prevent thundering herd on server restart
- Maximum retry limit
- "Reconnecting..." indicator distinct from "Disconnected"

**What a production reconnection hook would look like:**

```typescript
// pseudocode — not in the codebase
const backoff = useRef(1000);

ws.onclose = () => {
  setStatus("reconnecting");
  const delay = Math.min(backoff.current, 30_000);
  backoff.current = delay * 2 + Math.random() * 1000; // jitter
  reconnectTimer.current = setTimeout(connect, delay);
};

ws.onopen = () => {
  backoff.current = 1000; // reset on success
  setStatus("connected");
};
```

**Current behavior summary:** A dropped WebSocket shows as disconnected and stays disconnected until the user navigates away and back, changes the symbol, or manually calls `reconnect`. For a live trading dashboard this is a meaningful UX gap.

---

## 9. Where are the goroutines in this system, and what happens if one panics — does it crash the whole server?

### Goroutine inventory

**`market-service`:**
| Goroutine | Spawned in | Purpose |
|---|---|---|
| `go srv.ListenAndServe()` | `cmd/main.go` | HTTP server for REST quote endpoints |
| `go ingestion.Run(ctx)` | `cmd/main.go` | Drives the entire ingest pipeline |
| `go p.readLoop(conn)` | `finnhub.Provider.Connect()` | Reads frames from Finnhub WebSocket, pushes to `trades` channel |

**`websocket-service`:**
| Goroutine | Spawned in | Purpose |
|---|---|---|
| `go srv.ListenAndServe()` | `cmd/main.go` | HTTP/WebSocket server |
| `go sub.RunWithReconnect(ctx)` | `cmd/main.go` | Subscribes to Redis `market:updates`, calls `hub.Broadcast` |
| `go client.WritePump()` | `ws_handler.go` per connection | Drains `send` channel → `conn.WriteMessage`; **one goroutine per connected client** |

**`auth-service` / `api-gateway`:** Only HTTP server goroutine + Gin's per-request goroutines (pooled by the net/http runtime).

### Panic behavior

**HTTP handler goroutines (Gin):** `gin.Default()` includes Gin's built-in `Recovery()` middleware. A panic in any HTTP handler is caught, logged, and returns a 500 — the server keeps running.

**Non-HTTP goroutines — no recovery:**
- `readLoop` in finnhub provider
- `consume` inside `IngestionService.Run`
- `RunWithReconnect` / `Run` in RedisSubscriber
- `client.WritePump()` for each WebSocket client

**None of these have a `recover()` call.** A panic in any of them crashes the entire process. For example, a nil pointer dereference in `WritePump` would kill the websocket-service, disconnecting all current clients.

**What saves you at the process level:** Every service in `docker-compose.yml` has `restart: unless-stopped`. Docker will restart the crashed container, but there's a gap: all current WebSocket connections are killed, Redis pub/sub subscription is lost (and re-established on restart), and any clients without reconnect logic will sit disconnected.

**Production hardening:** Wrap long-running goroutines in a `recover()` + restart loop:
```go
go func() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("WritePump panic: %v", r)
            // optionally restart or just log
        }
    }()
    client.WritePump()
}()
```

---

## 10. What metrics did you actually expose to Prometheus, and what alert would fire first if the system degraded?

### What's instrumented

All four services share `shared/metrics/metrics.go` which registers exactly **two metric families**:

**`http_requests_total` (CounterVec)**
- Labels: `service`, `method`, `path`, `status`
- Scraped from `/metrics` on each service
- Skips `/metrics` and `/health` paths

**`http_request_duration_seconds` (HistogramVec)**
- Labels: `service`, `method`, `path`
- Default Prometheus buckets: `0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10s`

Scraped by Prometheus every **15 seconds** from all four services. Grafana shows three panels:
1. Request rate by service — `sum by (service)(rate(http_requests_total[5m]))`
2. p95 latency by service — `histogram_quantile(0.95, sum by (le, service)(rate(http_request_duration_seconds_bucket[5m])))`
3. Request rate by status — `sum by (service, status)(rate(http_requests_total[5m]))`

### What's NOT instrumented (notable gaps)

- Active WebSocket connections (no gauge)
- Messages dropped by `TrySend` (no counter)
- Redis pub/sub message rate or lag
- Finnhub reconnect attempts
- Hub broadcast duration
- JWT validation failures per minute
- Cache hit/miss ratio on `market:price:*`

### What alert fires first in degradation

Given only HTTP metrics are exposed, the first observable signal would be a spike in `http_requests_total{status="5xx"}` on the `auth-service` if Redis goes down (token refreshes fail with 500). If the Finnhub feed drops silently, **nothing alerts** — there are no pub/sub lag or tick-rate metrics.

**The most operationally dangerous gap:** A silent Finnhub disconnect (readLoop exits, `trades` channel drains, ingestion stops) would produce zero change in any current Prometheus metric. WebSocket clients would see frozen prices with no server-side alert firing.

**First alert to add:**
```
# fires if no ticks have been published in the last 60 seconds
rate(market_ticks_published_total[60s]) == 0
```
That requires adding a `promauto.NewCounter` increment in `IngestionService.consume()` — currently absent.

---

## 11. If concurrent connections went from 1K to 1M, what's the first component that breaks?

**In order of likely failure:**

### 1. `Hub.Broadcast()` — O(N) under read lock (breaks first)
Every tick from Redis triggers a full iteration of all registered clients under `sync.RWMutex`. At 1M clients and hundreds of ticks per second, you have `10^8` `IsSubscribed` calls per second, all under a single read lock. Multiple goroutines calling `Broadcast` concurrently will contend even on a read lock when the map is large. This is the **primary bottleneck**.

Fix: Shard the hub into N sub-hubs (e.g. 256 shards keyed by `client.id[:2]`), reducing contention by 256×.

### 2. Per-client `WritePump` goroutines — 1M goroutines
Each connection spawns exactly one goroutine. Go can handle millions of goroutines (2KB initial stack each), but 1M goroutines = ~2GB of stack memory minimum, plus GC scanning pressure proportional to the number of live goroutines. This likely pushes GC pause times into the tens-of-milliseconds range, causing latency spikes for all clients.

### 3. `send` channel heap allocation — ~512MB
Each `send chan []byte` has capacity 64. A buffered channel's backing array is allocated on the heap. At 1M clients: `1,000,000 × 64 × 8 bytes (pointer) = 512MB` just for channel backing arrays, before any actual payloads.

### 4. Single Redis pub/sub connection
The `websocket-service` holds **one** Redis pub/sub subscription. The `pubsub.Channel()` goroutine receives messages sequentially. At high symbol counts and tick rates, the receive loop can lag behind Redis, building up a backlog. Redis pub/sub is not designed for millions of downstream consumers on a single channel — the right answer at scale is to fan out at the application layer using a topic-per-symbol model, or switch to NATS/Kafka with consumer groups.

### 5. Single websocket-service container
Docker Compose has one instance. A single Go process on typical cloud hardware (4 vCPU, 8GB RAM) handles roughly 100K–500K idle WebSocket connections. 1M connections requires horizontal scaling, which requires all the load balancing work described in question 5.

**TL;DR:** The Hub's O(N) broadcast under a single mutex breaks first, then goroutine/memory pressure, then the single Redis consumer.

---

## 12. Did you load-test this? What was the actual latency/throughput you measured?

**No load testing results exist in the codebase.** There are no benchmark files (`*_bench_test.go`), no k6/Locust/Artillery scripts, and no results documents.

What can be reasoned from the architecture:

- **Single-instance ceiling for WebSocket connections:** Typically 50K–200K on a modest VM before the Hub's broadcast loop becomes the bottleneck (depends on tick rate and subscription overlap).
- **Redis pub/sub throughput:** Redis can handle ~1M messages/sec on a single node; at typical market tick rates (hundreds/sec for a handful of symbols) Redis is not the bottleneck.
- **p99 WebSocket write latency:** With no disk I/O in the hot path (Redis → memory → socket), expect sub-5ms in a local Docker network; tens of milliseconds in a real datacenter depending on Redis RTT.

**To actually measure this**, the right tools for this stack would be:
- [k6](https://k6.io/) with the WebSocket extension for client load
- `go test -bench` for Hub.Broadcast micro-benchmarks
- Redis `MONITOR` or `LATENCY HISTORY` for pub/sub lag

---

## 13. How do you guarantee message ordering per symbol if updates are fanned out across multiple goroutines/instances?

**You don't — and in most cases you don't need to. Here's the precise ordering contract:**

### Within a single `market-service` instance

Finnhub's WebSocket delivers frames in order. `readLoop` pushes each trade from a batch sequentially into `trades chan` (buffer 256). `consume()` drains the channel sequentially in a single goroutine — there is no parallel processing of trades. Redis `PUBLISH` calls are sequential within `consume`. So **publish order equals arrival order** for a single instance.

### Within Redis pub/sub

Redis delivers messages to a subscriber in publish order on a single channel. The `pubsub.Channel()` goroutine in `websocket-service` processes messages sequentially — one message at a time. So `hub.Broadcast()` is called in the same order messages were published.

### Within a single Hub

`Broadcast` is called from the single Redis subscriber goroutine, sequentially. All clients for a given symbol receive ticks in the same order.

### The ordering risks

1. **Batched Finnhub frames:** A single Finnhub frame can contain multiple trades with different timestamps (e.g. `t=100, t=99, t=101`). They're pushed into the channel in array order, not timestamp order. No sorting is applied — if Finnhub delivers them out of order, they arrive out of order.

2. **Horizontal scaling:** With multiple `websocket-service` instances each subscribed to `market:updates`, each instance broadcasts independently to its local clients. Two clients on different instances are subject to the same Redis delivery order (since Redis delivers in publish order to all subscribers), so this is fine.

3. **Multiple `market-service` instances:** If you ever ran two `market-service` instances both publishing to `market:updates`, you'd get interleaved publishes with no coordination — potentially out-of-order ticks for the same symbol. Currently only one instance is configured.

4. **No sequence numbers:** The `NormalizedTick` wire format has no sequence number or version field. Clients can't detect gaps or reorder out-of-sequence frames.

**Bottom line:** Ordering is preserved in the happy path (single instances, sequential processing). It's not guaranteed if you scale `market-service` horizontally or if Finnhub delivers batches with non-monotonic timestamps.

---

## 14. What does your Docker setup look like — one container per service, or a single binary?

**One container per service, 8 containers total.**

```
stockflow-postgres          postgres:17-alpine       port 5432
stockflow-redis             redis:7-alpine           port 6379
stockflow-api-gateway       Built: api-gateway/      port 8080
stockflow-auth-service      Built: auth-service/     port 8084
stockflow-market-service    Built: market-service/   port 8082
stockflow-websocket-service Built: websocket-service/ port 8083
stockflow-prometheus        prom/prometheus:v2.54.1  port 9090
stockflow-grafana           grafana/grafana:11.3.0   port 3001
```

### How the build contexts work

Each Go service has its own `Dockerfile`. The build context is `./backend` (the entire backend directory), and the Dockerfile path selects which service to build:

```yaml
api-gateway:
  build:
    context: ./backend
    dockerfile: api-gateway/Dockerfile
```

This allows the Dockerfiles to reference a shared `shared/` Go package across service boundaries without needing to copy files — the entire `backend/` directory is in the build context.

### Networking

All containers join `stockflow-network` (Docker bridge). Services reference each other by **container service name** as DNS (e.g. `http://auth-service:8084`), not by `localhost` or IP. The api-gateway is the only service that exposes port 8080 to the host; all other ports are mapped for local development but would be unexposed in production.

### Startup ordering

`depends_on` with `condition: service_healthy` ensures Postgres and Redis are ready before any Go service starts. Healthchecks use `pg_isready` for Postgres and `redis-cli ping` for Redis. Application services use `condition: service_started` (not `service_healthy` — no healthcheck endpoint is wired into `depends_on` for Go services, though `/health` and `/ready` endpoints exist).

### Shared environment

A YAML anchor `x-backend-env` injects common variables into every service:

```yaml
x-backend-env: &backend-env
  DATABASE_URL: postgres://...@postgres:5432/stockflow
  REDIS_URL: redis://redis:6379
  GIN_MODE: release
  SENTRY_DSN: ${SENTRY_DSN:-}
  POSTHOG_API_KEY: ${POSTHOG_API_KEY:-}
```

Sensitive values (`JWT_SECRET`, `FINNHUB_API_KEY`, `POSTGRES_PASSWORD`) are expected via a `.env` file at the project root and are never hardcoded.

### Persistence

Two named volumes provide data durability across container restarts:
- `stockflow-postgres-data` — all user data
- `stockflow-redis-data` — Redis AOF log (append-only file persistence is enabled via `--appendonly yes`)

### To run everything

```bash
# from /StockFlow
cp .env.example .env          # fill in secrets
docker compose up --build     # first time
docker compose up             # subsequent runs
docker compose logs -f        # tail all logs
docker compose down           # stop and remove containers
docker compose down -v        # also wipe volumes (destroys data)
```

---

## 15. Walk me through the auth flow end-to-end — register, login, make an authenticated request, refresh, logout.

**Register:** `POST /register` → `UserService.Register()` hashes the password with bcrypt at `DefaultCost` (10 rounds) via `model.HashPassword()`, persists the `users` row in Postgres via GORM, then calls `IssueTokens()` which generates an access token (HS256, 15 min) and a refresh token (HS256, 7 days) and writes `session:<userID>:<tokenID>` to Redis with a 7-day TTL. Returns both tokens to the client.

**Login:** `POST /login` → fetches user by email, calls `model.CheckPassword(hash, plaintext)` (bcrypt compare), on success calls `IssueTokens()` identically to register. A new Redis session key is written; old sessions from previous logins are **not** invalidated (no single-session enforcement).

**Authenticated request:** Client sends `Authorization: Bearer <accessToken>`. In the auth-service, `AuthRequired()` middleware calls `jwt.VerifyAccessToken(token)` which validates the HS256 signature and `exp` claim. No Redis lookup — access tokens are stateless. This means a logout does not immediately invalidate an unexpired access token; it remains valid for up to 15 minutes.

**Refresh:** `POST /refresh` with the refresh token in the body → `VerifyRefreshToken()` validates the signature and expiry, extracts `{userID, tokenID}` from claims → `SessionExists(userID, tokenID)` checks Redis → `DeleteSession` deletes the old key (one-time use, prevents replay) → `IssueTokens()` issues new pair and writes a new Redis session key. This is a **rotating refresh token** pattern: each refresh invalidates the previous token.

**Logout:** `POST /logout` → `VerifyRefreshToken()` + `DeleteSession()` removes the Redis key. The refresh token is dead. The access token remains valid until its 15-minute expiry (no blocklist). A stolen access token is valid for at most 15 minutes post-logout.

**The security tradeoff:** Short-lived access tokens (15 min) are the only mitigation for token theft after logout. Adding a Redis-backed access token blocklist would close this gap but adds a Redis lookup to every authenticated request.

---

## 16. Your `RateLimitMiddleware` in the api-gateway is a stub — what would a real implementation look like, and how would it handle distributed rate limiting across multiple gateway instances?

The current middleware in `shared/gateway/middleware.go` calls `c.Next()` with no logic — it's a placeholder.

**Local (single instance) implementation** using a token bucket:
```go
// one bucket per client IP, using golang.org/x/time/rate
var limiters sync.Map

func RateLimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        ip := c.ClientIP()
        v, _ := limiters.LoadOrStore(ip, rate.NewLimiter(rate.Every(time.Second), 100))
        limiter := v.(*rate.Limiter)
        if !limiter.Allow() {
            c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
            return
        }
        c.Next()
    }
}
```

**Distributed implementation using Redis** (required once you have multiple gateway instances):

The canonical approach is the **sliding window log** or **fixed window counter** in Redis using `INCR` + `EXPIRE`:

```
key:   ratelimit:<clientIP>:<window_unix_second>
value: INCR (atomic counter)
TTL:   set on first write to window duration
```

If the counter exceeds the limit, return 429. This is accurate to the window but has a burst problem at window boundaries. A **sliding window** using a sorted set (`ZADD` + `ZREMRANGEBYSCORE` + `ZCARD`) is more precise but more expensive.

Since Redis is already in the stack, the distributed version costs nothing operationally. You'd also want to rate-limit by `X-Request-ID` origin or API key rather than just IP for authenticated endpoints.

---

## 17. How does the api-gateway proxy WebSocket connections? The frontend connects directly to `websocket-service:8083` — is that intentional, and what are the security implications?

**It's intentional — and it's a deliberate architectural tradeoff with real security implications.**

Looking at `api-gateway/cmd/main.go`, the routes are:
```go
r.Any("/api/auth/*proxyPath",   makePathProxy(authURL, "/api/auth"))
r.Any("/api/market/*proxyPath", makePathProxy(marketURL, "/api/market"))
```

There is no `/api/ws` or `/ws` route. `httputil.ReverseProxy` can proxy WebSocket connections (it handles the `Upgrade` header correctly), but it was never wired in. The frontend connects directly to the websocket-service using `NEXT_PUBLIC_WS_URL` (port 8083).

**Implications:**

*Security:* The api-gateway's middleware chain — request ID injection, timeout enforcement, CORS policy, and (once implemented) rate limiting — is completely bypassed for WebSocket connections. Each WebSocket client connects directly to port 8083, which has its own auth check (`WS_REQUIRE_AUTH` env flag) but none of the gateway cross-cutting concerns.

*CORS:* The WebSocket protocol doesn't enforce CORS the same way HTTP does. The `CheckOrigin: func(r *http.Request) bool { return true }` in the gorilla upgrader accepts connections from any origin — this would need to be tightened in production.

*Rate limiting:* A single IP can open unlimited WebSocket connections since there's no gateway in front of them.

**To fix:** Add a WebSocket proxy route to the gateway:
```go
r.GET("/api/ws", makeWSProxy(wsURL))
```
Or accept the direct connection model but add rate limiting and origin checking at the websocket-service level.

---

## 18. What happens to in-flight WebSocket writes when the server shuts down — do clients get a clean close frame?

Looking at `websocket-service/cmd/main.go`, shutdown follows the standard Go pattern:
```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
srv.Shutdown(ctx)
```

`srv.Shutdown()` stops accepting new connections and waits up to 5 seconds for in-flight HTTP requests to complete. **But WebSocket connections are long-lived HTTP connections** — from `net/http`'s perspective they are hijacked connections and are not tracked by `Shutdown`. They are closed abruptly when the process exits.

**What clients experience:** The TCP connection is closed without a WebSocket close frame (opcode `0x8`). The browser's `WebSocket` object fires `onclose` with `code 1006` (abnormal closure) rather than the clean `1000` or `1001`. The frontend sets status to `"disconnected"` — same as an abrupt network drop.

**To send clean close frames on shutdown:**
1. Pass a cancellable `context.Context` to `ws_handler.Handle()` tied to the shutdown signal.
2. On shutdown, cancel the context, which causes `conn.ReadMessage()` to return.
3. In the deferred cleanup, send `conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "server shutting down"))` before calling `client.Close()`.

Without this, every rolling deploy or restart appears to clients as a network failure rather than a server restart — relevant because it affects reconnection behavior.

---

## 19. The `Hub` uses `sync.RWMutex` — walk me through a potential deadlock scenario in the current code, and how would you detect it in production?

**No deadlock exists in the current code**, but there are two patterns that could introduce one if the code evolves:

**Scenario 1 — Callback holding hub lock calls back into hub:**
If `TrySend` were changed to call any method that acquires `hub.mu` (e.g. `Unregister`), and `Unregister` was called while `Broadcast` holds a read lock that another goroutine is trying to upgrade to a write lock, you'd get a deadlock. Go's `sync.RWMutex` does not support lock upgrading.

**Scenario 2 — Client lock acquired inside hub lock:**
`Hub.Broadcast` holds `hub.mu.RLock()` and then calls `client.IsSubscribed()` which acquires `client.mu.RLock()`. If any code path acquires `client.mu` first and then tries to acquire `hub.mu`, you have a classic lock-ordering inversion. Currently `ws_handler.go` calls `client.Subscribe()` (acquires `client.mu`) without holding `hub.mu`, so the ordering is consistent. But it's fragile.

**The safe rule:** Always acquire `hub.mu` before `client.mu`, never the reverse. This is implicit in the current code but not documented.

**Detecting deadlocks in production:**
- `go tool pprof` goroutine profiles: a deadlock shows all goroutines blocked on mutex operations.
- `SIGQUIT` (Ctrl+\\) prints a full goroutine dump to stderr.
- [golangci-lint](https://github.com/golangci/golangci-lint) with `govet` catches some lock copying issues.
- The Go race detector (`go test -race`) catches data races that can precede deadlocks.
- In production: expose `runtime.NumGoroutine()` as a Prometheus gauge — a monotonically increasing goroutine count that never decreases is a strong deadlock signal.

---

## 20. How would you add per-symbol subscription counts and a "top symbols by subscriber count" API endpoint?

This is a product-useful feature (showing trending symbols) that requires coordinating data across the Hub.

**Step 1 — Track counts in the Hub:**
```go
type Hub struct {
    mu          sync.RWMutex
    clients     map[string]*Client
    symbolCount map[string]int  // symbol → subscriber count
}

func (h *Hub) Register(client *Client) {
    h.mu.Lock()
    h.clients[client.id] = client
    h.mu.Unlock()
}

// Called when client.Subscribe(symbol) fires:
func (h *Hub) IncrSymbol(symbol string) {
    h.mu.Lock()
    h.symbolCount[symbol]++
    h.mu.Unlock()
}

func (h *Hub) DecrSymbol(symbol string) {
    h.mu.Lock()
    h.symbolCount[symbol]--
    if h.symbolCount[symbol] <= 0 {
        delete(h.symbolCount, symbol)
    }
    h.mu.Unlock()
}
```

**Step 2 — Expose an endpoint:**
```go
// GET /internal/symbols/top?n=10
func (h *Hub) TopSymbols(n int) []SymbolCount {
    h.mu.RLock()
    defer h.mu.RUnlock()
    // copy, sort by count descending, return top N
}
```

**Step 3 — Multi-instance consideration:**
With one websocket-service instance, the Hub's in-memory map is accurate. With multiple instances, each hub only knows its local subscribers. To get a global count you'd need to either:
- Publish subscribe/unsubscribe events to a Redis sorted set (`ZINCRBY market:subscriptions <symbol> 1`) — accurate but adds Redis writes on every subscribe action.
- Accept approximate counts by summing across instances via an internal aggregation call.

The Redis sorted set approach also naturally gives you `ZREVRANGE market:subscriptions 0 9 WITHSCORES` for the top-10 query — a single Redis command.

---

## 21. The `NormalizedTick` has a `Source` field set to `"finnhub"`. Why is that there, and how would adding a second data provider (e.g. Polygon.io) change the architecture?

The `Source` field is forward-looking — it marks the origin of each tick so consumers can differentiate feeds without needing to infer it from context.

**Adding a second provider without `Source`:** You'd have two providers publishing ticks to `market:updates` with no way to distinguish them. A consumer that wants to prefer one provider over another (for latency, accuracy, or failover) has no signal to act on.

**With `Source`:**
```json
{"symbol":"AAPL","price":182.50,"source":"finnhub","timestamp":1713000000}
{"symbol":"AAPL","price":182.51,"source":"polygon","timestamp":1713000001}
```
Consumers can implement source preference, deduplication by `(symbol, timestamp, source)`, or latency comparison.

**Architectural changes to add Polygon.io:**

The `provider.MarketProvider` interface already exists:
```go
type MarketProvider interface {
    Connect(ctx context.Context) error
    Subscribe(symbols []string) error
    Trades() <-chan RawTrade
    Reconnect(ctx context.Context) error
    Close() error
}
```

You'd implement `PolygonProvider` satisfying this interface, setting `Source: "polygon"` in `RawTrade`. Then in `main.go`, run two `IngestionService` instances — one per provider — both publishing to `market:updates`.

**Deduplication challenge:** Both providers will emit ticks for the same symbol at nearly the same time. The websocket-service would broadcast two ticks per trade. The frontend takes the last element of `payload.data` — so the ordering between providers becomes meaningful. You'd likely want the ingestion service to implement a dedup window: if a tick for `(symbol, price, timestamp)` was published within the last N milliseconds, drop the duplicate.

---

## 22. How is the `JWT_SECRET` used — what breaks if it rotates, and how would you implement zero-downtime secret rotation?

**How it's used:** `shared/jwt` package signs and verifies all access and refresh tokens with HS256 using the `JWT_SECRET` environment variable. Both signing and verification use the same secret (symmetric).

**What breaks on rotation:**
- All currently issued access tokens (up to 15 min old) become immediately invalid — every in-flight authenticated request returns 401.
- All refresh tokens (up to 7 days old) become immediately invalid — all users are effectively logged out and must re-authenticate with username/password.
- The Redis session keys (`session:<uid>:<jti>`) become orphaned — they still exist but the JTIs they reference can no longer be decoded from a token.

**Zero-downtime rotation strategy:**

1. **Versioned secrets:** Add a `kid` (key ID) claim to the JWT header. Maintain a map of `kid → secret`.
   ```go
   // sign with new key
   token.Header["kid"] = "v2"
   // verify: read kid from header, look up secret
   secret := secrets[claims.KeyID]
   ```

2. **Overlap window:** Deploy with both `v1` and `v2` secrets active. New tokens are signed with `v2`. The verifier accepts tokens signed by either `v1` or `v2`. After the longest token TTL (7 days for refresh tokens) has elapsed, all `v1` tokens have expired naturally. Remove `v1` from the accepted set.

3. **Config-driven:** Store `{"v1": "old-secret", "v2": "new-secret"}` in a secrets manager (AWS Secrets Manager, Vault) and hot-reload on a signal or poll interval — no restart required.

**Current gap:** There is only one secret, no `kid` in the JWT header, and no hot-reload mechanism. A secret rotation today requires a restart and forces all users to log in again.

---

## 23. What's missing before this project is production-ready? Give your honest assessment.

Based on the actual codebase, ranked by severity:

**Critical:**
1. **No panic recovery in background goroutines** — `WritePump`, `readLoop`, `RunWithReconnect` can crash the process. One nil pointer dereference takes down all connected clients.
2. **`RateLimitMiddleware` is a no-op** — the api-gateway has no actual rate limiting, making it trivially DoS-able.
3. **WebSocket connections bypass the api-gateway entirely** — no CORS enforcement, no rate limiting, `CheckOrigin` returns `true` for all origins.
4. **No staleness signal to the client** — a dead Finnhub connection looks identical to a live one with no trades.

**High:**
5. **No frontend reconnection backoff** — a server restart triggers all clients to reconnect simultaneously (thundering herd).
6. **No tick-rate metrics** — the most important alert (feed went silent) can't be written against current Prometheus instrumentation.
7. **Access tokens are not revocable** — a compromised token is valid for 15 minutes after logout with no way to invalidate it sooner.
8. **No clean WebSocket shutdown** — rolling deploys cause `1006` abnormal closure on all clients instead of a graceful `1001 Going Away`.

**Medium:**
9. **No horizontal scaling infrastructure** — no load balancer config, no sticky session policy, no Kubernetes manifests.
10. **No time-series storage** — historical prices, portfolio history, and trade audit logs have nowhere to live.
11. **Hub is a single-mutex O(N) broadcast** — will degrade significantly above ~50K concurrent connections.
12. **No distributed tracing** — `X-Request-ID` is propagated but not exported to any trace collector (Jaeger, Tempo). Cross-service debugging requires log correlation by hand.

**Low / Nice to have:**
13. **No integration tests** — only `user_service_test.go` and `finnhub_test.go` exist; no end-to-end flow tests.
14. **No message deduplication** — adding a second data provider would send duplicate ticks to clients.
15. **No per-symbol subscription count API** — useful for a "trending symbols" feature.
