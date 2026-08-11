FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Three binaries from one image: livestream-service (monolithic, dev/
# docker-compose convenience), and the two independently-scalable roles
# the Helm chart actually deploys - livestream-service-node (SRS sidecar +
# ffmpeg supervisor) and livestream-service-edge (stateless rooms/chat/
# viewer API). Building all three here avoids doubling the CI/CD pipeline
# for what's otherwise identical code with different fx wiring.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o livestream-service ./cmd && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o livestream-service-node ./cmd/node && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o livestream-service-edge ./cmd/edge

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata ffmpeg

COPY --from=builder /app/livestream-service .
COPY --from=builder /app/livestream-service-node .
COPY --from=builder /app/livestream-service-edge .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

ENTRYPOINT ["./livestream-service"]
