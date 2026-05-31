#!/usr/bin/env bash
#
# deploy.sh - Builds and deploys the MyKanban application to a VPS via SSH/SCP.
#
# Usage: ./deploy.sh <user@vps-ip>
# Example: ./deploy.sh root@192.168.1.100
#
set -euo pipefail

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <user@vps-ip>"
    echo "Example: $0 root@192.168.1.100"
    exit 1
fi

VPS_TARGET=$1
REMOTE_CHITTA_DIR="/apps/chitta"
REMOTE_MAYA_DIR="/apps/maya"

log_info() {
    printf "\033[1;34m[INFO]\033[0m %s\n" "$1"
}

log_success() {
    printf "\033[1;32m[SUCCESS]\033[0m %s\n" "$1"
}

# --- 1. Build Backend ---
log_info "Building backend binary (chitta)..."
cd backend
if [ -f "scripts/build.sh" ]; then
    bash scripts/build.sh
else
    CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/tracker-server .
fi
cd ..
log_success "Backend built successfully."

# --- 2. Package Frontend (Static Export) ---
log_info "Building frontend (maya) as a static site..."
npm install
cd frontend
npm run build
cd ..
log_success "Frontend built successfully."

# --- 3. Deploy to VPS ---
log_info "Ensuring remote directories exist..."
ssh "$VPS_TARGET" "mkdir -p $REMOTE_CHITTA_DIR/bin $REMOTE_MAYA_DIR"

log_info "Transferring backend binary to $REMOTE_CHITTA_DIR..."
scp backend/bin/tracker-server "$VPS_TARGET:$REMOTE_CHITTA_DIR/bin/tracker-server"
if [ -f "backend/config.example.json" ]; then
    scp backend/config.example.json "$VPS_TARGET:$REMOTE_CHITTA_DIR/"
fi

log_info "Transferring frontend static files to $REMOTE_MAYA_DIR..."
scp -r frontend/out/* "$VPS_TARGET:$REMOTE_MAYA_DIR/"

log_success "Deployment to $VPS_TARGET completed!"
echo "---------------------------------------------------------"
echo "On your VPS, you can now start your backend service:"
echo "Backend: cd $REMOTE_CHITTA_DIR && ./bin/tracker-server"
echo "Your frontend is now served from $REMOTE_MAYA_DIR"
echo "---------------------------------------------------------"
