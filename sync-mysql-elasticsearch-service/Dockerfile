FROM golang:1.26-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /out/sync-mysql-elasticsearch ./cmd

FROM debian:bookworm-slim
RUN apt-get update && \
    apt-get install -y --no-install-recommends tini default-mysql-client ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/sync-mysql-elasticsearch /usr/local/bin/sync-mysql-elasticsearch

RUN useradd -u 10001 -M app && mkdir -p /var/lib/river && chown app /var/lib/river
USER 10001
ENTRYPOINT ["/usr/bin/tini", "--", "/usr/local/bin/sync-mysql-elasticsearch"]
