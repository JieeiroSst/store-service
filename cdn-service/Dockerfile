FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o cdn-service ./cmd

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/cdn-service .

# gRPC, grpc-gateway (REST), and Gin (static/media) servers respectively.
EXPOSE 2231 2234 2232

ENTRYPOINT ["./cdn-service"]
