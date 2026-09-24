# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git

# Cache module downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o car-rental-service .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app
WORKDIR /app

COPY --from=builder /app/car-rental-service .
# The service does not create its own tables: the Helm chart applies this
# idempotent schema from an init container before the server starts.
COPY database.sql ./database.sql
USER app

# HTTP gateway and gRPC ports come from Consul (see consul.json).
EXPOSE 2222 2223

ENTRYPOINT ["./car-rental-service", "api"]
