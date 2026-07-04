#!/usr/bin/env bash
# Redeploys VortexUI on a native (non-Docker) install: pulls latest code,
# rebuilds the frontend, rebuilds the panel binary, and restarts services.
#
# Usage: sudo ./deploy.sh [--skip-backend]
set -euo pipefail

REPO_DIR="${VORTEX_REPO_DIR:-/opt/vortexui}"
WEB_ROOT="${VORTEX_WEB_ROOT:-/var/www/vortexui}"
SERVICE="${VORTEX_SERVICE:-vortexui-panel}"
SKIP_BACKEND=0
[[ "${1:-}" == "--skip-backend" ]] && SKIP_BACKEND=1

cd "$REPO_DIR"

echo "==> git pull"
git pull origin master

echo "==> building frontend"
cd "$REPO_DIR/web"
npm install
npm run build

echo "==> deploying static files to $WEB_ROOT"
mkdir -p "$WEB_ROOT"
cp -r dist/* "$WEB_ROOT/"

if [[ "$SKIP_BACKEND" -eq 0 ]]; then
  echo "==> building backend"
  cd "$REPO_DIR"
  go build -o /usr/local/bin/vortex-panel ./cmd/panel

  echo "==> restarting $SERVICE"
  systemctl restart "$SERVICE"
fi

echo "==> reloading caddy"
systemctl reload caddy

echo "==> done. built files:"
ls -la "$WEB_ROOT"/assets | head -5
