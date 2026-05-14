#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

# Check config.json
if [ ! -f "config.json" ]; then
    if [ -f "config.example.json" ]; then
        cp config.example.json config.json
        echo "[INFO] config.json created from config.example.json."
        echo "[ACTION] Please edit config.json with your settings, then re-run this script."
        exit 1
    else
        echo "[ERROR] config.json and config.example.json not found."
        exit 1
    fi
fi

# Build if binary missing
if [ ! -f "bin/tracker-server" ]; then
    echo "==> Binary not found, building..."
    bash scripts/build.sh
fi

echo "==> Starting MyKanban server..."
exec ./bin/tracker-server
