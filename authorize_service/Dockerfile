FROM golang
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o /out/main ./cmd/main.go
EXPOSE 1234
ENTRYPOINT ["/out/main"]