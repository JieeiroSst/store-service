FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/starter ./cmd/starter \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/cron ./cmd/cron

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 app

COPY --from=builder /out/ ./

USER app

ENTRYPOINT ["./worker"]
