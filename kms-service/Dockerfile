FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o kms-service .

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 app \
    && mkdir -p /etc/kms && chown app /etc/kms

COPY --from=builder /app/kms-service .

USER app

ENV SERVER_PORT=8080 \
    MASTER_KEY_PATH=/etc/kms/master.key
EXPOSE 8080
VOLUME ["/etc/kms"]

ENTRYPOINT ["./kms-service"]
