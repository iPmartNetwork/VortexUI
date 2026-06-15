#!/usr/bin/env bash
#
# VortexUI installer — provisions the full stack (web UI + panel + node +
# PostgreSQL/TimescaleDB + Redis) on a fresh Linux host using Docker Compose.
#
#   bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
#
# Re-running is safe: it pulls the latest code and recreates the stack without
# touching existing data or credentials.
set -euo pipefail

REPO_URL="${VORTEXUI_REPO:-https://github.com/iPmartNetwork/VortexUI}"
BRANCH="${VORTEXUI_BRANCH:-master}"
INSTALL_DIR="${VORTEXUI_DIR:-/opt/vortexui}"
WEB_PORT="${VORTEXUI_WEB_PORT:-80}"

c_green=$'\e[32m'; c_blue=$'\e[34m'; c_yellow=$'\e[33m'; c_red=$'\e[31m'; c_reset=$'\e[0m'
info()  { echo "${c_blue}==>${c_reset} $*"; }
ok()    { echo "${c_green}✓${c_reset} $*"; }
warn()  { echo "${c_yellow}!${c_reset} $*"; }
die()   { echo "${c_red}✗ $*${c_reset}" >&2; exit 1; }

[ "$(id -u)" -eq 0 ] || die "please run as root (sudo)."

# --- 1. Dependencies: Docker + Compose plugin ---------------------------------
if ! command -v docker >/dev/null 2>&1; then
  info "installing Docker…"
  curl -fsSL https://get.docker.com | sh
  systemctl enable --now docker || true
fi
docker compose version >/dev/null 2>&1 || die "Docker Compose plugin not found. Install Docker Compose v2."
command -v git >/dev/null 2>&1 || { info "installing git…"; (apt-get update -y && apt-get install -y git) || yum install -y git || apk add git; }
ok "Docker and git are present."

# --- 2. Source checkout -------------------------------------------------------
if [ -d "$INSTALL_DIR/.git" ]; then
  info "updating existing checkout in $INSTALL_DIR…"
  git -C "$INSTALL_DIR" fetch --depth 1 origin "$BRANCH"
  git -C "$INSTALL_DIR" checkout -q "$BRANCH"
  git -C "$INSTALL_DIR" reset --hard "origin/$BRANCH"
else
  info "cloning $REPO_URL ($BRANCH) into $INSTALL_DIR…"
  git clone --depth 1 --branch "$BRANCH" "$REPO_URL" "$INSTALL_DIR"
fi
cd "$INSTALL_DIR"
ok "source ready."

# --- 3. Access mode: domain + automatic SSL, or IP-only HTTP ------------------
SITE_ADDRESS=":${WEB_PORT}"
ACME_EMAIL=""
if [ -z "${VORTEXUI_NONINTERACTIVE:-}" ]; then
  echo
  echo "  ${b}How should the panel be reached?${n}"
  echo "   ${b}1)${n} Domain with automatic HTTPS (Let's Encrypt)  ${d}— recommended${n}"
  echo "   ${b}2)${n} IP address, plain HTTP on a port"
  read -r -p "  choose [1/2]: " mode
  if [ "$mode" = "1" ]; then
    read -r -p "  domain/subdomain (e.g. panel.example.com): " DOMAIN
    [ -n "$DOMAIN" ] || die "a domain is required for HTTPS mode."
    read -r -p "  email for Let's Encrypt (optional): " ACME_EMAIL
    SITE_ADDRESS="$DOMAIN"
    warn "point $DOMAIN's DNS A record to this server and open ports 80 + 443."
  else
    read -r -p "  HTTP port [${WEB_PORT}]: " p
    WEB_PORT="${p:-$WEB_PORT}"
    SITE_ADDRESS=":${WEB_PORT}"
  fi
fi

# --- 4. Environment + secrets -------------------------------------------------
ENV_FILE="deploy/.env"
if [ ! -f "$ENV_FILE" ]; then
  info "generating secrets…"
  JWT_SECRET="$(openssl rand -hex 32 2>/dev/null || head -c32 /dev/urandom | xxd -p -c64)"
  DB_PASSWORD="$(openssl rand -hex 16 2>/dev/null || head -c16 /dev/urandom | xxd -p -c64)"
  cat > "$ENV_FILE" <<EOF
JWT_SECRET=$JWT_SECRET
DB_PASSWORD=$DB_PASSWORD
WEB_PORT=$WEB_PORT
SITE_ADDRESS=$SITE_ADDRESS
ACME_EMAIL=$ACME_EMAIL
EOF
  chmod 600 "$ENV_FILE"
  ok "wrote $ENV_FILE (keep it safe)."
else
  warn "$ENV_FILE exists — updating access settings, keeping secrets."
  # refresh access-mode keys without touching JWT/DB secrets
  sed -i "/^WEB_PORT=/d;/^SITE_ADDRESS=/d;/^ACME_EMAIL=/d" "$ENV_FILE"
  printf 'WEB_PORT=%s\nSITE_ADDRESS=%s\nACME_EMAIL=%s\n' "$WEB_PORT" "$SITE_ADDRESS" "$ACME_EMAIL" >> "$ENV_FILE"
fi

# --- 4. mTLS chain for the panel↔node hub -------------------------------------
PUBLIC_HOST="$(curl -fsS https://api.ipify.org 2>/dev/null || hostname -I 2>/dev/null | awk '{print $1}' || echo 127.0.0.1)"
if [ ! -f deploy/certs/ca.crt ]; then
  info "generating mTLS certificates (SAN: localhost,127.0.0.1,$PUBLIC_HOST)…"
  docker run --rm -v "$PWD":/src -w /src golang:1.26-alpine \
    go run ./cmd/gencerts -out deploy/certs -san "localhost,127.0.0.1,$PUBLIC_HOST"
  ok "certificates written to deploy/certs."
else
  warn "deploy/certs already present — reusing."
fi

COMPOSE="docker compose --env-file $ENV_FILE -f deploy/compose.yml"

# --- 5. Build + start ---------------------------------------------------------
info "building and starting the stack (first run pulls images, be patient)…"
$COMPOSE up -d --build

info "waiting for the panel to become healthy…"
for _ in $(seq 1 60); do
  if $COMPOSE exec -T panel /usr/local/bin/panel admin create --help >/dev/null 2>&1; then break; fi
  sleep 2
done

# --- 6. Bootstrap admin (only if none requested before) -----------------------
ADMIN_USER="${VORTEXUI_ADMIN_USER:-admin}"
ADMIN_PASS="${VORTEXUI_ADMIN_PASS:-$(openssl rand -hex 8 2>/dev/null || echo changeme$RANDOM)}"
if [ ! -f deploy/.admin-created ]; then
  info "creating the initial admin account…"
  if $COMPOSE exec -T panel /usr/local/bin/panel admin create --username "$ADMIN_USER" --password "$ADMIN_PASS" --sudo; then
    touch deploy/.admin-created
    ok "admin '$ADMIN_USER' created."
  else
    warn "admin creation failed (it may already exist) — create one with: vortexui admin create --username U --password P --sudo"
    ADMIN_PASS="(unchanged)"
  fi
else
  warn "admin already bootstrapped on a previous run."
  ADMIN_PASS="(unchanged)"
fi

# --- 7. Management CLI ---------------------------------------------------------
install -m 0755 scripts/vortexui /usr/local/bin/vortexui 2>/dev/null && ok "installed 'vortexui' management command."

case "$SITE_ADDRESS" in
  :*) ACCESS_URL="http://$PUBLIC_HOST:$WEB_PORT" ;;
  *)  ACCESS_URL="https://$SITE_ADDRESS" ;;
esac

echo
ok "VortexUI is up."
echo "   ${c_blue}URL:${c_reset}      $ACCESS_URL"
echo "   ${c_blue}Username:${c_reset} $ADMIN_USER"
echo "   ${c_blue}Password:${c_reset} $ADMIN_PASS"
echo
echo "   Manage with: ${c_green}vortexui {start|stop|restart|status|logs|update|admin|uninstall}${c_reset}"
