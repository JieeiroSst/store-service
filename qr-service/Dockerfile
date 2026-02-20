FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o qr-service ./cmd/main.go

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/qr-service .
COPY --from=builder /app/config.yaml .

EXPOSE 8080

CMD ["./qr-service"]
