run:
	go run cmd/main.go
gen:
	protoc --go_out=. --go-grpc_out=. ./proto/api.proto
swag:
	 swag init -o docs  -g internal/delivery/http/v1/*.go