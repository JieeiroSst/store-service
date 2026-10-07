FROM golang:1.21-alpine3.19 AS builder

WORKDIR /app

RUN apk add --no-cache build-base pkgconf librdkafka-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -tags dynamic -trimpath -ldflags="-s -w" -o message-service ./cmd

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata librdkafka \
    && adduser -D -u 10001 app

COPY --from=builder /app/message-service .

USER app

EXPOSE 8080

ENTRYPOINT ["./message-service"]
