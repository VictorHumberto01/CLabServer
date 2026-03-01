# Build stage
FROM golang:1.24-bookworm AS builder

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

# Install runtime dependencies: GCC for local compilation
# Docker CLI for spawning sandbox containers
RUN apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    ca-certificates \
    curl \
    gnupg \
    && install -m 0755 -d /etc/apt/keyrings \
    && curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg \
    && chmod a+r /etc/apt/keyrings/docker.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian bookworm stable" > /etc/apt/sources.list.d/docker.list \
    && apt-get update \
    && apt-get install -y docker-ce-cli \
    && rm -rf /var/lib/apt/lists/*

# Copy binary from builder
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
