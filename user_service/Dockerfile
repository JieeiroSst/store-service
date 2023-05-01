FROM golang
WORKDIR /app
COPY . .
ENV PORT 1235
ENV PORT 1236
EXPOSE 1235 1236
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o /out/main ./cmd/main.go
ENTRYPOINT ["/out/main"]