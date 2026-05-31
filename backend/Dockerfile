# =============================================================================
# Dockerfile - MyKanban Backend (Varam)
#
# Multi-stage build: compile the Go binary in a full toolchain image, then
# copy it into a tiny distroless runtime for a minimal attack surface.
#
# Service name (Compose):  varam
# Exposed port:            8080
# =============================================================================

# ---------------------------------------------------------------------------
# Stage 1 — Builder
# ---------------------------------------------------------------------------
FROM golang:1.22-alpine AS builder

# Install CA certificates for any HTTPS calls the binary might make
RUN apk add --no-cache ca-certificates

WORKDIR /src

# Leverage Docker layer caching: copy dependency files first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source tree
COPY . .

# Build a statically-linked binary with debug symbols stripped
#   CGO_ENABLED=0  → fully static binary (no libc dependency)
#   -ldflags …     → strip DWARF & symbol table (smaller image)
RUN CGO_ENABLED=0 \
    go build \
        -ldflags="-w -s" \
        -o /bin/tracker-server \
        .

# ---------------------------------------------------------------------------
# Stage 2 — Runtime (distroless)
# ---------------------------------------------------------------------------
FROM gcr.io/distroless/base-debian12:nonroot

# Copy the compiled binary from the builder stage
COPY --from=builder /bin/tracker-server /bin/tracker-server

# Copy CA certs so HTTPS / TLS works inside the container
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# The application data directory — a volume mount point for persistence
RUN mkdir -p /data/storage

# Metadata
EXPOSE 8080

# Default config path; Compose overrides this via environment or bind-mount
ENV CONFIG_PATH=/etc/mycan/config.json

ENTRYPOINT ["/bin/tracker-server"]