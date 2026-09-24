FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o hospital-patient-management-service ./cmd

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

RUN adduser -D -u 10001 app
USER app

COPY --from=builder /app/hospital-patient-management-service .

# 8080: REST gateway + /health, 9090: gRPC
EXPOSE 8080 9090

ENTRYPOINT ["./hospital-patient-management-service"]
