#!/usr/bin/env bash
#
# chitta-certbot.sh - Provisions SSL certificate for the chitta service.
#
# This script runs Certbot to obtain a certificate for chitta.jaganathan.co.uk
# and automatically configures the corresponding Nginx server block.
#
set -euo pipefail

DOMAIN="chitta.jaganathan.co.uk"
EMAIL="admin@jaganathan.co.uk" # CHANGE THIS to a valid email address

log_info() {
    printf "\033[1;34m[INFO]\033[0m %s\n" "$1"
}

if ! command -v certbot >/dev/null 2>&1; then
    printf "\033[1;31m[ERROR]\033[0m Certbot is not installed. Please install it first.\n"
    exit 1
fi

log_info "Requesting SSL certificate for $DOMAIN..."
log_info "You may be prompted for your sudo password."

# The --nginx plugin will automatically find and modify the correct Nginx config
sudo certbot --nginx \
    -d "$DOMAIN" \
    --non-interactive \
    --agree-tos \
    -m "$EMAIL" \
    --redirect

log_info "Certificate provisioning for $DOMAIN complete."
log_info "To test renewal, run: sudo certbot renew --dry-run"
