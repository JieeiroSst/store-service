BINARY    := notifyhub
CMD       := ./cmd/server
BUILD_DIR := bin

# ── Build ─────────────────────────────────────────────────────────────────────
.PHONY: build
build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(shell git describe --tags --always 2>/dev/null || echo dev)" \
		-o $(BUILD_DIR)/$(BINARY) $(CMD)
	@echo "✅ Binary: $(BUILD_DIR)/$(BINARY)"

.PHONY: run
run: build
	CONFIG_PATH=.env ./$(BUILD_DIR)/$(BINARY)

.PHONY: dev
dev:
	CONFIG_PATH=.env go run $(CMD)

# ── Test ──────────────────────────────────────────────────────────────────────
.PHONY: test
test:
	go test ./... -v -race -count=1 -timeout=60s

.PHONY: test-short
test-short:
	go test ./... -short -count=1

.PHONY: coverage
coverage:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

# ── Code Quality ──────────────────────────────────────────────────────────────
.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: fmt
fmt:
	gofmt -w .
	goimports -w .

.PHONY: vet
vet:
	go vet ./...

.PHONY: tidy
tidy:
	go mod tidy
	go mod verify

# ── Docker ────────────────────────────────────────────────────────────────────
.PHONY: docker-build
docker-build:
	docker build -t notifyhub:latest .

.PHONY: docker-up
docker-up:
	docker compose up -d --build
	@echo "✅ Services started — API at http://localhost:8080"

.PHONY: docker-down
docker-down:
	docker compose down

.PHONY: docker-logs
docker-logs:
	docker compose logs -f notifyhub

.PHONY: docker-restart
docker-restart:
	docker compose restart notifyhub

# ── Database ──────────────────────────────────────────────────────────────────
.PHONY: migrate
migrate:
	bash scripts/run_migrations.sh

.PHONY: seed
seed:
	bash scripts/seed.sh

.PHONY: db-shell
db-shell:
	docker compose exec mysql mysql -u notifyhub -ppassword notifyhub

# ── API Quick-Test (requires running service) ─────────────────────────────────
API_URL ?= http://localhost:8080
API_KEY ?= your-secret-api-key-here

.PHONY: api-health
api-health:
	curl -s $(API_URL)/health | python3 -m json.tool

.PHONY: api-channels
api-channels:
	curl -s -H "X-API-Key: $(API_KEY)" $(API_URL)/api/v1/channels | python3 -m json.tool

.PHONY: api-jobs
api-jobs:
	curl -s -H "X-API-Key: $(API_KEY)" "$(API_URL)/api/v1/jobs?page=1&page_size=10" | python3 -m json.tool

.PHONY: api-scheduler-status
api-scheduler-status:
	curl -s -H "X-API-Key: $(API_KEY)" $(API_URL)/api/v1/scheduler/status | python3 -m json.tool

# ── Example: Create a cron job via API ───────────────────────────────────────
.PHONY: example-cron-job
example-cron-job:
	@echo "Creating a cron job — edit Makefile to set CHANNEL_ID and TEMPLATE_ID"
	curl -s -X POST $(API_URL)/api/v1/jobs \
	  -H "X-API-Key: $(API_KEY)" \
	  -H "Content-Type: application/json" \
	  -d '{ \
	    "name": "Daily Report", \
	    "channel_id": "CHANNEL_ID", \
	    "template_id": "TEMPLATE_ID", \
	    "schedule_type": "cron", \
	    "cron_expr": "0 8 * * *", \
	    "recipients": ["admin@example.com"], \
	    "static_payload": {"env": "production"} \
	  }' | python3 -m json.tool

.PHONY: help
help:
	@echo "NotifyHub Service — available targets:"
	@echo "  build               Build binary"
	@echo "  dev                 Run from source with .env"
	@echo "  test                Run all tests with race detector"
	@echo "  coverage            Generate HTML coverage report"
	@echo "  fmt                 Format code"
	@echo "  vet                 Run go vet"
	@echo "  tidy                Tidy and verify go.mod"
	@echo "  docker-up           Start MySQL + service with Docker Compose"
	@echo "  docker-down         Stop all containers"
	@echo "  docker-logs         Tail service logs"
	@echo "  migrate             Apply SQL migrations"
	@echo "  seed                Seed example data via API"
	@echo "  api-health          Check /health endpoint"
	@echo "  api-jobs            List scheduled jobs"
	@echo "  api-scheduler-status  Show scheduler state"
