run:
	go run cmd/main.go

gen:
	protoc --go_out=. --go-grpc_out=. ./internal/proto/apis.proto

.PHONY: run gen