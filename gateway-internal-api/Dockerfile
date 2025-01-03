FROM golang:latest as builder

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o server .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .

CMD ["./server"]