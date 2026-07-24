#!/usr/bin/env bash
# ───────────────────────────────────────────────────────────────────────────────
# VortexUI Bridge Installer — Installs a single tunnel backend on an Iran VPS
# and runs it as a systemd service. Designed to be executed via a one-liner from
# the VortexUI panel (base64-encoded config piped through curl | bash).
#
# Supported backends: backhaul, rathole, wstunnel
#
# Usage:
#   vortex-bridge.sh --backend <backhaul|rathole|wstunnel> \
#     [--exec-b64 <base64_binary>] \--config-b64 <base64_config> \
#     [--config-path /etc/vortex-bridge/<name>.toml] \
#     [--cert <base64_cert>] [--port <listen_port>]
# ───────────────────────────────────────────────────────────────────────────────
set -euo pipefail

# ─── Defaults ─────────────────────────────────────────────────────────────────
BACKEND=""
EXEC_B64=""
CONFIG_B64=""
CONFIG_PATH=""
CERT_B64=""
PORT=""
INSTALL_DIR="/opt/vortex-bridge"
SERVICE_NAME="vortex-bridge"

# ─── Colors ───────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()  { echo -e "${GREEN}[vortex-bridge]${NC} $*"; }
warn() { echo -e "${YELLOW}[vortex-bridge]${NC} $*"; }
die()  { echo -e "${RED}[vortex-bridge] ERROR:${NC} $*" >&2; exit 1; }

# ─── Parse Arguments ──────────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case "$1" in
    --backend)     BACKEND="$2"; shift 2 ;;
    --exec-b64)    EXEC_B64="$2"; shift 2 ;;
    --config-b64)  CONFIG_B64="$2"; shift 2 ;;
    --config-path) CONFIG_PATH="$2"; shift 2 ;;
    --cert)        CERT_B64="$2"; shift 2 ;;
    --port)        PORT="$2"; shift 2 ;;
    *) die "Unknown argument: $1" ;;
  esac
done

[[ -z "$BACKEND" ]] && die "--backend is required (backhaul|rathole|wstunnel)"
[[ -z "$CONFIG_B64" ]] && die "--config-b64 is required"

# Validate backend
case "$BACKEND" in
  backhaul|rathole|wstunnel) ;;
  *) die "Unsupported backend: $BACKEND. Use backhaul, rathole, or wstunnel." ;;
esac

# ─── Detect Architecture ──────────────────────────────────────────────────────
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH_TAG="amd64" ;;
  aarch64) ARCH_TAG="arm64" ;;
  armv7l)  ARCH_TAG="armv7" ;;
  *) die "Unsupported architecture: $ARCH" ;;
esac

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
[[ "$OS" != "linux" ]] && die "This script only supports Linux."

# ─── Install Prerequisites ────────────────────────────────────────────────────
install_deps() {
  if command -v apt-get &>/dev/null; then
    apt-get update -qq && apt-get install -y -qq curl wget tar jq >/dev/null 2>&1
  elif command -v yum &>/dev/null; then
    yum install -y curl wget tar jq >/dev/null 2>&1
  fi
}

log "Installing prerequisites..."
install_deps

# ─── Create directories ──────────────────────────────────────────────────────
mkdir -p "$INSTALL_DIR"
mkdir -p /etc/vortex-bridge

# Set default config path if not provided
[[ -z "$CONFIG_PATH" ]] && CONFIG_PATH="/etc/vortex-bridge/${BACKEND}.toml"

# ─── Download / Install Backend Binary ────────────────────────────────────────
install_backhaul() {
  log "Installing Backhaul..."
  local version
  version=$(curl -sL "https://api.github.com/repos/Musixal/Backhaul/releases/latest" | jq -r '.tag_name' 2>/dev/null || echo "v0.6.5")
  local url="https://github.com/Musixal/Backhaul/releases/download/${version}/backhaul_linux_${ARCH_TAG}.tar.gz"
  log "Downloading ${url}..."
  wget -qO /tmp/backhaul.tar.gz "$url" || die "Failed to download backhaul"
  tar -xzf /tmp/backhaul.tar.gz -C "$INSTALL_DIR/" 2>/dev/null || tar -xzf /tmp/backhaul.tar.gz -C "$INSTALL_DIR/"
  chmod +x "$INSTALL_DIR/backhaul" 2>/dev/null || chmod +x "$INSTALL_DIR/Backhaul" 2>/dev/null
  # Normalize binary name
  [[ -f "$INSTALL_DIR/Backhaul" ]] && mv "$INSTALL_DIR/Backhaul" "$INSTALL_DIR/backhaul"
  rm -f /tmp/backhaul.tar.gz
}

install_rathole() {
  log "Installing Rathole..."
  local version
  version=$(curl -sL "https://api.github.com/repos/rapiz1/rathole/releases/latest" | jq -r '.tag_name' 2>/dev/null || echo "v0.5.0")
  local target
  case "$ARCH_TAG" in
    amd64) target="x86_64-unknown-linux-gnu" ;;
    arm64) target="aarch64-unknown-linux-gnu" ;;
    armv7) target="armv7-unknown-linux-gnueabihf" ;;
  esac
  local url="https://github.com/rapiz1/rathole/releases/download/${version}/rathole-${target}.zip"
  log "Downloading ${url}..."
  wget -qO /tmp/rathole.zip "$url" || die "Failed to download rathole"
  if command -v unzip &>/dev/null; then
    unzip -o /tmp/rathole.zip -d "$INSTALL_DIR/" >/dev/null
  else
    apt-get install -y -qq unzip >/dev/null 2>&1 || yum install -y unzip >/dev/null 2>&1
    unzip -o /tmp/rathole.zip -d "$INSTALL_DIR/" >/dev/null
  fi
  chmod +x "$INSTALL_DIR/rathole"
  rm -f /tmp/rathole.zip
}

install_wstunnel() {
  log "Installing wstunnel..."
  local version
  version=$(curl -sL "https://api.github.com/repos/erebe/wstunnel/releases/latest" | jq -r '.tag_name' 2>/dev/null || echo "v10.1.0")
  local target
  case "$ARCH_TAG" in
    amd64) target="x86_64-unknown-linux-gnu" ;;
    arm64) target="aarch64-unknown-linux-gnu" ;;
    armv7) target="armv7-unknown-linux-gnueabihf" ;;
  esac
  local url="https://github.com/erebe/wstunnel/releases/download/${version}/wstunnel_${version#v}_linux_${ARCH_TAG}.tar.gz"
  log "Downloading ${url}..."
  wget -qO /tmp/wstunnel.tar.gz "$url" || die "Failed to download wstunnel"
  tar -xzf /tmp/wstunnel.tar.gz -C "$INSTALL_DIR/" 2>/dev/null || true
  chmod +x "$INSTALL_DIR/wstunnel"
  rm -f /tmp/wstunnel.tar.gz
}

# If binary provided as base64, decode it; otherwise download from GitHub
if [[ -n "$EXEC_B64" ]]; then
  log "Decoding provided binary..."
  echo "$EXEC_B64" | base64 -d > "$INSTALL_DIR/$BACKEND"
  chmod +x "$INSTALL_DIR/$BACKEND"
else
  case "$BACKEND" in
    backhaul) install_backhaul ;;
    rathole)  install_rathole ;;
    wstunnel) install_wstunnel ;;
  esac
fi

# ─── Write Configuration ──────────────────────────────────────────────────────
log "Writing config to ${CONFIG_PATH}..."
echo "$CONFIG_B64" | base64 -d > "$CONFIG_PATH"

# ─── Write Certificate (optional) ────────────────────────────────────────────
if [[ -n "$CERT_B64" ]]; then
  log "Writing TLS certificate..."
  echo "$CERT_B64" | base64 -d > /etc/vortex-bridge/cert.pem
fi

# ─── Build ExecStart Command ──────────────────────────────────────────────────
build_exec_start() {
  case "$BACKEND" in
    backhaul)
      echo "$INSTALL_DIR/backhaul -c $CONFIG_PATH"
      ;;
    rathole)
      echo "$INSTALL_DIR/rathole $CONFIG_PATH"
      ;;
    wstunnel)
      # wstunnel uses command-line flags; the config file contains the full command
      # If config is a TOML/YAML, pass as --config; otherwise treat as args file
      echo "$INSTALL_DIR/wstunnel server --config-file $CONFIG_PATH"
      ;;
  esac
}

EXEC_START=$(build_exec_start)

# ─── Create systemd Service ──────────────────────────────────────────────────
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

log "Creating systemd service: ${SERVICE_NAME}..."
cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=VortexUI Bridge Tunnel (${BACKEND})
After=network.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=${EXEC_START}
Restart=always
RestartSec=5
LimitNOFILE=65535
StandardOutput=journal
StandardError=journal
SyslogIdentifier=${SERVICE_NAME}

[Install]
WantedBy=multi-user.target
EOF

# ─── Enable & Start ──────────────────────────────────────────────────────────
log "Reloading systemd and starting ${SERVICE_NAME}..."
systemctl daemon-reload
systemctl enable "$SERVICE_NAME"
systemctl restart "$SERVICE_NAME"

# ─── Verify ──────────────────────────────────────────────────────────────────
sleep 2
if systemctl is-active --quiet "$SERVICE_NAME"; then
  log "Service ${SERVICE_NAME} is running."
  echo ""
  echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
  echo -e "${GREEN}  VortexUI Bridge installed successfully!${NC}"
  echo -e "${GREEN}  Backend:  ${BACKEND}${NC}"
  echo -e "${GREEN}  Config:   ${CONFIG_PATH}${NC}"
  echo -e "${GREEN}  Service:  ${SERVICE_NAME}${NC}"
  [[ -n "$PORT" ]] && echo -e "${GREEN}  Port:     ${PORT}${NC}"
  echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
else
  warn "Service may have failed to start. Check logs with:"
  echo "  journalctl -u ${SERVICE_NAME} --no-pager -n 30"
  exit 1
fi
