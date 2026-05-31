#!/usr/bin/env bash
#
# create-sysChitta.sh - Creates a dedicated system user for running MyKanban services.
#
# This script performs the following actions:
# 1. Creates a system user 'apprunner' with no login shell.
# 2. Creates the storage directory for the 'chitta' service.
# 3. Sets exclusive ownership and permissions for the storage directory.
# 4. Provides instructions to grant Nginx control via a sudoers file.
#
set -euo pipefail

APP_USER="sysChitta"
CHITTA_DIR="/apps/chitta"
MAYA_DIR="/apps/maya"
CHITTA_STORAGE_DIR="$CHITTA_DIR/storage"

log_info() {
    printf "\033[1;34m[INFO]\033[0m %s\n" "$1"
}

log_success() {
    printf "\033[1;32m[SUCCESS]\033[0m %s\n" "$1"
}

log_warn() {
    printf "\033[1;33m[WARNING]\033[0m %s\n" "$1"
}

# --- 1. Create System User ---
if id "$APP_USER" &>/dev/null; then
    log_warn "User '$APP_USER' already exists. Skipping creation."
else
    log_info "Creating system user '$APP_USER'..."
    # --system: create a system user
    # --no-create-home: don't create a home directory
    # --shell /usr/sbin/nologin: prevent shell logins for security
    sudo useradd --system --no-create-home --shell /usr/sbin/nologin "$APP_USER"
    log_success "User '$APP_USER' created."
fi

# --- 2. Create and Secure App Directories ---
log_info "Configuring application directories..."
sudo mkdir -p "$CHITTA_STORAGE_DIR"
sudo mkdir -p "$MAYA_DIR"

log_info "Setting ownership for app directories..."
# Grant ownership of both app directories to the apprunner user
sudo chown -R "$APP_USER":"$APP_USER" "$CHITTA_DIR"
sudo chown -R "$APP_USER":"$APP_USER" "$MAYA_DIR"

# Set permissions for the maya directory to allow nginx to read files
# User: rwx, Group: r-x, Other: r-x
sudo chmod -R 755 "$MAYA_DIR"

# Set exclusive permissions for the chitta storage directory
log_info "Setting exclusive permissions for $CHITTA_STORAGE_DIR..."
sudo chmod -R 700 "$CHITTA_STORAGE_DIR"

log_success "Directory permissions configured."

# --- 3. Grant Nginx Control via Sudo ---
log_info "To grant Nginx control, please create a sudoers file with the following content."
echo "--------------------------------------------------------------------------------"
echo "Run this command to create the sudoers file:"
echo
echo "  sudo tee /etc/sudoers.d/apprunner-nginx > /dev/null <<'EOF'"
echo "# Allow the apprunner user to manage the nginx service"
echo "apprunner ALL=(ALL) NOPASSWD: /usr/bin/systemctl start nginx"
echo "apprunner ALL=(ALL) NOPASSWD: /usr/bin/systemctl stop nginx"
echo "apprunner ALL=(ALL) NOPASSWD: /usr/bin/systemctl restart nginx"
echo "apprunner ALL=(ALL) NOPASSWD: /usr/bin/systemctl status nginx"
echo "# Allow viewing nginx logs via journalctl"
echo "apprunner ALL=(ALL) NOPASSWD: /usr/bin/journalctl -u nginx"
echo "apprunner ALL=(ALL) NOPASSWD: /usr/bin/journalctl -u nginx -f"
echo "EOF"
echo
echo "And then set the correct permissions for the file:"
echo "  sudo chmod 440 /etc/sudoers.d/apprunner-nginx"
echo "--------------------------------------------------------------------------------"

log_success "Script finished. Manual sudoers configuration is required to complete setup."
