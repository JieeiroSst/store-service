.PHONY: run build test test-race vet fmt-check migrate seed lint tidy docker-up docker-down

APP_NAME = wallet-service
BUILD_DIR = ./bin
DB_URL ?= postgres://wallet:wallet@localhost:5432/wallet_db?sslmode=disable

run:
	go run ./cmd/server

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

fmt-check:
	test -z "$$(gofmt -l .)"

migrate:
	for f in migrations/*.sql; do psql "$(DB_URL)" -v ON_ERROR_STOP=1 --single-transaction -f $$f || exit 1; done

seed:
	psql "$(DB_URL)" -v ON_ERROR_STOP=1 --single-transaction -f migrations/seed/dev_seed.sql

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down
