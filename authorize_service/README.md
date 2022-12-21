protoc --go_out=. --go-grpc_out=. ./internal/proto/api.proto

http://127.0.0.1:8500/v1/kv/authorize