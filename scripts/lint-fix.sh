#!/bin/bash

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo "Error: golangci-lint is not installed"
    echo "Please run ./scripts/install-linters.sh first"
    exit 1
fi

# Run linter with auto-fix
echo "Running linter with auto-fix..."
golangci-lint run --fix 