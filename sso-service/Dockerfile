FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o sso-service .

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 app

COPY --from=builder /app/sso-service .

USER app

EXPOSE 8090

ENTRYPOINT ["./sso-service"]
