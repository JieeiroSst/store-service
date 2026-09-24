# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Cache module downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o catalogues-service ./cmd

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

RUN adduser -D -u 10001 app
USER app

COPY --from=builder /app/catalogues-service .

# 8080: REST API + /health
EXPOSE 8080

ENTRYPOINT ["./catalogues-service"]
