# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY . .
RUN go build -o backhub

# Final stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /build/backhub /usr/local/bin/

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
