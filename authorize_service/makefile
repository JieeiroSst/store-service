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

mod:
	go mod vendor
	go mod tidy

mockey-gen:
	mockgen -source=./internal/repository/casbin.go
	mockgen -source=./internal/usecase/casbin.go
	mockgen -source=./internal/usecase/otp.go

<<<<<<< HEAD
	mockgen -destination=mocks/mock_repository.go -package=mocks github.com/JieeiroSst/authorize-service/internal/repository Casbins
	
=======
>>>>>>> 035c5e087 (.)
.PHONY: migrate-status migrate-up migrate-down gen run