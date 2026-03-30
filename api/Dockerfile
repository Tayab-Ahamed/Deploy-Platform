# ── Build stage ──────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /src
# go.mod only — no go.sum needed when there are zero external deps
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /deploy-platform-api .

# ── Runtime stage ─────────────────────────────────────────────────────
FROM alpine:3.19

# docker CLI so the API can run `docker build` and `docker push`
RUN apk add --no-cache docker-cli ca-certificates

COPY --from=builder /deploy-platform-api /usr/local/bin/deploy-platform-api

EXPOSE 8080
ENTRYPOINT ["deploy-platform-api"]
