include .env

MIGRATE=goose -dir=migrations -allow-missing mysql "${DB_USER}:${DB_PASSWORD}@/${DB_NAME}?parseTime=true"

migrate-status:
	$(MIGRATE) status

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down 

gen:
	protoc --go_out=. --go-grpc_opt=. ./internal/proto/api.proto
run:
	go run cmd/main.go

mod:
	go mod vendor
	go mod tidy

mockey-gen:
	mockgen -source=./internal/repository/casbin.go
	mockgen -source=./internal/usecase/casbin.go
	mockgen -source=./internal/usecase/otp.go

	mockgen -destination=mocks/mock_repository.go -package=mocks github.com/JieeiroSst/authorize-service/internal/repository Casbins

generate: ## Generate proto
	@third_party/protoc/bin/protoc \
		-I proto/ \
		-I $(GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis/ \
		-I $(GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/ \
		-I $(GOPATH)/pkg/mod/github.com/gogo/protobuf@v1.3.1/protoc-gen-gogo/grpc \
		-I $(GOPATH)/pkg/mod/github.com/gogo/protobuf@v1.3.1 \
		-I $(GOPATH)/pkg/mod/github.com/mwitkow/go-proto-validators@v0.3.2/ \
		--gogo_out=plugins=grpc:pb \
		--govalidators_out=gogoimport=true:pb \
		--grpc-gateway_out=pb \
		--swagger_out=docs \
		proto/*.proto
	
.PHONY: migrate-status migrate-up migrate-down gen run