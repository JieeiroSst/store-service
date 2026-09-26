# Both tools in one small image: `/flash` (the load generator) and `/fakeus` (a user-service stand-in).
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY loadtest ./loadtest
RUN CGO_ENABLED=0 go build -o /flash ./loadtest/flash && CGO_ENABLED=0 go build -o /fakeus ./loadtest/fakeus

FROM alpine:3.19
COPY --from=build /flash /flash
COPY --from=build /fakeus /fakeus
EXPOSE 18999
ENTRYPOINT ["/flash"]
