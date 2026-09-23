FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o referral-service ./cmd/server

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

# The service does not migrate its own schema: the Helm chart runs this CLI
# from an init container before the server starts.
COPY --from=migrate/migrate:v4.17.1 /usr/local/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/referral-service .
COPY migrations ./migrations

EXPOSE 8080

ENTRYPOINT ["./referral-service"]
