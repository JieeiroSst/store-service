# NotifyHub Service

A high-performance, multi-channel notification microservice built in Go.

## Features

- **3 channels**: Email (SMTP/SendGrid), SMS (Twilio), Firebase Push (FCM)
- **3 schedule types**: Cron, One-time, Interval
- **External data fetching**: Pull live data from any REST API before rendering
- **Go template engine**: Full `text/html/template` with variable injection
- **Worker pool**: Bounded goroutine pool with exponential backoff retry
- **MySQL persistence**: Jobs, templates, channels, data sources, delivery history
- **Prometheus metrics**: Job executions, notification counts, fetch latency, queue depth
- **REST API**: Full CRUD + pause/resume/trigger for all resources
- **Graceful shutdown**: Drains in-flight notifications before exit

## Project Structure

```
notifyhub-service/
├── cmd/server/main.go              ← Entrypoint, wires all dependencies
├── internal/
│   ├── api/
│   │   ├── handler/handler.go     ← HTTP handlers (CRUD + lifecycle)
│   │   ├── middleware/middleware.go← Auth, rate-limit, logger, CORS, recovery
│   │   └── router.go              ← Gin route registration
│   ├── channel/
│   │   ├── registry.go            ← Plugin registry (Sender interface)
│   │   ├── email/sender.go        ← SMTP / SendGrid
│   │   ├── sms/sender.go          ← Twilio
│   │   └── firebase/sender.go     ← FCM push notifications
│   ├── config/config.go           ← Viper config from .env / env vars
│   ├── fetcher/fetcher.go         ← Pooled HTTP client for external APIs
│   ├── job/
│   │   ├── service.go             ← Job business logic + validation
│   │   ├── channel_service.go     ← Channel / DataSource / Template services
│   │   └── export_test.go         ← Test exports
│   ├── model/model.go             ← GORM models + custom JSON types
│   ├── pkg/
│   │   ├── logger/logger.go       ← Zap logger factory
│   │   └── metrics/metrics.go     ← Prometheus metric definitions
│   ├── scheduler/scheduler.go     ← gocron v2 scheduler with singleton mode
│   ├── store/mysql/store.go       ← All DB queries (GORM)
│   ├── template/engine.go         ← Go html/template compile + render cache
│   └── worker/pool.go             ← Goroutine pool, retry, dispatch
├── migrations/                    ← Raw SQL schema files
├── scripts/
│   ├── seed.sh                    ← Demo data seeding via API
│   ├── healthcheck.sh             ← Quick liveness check
│   └── run_migrations.sh          ← Apply SQL files in order
├── test/                          ← Unit + integration tests
├── .env.example                   ← All config variables documented
├── docker-compose.yml             ← MySQL + service
├── Dockerfile                     ← Multi-stage scratch build
└── Makefile                       ← Build, test, Docker, API commands
```

## Quick Start

### Option A — Docker (recommended)

```bash
# 1. Configure
cp .env.example .env
# Edit .env with your SMTP/Twilio/Firebase credentials

# 2. Start everything
make docker-up

# 3. Seed demo data
make seed
```

### Option B — Local (requires MySQL 8)

```bash
# 1. Start MySQL and create DB
mysql -u root -e "CREATE DATABASE notifyhub; CREATE USER notifyhub@localhost IDENTIFIED BY 'password'; GRANT ALL ON notifyhub.* TO notifyhub@localhost;"

# 2. Configure
cp .env.example .env   # set DB_DSN and credentials

# 3. Run
make dev
```

## API Reference

All authenticated endpoints require `X-API-Key: <your-key>` (or `Authorization: Bearer <key>`).

### Resources

| Resource | Base Path | Actions |
|----------|-----------|---------|
| Channels | `/api/v1/channels` | CRUD |
| Data Sources | `/api/v1/data-sources` | CRUD |
| Templates | `/api/v1/templates` | CRUD |
| Jobs | `/api/v1/jobs` | CRUD + pause/resume/trigger |
| History | `/api/v1/history` | List (read-only) |
| Scheduler | `/api/v1/scheduler/status` | Read-only |

### Job Lifecycle

```
POST /api/v1/jobs          → create + auto-register in scheduler
POST /api/v1/jobs/:id/pause   → stop scheduling (keeps DB record)
POST /api/v1/jobs/:id/resume  → re-register in scheduler
POST /api/v1/jobs/:id/trigger → fire immediately (bypasses schedule)
DELETE /api/v1/jobs/:id    → remove from scheduler + DB
```

### Complete Example

```bash
BASE=http://localhost:8080
KEY=your-secret-api-key-here

# 1. Create a channel
CH=$(curl -s -X POST $BASE/api/v1/channels \
  -H "X-API-Key: $KEY" -H "Content-Type: application/json" \
  -d '{"name":"main-email","type":"email","is_active":true}')
CH_ID=$(echo $CH | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")

# 2. Create a template
TM=$(curl -s -X POST $BASE/api/v1/templates \
  -H "X-API-Key: $KEY" -H "Content-Type: application/json" \
  -d '{"name":"report","channel":"email","subject":"Report for {{.Date}}","body":"<p>Hello {{.Name}}</p>","is_active":true}')
TM_ID=$(echo $TM | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")

# 3. Register an external data source (optional)
DS=$(curl -s -X POST $BASE/api/v1/data-sources \
  -H "X-API-Key: $KEY" -H "Content-Type: application/json" \
  -d '{"name":"stats","url":"https://api.example.com/stats","method":"GET","auth_type":"bearer","auth_config":{"token":"MY_TOKEN"},"json_path":"$.data"}')
DS_ID=$(echo $DS | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")

# 4. Create a cron job (every day at 08:00 UTC)
curl -s -X POST $BASE/api/v1/jobs \
  -H "X-API-Key: $KEY" -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Daily Report\",
    \"channel_id\": \"$CH_ID\",
    \"template_id\": \"$TM_ID\",
    \"data_source_id\": \"$DS_ID\",
    \"schedule_type\": \"cron\",
    \"cron_expr\": \"0 8 * * *\",
    \"recipients\": [\"admin@company.com\"],
    \"static_payload\": {\"Name\": \"Admin\", \"Date\": \"today\"}
  }" | python3 -m json.tool

# 5. Trigger immediately to test
curl -s -X POST $BASE/api/v1/jobs/JOB_ID/trigger -H "X-API-Key: $KEY"

# 6. Check delivery history
curl -s "$BASE/api/v1/history?job_id=JOB_ID" -H "X-API-Key: $KEY" | python3 -m json.tool
```

## Schedule Types

| Type | Field | Example |
|------|-------|---------|
| `cron` | `cron_expr` | `"0 8 * * *"` = daily 8AM UTC |
| `once` | `run_at` | `"2025-12-25T09:00:00Z"` |
| `interval` | `interval_sec` | `3600` = every hour |

## Configuration

See `.env.example` for all variables. Key ones:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP listen port |
| `JWT_SECRET` | — | API key for authentication |
| `DB_DSN` | — | MySQL DSN |
| `WORKER_POOL_SIZE` | `10` | Goroutine count |
| `WORKER_QUEUE_SIZE` | `1000` | Buffered task channel depth |
| `WORKER_RETRY_MAX` | `3` | Max retries per message |
| `FETCHER_POOL_SIZE` | `20` | HTTP connection pool size |
| `EMAIL_PROVIDER` | `smtp` | `smtp` or `sendgrid` |
| `SMS_PROVIDER` | `twilio` | `twilio` or `vonage` |

## Monitoring

Prometheus metrics at `GET /metrics`:

| Metric | Type | Description |
|--------|------|-------------|
| `notifyhub_jobs_scheduled_total` | Gauge | Jobs in scheduler |
| `notifyhub_job_executions_total` | Counter | By schedule_type, status |
| `notifyhub_notifications_sent_total` | Counter | By channel, status |
| `notifyhub_worker_queue_depth` | Gauge | Pending tasks |
| `notifyhub_fetch_duration_seconds` | Histogram | External API latency |
| `notifyhub_http_request_duration_seconds` | Histogram | API latency |
| `notifyhub_retry_total` | Counter | Retry attempts by channel |

## Testing

```bash
make test            # all tests
make test-race       # with race detector
make test-cover      # with HTML coverage report
```
