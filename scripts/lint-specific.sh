#!/bin/bash

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo "Error: golangci-lint is not installed"
    echo "Please run ./scripts/install-linters.sh first"
    exit 1
fi

# Parse command line arguments
VERBOSE=""
FIX=""
LINTER=""

while getopts "vfl:" opt; do
    case $opt in
        v) VERBOSE="-v" ;;
        f) FIX="--fix" ;;
        l) LINTER="$OPTARG" ;;
        *) echo "Usage: $0 [-v] [-f] -l <linter>"
           echo "  -v: Enable verbose output"
           echo "  -f: Enable auto-fix"
           echo "  -l: Specify linter to run"
           exit 1 ;;
    esac
done

# Check if a linter was specified
if [ -z "$LINTER" ]; then
    echo "Error: No linter specified"
    echo "Usage: $0 [-v] [-f] -l <linter>"
    echo "Available linters:"
    echo "  - gofmt"
    echo "  - golint"
    echo "  - govet"
    echo "  - errcheck"
    echo "  - staticcheck"
    echo "  - gosimple"
    echo "  - ineffassign"
    echo "  - unused"
    echo "  - misspell"
    echo "  - gosec"
    echo "  - goconst"
    echo "  - gocyclo"
    echo "  - dupl"
    echo "  - gocritic"
    echo "  - revive"
    echo "  - unconvert"
    echo "  - unparam"
    echo "  - whitespace"
    echo "  - prealloc"
    echo "  - gocognit"
    echo "  - interfacer"
    echo "  - maligned"
    echo "  - scopelint"
    echo "  - stylecheck"
    echo "  - thelper"
    echo "  - tparallel"
    echo "  - wsl"
    exit 1
fi

# Run linter with specific linter
echo "Running linter with specific linter: $LINTER"
golangci-lint run --disable-all -E "$LINTER" $FIX $VERBOSE 