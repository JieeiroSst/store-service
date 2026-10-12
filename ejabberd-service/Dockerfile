FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-s -w -X github.com/JIeeiroSst/ejabberd-service/internal/config.Version=${VERSION}" \
    -o ejabberd-service ./cmd/ejabberd-service

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 app

COPY --from=builder /app/ejabberd-service .
COPY conf/ejabberd.yml /etc/ejabberd/ejabberd.yml

USER app

ENV EJABBERD_CONFIG=/etc/ejabberd/ejabberd.yml
EXPOSE 5222 5223 5280

ENTRYPOINT ["./ejabberd-service"]
