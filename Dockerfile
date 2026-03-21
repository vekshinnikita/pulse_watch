FROM golang:1.24 AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y gcc libc6-dev librdkafka-dev pkg-config

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o pulse_watch ./cmd/pulse_watch && \
    go build -o pulse_watch_background ./cmd/pulse_watch_background && \
    go build -o pulse_watch_queues ./cmd/pulse_watch_queues

FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y librdkafka1 ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /root/

COPY --from=builder /app/configs ./configs
COPY --from=builder /app/.env ./

ENV CONFIG_PATH=/root/configs/config.yaml

COPY --from=builder /app/pulse_watch .
COPY --from=builder /app/pulse_watch_background .
COPY --from=builder /app/pulse_watch_queues .