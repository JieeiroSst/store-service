FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o account-transaction-service .

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/account-transaction-service .

# .env is not baked into the image (it only carries Consul bootstrap vars,
# not secrets, but env config is still runtime-provided) — mount/pass it via
# `docker run --env-file .env` or an orchestrator secret at /app/.env.

# Actual ports are resolved at runtime from config (Consul), these are the
# conventional gRPC / HTTP-gateway defaults used across services in this repo.
EXPOSE 50051 8080

ENTRYPOINT ["./account-transaction-service"]
CMD ["api"]
