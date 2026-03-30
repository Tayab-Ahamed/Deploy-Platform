FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 go build -o /app .

FROM alpine:3.19
COPY --from=builder /app /app
EXPOSE 8080
ENV APP_NAME=hello-world
CMD ["/app"]
