FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-s -w -X github.com/JIeeiroSst/lucky-service/internal/config.Version=${VERSION}" \
    -o lucky-service ./cmd

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 app

COPY --from=builder /app/lucky-service .

USER app

ENV HTTP_PORT=9000
EXPOSE 9000

ENTRYPOINT ["./lucky-service"]
