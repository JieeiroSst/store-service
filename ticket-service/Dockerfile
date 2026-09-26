FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ticket-service ./cmd

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

RUN adduser -D -u 10001 app
USER app

COPY --from=builder /app/ticket-service .

EXPOSE 8080

ENTRYPOINT ["./ticket-service"]
