#!/usr/bin/env bash
#
# deploy.sh - Builds and deploys the MyKanban application to a VPS via SSH/SCP.
#
# Usage: ./deploy.sh [user@vps-ip]
#   - If no target is provided, it defaults to 'ravi@ravinath-prod'.
#
# Example:
#   ./deploy.sh                    # Deploys to the default target
#   ./deploy.sh other@server       # Deploys to a different server
#
set -euo pipefail

# --- Configuration ---
DEFAULT_VPS_TARGET="ravi@ravinath-prod"
REMOTE_CHITTA_DIR="/apps/chitta"
REMOTE_MAYA_DIR="/apps/maya"

# Use the provided argument or fall back to the default
VPS_TARGET="${1:-$DEFAULT_VPS_TARGET}"

log_info() {
    printf "\033[1;34m[INFO]\033[0m %s\n" "$1"
}

log_success() {
    printf "\033[1;32m[SUCCESS]\033[0m %s\n" "$1"
}

# --- 1. Build Backend ---
log_info "Building backend binary (chitta)..."
cd chitta
if [ -f "scripts/build.sh" ]; then
    bash scripts/build.sh
else
    # Ensure dependencies are downloaded and docs are generated
    go mod tidy
    # swag init # Uncomment this if you are using swaggo/swag for docs
    CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/chitta .
fi
cd ..
log_success "Backend built successfully."

# --- 2. Package Frontend (Static Export) ---
log_info "Building frontend (maya) as a static site..."
npm install
cd maya
npm run build
cd ..
log_success "Frontend built successfully."

# --- 3. Deploy to VPS ---
log_info "Deploying to target: $VPS_TARGET"
log_info "Ensuring remote directories exist..."
ssh "$VPS_TARGET" "mkdir -p $REMOTE_CHITTA_DIR/bin $REMOTE_MAYA_DIR"

log_info "Transferring backend binary to $REMOTE_CHITTA_DIR..."
scp chitta/bin/chitta "$VPS_TARGET:$REMOTE_CHITTA_DIR/bin/chitta"
if [ -f "chitta/config.example.json" ]; then
    scp chitta/config.example.json "$VPS_TARGET:$REMOTE_CHITTA_DIR/"
fi

log_info "Transferring frontend static files to $REMOTE_MAYA_DIR..."
scp -r maya/dist/* "$VPS_TARGET:$REMOTE_MAYA_DIR/"


log_success "Deployment to $VPS_TARGET completed!"
echo "---------------------------------------------------------"
echo "On your VPS, you can now start your backend service:"
echo "Backend: cd $REMOTE_CHITTA_DIR && ./bin/chitta"
echo "Your frontend is now served from $REMOTE_MAYA_DIR"
echo "---------------------------------------------------------"
