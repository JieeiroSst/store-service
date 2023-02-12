run:
	go run cmd/main.go

mod:
	go mod tidy
gen:
	protoc --go_out=. --go-grpc_out=. ./internal/proto/apis.proto

.PHONY: run gen