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

# Check if air is installed
if ! command -v air &> /dev/null; then
    echo "Error: air is not installed"
    echo "Please install it using: go install github.com/cosmtrek/air@latest"
    exit 1
fi

# Install dependencies
echo "Installing dependencies..."
go mod download

# Run the server with hot reload
echo "Starting server in development mode with hot reload..."
air 