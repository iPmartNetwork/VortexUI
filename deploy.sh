#!/usr/bin/env bash
# Redeploys VortexUI on a native (non-Docker) install: pulls latest code,
# applies database migrations, rebuilds the frontend, rebuilds the panel binary,
# and restarts services.
#
# Usage: sudo ./deploy.sh [--skip-backend] [--skip-frontend] [--skip-migrate]
set -euo pipefail

VERSION="1.4.1"
REPO_DIR="${VORTEX_REPO_DIR:-/opt/vortexui}"
WEB_ROOT="${VORTEX_WEB_ROOT:-/var/www/vortexui}"
SERVICE="${VORTEX_SERVICE:-vortexui-panel}"
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

cd "$REPO_DIR"

echo "==> VortexUI v${VERSION} deploy"
echo "==> git pull"
git pull origin master

# --- Database Migrations ---
if [[ "$SKIP_MIGRATE" -eq 0 ]]; then
  echo "==> running database migrations"
  if command -v goose >/dev/null 2>&1 && [[ -n "${VORTEX_DATABASE_URL:-}" ]]; then
    goose -dir migrations postgres "$VORTEX_DATABASE_URL" up
  else
    echo "    (skipped: goose not found or VORTEX_DATABASE_URL not set)"
    echo "    Migrations will auto-apply on panel startup."
  fi
fi

# --- Frontend ---
if [[ "$SKIP_FRONTEND" -eq 0 ]]; then
  echo "==> building frontend"
  cd "$REPO_DIR/web"
  npm install --prefer-offline
  npm run build

  echo "==> deploying static files to $WEB_ROOT"
  mkdir -p "$WEB_ROOT"
  # Wipe old hashed assets first so stale bundles never linger and mask a fresh
  # deploy (Vite renames every build, cp alone would just keep piling files up).
  if command -v rsync >/dev/null 2>&1; then
    rsync -a --delete dist/ "$WEB_ROOT/"
  else
    rm -rf "${WEB_ROOT:?}"/*
    cp -r dist/* "$WEB_ROOT/"
  fi
fi

# --- Backend ---
if [[ "$SKIP_BACKEND" -eq 0 ]]; then
  echo "==> building backend"
  cd "$REPO_DIR"
  CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" \
    -o /usr/local/bin/vortex-panel ./cmd/panel

  echo "==> running doctor check"
  /usr/local/bin/vortex-panel doctor || true

  echo "==> restarting $SERVICE"
  systemctl restart "$SERVICE"
fi

# --- Caddy ---
if systemctl is-active --quiet caddy 2>/dev/null; then
  echo "==> reloading caddy"
  systemctl reload caddy
fi

echo ""
echo "==> VortexUI v${VERSION} deployed successfully!"
if [[ "$SKIP_FRONTEND" -eq 0 ]]; then
  echo "    Frontend assets:"
  ls -la "$WEB_ROOT"/assets/ 2>/dev/null | head -5
fi
echo "    Panel binary: $(vortex-panel --version 2>/dev/null || echo '/usr/local/bin/vortex-panel')"
echo ""
