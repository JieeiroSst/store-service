FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o authorize-service .

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/authorize-service .
# Casbin loads its RBAC model from this relative path at runtime
# (common.RBACModelPath), so it must ship alongside the binary.
COPY --from=builder /app/config/conf ./config/conf

# .env only carries Consul bootstrap vars (HostConsul/KeyConsul/ServiceConsul),
# not secrets — actual config is resolved at runtime from Consul, falling
# back to environment variables (see config/config.go).
EXPOSE 8080 9090

ENTRYPOINT ["./authorize-service"]
CMD ["api"]
