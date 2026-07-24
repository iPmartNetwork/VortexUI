#!/usr/bin/env bash
# VortexUI v1.4.1 — Deploy/Update Script (systemd)
# Used by: vortexui update (option 6 in CLI menu)
set -euo pipefail

VERSION="1.4.1"
INSTALL_DIR="${VORTEX_REPO_DIR:-/opt/vortexui}"
WEB_ROOT="${VORTEX_WEB_ROOT:-/var/www/vortexui}"
SERVICE="${VORTEX_SERVICE:-vortexui-panel}"
BRANCH="master"

SKIP_BACKEND=0
SKIP_FRONTEND=0
SKIP_MIGRATE=0

for arg in "$@"; do
    case "$arg" in
        --skip-backend)  SKIP_BACKEND=1 ;;
        --skip-frontend) SKIP_FRONTEND=1 ;;
        --skip-migrate)  SKIP_MIGRATE=1 ;;
    esac
done

cd "$INSTALL_DIR"

echo "==> VortexUI v${VERSION} update"
echo "==> git pull"
git pull origin "$BRANCH"

# Migrations
if [[ "$SKIP_MIGRATE" -eq 0 ]]; then
    echo "==> database migrations"
    if command -v goose >/dev/null 2>&1 && [[ -n "${VORTEX_DATABASE_URL:-}" ]]; then
        goose -dir migrations postgres "$VORTEX_DATABASE_URL" up
    else
        echo "    (auto-apply on startup)"
    fi
fi

# Frontend
if [[ "$SKIP_FRONTEND" -eq 0 ]]; then
    echo "==> building frontend"
    cd "$INSTALL_DIR/web"
    npm install --prefer-offline
    npm run build
    mkdir -p "$WEB_ROOT"
    if command -v rsync >/dev/null 2>&1; then
        rsync -a --delete dist/ "$WEB_ROOT/"
    else
        rm -rf "${WEB_ROOT:?}"/*
        cp -r dist/* "$WEB_ROOT/"
    fi
fi

# Backend
if [[ "$SKIP_BACKEND" -eq 0 ]]; then
    echo "==> building backend"
    cd "$INSTALL_DIR"

    # Install Go if not available
    if ! command -v go >/dev/null 2>&1; then
        echo "    Go not found, installing..."
        GO_VERSION="1.23.4"
        ARCH=$(uname -m)
        case "$ARCH" in
            x86_64)  GO_ARCH="amd64" ;;
            aarch64) GO_ARCH="arm64" ;;
            *)       echo "ERROR: Unsupported arch: $ARCH"; exit 1 ;;
        esac
        curl -sL "https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz" | tar -C /usr/local -xzf -
        export PATH="/usr/local/go/bin:$PATH"
        echo "    Go ${GO_VERSION} installed"
    fi

    CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" \
        -o /usr/local/bin/vortex-panel ./cmd/panel
fi

# UNMASK IF MASKED (this is the critical fix)
echo "==> checking service state"
STATE=$(systemctl is-enabled "$SERVICE" 2>/dev/null || true)
if [[ "$STATE" == "masked" ]]; then
    echo "==> $SERVICE is masked — unmasking..."
    systemctl unmask "$SERVICE"
fi

# Restart
echo "==> restarting $SERVICE"
systemctl daemon-reload
systemctl restart "$SERVICE"

# Caddy
if systemctl is-active --quiet caddy 2>/dev/null; then
    echo "==> reloading caddy"
    systemctl reload caddy
fi

echo ""
echo "==> Done! VortexUI v${VERSION} is running."
echo "    Status: systemctl status $SERVICE"
