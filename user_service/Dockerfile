# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS builder
WORKDIR /app

# Cache module downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/user-service .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates && \
    addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=builder /out/user-service ./user-service
COPY --from=builder /app/docs ./docs
USER app

ENV PORT_SERVER=1235
ENV PORT_SERVER_GRPC=1236
EXPOSE 1235 1236

ENTRYPOINT ["./user-service", "api"]
