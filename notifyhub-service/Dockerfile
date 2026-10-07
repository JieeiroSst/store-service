FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o notifyhub-service ./cmd/server

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 app

COPY --from=builder /app/notifyhub-service .

USER app

ENV PORT=8095
EXPOSE 8095

ENTRYPOINT ["./notifyhub-service"]
