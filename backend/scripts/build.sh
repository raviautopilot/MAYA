#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

echo "==> Formatting code..."
go fmt ./...

echo "==> Running go vet..."
go vet ./...

echo "==> Generating Swagger docs..."
if command -v swag &> /dev/null; then
    swag init --parseDependency --parseInternal -g main.go
    echo "    Swagger docs generated."
else
    echo "    [WARN] swag not installed. Skipping swagger generation."
    echo "    Install: go install github.com/swaggo/swag/cmd/swag@latest"
fi

echo "==> Building binary..."
mkdir -p bin
go build -o bin/tracker-server main.go

echo "==> Build success. Binary at bin/tracker-server"
