FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o tool-service .

FROM chromedp/headless-shell:157.0.8103.15
WORKDIR /app
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates tzdata fonts-noto-core fonts-liberation adb \
 && rm -rf /var/lib/apt/lists/* \
 && useradd -u 10001 -m -d /home/app app \
 && mkdir -p /app/data && chown 10001:10001 /app/data
COPY --from=builder /app/tool-service .
ENV CHROME_PATH=/headless-shell/headless-shell \
    CHROME_NO_SANDBOX=true
EXPOSE 8080
USER 10001
ENTRYPOINT ["./tool-service"]
