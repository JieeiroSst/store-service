FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o real-time-service ./cmd

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 app

COPY --from=builder /app/real-time-service .

USER app

EXPOSE 8080

ENTRYPOINT ["./real-time-service"]
