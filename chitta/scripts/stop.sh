#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

# Read port from config.json if jq is available
PORT=8080
if command -v jq &> /dev/null && [ -f "config.json" ]; then
    PORT=$(jq -r '.port // 8080' config.json)
fi

echo "==> Attempting to stop chitta on port $PORT..."

# Try to find PID by port
PID=$(lsof -ti ":$PORT" 2>/dev/null || true)
if [ -n "$PID" ]; then
    echo "    Found PID $PID on port $PORT. Sending SIGTERM..."
    kill "$PID" 2>/dev/null || true
    sleep 2
    # Check if still running
    if kill -0 "$PID" 2>/dev/null; then
        echo "    Process still running. Sending SIGKILL..."
        kill -9 "$PID" 2>/dev/null || true
    fi
    echo "    Server stopped."
else
    # Fallback: pkill
    if pkill chitta 2>/dev/null; then
        echo "    Server stopped via pkill."
    else
        echo "    No running chitta process found."
    fi
fi
