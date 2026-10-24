FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o wallet-service ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 app
WORKDIR /app
COPY --from=builder /app/wallet-service .
COPY --from=builder /app/configs ./configs
USER app
EXPOSE 8080
CMD ["./wallet-service"]
