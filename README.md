# Image Analysis Platform Lite

Repository: [github.com/droskat/BH-Backend](https://github.com/droskat/BH-Backend)

A high-performance Go + Gin REST API for image metadata management, built for 1M DAU with P99 < 100ms latency.

Kafka ideally  can be and should be de-coupled to run separate consumer flow. To maintain the simplicity of running it all together has been kept coupled. 

## Architecture

```
cmd/server/          → Entry point, router setup, graceful shutdown
config/              → Environment-driven configuration
controllers/         → HTTP handlers with input validation
middlewares/         → JWT auth engine, rate limiting, panic recovery
models/              → Domain models, DTOs, Kafka events
services/            → Business logic, repository interfaces, cache layer
connectors/          → Cassandra, Redis, Kafka client initialization
internal/worker/     → Background Kafka consumer for image processing
migrations/          → CQL schema scripts
```

## Technology Stack

| Layer | Technology |
|-------|-----------|
| API Runtime | Go 1.22+ / Gin |
| Primary DB | Apache Cassandra (gocql) |
| Cache | Redis Cluster (go-redis/v9) |
| Message Broker | Apache Kafka (kafka-go) |
| Auth | Dual-token JWT (golang-jwt/v5) |

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/login` | Generate token pair |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/images/bulk` | Bulk metadata upload (max 50) |
| GET | `/api/v1/images` | List user's images |
| GET | `/api/v1/images/:id` | Get image details |
| GET | `/api/v1/images/:id/download` | Generate presigned download URL |
| PUT | `/api/v1/images/:id` | Update image metadata |
| DELETE | `/api/v1/images/:id` | Delete image |

## Quick Start

```bash
# Run Cassandra migration
cqlsh -f migrations/001_init_schema.cql

# Set environment variables (or use defaults)
export CASSANDRA_HOSTS=127.0.0.1
export REDIS_ADDR=127.0.0.1:6379
export KAFKA_BROKERS=127.0.0.1:9092

# Build and run
go build -o server ./cmd/server
./server
```

## API Verification

Sample curl commands live in:

- `scripts/api-curl-samples.sh` — runnable script (full flow or single endpoint)
- `scripts/api-curl-samples.md` — copy-paste curl snippets

Run the full happy-path flow (login → bulk upload → list → get → download → update → delete):

```bash
chmod +x scripts/api-curl-samples.sh
./scripts/api-curl-samples.sh
```

Run a single endpoint:

```bash
./scripts/api-curl-samples.sh health
./scripts/api-curl-samples.sh login
```

Quick manual check after the server is running:

```bash
export BASE_URL=http://127.0.0.1:8080
export USER_ID=4a2bc1d8-7e3f-412e-a19b-625d91c84f32

curl -sS -X POST "${BASE_URL}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"${USER_ID}\"}"
```

Use the `access_token` from the response as `Authorization: Bearer <token>` on the image endpoints. See `scripts/api-curl-samples.md` for the remaining requests.

## Running Tests

```bash
go test ./... -v
```

## Key Design Decisions

- **Cache-aside with jitter**: Prevents thundering herd via randomized TTL (48h + 1-6h jitter)
- **BOLA protection**: JWT user_id verified against resource ownership on every access
- **Redis pipeline**: Atomic batch operations for cache invalidation during bulk uploads
- **Kafka idempotent writes**: RequiredAcks=All ensures exactly-once semantics
- **Non-logged Cassandra batches**: Independent concurrent writes avoid coordinator overhead
- **Graceful degradation**: Redis failures fall through to Cassandra transparently
