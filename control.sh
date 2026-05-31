#!/usr/bin/env bash
#
# control.sh - Senior DevOps & Systems Automation Controller for MyKanban
#
# This script orchestrates and manages the local development lifecycle of the
# MyKanban Go API Server (Backend) and the Next.js Web App (Frontend).
#
# POSIX-compliant, shell-safe, and self-documenting.
#
set -euo pipefail

# Define variables and resolve absolute paths to support execution from any directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOCKFILE="$SCRIPT_DIR/.dev_pids"
LOGS_DIR="$SCRIPT_DIR/logs"
BACKEND_DIR="$SCRIPT_DIR/backend"
FRONTEND_DIR="$SCRIPT_DIR/frontend"

# Ports associated with the services for validation and forceful sweeps
BACKEND_PORT=8080
FRONTEND_PORTS=(3000 3001 3002)

# Ensure logs directory exists
mkdir -p "$LOGS_DIR"

# Print colored messages for terminal feedback
log_info() {
    printf "\033[1;34m[INFO]\033[0m %s\n" "$1"
}

log_success() {
    printf "\033[1;32m[SUCCESS]\033[0m %s\n" "$1"
}

log_warn() {
    printf "\033[1;33m[WARNING]\033[0m %s\n" "$1"
}

log_error() {
    printf "\033[1;31m[ERROR]\033[0m %s\n" "$1"
}

# Sourced lockfile checker
load_pids() {
    BACKEND_PID=""
    FRONTEND_PID=""
    if [ -f "$LOCKFILE" ]; then
        # Parse lockfile securely without arbitrary execution
        BACKEND_PID=$(grep -E '^BACKEND_PID=' "$LOCKFILE" | cut -d'=' -f2 || true)
        FRONTEND_PID=$(grep -E '^FRONTEND_PID=' "$LOCKFILE" | cut -d'=' -f2 || true)
    fi
}

# Helper to check if a process is active
is_alive() {
    local pid="$1"
    if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
        return 0 # Alive
    fi
    return 1 # Dead
}

# Terminal User Manual
show_help() {
    cat << 'EOF'
MyKanban Development Environment Controller
===========================================
Usage: ./control.sh [command]

A production-grade developer tool to automate orchestrating local backend (Go)
and frontend (Next.js) servers simultaneously in the background.

Available Commands:
  start     Concurrently start Go API server and Next.js dev server in the background.
            Captures process IDs and tracks them in a hidden lockfile.
  stop      Gracefully terminate running servers using SIGTERM and clean lockfile.
  kill      Forcefully terminate servers using SIGKILL and sweep port conflicts.
  restart   Sequence graceful stop (or kill) followed by start with a 2-second delay.
  status    Check OS kernel for process states and display uptime statistics.
  logs      Aggregated interactive tail of backend and frontend logs.
  help      Display this user manual.

Log Files:
  - Backend:   logs/backend.log
  - Frontend:  logs/frontend.log
EOF
}

# Display service access endpoints
show_access_endpoints() {
    log_info "MyKanban is now running. Access it here:"
    echo "------------------------------------------"
    printf "Backend API:   http://localhost:%s\n" "$BACKEND_PORT"
    printf "Frontend Web:  http://localhost:%s\n" "${FRONTEND_PORTS[0]}"
    echo "------------------------------------------"
}

# Start subcommand
do_start() {
    load_pids
    
    local already_running=0
    if is_alive "$BACKEND_PID"; then
        log_warn "Backend Go Server is already running on PID $BACKEND_PID."
        already_running=1
    fi
    if is_alive "$FRONTEND_PID"; then
        log_warn "Frontend Next.js Server is already running on PID $FRONTEND_PID."
        already_running=1
    fi

    if [ "$already_running" -eq 1 ]; then
        log_info "Use './control.sh status' to inspect, or './control.sh restart' to reload."
        return 0
    fi

    # 1. Start Go API Backend
    log_info "Launching Backend Go Server on port $BACKEND_PORT..."
    if [ ! -f "$BACKEND_DIR/bin/tracker-server" ]; then
        log_warn "Backend binary not found. Compiling first..."
        (cd "$BACKEND_DIR" && bash scripts/build.sh)
    fi
    
    cd "$BACKEND_DIR"
    nohup ./bin/tracker-server > "$LOGS_DIR/backend.log" 2>&1 &
    NEW_BACKEND_PID=$!
    disown "$NEW_BACKEND_PID" 2>/dev/null || true
    cd "$SCRIPT_DIR"

    # 2. Start Next.js Frontend
    log_info "Launching Frontend Next.js Dev Server..."
    cd "$FRONTEND_DIR"
    # Set PORT=3000 but Node will fallback to 3001/3002 if 3000 is occupied.
    nohup npx next dev > "$LOGS_DIR/frontend.log" 2>&1 &
    NEW_FRONTEND_PID=$!
    disown "$NEW_FRONTEND_PID" 2>/dev/null || true
    cd "$SCRIPT_DIR"

    # Give them a split second to spin up, then verify
    sleep 1.5

    local backend_started=0
    local frontend_started=0

    if is_alive "$NEW_BACKEND_PID"; then
        log_success "Backend Go Server started (PID: $NEW_BACKEND_PID)."
        backend_started=1
    else
        log_error "Backend failed to start. Check logs/backend.log for details."
    fi

    if is_alive "$NEW_FRONTEND_PID"; then
        log_success "Frontend Next.js Server started (PID: $NEW_FRONTEND_PID)."
        frontend_started=1
    else
        log_error "Frontend failed to start. Check logs/frontend.log for details."
    fi

    # Save to Dev lockfile
    echo "BACKEND_PID=$NEW_BACKEND_PID" > "$LOCKFILE"
    echo "FRONTEND_PID=$NEW_FRONTEND_PID" >> "$LOCKFILE"
    log_info "PIDs successfully committed to .dev_pids lockfile."

    if [ "$backend_started" -eq 1 ] && [ "$frontend_started" -eq 1 ]; then
        show_access_endpoints
    fi
}

# Stop subcommand
do_stop() {
    load_pids
    local stopped=0

    # Stop Backend
    if is_alive "$BACKEND_PID"; then
        log_info "Gracefully stopping Backend (PID: $BACKEND_PID) with SIGTERM..."
        kill -15 "$BACKEND_PID" 2>/dev/null || true
        stopped=1
    else
        log_info "Backend Go Server is not running."
    fi

    # Stop Frontend
    if is_alive "$FRONTEND_PID"; then
        log_info "Gracefully stopping Frontend (PID: $FRONTEND_PID) with SIGTERM..."
        kill -15 "$FRONTEND_PID" 2>/dev/null || true
        stopped=1
    else
        log_info "Frontend Next.js Server is not running."
    fi

    # Wait up to 5 seconds for clean exit
    if [ "$stopped" -eq 1 ]; then
        log_info "Waiting for processes to exit..."
        for _ in {1..10}; do
            if ! is_alive "$BACKEND_PID" && ! is_alive "$FRONTEND_PID"; then
                break
            fi
            sleep 0.5
        done
    fi

    # Post-check and cleanup
    if ! is_alive "$BACKEND_PID" && ! is_alive "$FRONTEND_PID"; then
        rm -f "$LOCKFILE"
        log_success "All development services gracefully stopped and lockfile cleared."
    else
        log_warn "Some services did not stop gracefully. Run './control.sh kill' to force quit."
    fi
}

# Kill subcommand
do_kill() {
    load_pids

    # Hard-kill registered PIDs
    if is_alive "$BACKEND_PID"; then
        log_warn "Force killing Backend Go process $BACKEND_PID..."
        kill -9 "$BACKEND_PID" 2>/dev/null || true
    fi

    if is_alive "$FRONTEND_PID"; then
        log_warn "Force killing Frontend Node process $FRONTEND_PID..."
        kill -9 "$FRONTEND_PID" 2>/dev/null || true
    fi

    # Sweep port conflicts using lsof
    if command -v lsof >/dev/null 2>&1; then
        log_info "Sweeping port conflicts using lsof..."
        # Check backend port
        local bp_pids
        bp_pids=$(lsof -t -i :$BACKEND_PORT 2>/dev/null || true)
        if [ -n "$bp_pids" ]; then
            log_warn "Found lingering processes on backend port $BACKEND_PORT: $bp_pids. Killing..."
            kill -9 $bp_pids 2>/dev/null || true
        fi

        # Check frontend ports
        for port in "${FRONTEND_PORTS[@]}"; do
            local fp_pids
            fp_pids=$(lsof -t -i :"$port" 2>/dev/null || true)
            if [ -n "$fp_pids" ]; then
                log_warn "Found lingering processes on frontend port $port: $fp_pids. Killing..."
                kill -9 $fp_pids 2>/dev/null || true
            fi
        done
    else
        log_warn "lsof command not found. Port sweep skipped. Please install lsof if port conflicts persist."
    fi

    rm -f "$LOCKFILE"
    log_success "Environment forcefully cleared and lockfile purged."
}

# Status subcommand
do_status() {
    load_pids
    local backend_status="\033[1;31mSTOPPED\033[0m"
    local frontend_status="\033[1;31mSTOPPED\033[0m"
    local backend_uptime="N/A"
    local frontend_uptime="N/A"
    local backend_running=0
    local frontend_running=0

    if is_alive "$BACKEND_PID"; then
        backend_status="\033[1;32mRUNNING\033[0m"
        backend_uptime=$(ps -o etime= -p "$BACKEND_PID" | xargs || echo "Unknown")
        backend_running=1
    fi

    if is_alive "$FRONTEND_PID"; then
        frontend_status="\033[1;32mRUNNING\033[0m"
        frontend_uptime=$(ps -o etime= -p "$FRONTEND_PID" | xargs || echo "Unknown")
        frontend_running=1
    fi

    echo "MyKanban Development Environment Status"
    echo "========================================="
    printf "Backend Go Server:  [%b] (PID: %s, Uptime: %s)\n" "$backend_status" "${BACKEND_PID:-None}" "$backend_uptime"
    printf "Frontend Next.js:   [%b] (PID: %s, Uptime: %s)\n" "$frontend_status" "${FRONTEND_PID:-None}" "$frontend_uptime"
    echo "========================================="

    if [ "$backend_running" -eq 1 ] && [ "$frontend_running" -eq 1 ]; then
        show_access_endpoints
    fi
}

# Logs subcommand
do_logs() {
    log_info "Tailing backend and frontend development logs. Press Ctrl+C to exit."
    tail -n 50 -F "$LOGS_DIR/backend.log" "$LOGS_DIR/frontend.log"
}

# Main Command Dispatcher
COMMAND="${1:-help}"

case "$COMMAND" in
    start)
        do_start
        ;;
    stop)
        do_stop
        ;;
    kill)
        do_kill
        ;;
    restart)
        log_info "Initiating development environment restart..."
        # Attempt graceful stop, fallback to kill if lockfile remains
        do_stop
        load_pids
        if [ -f "$LOCKFILE" ]; then
            log_warn "Lockfile remains. Executing forceful sweep..."
            do_kill
        fi
        sleep 2
        do_start
        ;;
    status)
        do_status
        ;;
    logs)
        do_logs
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
