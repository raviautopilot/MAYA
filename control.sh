#!/usr/bin/env bash
#
# control.sh - Senior DevOps & Systems Automation Controller for MyKanban
#
# This script orchestrates and manages the local development lifecycle of the
# MyKanban Go API Server (chitta) and the Next.js Web App (maya).
#
# POSIX-compliant, shell-safe, and self-documenting.
#
set -euo pipefail

# Define variables and resolve absolute paths to support execution from any directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOCKFILE="$SCRIPT_DIR/.dev_pids"
LOGS_DIR="$SCRIPT_DIR/logs"
CHITTA_DIR="$SCRIPT_DIR/chitta"
MAYA_DIR="$SCRIPT_DIR/maya"

# Ports associated with the services for validation and forceful sweeps
CHITTA_PORT=8080
MAYA_PORTS=(3000 3001 3002)

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
    CHITTA_PID=""
    MAYA_PID=""
    if [ -f "$LOCKFILE" ]; then
        # Parse lockfile securely without arbitrary execution
        CHITTA_PID=$(grep -E '^CHITTA_PID=' "$LOCKFILE" | cut -d'=' -f2 || true)
        MAYA_PID=$(grep -E '^MAYA_PID=' "$LOCKFILE" | cut -d'=' -f2 || true)
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
  - Backend:   logs/chitta.log
  - Frontend:  logs/maya.log
EOF
}

# Display service access endpoints
show_access_endpoints() {
    log_info "MyKanban is now running. Access it here:"
    echo "------------------------------------------"
    printf "Backend API:   http://localhost:%s\n" "$CHITTA_PORT"
    printf "Frontend Web:  http://localhost:%s\n" "${MAYA_PORTS[0]}"
    echo "------------------------------------------"
}

# Start subcommand
do_start() {
    load_pids
    
    local already_running=0
    if is_alive "$CHITTA_PID"; then
        log_warn "Chitta Go Server is already running on PID $CHITTA_PID."
        already_running=1
    fi
    if is_alive "$MAYA_PID"; then
        log_warn "Maya Next.js Server is already running on PID $MAYA_PID."
        already_running=1
    fi

    if [ "$already_running" -eq 1 ]; then
        log_info "Use './control.sh status' to inspect, or './control.sh restart' to reload."
        return 0
    fi

    # 1. Start Go API Backend
    log_info "Launching Chitta Go Server on port $CHITTA_PORT..."
    if [ ! -f "$CHITTA_DIR/bin/chitta" ]; then
        log_warn "Chitta binary not found. Compiling first..."
        (cd "$CHITTA_DIR" && bash scripts/build.sh)
    fi
    
    cd "$CHITTA_DIR"
    nohup ./bin/chitta > "$LOGS_DIR/chitta.log" 2>&1 &
    NEW_CHITTA_PID=$!
    disown "$NEW_CHITTA_PID" 2>/dev/null || true
    cd "$SCRIPT_DIR"

    # 2. Start Next.js Frontend
    log_info "Launching Maya Next.js Dev Server..."
    cd "$MAYA_DIR"
    # Set PORT=3000 but Node will fallback to 3001/3002 if 3000 is occupied.
    nohup ../node_modules/.bin/next dev > "$LOGS_DIR/maya.log" 2>&1 &
    NEW_MAYA_PID=$!
    disown "$NEW_MAYA_PID" 2>/dev/null || true
    cd "$SCRIPT_DIR"

    # Give them a split second to spin up, then verify
    sleep 1.5

    local chitta_started=0
    local maya_started=0

    if is_alive "$NEW_CHITTA_PID"; then
        log_success "Chitta Go Server started (PID: $NEW_CHITTA_PID)."
        chitta_started=1
    else
        log_error "Chitta failed to start. Check logs/chitta.log for details."
    fi

    if is_alive "$NEW_MAYA_PID"; then
        log_success "Maya Next.js Server started (PID: $NEW_MAYA_PID)."
        maya_started=1
    else
        log_error "Maya failed to start. Check logs/maya.log for details."
    fi

    # Save to Dev lockfile
    echo "CHITTA_PID=$NEW_CHITTA_PID" > "$LOCKFILE"
    echo "MAYA_PID=$NEW_MAYA_PID" >> "$LOCKFILE"
    log_info "PIDs successfully committed to .dev_pids lockfile."

    if [ "$chitta_started" -eq 1 ] && [ "$maya_started" -eq 1 ]; then
        show_access_endpoints
    fi
}

# Stop subcommand
do_stop() {
    load_pids
    local stopped=0

    # Stop Backend
    if is_alive "$CHITTA_PID"; then
        log_info "Gracefully stopping Chitta (PID: $CHITTA_PID) with SIGTERM..."
        kill -15 "$CHITTA_PID" 2>/dev/null || true
        stopped=1
    else
        log_info "Chitta Go Server is not running."
    fi

    # Stop Frontend
    if is_alive "$MAYA_PID"; then
        log_info "Gracefully stopping Maya (PID: $MAYA_PID) with SIGTERM..."
        kill -15 "$MAYA_PID" 2>/dev/null || true
        stopped=1
    else
        log_info "Maya Next.js Server is not running."
    fi

    # Wait up to 5 seconds for clean exit
    if [ "$stopped" -eq 1 ]; then
        log_info "Waiting for processes to exit..."
        for _ in {1..10}; do
            if ! is_alive "$CHITTA_PID" && ! is_alive "$MAYA_PID"; then
                break
            fi
            sleep 0.5
        done
    fi

    # Post-check and cleanup
    if ! is_alive "$CHITTA_PID" && ! is_alive "$MAYA_PID"; then
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
    if is_alive "$CHITTA_PID"; then
        log_warn "Force killing Chitta Go process $CHITTA_PID..."
        kill -9 "$CHITTA_PID" 2>/dev/null || true
    fi

    if is_alive "$MAYA_PID"; then
        log_warn "Force killing Maya Node process $MAYA_PID..."
        kill -9 "$MAYA_PID" 2>/dev/null || true
    fi

    # Sweep port conflicts using lsof
    if command -v lsof >/dev/null 2>&1; then
        log_info "Sweeping port conflicts using lsof..."
        # Check backend port
        local bp_pids
        bp_pids=$(lsof -t -i :$CHITTA_PORT 2>/dev/null || true)
        if [ -n "$bp_pids" ]; then
            log_warn "Found lingering processes on chitta port $CHITTA_PORT: $bp_pids. Killing..."
            kill -9 $bp_pids 2>/dev/null || true
        fi

        # Check frontend ports
        for port in "${MAYA_PORTS[@]}"; do
            local fp_pids
            fp_pids=$(lsof -t -i :"$port" 2>/dev/null || true)
            if [ -n "$fp_pids" ]; then
                log_warn "Found lingering processes on maya port $port: $fp_pids. Killing..."
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
    local chitta_status="\033[1;31mSTOPPED\033[0m"
    local maya_status="\033[1;31mSTOPPED\033[0m"
    local chitta_uptime="N/A"
    local maya_uptime="N/A"
    local chitta_running=0
    local maya_running=0

    if is_alive "$CHITTA_PID"; then
        chitta_status="\033[1;32mRUNNING\033[0m"
        chitta_uptime=$(ps -o etime= -p "$CHITTA_PID" | xargs || echo "Unknown")
        chitta_running=1
    fi

    if is_alive "$MAYA_PID"; then
        maya_status="\033[1;32mRUNNING\033[0m"
        maya_uptime=$(ps -o etime= -p "$MAYA_PID" | xargs || echo "Unknown")
        maya_running=1
    fi

    echo "MyKanban Development Environment Status"
    echo "========================================="
    printf "Chitta Go Server:  [%b] (PID: %s, Uptime: %s)\n" "$chitta_status" "${CHITTA_PID:-None}" "$chitta_uptime"
    printf "Maya Next.js:   [%b] (PID: %s, Uptime: %s)\n" "$maya_status" "${MAYA_PID:-None}" "$maya_uptime"
    echo "========================================="

    if [ "$chitta_running" -eq 1 ] && [ "$maya_running" -eq 1 ]; then
        show_access_endpoints
    fi
}

# Logs subcommand
do_logs() {
    log_info "Tailing chitta and maya development logs. Press Ctrl+C to exit."
    tail -n 50 -F "$LOGS_DIR/chitta.log" "$LOGS_DIR/maya.log"
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
