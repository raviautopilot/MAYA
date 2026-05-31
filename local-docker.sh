#!/usr/bin/env bash
#
# local-docker.sh - Local Docker Environment Controller
#
# This script provides a convenient wrapper for managing the local Docker
# containers using docker-compose. It's designed for local testing and
# validation before promoting images.
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
Usage: ./local-docker.sh [command]

Manages the local Docker environment for MyKanban.

Commands:
  start     Build images if they don't exist and start services.
  stop      Gracefully stop the running services.
  kill      Force kill the services immediately.
  delete    Stop and remove containers, networks, and volumes.
  restart   Restart the services.
  status    Show the status of the containers.
  logs      Follow the logs of all services.
  build     Force a rebuild of the images.
  help      Show this help message.
EOF
}

COMMAND="${1:-help}"

case "$COMMAND" in
    start)
        log_info "Starting services..."
        $DOCKER_CMD up -d
        log_success "Services started. Use './local-docker.sh status' to check."
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
        log_info "Removing services (containers, networks, volumes)..."
        $DOCKER_CMD down --volumes
        log_success "Services and associated volumes deleted."
        ;;
    restart)
        log_info "Restarting services..."
        $DOCKER_CMD restart
        log_success "Services restarted."
        ;;
    status|ps)
        log_info "Local deployment status:"
        $DOCKER_CMD ps
        ;;
    logs)
        log_info "Tailing logs (Ctrl+C to exit)..."
        $DOCKER_CMD logs -f
        ;;
    build)
        log_info "Forcing a rebuild of all images..."
        $DOCKER_CMD build
        log_success "Images rebuilt."
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
