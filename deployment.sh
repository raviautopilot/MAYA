#!/usr/bin/env bash
#
# deployment.sh - VPS Deployment Controller
#
# This script manages the deployment lifecycle using docker compose.
# It should be run directly on the VPS.
#
set -euo pipefail

# Ensure docker compose is available
if ! command -v docker >/dev/null 2>&1; then
    printf "\033[1;31m[ERROR]\033[0m Docker is not installed or not in PATH.\n"
    exit 1
fi
DOCKER_CMD="docker compose"

log_info() {
    printf "\033[1;34m[INFO]\033[0m %s\n" "$1"
}

log_success() {
    printf "\033[1;32m[SUCCESS]\033[0m %s\n" "$1"
}

log_error() {
    printf "\033[1;31m[ERROR]\033[0m %s\n" "$1"
}

show_help() {
    cat << 'EOF'
Usage: ./deployment.sh [command]

Commands:
  start     Start the services in the background (detached mode)
  stop      Gracefully stop the running services
  kill      Force kill the services immediately
  delete    Stop and remove containers, networks, and volumes (down)
  restart   Restart the services
  status    Show the status of the containers
  logs      Follow the logs of all services
  help      Show this help message
EOF
}

COMMAND="${1:-help}"

case "$COMMAND" in
    start)
        log_info "Starting services..."
        $DOCKER_CMD up -d
        log_success "Services started."
        ;;
    stop)
        log_info "Stopping services..."
        $DOCKER_CMD stop
        log_success "Services stopped."
        ;;
    kill)
        log_info "Force killing services..."
        $DOCKER_CMD kill
        log_success "Services killed."
        ;;
    delete|down)
        log_info "Removing services (containers, networks)..."
        $DOCKER_CMD down
        log_success "Services deleted."
        ;;
    restart)
        log_info "Restarting services..."
        $DOCKER_CMD restart
        log_success "Services restarted."
        ;;
    status|ps)
        log_info "Current deployment status:"
        $DOCKER_CMD ps
        ;;
    logs)
        log_info "Tailing logs (Ctrl+C to exit)..."
        $DOCKER_CMD logs -f
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        log_error "Unknown command: '$COMMAND'"
        show_help
        exit 1
        ;;
esac
