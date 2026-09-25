FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o movie-recommendation-service ./cmd

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

RUN adduser -D -u 10001 app
USER app

COPY --from=builder /app/movie-recommendation-service .

EXPOSE 8080

ENTRYPOINT ["./movie-recommendation-service"]
