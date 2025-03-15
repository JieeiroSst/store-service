api:
	go run main.go api
gen-client:
	protoc --go_out=. --go-grpc_out=. ./proto/api.proto
gen-server:
	protoc --go_out=. --go-grpc_out=. ./internal/proto/api.proto
swag:
	 swag init -o docs  -g internal/delivery/http/v1/*.go
mode:
	go mod tidy 
	go mod vendor

image_grpc:
	docker build -t user_service_grpc:${version} -f Dockerfile.grpc .
image_grpc_tag:
	docker tag user_service_grpc:${version}  1760151/user_service_grpc:${version} 
image_grpc_push:
	docker push 1760151/user_service_grpc:${version}

image_api:
	docker build -t  user_service_api:${version} -f Dockerfile.api .
image_api_tag:
	docker tag user_service_api:${version}  1760151/user_service_api:${version} 
image_api_push:
	docker push 1760151/user_service_api:${version}