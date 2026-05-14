# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

# Final stage
FROM alpine:latest

WORKDIR /root/

# Install runtime dependencies for SQLite
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/main .
COPY --from=builder /app/.env.example .env

EXPOSE 7000

CMD ["./main"]
