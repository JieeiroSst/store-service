FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-s -w -X github.com/JIeeiroSst/serpapi-service/config.Version=${VERSION}" \
    -o serpapi-service ./cmd

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 app

COPY --from=builder /app/serpapi-service .

USER app

ENV PORT_HTTP_SERVER=8080
EXPOSE 8080

ENTRYPOINT ["./serpapi-service"]
