FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o video-service ./cmd

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata ffmpeg \
    && addgroup -S app && adduser -S app -G app

COPY --from=builder /app/video-service .

USER app
# ffmpeg is only used when ROLE=worker (transcoding); the API ignores it.
EXPOSE 8080

ENTRYPOINT ["./video-service"]
