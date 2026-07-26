FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o airflow-service ./cmd

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/airflow-service .

EXPOSE 8000

ENTRYPOINT ["./airflow-service"]
