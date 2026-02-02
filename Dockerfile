# Build stage
FROM golang:1.23-bookworm AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application

RUN CGO_ENABLED=1 GOOS=linux go build -o server ./cmd/server/main.go

# Final stage
FROM debian:bookworm-slim

WORKDIR /app

# Install runtime dependencies: GCC and Firejail
RUN apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    firejail \
    && rm -rf /var/lib/apt/lists/*

# Copy binary from builder
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
