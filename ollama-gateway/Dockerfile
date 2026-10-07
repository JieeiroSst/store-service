FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o ollama-gateway ./cmd

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 app

COPY --from=builder /app/ollama-gateway .
COPY --chown=app presets ./presets

USER app

EXPOSE 8080

ENTRYPOINT ["./ollama-gateway"]
CMD ["-addr", ":8080", "-presets", "./presets"]
