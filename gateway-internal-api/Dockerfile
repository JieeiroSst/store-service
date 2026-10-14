FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN if [ -d .upstream/gateway-api/graph ]; then \
      go run ./cmd/catalogsync -src .upstream/gateway-api/graph -dst catalog/schema; \
    fi
RUN CGO_ENABLED=0 go test ./catalog/...
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o gateway-internal-api .

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 app

COPY --from=builder /app/gateway-internal-api .

USER app

ENV PORT=8081 GIN_MODE=release
EXPOSE 8081

ENTRYPOINT ["./gateway-internal-api"]
