#!/bin/bash

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo "Error: golangci-lint is not installed"
    echo "Please run ./scripts/install-linters.sh first"
    exit 1
fi

# Parse command line arguments
VERBOSE=""
while getopts "v" opt; do
    case $opt in
        v) VERBOSE="-v" ;;
        *) echo "Usage: $0 [-v]"
           echo "  -v: Enable verbose output"
           exit 1 ;;
    esac
done

# Run linter
echo "Running linter..."
golangci-lint run $VERBOSE 