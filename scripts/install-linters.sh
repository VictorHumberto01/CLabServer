#!/bin/bash

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    exit 1
fi

# Install golangci-lint
echo "Installing golangci-lint..."
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2

# Install required linters
echo "Installing required linters..."
go install github.com/mibk/dupl@latest
go install github.com/go-critic/go-critic/cmd/gocritic@latest
go install github.com/mgechev/revive@latest
go install github.com/mdempsky/unconvert@latest
go install github.com/mvdan/unparam@latest
go install github.com/ultraware/whitespace/cmd/whitespace@latest
go install github.com/alexkohler/prealloc@latest
go install github.com/uudashr/gocognit/cmd/gocognit@latest
go install github.com/mbilski/exhaustivestruct/cmd/exhaustivestruct@latest
go install github.com/mdempsky/maligned@latest
go install github.com/kyoh86/scopelint@latest
go install github.com/ssgreg/nlreturn/v2/cmd/nlreturn@latest
go install github.com/tdakkota/asciicheck/cmd/asciicheck@latest
go install github.com/tomarrell/wrapcheck/v2/cmd/wrapcheck@latest
go install github.com/charithe/durationcheck/cmd/durationcheck@latest
go install github.com/leonklingele/grouper/cmd/grouper@latest
go install github.com/butuzov/ireturn/cmd/ireturn@latest
go install github.com/butuzov/mirror/cmd/mirror@latest
go install github.com/daixiang0/gci/cmd/gci@latest
go install github.com/denis-tingaikin/go-header/cmd/go-header@latest
go install github.com/kkHAIKE/contextcheck/cmd/contextcheck@latest
go install github.com/kyoh86/exportloopref/cmd/exportloopref@latest
go install github.com/ldez/gomoddirectives/cmd/gomoddirectives@latest
go install github.com/ldez/tagliatelle/cmd/tagliatelle@latest
go install github.com/loov/loov/cmd/loov@latest
go install github.com/maratori/testpackage/cmd/testpackage@latest
go install github.com/matoous/godox/cmd/godox@latest
go install github.com/mgechev/dots/cmd/dots@latest
go install github.com/mitchellh/go-ps/cmd/ps@latest
go install github.com/moricho/tparallel/cmd/tparallel@latest
go install github.com/nakabonne/nestif/cmd/nestif@latest
go install github.com/nishanths/exhaustive/cmd/exhaustive@latest
go install github.com/nishanths/predeclared/cmd/predeclared@latest
go install github.com/polyfloyd/go-errorlint/cmd/go-errorlint@latest

echo "Linters installed successfully!" 