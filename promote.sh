#!/usr/bin/env bash
#
# promote.sh - Builds and promotes Docker images to a remote VPS.
#
# This script builds the chitta (backend) and maya (frontend) Docker images locally, then
# transfers them to the specified remote host using `docker save` and `ssh`.
#
set -euo pipefail

# --- Configuration ---
# Image names
CHITTA_IMAGE="mycan-chitta:latest"
MAYA_IMAGE="mycan-maya:latest"

# Source directories
BACKEND_DIR="./backend"
FRONTEND_DIR="./frontend"

# --- Main Logic ---
# Validate input
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <user@vps-ip>"
    echo "Example: $0 root@192.168.1.100"
    exit 1
fi
VPS_TARGET=$1

log_info() {
    printf "\033[1;34m[INFO]\033[0m %s\n" "$1"
}

log_success() {
    printf "\033[1;32m[SUCCESS]\033[0m %s\n" "$1"
}

# 1. Build chitta (backend) image
log_info "Building chitta (backend) image: $CHITTA_IMAGE..."
docker build -t "$CHITTA_IMAGE" "$BACKEND_DIR"

# 2. Build maya (frontend) image
log_info "Building maya (frontend) image: $MAYA_IMAGE..."
docker build -t "$MAYA_IMAGE" "$FRONTEND_DIR"

# 3. Promote chitta image
log_info "Promoting chitta image to $VPS_TARGET..."
docker save "$CHITTA_IMAGE" | ssh "$VPS_TARGET" "docker load"

# 4. Promote maya image
log_info "Promoting maya image to $VPS_TARGET..."
docker save "$MAYA_IMAGE" | ssh "$VPS_TARGET" "docker load"

log_success "All images promoted successfully to $VPS_TARGET!"
