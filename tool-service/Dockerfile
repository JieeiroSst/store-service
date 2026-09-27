FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o tool-service .

FROM chromedp/headless-shell:latest
WORKDIR /app
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates tzdata fonts-noto-core fonts-liberation adb \
 && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/tool-service .
ENV CHROME_PATH=/headless-shell/headless-shell \
    CHROME_NO_SANDBOX=true
EXPOSE 8080
ENTRYPOINT ["./tool-service"]
