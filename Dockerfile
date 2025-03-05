# -------- STAGE 1: BUILD --------
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN go build -o server ./cmd/main.go

FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    chromium \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/server /app/server

COPY --from=builder /app/templates /app/templates


EXPOSE 8080

ENTRYPOINT ["/app/server"]