FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-s -w -X github.com/JIeeiroSst/networking-service/config.Version=${VERSION}" \
    -o networking-service ./cmd

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 app \
    && mkdir -p /app/data && chown app /app/data

COPY --from=builder /app/networking-service .

USER app

# Consul's ports: HTTP API + UI on 8500, DNS on 8600 (UDP and TCP).
ENV PORT_HTTP_SERVER=8500 \
    PORT_DNS_SERVER=8600 \
    DATA_DIR=/app/data
EXPOSE 8500 8600/tcp 8600/udp
VOLUME ["/app/data"]

ENTRYPOINT ["./networking-service"]
