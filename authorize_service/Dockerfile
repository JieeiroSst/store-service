FROM golang
WORKDIR /app
COPY . .
ENV PORT_HTTP_SERVER 1234
ENV PORT_GRPC_SERVER 1231
EXPOSE 1234 1231
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o /out/main ./cmd/main.go
ENTRYPOINT ["/out/main"]