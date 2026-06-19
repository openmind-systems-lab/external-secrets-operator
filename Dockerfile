FROM golang:1.22-alpine AS build
WORKDIR /app
COPY main.go .
RUN go mod init eso-logger && go build -o eso-logger main.go

FROM alpine:3.20
COPY --from=build /app/eso-logger /eso-logger
CMD ["/eso-logger"]
