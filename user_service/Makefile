run:
	go run cmd/main.go
gen-client:
	protoc --go_out=. --go-grpc_out=. ./proto/api.proto
gen-server:
	protoc --go_out=. --go-grpc_out=. ./internal/proto/api.proto
swag:
	 swag init -o docs  -g internal/delivery/http/v1/*.go
mode:
	go mod tidy 
	go mod vendor