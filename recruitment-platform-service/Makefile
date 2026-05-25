.PHONY: build run test lint migrate migrate-down docker-up docker-down tidy

# ─── Build ────────────────────────────────────────────────────────────────────
build:
	go build -ldflags="-s -w" -o bin/recruitment-platform ./cmd/server

run:
	CONFIG_PATH=config/config.yaml go run ./cmd/server

# ─── Test ─────────────────────────────────────────────────────────────────────
test:
	go test ./... -v -race -count=1

test-domain:
	go test ./internal/domain/... -v -race

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ─── Code quality ─────────────────────────────────────────────────────────────
lint:
	golangci-lint run ./...

tidy:
	go mod tidy

generate:
	go generate ./...

# ─── Database ─────────────────────────────────────────────────────────────────
DB_URL=postgres://postgres:secret@localhost:5432/recruitment_db?sslmode=disable

migrate:
	goose -dir migrations postgres "$(DB_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DB_URL)" down

migrate-status:
	goose -dir migrations postgres "$(DB_URL)" status

migrate-reset:
	goose -dir migrations postgres "$(DB_URL)" reset

# ─── Docker ───────────────────────────────────────────────────────────────────
docker-up:
	docker compose up -d postgres temporal temporal-ui
	@echo "Waiting for services..."
	@sleep 3
	$(MAKE) migrate

docker-down:
	docker compose down -v

docker-build:
	docker compose build

# ─── Temporal ─────────────────────────────────────────────────────────────────
# Register the namespace (run once after temporal starts)
temporal-namespace:
	temporal operator namespace create --namespace recruitment

# ─── Setup ────────────────────────────────────────────────────────────────────
setup: docker-up
	@echo "✅ Infrastructure ready. Run 'make run' to start the server."

# ─── Generate mock for testing ────────────────────────────────────────────────
mocks:
	mockgen -source=internal/domain/candidate/repository.go \
		-destination=internal/mock/candidate_repo_mock.go \
		-package=mock
	mockgen -source=internal/port/ports.go \
		-destination=internal/mock/ports_mock.go \
		-package=mock
