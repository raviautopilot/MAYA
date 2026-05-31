#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

PORT=8080
if command -v jq &> /dev/null && [ -f "config.json" ]; then
    PORT=$(jq -r '.port // 8080' config.json)
fi

echo "===== MyKanban Troubleshoot ====="
echo ""

echo "1. Go version:"
go version 2>/dev/null || echo "   [ERROR] Go not found in PATH"
echo ""

echo "2. Go environment:"
go env GOROOT GOPATH GOOS GOARCH 2>/dev/null || echo "   [ERROR] Cannot read Go env"
echo ""

echo "3. Port $PORT status:"
lsof -i ":$PORT" 2>/dev/null || echo "   Port $PORT is free (no process listening)"
echo ""

echo "4. Validating config.json..."
if [ -f "config.json" ]; then
    if command -v jq &> /dev/null; then
        jq . config.json > /dev/null 2>&1 && echo "   config.json: VALID" || echo "   config.json: INVALID JSON"
    elif command -v python3 &> /dev/null; then
        python3 -m json.tool config.json > /dev/null 2>&1 && echo "   config.json: VALID" || echo "   config.json: INVALID JSON"
    else
        echo "   [WARN] No JSON validator (jq/python3) available"
    fi
else
    echo "   [WARN] config.json not found"
fi
echo ""

echo "5. Validating storage JSON files..."
if [ -d "storage" ]; then
    for f in storage/*.json; do
        if [ -f "$f" ]; then
            if command -v jq &> /dev/null; then
                jq . "$f" > /dev/null 2>&1 && echo "   $f: VALID" || echo "   $f: INVALID JSON"
            elif command -v python3 &> /dev/null; then
                python3 -m json.tool "$f" > /dev/null 2>&1 && echo "   $f: VALID" || echo "   $f: INVALID JSON"
            fi
        fi
    done
else
    echo "   [INFO] storage/ directory not found (will be created on first run)"
fi
echo ""

echo "6. Server logs (last 20 lines):"
LOG_FILE="server.log"
if command -v jq &> /dev/null && [ -f "config.json" ]; then
    LOG_FILE=$(jq -r '.log_file // "server.log"' config.json)
fi
if [ -f "$LOG_FILE" ]; then
    tail -20 "$LOG_FILE"
else
    echo "   No log file found at $LOG_FILE"
fi

echo ""
echo "===== Troubleshoot Complete ====="
