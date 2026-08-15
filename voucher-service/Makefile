.PHONY: up down run build test test-integration tidy migrate-up migrate-down

up:
	docker-compose up -d postgres redis kafka

down:
	docker-compose down

run:
	go run ./cmd/api

build:
	go build -o bin/voucher-api ./cmd/api

test:
	go test ./...

# Requires Docker (spins up real Postgres via testcontainers-go).
test-integration:
	go test -tags=integration ./...

tidy:
	go mod tidy

migrate-up:
	migrate -path migrations -database "postgres://voucher:voucher@localhost:5432/voucher?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://voucher:voucher@localhost:5432/voucher?sslmode=disable" down
