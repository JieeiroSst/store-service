.PHONY: run dev deps tidy docker-up docker-down test lint

## Run the service (requires .env file)
run:
	go run ./cmd/server/main.go

## Start local DynamoDB + create tables, then run service
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

## Spin up local DynamoDB via Docker Compose
docker-up:
	docker compose up -d

## Tear down local DynamoDB
docker-down:
	docker compose down -v

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
