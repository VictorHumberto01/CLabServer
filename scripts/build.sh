#!/bin/bash

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    exit 1
fi

# Install dependencies
echo "Installing dependencies..."
go mod download

# Build the server
echo "Building server..."
go build -o bin/server cmd/server/main.go

echo "Build completed successfully!" 