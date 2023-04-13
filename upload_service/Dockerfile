FROM golang
WORKDIR /app
COPY . .
ENV PORT 1985
EXPOSE 1985
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o /out/main ./cmd/main.go
ENTRYPOINT ["/out/main"]