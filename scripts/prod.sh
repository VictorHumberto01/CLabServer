#!/bin/bash

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Check if the binary exists
if [ ! -f bin/server ]; then
    echo "Error: Server binary not found. Please run ./scripts/build.sh first."
    exit 1
fi

# Run the server
echo "Starting server in production mode..."
./bin/server 