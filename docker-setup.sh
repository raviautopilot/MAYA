#!/usr/bin/env bash
# =============================================================================
# docker-setup.sh — MyKanban Docker Compose Lifecycle Controller
#
# POSIX-compliant, production-grade Bash wrapper that orchestrates the full
# containerised environment lifecycle for the MyKanban project.
#
# Services managed:
#   potti   — PostgreSQL 16 database (persistent volume)
#   varam   — Go REST API backend (compiled Gin server)
#   laxmi   — Next.js 14 frontend (SSR production build)
#
# Usage:  ./docker-setup.sh <subcommand>
#
# Author:  DevOps Engineering
# Version: 1.0.0
# License: UNLICENSED (private)
# =============================================================================

# ---------------------------------------------------------------------------
# Strict mode — exit on error, undefined var reference, or pipeline failure.
# 'set -o pipefail' ensures a pipeline returns non-zero if ANY stage fails.
# ---------------------------------------------------------------------------
set -euo pipefail

# ---------------------------------------------------------------------------
# Paths & Constants
#
# SCRIPT_DIR   — absolute path so this script works when invoked from any cwd
# COMPOSE_FILE — canonical Compose file (both forms are accepted by Docker)
# ENV_FILE     — local .env that Docker Compose auto-loads for variable injection
# ---------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="${SCRIPT_DIR}/docker-compose.yml"
ENV_FILE="${SCRIPT_DIR}/.env"

# If the user also defines a second compose override file, Docker merges them
# automatically when the filename is 'docker-compose.override.yml'
COMPOSE_OVERRIDE="${SCRIPT_DIR}/docker-compose.override.yml"

# ---------------------------------------------------------------------------
# Colour & Formatting — sourced from terminal escape codes
#
# We define the full ANSI palette here and keep usage consistent throughout
# the script so colour-blind / no-color environments can be adjusted in one
# place.
# ---------------------------------------------------------------------------
NC='\033[0m'           # No Colour (reset)

# Foreground colours
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
BOLD='\033[1m'
DIM='\033[2m'

# ---------------------------------------------------------------------------
# Logging helpers — each prefixes a coloured tag so output is easy to scan
# ---------------------------------------------------------------------------
log_info()    { printf "${BLUE}[INFO]${NC}    %s\n" "$*"; }
log_success() { printf "${GREEN}[OK]${NC}      %s\n" "$*"; }
log_warn()    { printf "${YELLOW}[WARN]${NC}    %s\n" "$*" >&2; }
log_error()   { printf "${RED}[ERROR]${NC}   %s\n" "$*" >&2; }
log_fatal()   { printf "${RED}[FATAL]${NC}   %s\n" "$*" >&2; exit 1; }
log_task()    { printf "${MAGENTA}[TASK]${NC}    %s\n" "$*"; }
log_manual()  { printf "${CYAN}[ACTION]${NC}  %s\n" "$*"; }

# ---------------------------------------------------------------------------
# Section separator — visually marks phases in the terminal
# ---------------------------------------------------------------------------
separator() {
    printf "${DIM}──────────────────────────────────────────────────────────────${NC}\n"
}

# ---------------------------------------------------------------------------
# Pre-flight checks — run before EVERY subcommand to ensure:
#   1. Docker Engine is reachable
#   2. Docker Compose plugin (v2) is available
#   3. The compose YAML definition exists
#   4. Required credential / configuration files are present
#   5. Required environment variables are set
# ---------------------------------------------------------------------------
preflight_checks() {
    log_task "Running pre-flight checks..."

    # --- 1. Docker daemon reachable ---------------------------------------
    if ! docker info >/dev/null 2>&1; then
        log_fatal "Docker daemon is not running or is not accessible.\n" \
                  "  Start Docker with:  sudo systemctl start docker\n" \
                  "  Or check:           docker info"
    fi
    log_success "Docker Engine is reachable."

    # --- 2. Docker Compose plugin (v2) available --------------------------
    if ! docker compose version >/dev/null 2>&1; then
        log_fatal "Docker Compose v2 (plugin) is not installed.\n" \
                  "  Install: https://docs.docker.com/compose/install/"
    fi
    log_success "Docker Compose plugin detected."

    # --- 3. Compose file exists -------------------------------------------
    if [ ! -f "${COMPOSE_FILE}" ]; then
        log_fatal "Compose definition not found at: ${COMPOSE_FILE}\n" \
                  "  Ensure 'docker-compose.yml' exists in the project root."
    fi
    log_success "Compose file found: ${COMPOSE_FILE}"

    # --- 4. .env file exists ----------------------------------------------
    if [ ! -f "${ENV_FILE}" ]; then
        log_warn ".env file not found at: ${ENV_FILE}"
        log_manual "Create ${ENV_FILE} with the required variables:\n" \
                   "  ${CYAN}PG_PASSWORD=your_secure_password${NC}\n" \
                   "  ${CYAN}JWT_SECRET=your_jwt_secret${NC}\n" \
                   "\n  See '${COMPOSE_FILE}' for the full variable reference," \
                   "or copy from .env.example if one exists."
        log_fatal "Aborting — .env is mandatory for safe operation."
    fi

    # Source the .env so we can validate values directly in this script
    # shellcheck disable=SC1090
    set -a && source "${ENV_FILE}" && set +a
    log_success ".env file loaded (${ENV_FILE})"

    # --- 5. Validate critical variables -----------------------------------
    # PG_PASSWORD is marked with `:?` in compose, so we double-check it here
    # for a friendlier error message.
    if [ -z "${PG_PASSWORD:-}" ]; then
        log_fatal "PG_PASSWORD is not set in ${ENV_FILE}.\n" \
                  "  Add the following line to ${ENV_FILE}:\n" \
                  "    ${CYAN}PG_PASSWORD=$(openssl rand -base64 24)${NC}"
    fi

    # --- 6. Backend config.json exists for bind-mount ---------------------
    local BACKEND_CONFIG="${SCRIPT_DIR}/backend/config.json"
    if [ ! -f "${BACKEND_CONFIG}" ]; then
        local EXAMPLE_CONFIG="${SCRIPT_DIR}/backend/config.example.json"
        if [ -f "${EXAMPLE_CONFIG}" ]; then
            log_warn "backend/config.json not found."
            log_manual "Copy the example and edit:\n" \
                       "  cp ${EXAMPLE_CONFIG} ${BACKEND_CONFIG}\n" \
                       "  ${EDITOR:-vi} ${BACKEND_CONFIG}"
        fi
        log_fatal "Aborting — backend configuration is required."
    fi
    log_success "Backend config present: ${BACKEND_CONFIG}"

    # Safety check: ensure config.json has a real JWT secret (not the default)
    if grep -q 'change-me' "${BACKEND_CONFIG}" 2>/dev/null; then
        log_warn "backend/config.json still uses the DEFAULT JWT secret."
        log_manual "Edit ${BACKEND_CONFIG} and set a strong value for" \
                   "\"jwt_secret\" before deploying to anything beyond local dev."
    fi

    separator
    log_success "All pre-flight checks passed."
    echo
}

# ===========================================================================
# SUBCOMMAND: build
#
# Builds (or rebuilds) the Docker images for varam and laxmi.
#
# Flags:
#   --no-cache   Force a clean build, discarding Docker's layer cache.
#   --pull       Always pull fresh base images before building.
#
# Semantic: docker compose build [--no-cache] [--pull]
# ===========================================================================
do_build() {
    local BUILD_FLAGS=""

    # Parse all remaining positional arguments as flags (shifted in main())
    # Supports zero or more of: --no-cache, --clean, --pull
    # Using a while-loop over $@ avoids the case-pattern limitation (no spaces).
    local arg
    for arg in "$@"; do
        case "${arg}" in
            --no-cache|--clean)
                BUILD_FLAGS="${BUILD_FLAGS} --no-cache"
                log_warn "Cache disabled — performing a clean rebuild."
                ;;
            --pull)
                BUILD_FLAGS="${BUILD_FLAGS} --pull"
                log_info "Pulling latest base images before building."
                ;;
            --help|-h)
                do_help
                exit 0
                ;;
            # Silently skip the subcommand word itself
            build)
                ;;
            *)
                log_warn "Unknown build flag: '${arg}' — ignoring."
                ;;
        esac
    done
    # Trim leading space
    BUILD_FLAGS="${BUILD_FLAGS# }"

    log_task "Building Docker images for services: varam, laxmi"
    separator

    # Shell-expand variables in the compose file before building
    # shellcheck disable=SC2086
    docker compose \
        --file "${COMPOSE_FILE}" \
        build \
        ${BUILD_FLAGS} \
        varam laxmi

    local EXIT_CODE=$?
    separator

    if [ ${EXIT_CODE} -eq 0 ]; then
        log_success "Images built successfully."
        echo
        # Show image sizes for auditability
        docker images --filter "label=com.docker.compose.project=mycan" \
            --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}" 2>/dev/null || \
        docker images \
            --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}" \
            "$(basename "${SCRIPT_DIR}")_varam" \
            "$(basename "${SCRIPT_DIR}")_laxmi" 2>/dev/null || true
    else
        log_error "Build failed (exit code ${EXIT_CODE}). Check the output above."
    fi
}

# ===========================================================================
# SUBCOMMAND: start | run
#
# Launches the entire cluster in detached mode.
#
# Behaviour:
#   1. Runs pre-flight checks (config validation).
#   2. Creates missing volumes if this is the first deployment.
#   3. Starts all services in dependency order (potti → varam → laxmi).
#   4. Waits for health checks to pass before declaring success.
#   5. Prints a human-readable access summary table.
# ===========================================================================
do_start() {
    log_task "Launching MyKanban cluster in detached mode..."
    separator

    # -----------------------------------------------------------------------
    # Docker Compose 'up' — create & start all services
    # -----------------------------------------------------------------------
    log_info "Starting services: potti (DB) → varam (API) → laxmi (UI)"
    echo

    docker compose \
        --file "${COMPOSE_FILE}" \
        up \
        --detach \
        --remove-orphans \
        --wait \
        2>&1

    local EXIT_CODE=$?
    separator

    if [ ${EXIT_CODE} -ne 0 ]; then
        log_error "Failed to start the cluster (exit ${EXIT_CODE})."
        log_info "Inspect individual services:"
        echo "    docker compose logs --tail=50 potti"
        echo "    docker compose logs --tail=50 varam"
        echo "    docker compose logs --tail=50 laxmi"
        exit "${EXIT_CODE}"
    fi

    log_success "Cluster is live and healthy."
    echo

    # -----------------------------------------------------------------------
    # Print Access Summary — a terminal-friendly table showing how to reach
    # each service from the host.
    # -----------------------------------------------------------------------
    print_access_summary

    # -----------------------------------------------------------------------
    # Optional: confirm the backend health endpoint responds
    # -----------------------------------------------------------------------
    if command -v curl >/dev/null 2>&1; then
        local HEALTH
        HEALTH=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 \
                    http://localhost:8080/api/health 2>/dev/null || echo "timeout")
        if [ "${HEALTH}" = "200" ]; then
            log_success "Backend health check: HTTP 200 (healthy)"
        else
            log_warn "Backend health check returned HTTP ${HEALTH}."
            log_info "This may be normal during initial startup — retry in a few seconds:"
            echo "    curl -s http://localhost:8080/api/health"
        fi
    fi
}

# ===========================================================================
# Helper: Print Access Summary
#
# Queries Docker to determine the actual host port mappings (important when
# ports are dynamically assigned or when running in non-standard configs)
# and prints a clean table.
# ===========================================================================
print_access_summary() {
    printf "${BOLD}Application Access Summary${NC}\n"
    printf "${DIM}──────────────────────────────────────────────────────────────${NC}\n"

    # Define known service metadata
    # Format: service_name  container_port  description  url_path
    local SERVICES="varam:8080:Backend API (Gin):/api/health
laxmi:3000:Frontend (Next.js):/
potti:5432:Database (PostgreSQL):(internal only)"

    local LINE
    echo "${SERVICES}" | while IFS=: read -r SERVICE CPORT DESC URL; do
        # Query Docker for the actual host port mapping
        local HOST_PORT
        HOST_PORT=$(docker compose port "${SERVICE}" "${CPORT}" 2>/dev/null \
                    | sed 's/.*://' || echo "${CPORT}")

        local FULL_URL="http://localhost:${HOST_PORT}${URL}"

        if [ "${DESC}" = "Database (PostgreSQL)" ]; then
            printf "  ${CYAN}%-10s${NC}  ➜  ${BOLD}%-55s${NC}\n" \
                   "${SERVICE}" "${DESC}"
            printf "  ${DIM}%12s${NC}      postgresql://mycan@localhost:${HOST_PORT}/mykanban${NC}\n" ""
        else
            printf "  ${CYAN}%-10s${NC}  ➜  ${BOLD}%-55s${NC}\n" \
                   "${SERVICE}" "${FULL_URL}"
        fi
    done

    printf "${DIM}──────────────────────────────────────────────────────────────${NC}\n"
    echo
    log_manual "Open ${BOLD}http://localhost:3000${NC} in your browser for the application."
    log_info "Swagger API docs are at ${BOLD}http://localhost:8080/swagger/index.html${NC}"
    echo
}

# ===========================================================================
# SUBCOMMAND: stop
#
# Gracefully halts all running containers WITHOUT removing them.
#
# Uses: docker compose stop
#   - Sends SIGTERM → waits for graceful shutdown (default 10s per container)
#   - Containers remain in 'exited' state — data volumes are untouched
#   - No networks or volumes are deleted
#   - Use 'start' to resume from the stopped state instantly
# ===========================================================================
do_stop() {
    log_task "Gracefully stopping all cluster containers..."
    separator

    # Count running containers before stopping
    local RUNNING_COUNT
    RUNNING_COUNT=$(docker compose ps --status running --format json 2>/dev/null \
                    | grep -c '"State":"running"' || true)

    if [ "${RUNNING_COUNT}" -eq 0 ]; then
        log_info "No running containers to stop."
        log_info "Cluster is already stopped. Use './docker-setup.sh start' to launch."
        return 0
    fi

    log_info "Stopping ${RUNNING_COUNT} running container(s) with SIGTERM..."

    docker compose \
        --file "${COMPOSE_FILE}" \
        stop \
        --time 30

    local EXIT_CODE=$?
    separator

    if [ ${EXIT_CODE} -eq 0 ]; then
        log_success "All containers gracefully stopped."
        log_info "Containers are preserved in the 'exited' state."
        log_info "Resume with:  ./docker-setup.sh start"
    else
        log_error "Some containers may not have stopped cleanly (exit ${EXIT_CODE})."
        log_info "Inspect state:  docker compose ps"
    fi
}

# ===========================================================================
# SUBCOMMAND: delete | down
#
# Completely tears down the container cluster, removes networks, and prunes
# orphaned containers — while EXPLICITLY PRESERVING persistent data volumes.
#
# Safety:
#   - potti_data  (PostgreSQL data)    → PRESERVED
#   - backend_storage (JSON flatfiles) → PRESERVED
#   - Networks, containers, images     → REMOVED
#
# Uses: docker compose down --remove-orphans
#   -v flag is deliberately NOT used to avoid volume destruction.
# ===========================================================================
do_delete() {
    log_task "Tearing down the container cluster..."
    log_warn "This will STOP and REMOVE all containers, networks, and orphans."
    log_success "Persistent volumes (potti_data, backend_storage) are PRESERVED."
    separator

    # Warn about volumes that would be destroyed
    printf "${YELLOW}⚠  ${BOLD}Data preservation notice:${NC}\n"
    printf "   ${GREEN}✔${NC}  ${BOLD}potti_data${NC}       (PostgreSQL)  →  ${GREEN}KEPT${NC}\n"
    printf "   ${GREEN}✔${NC}  ${BOLD}backend_storage${NC}  (JSON files)  →  ${GREEN}KEPT${NC}\n"
    printf "   ${YELLOW}−${NC}  ${BOLD}All containers${NC}   (potti/varam/laxmi)  →  ${RED}REMOVED${NC}\n"
    printf "   ${YELLOW}−${NC}  ${BOLD}Networks${NC}         (mycan_net)          →  ${RED}REMOVED${NC}\n"
    printf "   ${YELLOW}−${NC}  ${BOLD}Orphans${NC}          (unreferenced)       →  ${RED}PRUNED${NC}\n"
    separator
    echo

    # Execute the teardown
    docker compose \
        --file "${COMPOSE_FILE}" \
        down \
        --remove-orphans

    local EXIT_CODE=$?
    separator

    if [ ${EXIT_CODE} -eq 0 ]; then
        log_success "Cluster torn down successfully."
        log_info "Persistent volumes remain intact."
        echo
        log_info "To REBUILD from scratch (including volume reset):"
        echo "    ./docker-setup.sh build --no-cache"
        echo "    ./docker-setup.sh start"
        echo
        log_warn "To DESTROY ALL DATA (irreversible):"
        echo "    docker compose down --volumes"
        echo "    docker volume rm mycan_potti_data mycan_backend_storage"
    else
        log_error "Teardown encountered issues (exit ${EXIT_CODE})."
    fi
}

# ===========================================================================
# SUBCOMMAND: status
#
# Renders a detailed audit table showing:
#   - Per-service container state (running / exited / paused)
#   - Uptime (elapsed wall-clock time since container start)
#   - Internal container port bindings
#   - Actual host port mapping (verified via live Docker queries)
#   - Resource usage (CPU / memory)
# ===========================================================================
do_status() {
    log_task "Cluster Status & Port Audit"
    separator

    # -----------------------------------------------------------------------
    # 1. Quick summary — how many containers are healthy?
    # -----------------------------------------------------------------------
    local SERVICE_COUNT
    local RUNNING_COUNT
    local SERVICE_COUNT
    SERVICE_COUNT=$(docker compose ps --format json 2>/dev/null | grep -c '"ID"' || true)
    RUNNING_COUNT=$(docker compose ps --status running --format json 2>/dev/null \
                    | grep -c '"State":"running"' || true)

    if [ "${SERVICE_COUNT}" -eq 0 ]; then
        log_warn "No containers are defined or running."
        log_info "Launch the cluster first:  ./docker-setup.sh start"
        return 0
    fi

    printf "${BOLD}  %-10s  %-9s  %-10s  %-22s  %-22s  %s${NC}\n" \
           "SERVICE" "STATE" "UPTIME" "CONTAINER PORT" "HOST PORT" "IMAGE"
    printf "${DIM}  ──────────  ─────────  ──────────  ──────────────────────  ──────────────────────  ─────────────────────────${NC}\n"

    # -----------------------------------------------------------------------
    # 2. Iterate over each service and build the audit row
    # -----------------------------------------------------------------------
    # We read the compose ps output — format uses Go templates for precision
    local PS_OUTPUT
    PS_OUTPUT=$(docker compose ps --format '{{.Name}}|{{.Status}}|{{.Ports}}|{{.Image}}' 2>/dev/null)

    if [ -z "${PS_OUTPUT}" ]; then
        log_warn "No container data returned from 'docker compose ps'."
        log_info "Try:  docker compose ps"
        return 0
    fi

    echo "${PS_OUTPUT}" | while IFS='|' read -r NAME STATUS_RAW PORTS_RAW IMAGE; do
        # Strip the Compose project prefix from the container name for brevity
        local SERVICE_NAME
        SERVICE_NAME="${NAME#mycan-}"

        # --- State & Uptime -----------------------------------------------
        local STATE="stopped"
        local UPTIME="—"

        # Status examples from Docker:
        #   "Up 2 hours"       → running
        #   "Up About a minute" → running
        #   "Exited (0)"        → stopped
        #   "Paused"            → paused
        if echo "${STATUS_RAW}" | grep -q "^Up "; then
            STATE="running"
            UPTIME=$(echo "${STATUS_RAW}" | sed 's/^Up //')
        elif echo "${STATUS_RAW}" | grep -q "^Exited"; then
            STATE="stopped"
            UPTIME=$(echo "${STATUS_RAW}" | sed 's/^Exited (//;s/).*//')
            STATE="exited(${UPTIME})"
            UPTIME="—"
        elif echo "${STATUS_RAW}" | grep -q "^Paused"; then
            STATE="paused"
            UPTIME=$(echo "${STATUS_RAW}" | sed 's/^Paused //')
        fi

        # --- Port Mappings ------------------------------------------------
        # Parse the PORTS column which looks like:
        #   0.0.0.0:8080->8080/tcp  or  3000/tcp  (no host mapping)
        local CONTAINER_PORTS="—"
        local HOST_PORTS="—"

        if [ -n "${PORTS_RAW}" ] && [ "${PORTS_RAW}" != "—" ]; then
            # Split multi-port entries (e.g. "0.0.0.0:5432->5432/tcp" or
            # "0.0.0.0:8080->8080/tcp, 8081/tcp") by comma
            CONTAINER_PORTS=""
            HOST_PORTS=""

            IFS=',' read -ra PORT_ENTRIES <<< "${PORTS_RAW}"
            for ENTRY in "${PORT_ENTRIES[@]}"; do
                ENTRY=$(echo "${ENTRY}" | xargs)  # trim whitespace

                local C_PORT H_PORT

                # Pattern: 0.0.0.0:5432->5432/tcp  (published)
                if echo "${ENTRY}" | grep -q '\->'; then
                    H_PORT=$(echo "${ENTRY}" | sed 's/.*://;s/->.*//')
                    C_PORT=$(echo "${ENTRY}" | sed 's/.*->//;s|/tcp.*||')
                # Pattern: 5432/tcp  (unpublished / internal only)
                else
                    C_PORT=$(echo "${ENTRY}" | sed 's|/tcp||;s|/udp||')
                    H_PORT="(internal)"
                fi

                if [ -z "${CONTAINER_PORTS}" ]; then
                    CONTAINER_PORTS="${C_PORT}"
                    HOST_PORTS="${H_PORT}"
                else
                    CONTAINER_PORTS="${CONTAINER_PORTS}, ${C_PORT}"
                    HOST_PORTS="${HOST_PORTS}, ${H_PORT}"
                fi
            done
        fi

        # --- Colourise the state cell --------------------------------------
        local STATE_COLORED
        case "${STATE}" in
            running*)   STATE_COLORED="${GREEN}${STATE}${NC}" ;;
            exited*)    STATE_COLORED="${RED}${STATE}${NC}" ;;
            paused*)    STATE_COLORED="${YELLOW}${STATE}${NC}" ;;
            *)          STATE_COLORED="${DIM}${STATE}${NC}" ;;
        esac

        printf "  ${CYAN}%-10s${NC}  %b  %-10s  %-22s  %-22s  ${DIM}%s${NC}\n" \
               "${SERVICE_NAME}" \
               "${STATE_COLORED}" \
               "${UPTIME:0:10}" \
               "${CONTAINER_PORT:0:22}" \
               "${HOST_PORTS:0:22}" \
               "${IMAGE##*/}"
    done

    printf "${DIM}  ──────────  ─────────  ──────────  ──────────────────────  ──────────────────────  ─────────────────────────${NC}\n"
    echo

    # -----------------------------------------------------------------------
    # 3. Host-level port binding verification
    # -----------------------------------------------------------------------
    # Cross-reference the Docker mappings against what ss (socket statistics)
    # reports the kernel is actually listening on. This catches cases where a
    # Docker port mapping is defined but the host port is already occupied.
    # -----------------------------------------------------------------------
    if command -v ss >/dev/null 2>&1; then
        printf "${BOLD}Host Port Audit (ss -tlnp)${NC}\n"
        printf "${DIM}──────────────────────────────────────────────────────────────${NC}\n"

        # The service ports we expect to be listening on the host
        local EXPECTED_PORTS="8080 3000 5432"
        local PORT
        for PORT in ${EXPECTED_PORTS}; do
            local LISTENER
            LISTENER=$(ss -tlnp "sport = :${PORT}" 2>/dev/null | tail -n +2 || true)

            if [ -n "${LISTENER}" ]; then
                local PROC_NAME
                PROC_NAME=$(echo "${LISTENER}" | grep -oP 'users:\(\(".*?(?=")' || echo "docker-proxy")
                printf "  ${GREEN}✔${NC}  Port ${BOLD}%-5s${NC}  ${DIM}%s${NC}\n" \
                       "${PORT}" "${PROC_NAME}"
            else
                # Port may still be used by a non-Docker process
                local OTHER
                OTHER=$(ss -tlnp "sport = :${PORT}" 2>/dev/null | grep -v 'docker' || true)
                if [ -n "${OTHER}" ]; then
                    printf "  ${YELLOW}⚠${NC}  Port ${BOLD}%-5s${NC}  ${YELLOW}occupied by non-Docker process${NC}\n" "${PORT}"
                else
                    printf "  ${RED}✘${NC}  Port ${BOLD}%-5s${NC}  ${RED}NOT listening on host${NC}\n" "${PORT}"
                fi
            fi
        done
        echo
    else
        log_warn "'ss' command not found — skipping host port audit."
        log_info "Install iproute2:  sudo apt install -y iproute2"
        echo
    fi

    # -----------------------------------------------------------------------
    # 4. Resource usage (optional — uses docker stats snapshot)
    # -----------------------------------------------------------------------
    if [ "${RUNNING_COUNT}" -gt 0 ]; then
        printf "${BOLD}Live Resource Usage (docker stats — no-stream)${NC}\n"
        printf "${DIM}──────────────────────────────────────────────────────────────${NC}\n"
        docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}" 2>/dev/null || \
        log_info "(docker stats unavailable — possibly running in a restricted context)"
        echo
    fi

    # -----------------------------------------------------------------------
    # 5. Volume sanity check
    # -----------------------------------------------------------------------
    printf "${BOLD}Persistent Volumes${NC}\n"
    printf "${DIM}──────────────────────────────────────────────────────────────${NC}\n"
    docker volume ls --filter "name=mycan_" --format "table {{.Name}}\t{{.Driver}}\t{{.Scope}}" 2>/dev/null || \
        log_info "No MyKanban volumes found."
    echo
}

# ===========================================================================
# SUBCOMMAND: logs
#
# Tails the aggregated, colourised, streaming log output for all services
# simultaneously. Each line is prefixed with the service name.
#
# Uses: docker compose logs --follow --tail=100
#
# Flags:
#   --tail=N   Number of lines from the end of each log to show initially
#              (default: 100).
#   --no-follow  Print the log tail and exit instead of streaming.
# ===========================================================================
do_logs() {
    local TAIL_LINES=100
    local FOLLOW_FLAG="--follow"

    # Parse optional flags from remaining arguments
    local arg
    for arg in "$@"; do
        case "${arg}" in
            --tail=*)
                TAIL_LINES="${arg#--tail=}"
                ;;
            --no-follow)
                FOLLOW_FLAG=""
                ;;
            logs|--help|-h)
                do_help
                exit 0
                ;;
            *)
                log_warn "Unknown logs flag: '${arg}' — ignoring."
                ;;
        esac
    done

    log_info "Aggregating logs for all services (last ${TAIL_LINES} lines)..."
    log_info "Press ${BOLD}Ctrl+C${NC} to stop following."
    separator
    echo

    # 'docker compose logs' with colourised output via the --no-log-prefix
    # flag conflicts with our prefix strategy, so we rely on Docker's default
    # prefix behaviour (service name in colour) which is already clear.
    docker compose \
        --file "${COMPOSE_FILE}" \
        logs \
        ${FOLLOW_FLAG} \
        --tail="${TAIL_LINES}" \
        --no-color=false \
        potti varam laxmi

    # Note: This command blocks indefinitely when --follow is active.
    # When the user presses Ctrl+C, control returns here.
    echo
    separator
    log_info "Log streaming ended."
}

# ===========================================================================
# SUBCOMMAND: help
#
# Prints a beautifully formatted terminal manual covering all subcommands,
# access URLs, and practical execution examples.
# ===========================================================================
do_help() {
    cat << 'EOF'

╔══════════════════════════════════════════════════════════════════════════╗
║              🐚  MyKanban Docker Controller  —  Manual                  ║
╚══════════════════════════════════════════════════════════════════════════╝

  A production-grade Bash wrapper for orchestrating the containerised
  MyKanban development environment via Docker Compose.

  Usage:  ./docker-setup.sh <subcommand> [options]

─── Commands ─────────────────────────────────────────────────────────────

  build [--no-cache|--pull]
        Build the Docker images (varam, laxmi) from their Dockerfiles.
          --no-cache    Force a clean rebuild (ignore layer cache).
          --pull        Pull fresh base images before building.

  start | run
        Launch the full cluster in detached mode:
          1. potti   → PostgreSQL database
          2. varam   → Go REST API backend
          3. laxmi   → Next.js frontend
        Automatically runs pre-flight checks and waits for health probes.
        Prints access URLs on success.

  stop
        Gracefully halt all containers with SIGTERM (30s timeout).
        Containers are preserved in the 'exited' state — data intact.
        Resume instantly with:  ./docker-setup.sh start

  delete | down
        FULL teardown: removes containers, networks, and orphaned
        containers.  Persistent data volumes (potti_data, backend_storage)
        are EXPLICITLY PRESERVED.

  status
        Live audit table showing:
          • Container state & uptime
          • Port bindings (container → host)
          • Host port verification (ss -tlnp cross-reference)
          • Resource usage (CPU / memory / I/O)
          • Persistent volume inventory

  logs [--tail=N|--no-follow]
        Stream aggregated, colourised logs from all services.
          --tail=50     Show last 50 lines per service (default: 100).
          --no-follow   Print tail and exit without streaming.

  help | --help | -h
        Display this manual.

─── Application URLs ─────────────────────────────────────────────────────

  Frontend (laxmi)     ➜  http://localhost:3000
  Backend API (varam)  ➜  http://localhost:8080/api/health
  Swagger Docs         ➜  http://localhost:8080/swagger/index.html
  Database (potti)     ➜  postgresql://mycan@localhost:5432/mykanban

─── Setup Example ─────────────────────────────────────────────────────────

  # 1. Create your environment credentials
  cp .env.example .env
  # Edit .env and set a strong PG_PASSWORD

  # 2. Build images (first time or after code changes)
  ./docker-setup.sh build

  # 3. Launch everything
  ./docker-setup.sh start

  # 4. Check cluster health
  ./docker-setup.sh status

  # 5. Inspect logs
  ./docker-setup.sh logs

  # 6. Stop gracefully
  ./docker-setup.sh stop

  # 7. Full teardown (data preserved)
  ./docker-setup.sh delete

─── Files ─────────────────────────────────────────────────────────────────

  docker-compose.yml          Service definitions for potti, varam, laxmi
  .env                        Credentials and runtime overrides
  backend/config.json         Backend configuration (JWT secret, CORS, etc.)
  backend/Dockerfile          Multi-stage Go build → distroless runtime
  frontend/Dockerfile         Multi-stage Next.js build → slim Node runtime

─── Notes ─────────────────────────────────────────────────────────────────

  • All subcommands run pre-flight checks before execution.
  • Data volumes survive 'delete/down' — use 'docker compose down -v'
    explicitly if you need a full reset.
  • PostgreSQL is configured via .env — PG_PASSWORD is mandatory.
  • Port bindings are verified at runtime by the 'status' subcommand.

EOF
}

# ===========================================================================
# Main Command Dispatcher
#
# Routes the first argument to the appropriate subcommand function.
# Falls through to 'help' when no argument is provided.
#
# Aliases are supported:
#   run  →  start     (semantic alias)
#   down →  delete    (semantic alias)
# ===========================================================================
main() {
    # Capture the subcommand, lowercased for case-insensitive matching
    local CMD="${1:-help}"
    CMD="$(echo "${CMD}" | tr '[:upper:]' '[:lower:]')"
    # Shift off the subcommand so remaining $@ contains only flags
    shift 2>/dev/null || true

    # -------------------------------------------------------------------
    # Dispatch table
    # -------------------------------------------------------------------
    case "${CMD}" in
        # --- Build -----------------------------------------------------
        build)
            preflight_checks
            do_build "$@"
            ;;

        # --- Start / Run ----------------------------------------------
        start|run)
            preflight_checks
            do_start
            ;;

        # --- Stop -----------------------------------------------------
        stop)
            preflight_checks
            do_stop
            ;;

        # --- Delete / Down --------------------------------------------
        delete|down)
            preflight_checks
            do_delete
            ;;

        # --- Status ---------------------------------------------------
        status)
            # status can run with partial failures — we still do preflight,
            # but if docker compose isn't running that's not a hard error.
            preflight_checks 2>/dev/null || true
            do_status
            ;;

        # --- Logs -----------------------------------------------------
        logs)
            # Logs can show historical output even without running containers
            preflight_checks 2>/dev/null || true
            do_logs "$@"
            ;;

        # --- Help -----------------------------------------------------
        help|--help|-h|*)
            do_help
            ;;
    esac
}

# ---------------------------------------------------------------------------
# Entry Point — runs main() with all positional arguments
# ---------------------------------------------------------------------------
main "$@"