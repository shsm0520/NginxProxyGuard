#!/bin/sh
set -e

CERT_DIR="/etc/nginx/ssl"
CERT_FILE="$CERT_DIR/server.crt"
KEY_FILE="$CERT_DIR/server.key"

# Create SSL directory if not exists
mkdir -p "$CERT_DIR"

# Generate 100-year self-signed certificate if not exists
if [ ! -f "$CERT_FILE" ] || [ ! -f "$KEY_FILE" ]; then
    echo "Generating 100-year self-signed SSL certificate for Nginx Proxy Guard UI..."

    openssl req -x509 -nodes -days 36500 \
        -newkey rsa:2048 \
        -keyout "$KEY_FILE" \
        -out "$CERT_FILE" \
        -subj "/C=KR/ST=Seoul/L=Seoul/O=Nginx Proxy Guard/OU=Security/CN=nginx-proxy-guard" \
        -addext "subjectAltName=DNS:localhost,DNS:nginx-proxy-guard,IP:127.0.0.1"

    chmod 644 "$CERT_FILE"
    chmod 600 "$KEY_FILE"

    echo "SSL certificate generated successfully!"
    echo "  - Certificate: $CERT_FILE"
    echo "  - Private Key: $KEY_FILE"
    echo "  - Validity: 100 years (36500 days)"
fi

# Start nginx
exec nginx -g 'daemon off;'
