gen:
	protoc -I./googleapis -I. \
	--go_out=gateway --go_opt=paths=source_relative \
	--go-grpc_out=gateway --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=gateway --grpc-gateway_opt=paths=source_relative \
	proto/*.proto

api:
	go run main.go api

cron:
	go run main.go cron

consumer:
	go run main.go consumer

subscriber:
	go run main.go subscriber

gen_proxy:
	protoc -I./googleapis -I. \
	--go_out=./internal/proxy/${folder}/gateway --go_opt=paths=source_relative \
	--go-grpc_out=./internal/proxy/${folder}/gateway --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=./internal/proxy/${folder}/gateway --grpc-gateway_opt=paths=source_relative \
	./internal/proxy/${folder}/proto/*.proto
