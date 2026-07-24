#!/usr/bin/env bash
# VortexUI v1.4.1 — One-line installer
# Usage: bash <(curl -sL https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
#
# Requirements: Linux (amd64/arm64), Docker + Docker Compose v2, curl, git
# This script:
#   1. Checks system requirements
#   2. Creates /opt/vortexui directory
#   3. Clones the repository (or pulls if exists)
#   4. Generates mTLS certificates
#   5. Creates .env from example
#   6. Starts the stack with Docker Compose
#   7. Prints access URL

set -euo pipefail

VERSION="1.4.1"
REPO_URL="https://github.com/iPmartNetwork/VortexUI.git"
INSTALL_DIR="/opt/vortexui"
BRANCH="master"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

log()   { echo -e "${GREEN}[VortexUI]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

header() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════╗"
    echo "║          VortexUI v${VERSION} Installer              ║"
    echo "║   Next-Gen Proxy Management Panel               ║"
    echo "╚══════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        error "This script must be run as root. Use: sudo bash install.sh"
    fi
}

check_requirements() {
    log "Checking system requirements..."

    # Architecture
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)  ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        *)       error "Unsupported architecture: $ARCH (need amd64 or arm64)" ;;
    esac
    log "Architecture: $ARCH"

    # Docker
    if ! command -v docker &>/dev/null; then
        warn "Docker not found. Installing..."
        curl -fsSL https://get.docker.com | sh
        systemctl enable --now docker
    fi
    log "Docker: $(docker --version | awk '{print $3}')"

    # Docker Compose v2
    if ! docker compose version &>/dev/null; then
        error "Docker Compose v2 is required. Install with: apt install docker-compose-plugin"
    fi
    log "Docker Compose: $(docker compose version --short)"

    # Git
    if ! command -v git &>/dev/null; then
        apt-get update -qq && apt-get install -y -qq git
    fi
}

clone_or_pull() {
    if [[ -d "$INSTALL_DIR/.git" ]]; then
        log "Existing installation found. Updating to v${VERSION}..."
        cd "$INSTALL_DIR"
        git fetch origin
        git checkout "$BRANCH"
        git pull origin "$BRANCH"
        IS_UPDATE=1
    else
        log "Fresh install — cloning VortexUI repository..."
        git clone --depth 1 --branch "$BRANCH" "$REPO_URL" "$INSTALL_DIR"
        cd "$INSTALL_DIR"
        IS_UPDATE=0
    fi
    log "Repository ready at $INSTALL_DIR (v${VERSION})"
}

generate_secrets() {
    if [[ ! -f "$INSTALL_DIR/.env" ]]; then
        log "Generating configuration..."
        cp "$INSTALL_DIR/.env.example" "$INSTALL_DIR/.env"

        # Generate random secrets
        JWT_SECRET=$(openssl rand -hex 32)
        PANEL_SECRET=$(openssl rand -hex 32)
        DB_PASSWORD=$(openssl rand -base64 16 | tr -d '=/+')

        sed -i "s|^JWT_SECRET=.*|JWT_SECRET=${JWT_SECRET}|" "$INSTALL_DIR/.env"
        sed -i "s|^PANEL_SECRET=.*|PANEL_SECRET=${PANEL_SECRET}|" "$INSTALL_DIR/.env"
        sed -i "s|^POSTGRES_PASSWORD=.*|POSTGRES_PASSWORD=${DB_PASSWORD}|" "$INSTALL_DIR/.env"
        sed -i "s|^DATABASE_URL=.*|DATABASE_URL=postgres://vortex:${DB_PASSWORD}@postgres:5432/vortex?sslmode=disable|" "$INSTALL_DIR/.env"

        log "Secrets generated and saved to .env"
    else
        log "Existing .env found, keeping current configuration"
    fi
}

generate_certs() {
    if [[ ! -d "$INSTALL_DIR/deploy/certs" ]]; then
        log "Generating mTLS certificates..."
        cd "$INSTALL_DIR"
        # Use the built-in cert generator if Go is available, otherwise use openssl
        if command -v go &>/dev/null; then
            go run ./cmd/gencerts -out deploy/certs -san localhost,127.0.0.1
        else
            mkdir -p deploy/certs
            # Generate a self-signed CA + panel + node cert with openssl
            openssl req -x509 -newkey ec -pkeyopt ec_paramgen_curve:P-256 \
                -days 3650 -nodes -subj "/CN=VortexUI CA" \
                -keyout deploy/certs/ca.key -out deploy/certs/ca.crt 2>/dev/null
            openssl req -newkey ec -pkeyopt ec_paramgen_curve:P-256 -nodes \
                -subj "/CN=panel" -addext "subjectAltName=DNS:localhost,IP:127.0.0.1" \
                -keyout deploy/certs/panel.key -out deploy/certs/panel.csr 2>/dev/null
            openssl x509 -req -in deploy/certs/panel.csr -CA deploy/certs/ca.crt \
                -CAkey deploy/certs/ca.key -CAcreateserial -days 3650 \
                -copy_extensions copyall -out deploy/certs/panel.crt 2>/dev/null
            cp deploy/certs/panel.key deploy/certs/node.key
            cp deploy/certs/panel.crt deploy/certs/node.crt
            rm -f deploy/certs/panel.csr deploy/certs/ca.srl
        fi
        log "Certificates generated in deploy/certs/"
    else
        log "Certificates already exist, skipping"
    fi
}

start_stack() {
    log "Starting VortexUI stack..."
    cd "$INSTALL_DIR"

    # Apply database migrations (panel does this on startup too, but explicit is safer)
    log "Applying database migrations..."
    docker compose -f deploy/compose.yml up -d postgres redis
    sleep 3
    docker compose -f deploy/compose.yml run --rm panel ./panel doctor || true

    docker compose -f deploy/compose.yml up --build -d

    log "Waiting for services to start..."
    sleep 5

    # Health check
    for i in {1..30}; do
        if curl -sf http://localhost:8080/api/health &>/dev/null; then
            break
        fi
        sleep 2
    done
}

print_success() {
    # Detect public IP
    PUBLIC_IP=$(curl -sf https://api.ipify.org || hostname -I | awk '{print $1}')

    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════════════════╗${NC}"
    if [[ "${IS_UPDATE:-0}" -eq 1 ]]; then
    echo -e "${GREEN}║      VortexUI updated to v${VERSION}!               ║${NC}"
    else
    echo -e "${GREEN}║        VortexUI v${VERSION} installed!               ║${NC}"
    fi
    echo -e "${GREEN}╚══════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "  ${BLUE}Panel URL:${NC}    http://${PUBLIC_IP}:8080"
    echo -e "  ${BLUE}Install Dir:${NC}  ${INSTALL_DIR}"
    echo -e "  ${BLUE}Config:${NC}       ${INSTALL_DIR}/.env"
    echo -e "  ${BLUE}Logs:${NC}         docker compose -f ${INSTALL_DIR}/deploy/compose.yml logs -f"
    echo ""
    if [[ "${IS_UPDATE:-0}" -eq 0 ]]; then
    echo -e "  ${YELLOW}First run:${NC} Create an admin account:"
    echo -e "    cd ${INSTALL_DIR} && docker compose -f deploy/compose.yml exec panel ./panel admin create"
    echo ""
    fi
    echo -e "  ${CYAN}Commands:${NC}"
    echo -e "    Stop:     docker compose -f ${INSTALL_DIR}/deploy/compose.yml down"
    echo -e "    Start:    docker compose -f ${INSTALL_DIR}/deploy/compose.yml up -d"
    echo -e "    Update:   bash <(curl -sL https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)"
    echo -e "    Doctor:   docker compose -f ${INSTALL_DIR}/deploy/compose.yml exec panel ./panel doctor"
    echo -e "    Backup:   docker compose -f ${INSTALL_DIR}/deploy/compose.yml exec panel ./panel backup create"
    echo -e "    CLI Help: docker compose -f ${INSTALL_DIR}/deploy/compose.yml exec panel ./panel --help"
    echo ""
}

# --- Main ---
header
check_root
check_requirements
clone_or_pull
generate_secrets
generate_certs
start_stack
print_success
