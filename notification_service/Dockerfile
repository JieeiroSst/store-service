FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o notification-service ./cmd

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/notification-service .

EXPOSE 1235

ENTRYPOINT ["./notification-service"]
