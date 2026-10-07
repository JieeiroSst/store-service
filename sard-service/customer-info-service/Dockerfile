FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o customer-info-service .

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata tesseract-ocr tesseract-ocr-data-vie tesseract-ocr-data-eng \
    && adduser -D -u 10001 app

COPY --from=builder /app/customer-info-service .

USER app

ENV PORT=8081
EXPOSE 8081

ENTRYPOINT ["./customer-info-service"]
