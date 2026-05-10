# Payment Retry Processing System

A concurrent backend processing system in Go to asynchronously manage failed payment retry workflows using goroutines and worker pools.

---

## What this project does

When a payment fails — due to network issues, insufficient funds, or provider downtime — it shouldn't just silently die. This system automatically retries failed payments on a schedule, backing off exponentially between attempts, tracking every state transition, and marking jobs as permanently dead only after all retries are exhausted.

The core idea is a **job queue backed by Postgres** — no Redis, no RabbitMQ, no external message broker. Just `SELECT FOR UPDATE SKIP LOCKED` and a worker pool of goroutines. This is the same pattern used by production systems like Sidekiq (Ruby) and GoodJob (Rails) — simple, reliable, and easy to reason about.

---

## Tech stack

- **Go** — primary language
- **Gin** — HTTP framework for REST APIs
- **GORM** — ORM for Postgres interaction
- **PostgreSQL 15** — persistence and job queue
- **Docker + Docker Compose** — containerised local development

---

## Project structure

```
payment-retry/
├── cmd/
│   └── server/         # main.go — app entry point
├── internal/
│   ├── api/            # HTTP handlers (Gin routes)
│   ├── models/         # DB structs + status constants
│   ├── repository/     # DB query layer (GORM)
│   ├── service/        # business logic + retry strategy
│   ├── worker/         # goroutines + worker pool
│   └── scheduler/      # periodic job polling
├── config/             # env + app config loader
├── migrations/         # SQL migration files
├── docker/             # Dockerfiles
├── docker-compose.yml
├── Makefile
└── go.mod
```

---

## Database schema

```sql
CREATE TABLE retry_jobs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id      VARCHAR(255) NOT NULL,
    amount          DECIMAL(10, 2) NOT NULL,
    currency        VARCHAR(3) NOT NULL DEFAULT 'INR',
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    retry_count     INT NOT NULL DEFAULT 0,
    max_retries     INT NOT NULL DEFAULT 3,
    next_retry_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_error      TEXT,
    webhook_url     VARCHAR(500),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_retry_jobs_status ON retry_jobs(status);
CREATE INDEX idx_retry_jobs_next_retry_at ON retry_jobs(next_retry_at);
CREATE INDEX idx_retry_jobs_status_next_retry ON retry_jobs(status, next_retry_at);
```

### Job lifecycle

```
pending → processing → succeeded
                    ↘
                      pending → processing → succeeded
                             ↘
                               pending → processing → dead
```

| Status | Meaning |
|---|---|
| `pending` | Waiting to be picked up by a worker |
| `processing` | Claimed by a worker, currently being attempted |
| `succeeded` | Payment processed successfully |
| `failed` | Attempt failed, will be retried |
| `dead` | All retries exhausted, permanently failed |

---

## Environment variables

Create a `.env` file in the project root:

```env
DB_DSN=postgres://admin:admin@localhost:5432/payment-retry?sslmode=disable
PORT=8080
MAX_WORKERS=5
MAX_RETRIES=3
```

---

## Getting started

### Prerequisites
- Go 1.21+
- Docker + Docker Compose

### Run locally

```bash
# Start Postgres and pgadmin
docker compose up -d

# Run the SQL migration manually to create the schema
docker exec -i payment-retry-postgres psql -U admin -d payment-retry < migrations/001_create_retry_jobs.sql

# Start the server — AutoMigrate runs on startup to sync the table structure
go run cmd/server/main.go
```

### How migrations work

Run the SQL migration file once on a fresh database — that's it:

```bash
docker exec -i payment-retry-postgres psql -U admin -d payment-retry < migrations/001_create_retry_jobs.sql
```

This creates the table, indexes, and triggers. Any future schema changes will be added as new SQL files in `migrations/` and run the same way.

### Other useful commands

```bash
# Stop all containers
docker compose down

# Run tests
go test ./...

# Run tests with race detector
go test -race ./...

# Build the binary
go build -o bin/server cmd/server/main.go
```

### pgadmin
Visit `http://localhost:5050` — use `postgres` (not `localhost`) as the host when adding the server connection.

---

## REST API

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/jobs` | Create a retry job |
| `GET` | `/api/v1/jobs/:id` | Get job by ID |
| `GET` | `/api/v1/jobs` | List jobs (filter by ?status=) |
| `POST` | `/api/v1/jobs/:id/cancel` | Cancel a pending job |
| `GET` | `/health` | Health check |
| `GET` | `/metrics` | System metrics |

### Create a job

```bash
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "payment_id": "pay_123",
    "amount": 99.99,
    "currency": "INR",
    "max_retries": 3
  }'
```

### Response

```json
{
  "id": "88c6f58e-1a8f-4979-9110-f766f6651224",
  "payment_id": "pay_123",
  "amount": 99.99,
  "currency": "INR",
  "status": "pending",
  "retry_count": 0,
  "max_retries": 3,
  "next_retry_at": "2026-05-11T00:45:18Z",
  "created_at": "2026-05-11T00:45:08Z"
}
```

---

## How the retry system works

### Worker pool
A configurable number of goroutines (`MAX_WORKERS`) run continuously, each waiting for jobs from a shared channel. When the scheduler finds eligible jobs, it submits them to the channel and an available worker picks them up.

### Claiming jobs safely
The scheduler uses `SELECT FOR UPDATE SKIP LOCKED` inside a transaction to atomically claim jobs. This prevents two workers from processing the same job simultaneously — even when running multiple instances of the server.

```sql
SELECT * FROM retry_jobs
WHERE status = 'pending' AND next_retry_at <= NOW()
ORDER BY next_retry_at ASC
LIMIT 10
FOR UPDATE SKIP LOCKED;
```

### Exponential backoff
After each failure, the next retry is scheduled using exponential backoff with jitter to avoid thundering herd:

```
attempt 1 → ~30 seconds
attempt 2 → ~60 seconds
attempt 3 → ~120 seconds
```

Formula: `baseDelay × 2^attempt + random jitter`

### Graceful shutdown
On `SIGINT` or `SIGTERM`, the scheduler stops polling, the worker pool drains all in-flight jobs, and the server exits cleanly — no jobs are lost mid-processing.

---

## Progress

### Completed
- [x] Project structure and module setup
- [x] Docker Compose with Postgres and pgadmin
- [x] Environment config loader
- [x] Database schema and migrations
- [x] `RetryJob` model with status constants
- [x] Repository layer with all methods:
  - `Create`
  - `GetByID`
  - `ClaimPendingJobs` (with `FOR UPDATE SKIP LOCKED`)
  - `UpdateStatus`
  - `IncrementRetry`
- [x] Gin server with health check endpoint
- [x] Manual repository testing via `cmd/test/main.go`

### In progress
- [ ] Service layer (business logic + retry strategy)
- [ ] REST API handlers
- [ ] Worker pool (goroutines + channels)
- [ ] Job scheduler (periodic polling)
- [ ] Graceful shutdown

### Upcoming
- [ ] Structured logging with `slog`
- [ ] Metrics endpoint (`/metrics`)
- [ ] Multi-stage Dockerfile
- [ ] Webhook notifications on job completion
- [ ] Dynamic worker pool resizing
- [ ] Idempotency keys for job creation

---

## Key concepts explored

| Concept | Where used |
|---|---|
| Goroutines | Worker pool — each worker is a goroutine |
| Channels | Job queue between scheduler and workers |
| `sync.WaitGroup` | Graceful shutdown — wait for all workers to finish |
| `sync/atomic` | Lock-free metrics counters |
| `FOR UPDATE SKIP LOCKED` | Safe concurrent job claiming from Postgres |
| Exponential backoff with jitter | Retry scheduling to avoid thundering herd |
| Interface-based design | Repository interface allows easy mocking for tests |
| Pointer receivers | All repository and service methods use pointer receivers |

---

## Architecture overview

```
HTTP Request
     ↓
  Gin Router
     ↓
  Handler  →  Service  →  Repository  →  Postgres
                ↑
           Scheduler  (ticks every N seconds)
                ↓
           Worker Pool
          /     |     \
       Worker Worker Worker  (goroutines)
          \     |     /
           Payment Processor (mock → real)
```