.PHONY: run dev deps tidy docker-up docker-down test lint migrate-up migrate-down

MYSQL_HOST ?= localhost
MYSQL_PORT ?= 3306
MYSQL_USER ?= root
MYSQL_PASSWORD ?= root
MYSQL_DATABASE ?= referral_service
MIGRATE_DSN := mysql://$(MYSQL_USER):$(MYSQL_PASSWORD)@tcp($(MYSQL_HOST):$(MYSQL_PORT))/$(MYSQL_DATABASE)

## Run the service (requires .env file)
run:
	go run ./cmd/server/main.go

## Start local MySQL + apply migrations, then run service
dev: docker-up
	sleep 3
	go run ./cmd/server/main.go

## Download all Go dependencies
deps:
	go mod download
	go mod verify

## Tidy go.mod / go.sum
tidy:
	go mod tidy

## Spin up local MySQL via Docker Compose (auto-applies migrations via the `migrate` service)
docker-up:
	docker compose up -d

## Tear down local MySQL
docker-down:
	docker compose down -v

## Apply migrations against MYSQL_* env vars (requires: go install github.com/golang-migrate/migrate/v4/cmd/migrate)
migrate-up:
	migrate -path migrations -database "$(MIGRATE_DSN)" up

## Roll back the last migration against MYSQL_* env vars
migrate-down:
	migrate -path migrations -database "$(MIGRATE_DSN)" down 1

## Run tests
test:
	go test ./... -v -count=1

## Lint (requires golangci-lint)
lint:
	golangci-lint run ./...

## Build binary
build:
	go build -o bin/referral-service ./cmd/server/main.go

## Example: generate a referral link
curl-generate:
	curl -s -X POST http://localhost:8080/api/v1/referral/generate \
	  -H "Content-Type: application/json" \
	  -d '{"owner_user_id":"user-123","channel":"copy","platform":"ios"}' | jq .

## Example: track a click event
curl-track:
	curl -s -X POST http://localhost:8080/api/v1/referral/event \
	  -H "Content-Type: application/json" \
	  -d '{"ref_code":"$(REF)","event_type":"link_clicked","platform":"ios","device_id":"dev-abc"}' | jq .

## Example: confirm install
curl-confirm:
	curl -s -X POST http://localhost:8080/api/v1/referral/confirm-install \
	  -H "Content-Type: application/json" \
	  -d '{"ref_code":"$(REF)","new_user_id":"user-456","platform":"ios","device_id":"dev-abc"}' | jq .
