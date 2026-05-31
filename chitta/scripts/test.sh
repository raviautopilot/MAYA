#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

echo "==> Running unit tests with coverage..."
go test ./... -coverprofile=coverage.out -v

echo ""
echo "==> Coverage summary:"
go tool cover -func=coverage.out | tail -1

echo ""
echo "==> Generating HTML coverage report..."
go tool cover -html=coverage.out -o coverage.html
echo "    Report saved to coverage.html"
