#!/usr/bin/env bash
# deploy.sh — deploy u to your Vultr server
# Usage: ./deploy.sh <ssh-host>
# Example: ./deploy.sh root@45.63.7.246

set -euo pipefail

HOST="${1:?Usage: ./deploy.sh <ssh-host>}"

echo "==> Building binary for linux/amd64..."
GOOS=linux GOARCH=amd64 go build -o u .

echo "==> Uploading binary and configs..."
scp u u.service u.fftp.io.nginx "$HOST":/tmp/

echo "==> Setting up on server..."
ssh "$HOST" bash -s <<'REMOTE'
set -euo pipefail

# Create directories
mkdir -p /opt/u /var/www/uploads
chown www-data:www-data /var/www/uploads

# Generate API key if not already set
if [ ! -f /etc/u.env ]; then
    KEY=$(openssl rand -hex 32)
    echo "UPLOAD_API_KEY=$KEY" > /etc/u.env
    chmod 600 /etc/u.env
    echo "✓ Generated API key: $KEY"
    echo "  (save this — it won't be shown again)"
else
    echo "✓ /etc/u.env already exists, keeping existing key"
fi

# Install binary
mv /tmp/u /opt/u/u
chmod +x /opt/u/u

# Install systemd service
mv /tmp/u.service /etc/systemd/system/u.service
systemctl daemon-reload
systemctl enable u
systemctl restart u
echo "✓ u service started"

# Install nginx config
mv /tmp/u.fftp.io.nginx /etc/nginx/sites-available/u.fftp.io
ln -sf /etc/nginx/sites-available/u.fftp.io /etc/nginx/sites-enabled/u.fftp.io

if nginx -t 2>&1; then
    systemctl reload nginx
    echo "✓ Nginx reloaded"
else
    echo "✗ Nginx config test failed!"
    exit 1
fi

# Set up SSL with Certbot
if command -v certbot &>/dev/null; then
    echo "==> Running certbot for u.fftp.io..."
    certbot --nginx -d u.fftp.io --non-interactive --agree-tos --redirect
    echo "✓ SSL configured"
else
    echo "⚠ Certbot not found. Install with:"
    echo "  apt install -y certbot python3-certbot-nginx"
    echo "  Then run: certbot --nginx -d u.fftp.io"
fi

echo "==> Done!"
REMOTE

# Cleanup local binary
rm -f u
echo "==> Deploy complete."
