FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o reservation-service ./cmd

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

RUN adduser -D -u 10001 app
USER app

COPY --from=builder /app/reservation-service .

EXPOSE 8080

ENTRYPOINT ["./reservation-service"]
