include .env

MIGRATE=goose -dir=migrations -allow-missing mysql "${DB_USER}:${DB_PASSWORD}@/${DB_NAME}?parseTime=true"

migrate-status:
	$(MIGRATE) status

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down 

gen:
	protoc --go_out=. --go-grpc_out=. ./internal/proto/api.proto
run:
	go run cmd/main.go

.PHONY: migrate-status migrate-up migrate-down gen run