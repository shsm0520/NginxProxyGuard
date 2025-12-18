#!/bin/sh
# GeoIP Database Auto-Update Script for Nginx Proxy Guard
# Downloads and updates MaxMind GeoLite2 databases

GEOIP_DIR="/etc/nginx/geoip"
GEOIP_CONF="/etc/nginx/geoip/GeoIP.conf"
EDITION_IDS="GeoLite2-Country GeoLite2-ASN"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[GeoIP]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[GeoIP]${NC} $1"
}

log_error() {
    echo -e "${RED}[GeoIP]${NC} $1"
}

# Create GeoIP directory
mkdir -p "$GEOIP_DIR"

# Check if license key is provided
if [ -z "$MAXMIND_LICENSE_KEY" ]; then
    log_warn "MAXMIND_LICENSE_KEY not set. GeoIP features will be disabled."
    log_warn "Get a free license key at: https://www.maxmind.com/en/geolite2/signup"
    log_warn ""
    log_warn "To enable GeoIP:"
    log_warn "  1. Sign up at https://www.maxmind.com/en/geolite2/signup"
    log_warn "  2. Generate a license key in your MaxMind account"
    log_warn "  3. Set MAXMIND_LICENSE_KEY in Settings or environment"
    log_warn ""
    exit 0
fi

# Check if geoipupdate is available
if ! command -v geoipupdate >/dev/null 2>&1; then
    log_error "geoipupdate not found. Cannot download GeoIP databases."
    exit 1
fi

# Create GeoIP.conf for geoipupdate
log_info "Configuring GeoIP update..."
cat > "$GEOIP_CONF" << EOF
# GeoIP.conf for Nginx Proxy Guard
# Auto-generated - Do not edit manually

AccountID ${MAXMIND_ACCOUNT_ID:-0}
LicenseKey ${MAXMIND_LICENSE_KEY}
EditionIDs ${EDITION_IDS}
DatabaseDirectory ${GEOIP_DIR}
EOF

chmod 600 "$GEOIP_CONF"

# Download/Update databases
log_info "Downloading GeoIP databases..."
if geoipupdate -f "$GEOIP_CONF" -v; then
    log_info "GeoIP databases updated successfully!"

    # List downloaded files
    log_info "Downloaded databases:"
    ls -lh "$GEOIP_DIR"/*.mmdb 2>/dev/null || log_warn "No .mmdb files found"
else
    log_error "Failed to download GeoIP databases"
    log_warn "Check your MAXMIND_LICENSE_KEY and network connection"
    exit 1
fi

log_info "GeoIP setup complete!"
