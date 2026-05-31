#!/usr/bin/env bash
#
# deploy.sh - Builds and deploys the MyKanban application to a VPS via SSH/SCP.
#
# Usage: ./deploy.sh <user@vps-ip> <remote-apps-folder>
# Example: ./deploy.sh root@192.168.1.100 /var/www/mykanban
#

set -euo pipefail

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <user@vps-ip> <remote-apps-folder>"
    echo "Example: $0 root@192.168.1.100 /var/www/mykanban"
    exit 1
fi

VPS_TARGET=$1
REMOTE_DIR=$2

log_info() {
    printf "\033[1;34m[INFO]\033[0m %s\n" "$1"
}

log_success() {
    printf "\033[1;32m[SUCCESS]\033[0m %s\n" "$1"
}

log_error() {
    printf "\033[1;31m[ERROR]\033[0m %s\n" "$1"
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

# --- 2. Package Frontend ---
log_info "Building frontend (maya)..."
# Assuming root package.json handles workspaces
npm install
cd frontend
npm run build

log_info "Packaging frontend files..."
# Create a tarball containing the built files and necessary configs
tar -czf frontend-release.tar.gz .next public package.json next.config.mjs
cd ..
log_success "Frontend packaged successfully."

# --- 3. Deploy to VPS ---
log_info "Ensuring remote directory exists: $REMOTE_DIR..."
ssh "$VPS_TARGET" "mkdir -p $REMOTE_DIR/backend/bin $REMOTE_DIR/frontend"

log_info "Transferring backend binary..."
scp backend/bin/tracker-server "$VPS_TARGET:$REMOTE_DIR/backend/bin/tracker-server"
if [ -f "backend/config.example.json" ]; then
    scp backend/config.example.json "$VPS_TARGET:$REMOTE_DIR/backend/"
fi

log_info "Transferring frontend package..."
scp frontend/frontend-release.tar.gz "$VPS_TARGET:$REMOTE_DIR/frontend/"
# Copy the root package-lock.json just in case it's needed for strict installs
scp package-lock.json "$VPS_TARGET:$REMOTE_DIR/frontend/"

log_info "Extracting frontend on VPS and installing production dependencies..."
ssh "$VPS_TARGET" "cd $REMOTE_DIR/frontend && tar -xzf frontend-release.tar.gz && rm frontend-release.tar.gz && npm install --omit=dev"

# Clean up local tarball
rm frontend/frontend-release.tar.gz

log_success "Deployment to $VPS_TARGET:$REMOTE_DIR completed!"
echo "---------------------------------------------------------"
echo "On your VPS, you can now start your services:"
echo "Backend: cd $REMOTE_DIR/backend && ./bin/tracker-server"
echo "Frontend: cd $REMOTE_DIR/frontend && npm start"
echo "---------------------------------------------------------"
