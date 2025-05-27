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

# Run the server
echo "Starting server in development mode..."
go run cmd/server/main.go 