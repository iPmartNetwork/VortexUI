#!/usr/bin/env bash
# VortexUI v1.4.1 — Unified Smart Deploy Script
# Detects Docker vs systemd mode and executes the appropriate deployment path.
#
# Usage:
#   sudo ./setup.sh [--docker|--systemd] [--skip-backend] [--skip-frontend] [--skip-migrate]
#
# Mode detection priority:
#   1. Explicit flag (--docker or --systemd)
#   2. Running VortexUI Docker containers detected
#   3. Existing systemd service file found
#   4. Docker available → Docker mode (default for fresh install)
set -euo pipefail

VERSION="1.4.1"
REPO_URL="https://github.com/iPmartNetwork/VortexUI.git"
INSTALL_DIR="${VORTEX_REPO_DIR:-/opt/vortexui}"
WEB_ROOT="${VORTEX_WEB_ROOT:-/var/www/vortexui}"
SERVICE="${VORTEX_SERVICE:-vortexui-panel}"
BRANCH="master"

MODE=""
SKIP_BACKEND=0
SKIP_FRONTEND=0
SKIP_MIGRATE=0
USER_DOMAIN=""
WEB_PORT="80"
INSTALL_TYPE=""

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

pass()  { echo -e "  ${GREEN}[✓]${NC} $1"; }
fail()  { echo -e "  ${RED}[✗]${NC} $1"; }
info()  { echo -e "  ${BLUE}[•]${NC} $1"; }

header() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════╗"
    echo "║        VortexUI v${VERSION} — Smart Deploy           ║"
    echo "║   Next-Gen Proxy Management Panel               ║"
    echo "╚══════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# --- Argument Parsing ---

parse_args() {
    for arg in "$@"; do
        case "$arg" in
            --docker)        MODE="docker" ;;
            --systemd)       MODE="systemd" ;;
            --skip-backend)  SKIP_BACKEND=1 ;;
            --skip-frontend) SKIP_FRONTEND=1 ;;
            --skip-migrate)  SKIP_MIGRATE=1 ;;
            --help|-h)
                echo "Usage: sudo ./setup.sh [--docker|--systemd] [--skip-backend] [--skip-frontend] [--skip-migrate]"
                echo ""
                echo "Modes:"
                echo "  --docker     Force Docker Compose deployment"
                echo "  --systemd    Force native systemd deployment"
                echo "  (omit)       Auto-detect based on environment"
                echo ""
                echo "Flags:"
                echo "  --skip-backend   Skip Go backend build (systemd mode)"
                echo "  --skip-frontend  Skip frontend build"
                echo "  --skip-migrate   Skip database migrations"
                exit 0
                ;;
            *)
                warn "Unknown argument: $arg"
                ;;
        esac
    done
}

# --- Mode Detection ---

detect_mode() {
    if [[ -n "$MODE" ]]; then
        log "Mode: $MODE (explicit flag)"
        return
    fi

    # Check for running VortexUI Docker containers
    if command -v docker &>/dev/null && docker ps --format '{{.Names}}' 2>/dev/null | grep -qi "vortex"; then
        MODE="docker"
        log "Mode: docker (running containers detected)"
        return
    fi

    # Check for existing systemd service
    if systemctl list-unit-files "${SERVICE}.service" &>/dev/null 2>&1; then
        MODE="systemd"
        log "Mode: systemd (service file found)"
        return
    fi

    # Default: Docker if available
    if command -v docker &>/dev/null && docker compose version &>/dev/null 2>&1; then
        MODE="docker"
        log "Mode: docker (Docker available, fresh install)"
        return
    fi

    error "Cannot determine deployment mode. Use --docker or --systemd explicitly."
}

# --- Pre-flight Doctor Check ---

doctor_check() {
    local phase="$1"
    local failures=0

    echo ""
    log "Running ${phase} health checks..."

    # Check disk space (need at least 1GB free)
    local free_kb
    free_kb=$(df -k / | tail -1 | awk '{print $4}')
    if [[ "$free_kb" -lt 1048576 ]]; then
        fail "Disk space: $(( free_kb / 1024 ))MB free (need ≥1GB)"
        failures=$((failures + 1))
    else
        pass "Disk space: $(( free_kb / 1024 ))MB free"
    fi

    # Check if required ports are available (or in use by us)
    if command -v ss &>/dev/null; then
        local port_8080
        port_8080=$(ss -tlnp 2>/dev/null | grep ":8080 " || true)
        if [[ -n "$port_8080" ]]; then
            if echo "$port_8080" | grep -q "vortex\|panel\|docker"; then
                pass "Port 8080: in use by VortexUI"
            else
                warn "Port 8080: in use by another process"
            fi
        else
            pass "Port 8080: available"
        fi
    fi

    # Check DNS resolution
    if command -v nslookup &>/dev/null; then
        if nslookup github.com &>/dev/null 2>&1; then
            pass "DNS resolution: OK"
        else
            fail "DNS resolution: cannot resolve github.com"
            failures=$((failures + 1))
        fi
    elif command -v dig &>/dev/null; then
        if dig +short github.com &>/dev/null 2>&1; then
            pass "DNS resolution: OK"
        else
            fail "DNS resolution: cannot resolve github.com"
            failures=$((failures + 1))
        fi
    else
        info "DNS resolution: cannot verify (no nslookup/dig)"
    fi

    # Check TLS certificates if they exist
    if [[ -f "$INSTALL_DIR/deploy/certs/panel.crt" ]]; then
        local expiry
        expiry=$(openssl x509 -enddate -noout -in "$INSTALL_DIR/deploy/certs/panel.crt" 2>/dev/null | cut -d= -f2)
        if [[ -n "$expiry" ]]; then
            local expiry_epoch
            expiry_epoch=$(date -d "$expiry" +%s 2>/dev/null || date -j -f "%b %d %T %Y %Z" "$expiry" +%s 2>/dev/null || echo 0)
            local now_epoch
            now_epoch=$(date +%s)
            local days_left=$(( (expiry_epoch - now_epoch) / 86400 ))
            if [[ "$days_left" -lt 0 ]]; then
                fail "TLS certificate: EXPIRED ($expiry)"
                failures=$((failures + 1))
            elif [[ "$days_left" -lt 30 ]]; then
                warn "TLS certificate: expires in ${days_left} days"
            else
                pass "TLS certificate: valid (expires in ${days_left} days)"
            fi
        fi
    fi

    # Docker-specific checks
    if [[ "$MODE" == "docker" ]]; then
        if command -v docker &>/dev/null && docker info &>/dev/null 2>&1; then
            pass "Docker daemon: running"
        else
            fail "Docker daemon: not running"
            failures=$((failures + 1))
        fi
    fi

    # Database check (if DATABASE_URL or VORTEX_DATABASE_URL is set)
    local db_url="${VORTEX_DATABASE_URL:-${DATABASE_URL:-}}"
    if [[ -n "$db_url" ]] && command -v pg_isready &>/dev/null; then
        if pg_isready -d "$db_url" &>/dev/null 2>&1; then
            pass "Database: reachable"
        else
            fail "Database: unreachable"
            failures=$((failures + 1))
        fi
    fi

    # Redis check (if REDIS_URL is set)
    local redis_url="${VORTEX_REDIS_URL:-${REDIS_URL:-}}"
    if [[ -n "$redis_url" ]] && command -v redis-cli &>/dev/null; then
        if redis-cli -u "$redis_url" ping &>/dev/null 2>&1; then
            pass "Redis: reachable"
        else
            fail "Redis: unreachable"
            failures=$((failures + 1))
        fi
    fi

    echo ""
    if [[ "$failures" -gt 0 ]]; then
        warn "${phase} checks: ${failures} issue(s) found"
        if [[ "$phase" == "Pre-flight" ]]; then
            warn "Continuing anyway — issues may cause deployment problems."
        fi
    else
        log "${phase} checks: all passed"
    fi

    return 0
}

# --- Docker Mode ---

check_root() {
    if [[ $EUID -ne 0 ]]; then
        error "This script must be run as root. Use: sudo ./setup.sh"
    fi
}

check_docker_requirements() {
    log "Checking system requirements..."

    ARCH=$(uname -m)
    case $ARCH in
        x86_64)  ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        *)       error "Unsupported architecture: $ARCH (need amd64 or arm64)" ;;
    esac
    log "Architecture: $ARCH"

    if ! command -v docker &>/dev/null; then
        warn "Docker not found. Installing..."
        curl -fsSL https://get.docker.com | sh
        systemctl enable --now docker
    fi
    log "Docker: $(docker --version | awk '{print $3}')"

    if ! docker compose version &>/dev/null; then
        error "Docker Compose v2 is required. Install with: apt install docker-compose-plugin"
    fi
    log "Docker Compose: $(docker compose version --short)"

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

        JWT_SECRET=$(openssl rand -hex 32)
        PANEL_SECRET=$(openssl rand -hex 32)
        DB_PASSWORD=$(openssl rand -base64 16 | tr -d '=/+')

        sed -i "s|^JWT_SECRET=.*|JWT_SECRET=${JWT_SECRET}|" "$INSTALL_DIR/.env"
        sed -i "s|^PANEL_SECRET=.*|PANEL_SECRET=${PANEL_SECRET}|" "$INSTALL_DIR/.env"
        sed -i "s|^POSTGRES_PASSWORD=.*|POSTGRES_PASSWORD=${DB_PASSWORD}|" "$INSTALL_DIR/.env"
        sed -i "s|^DATABASE_URL=.*|DATABASE_URL=postgres://vortex:${DB_PASSWORD}@postgres:5432/vortex?sslmode=disable|" "$INSTALL_DIR/.env"

        log "Secrets generated and saved to .env"

        # Also create deploy/.env for Docker Compose
        if [[ ! -f "$INSTALL_DIR/deploy/.env" ]]; then
            log "Generating deploy/.env for Docker Compose..."
            cat > "$INSTALL_DIR/deploy/.env" <<DEOF
JWT_SECRET=${JWT_SECRET}
DB_PASSWORD=${DB_PASSWORD}
LOCAL_NODE_HOST=$(curl -sf https://api.ipify.org || hostname -I | awk '{print $1}')
CORE=xray
SITE_ADDRESS=${USER_DOMAIN:-:${WEB_PORT:-80}}
DEOF
            log "deploy/.env created"
        fi
    else
        log "Existing .env found, keeping current configuration"

        # Ensure deploy/.env exists even if root .env already existed
        if [[ ! -f "$INSTALL_DIR/deploy/.env" ]]; then
            log "Generating deploy/.env from existing secrets..."
            JWT_SECRET=$(grep '^JWT_SECRET=' "$INSTALL_DIR/.env" 2>/dev/null | cut -d= -f2- || openssl rand -hex 32)
            DB_PASSWORD=$(grep '^POSTGRES_PASSWORD=' "$INSTALL_DIR/.env" 2>/dev/null | cut -d= -f2- || openssl rand -base64 16 | tr -d '=/+')
            cat > "$INSTALL_DIR/deploy/.env" <<DEOF
JWT_SECRET=${JWT_SECRET:-$(openssl rand -hex 32)}
DB_PASSWORD=${DB_PASSWORD:-vortex}
LOCAL_NODE_HOST=$(curl -sf https://api.ipify.org || hostname -I | awk '{print $1}')
CORE=xray
SITE_ADDRESS=${USER_DOMAIN:-:${WEB_PORT:-80}}
DEOF
            log "deploy/.env created"
        fi
    fi
}

generate_certs() {
    if [[ ! -d "$INSTALL_DIR/deploy/certs" ]]; then
        log "Generating mTLS certificates..."
        cd "$INSTALL_DIR"
        if command -v go &>/dev/null; then
            go run ./cmd/gencerts -out deploy/certs -san localhost,127.0.0.1
        else
            mkdir -p deploy/certs
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

start_docker_stack() {
    log "Starting VortexUI stack..."
    cd "$INSTALL_DIR"

    log "Applying database migrations..."
    docker compose -f deploy/compose.yml up -d db redis
    sleep 3
    docker compose -f deploy/compose.yml run --rm panel ./panel doctor || true

    docker compose -f deploy/compose.yml up --build -d

    log "Waiting for services to start..."
    sleep 5

    for i in {1..30}; do
        if curl -sf http://localhost:8080/api/health &>/dev/null; then
            break
        fi
        sleep 2
    done
}

deploy_docker() {
    check_docker_requirements
    clone_or_pull
    generate_secrets
    generate_certs
    start_docker_stack
}

create_admin_account() {
    echo ""
    echo -e "  ${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "  ${CYAN}  Create your admin account${NC}"
    echo -e "  ${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    read -r -p "  Admin username: " ADMIN_USER
    while [[ -z "$ADMIN_USER" ]]; do
        echo -e "  ${RED}Username cannot be empty${NC}"
        read -r -p "  Admin username: " ADMIN_USER
    done
    
    read -r -s -p "  Admin password: " ADMIN_PASS
    echo ""
    while [[ ${#ADMIN_PASS} -lt 6 ]]; do
        echo -e "  ${RED}Password must be at least 6 characters${NC}"
        read -r -s -p "  Admin password: " ADMIN_PASS
        echo ""
    done
    
    echo ""
    log "Creating admin account..."

    local created=0
    
    if [[ "$MODE" == "docker" ]]; then
        # Docker: exec into running container
        cd "$INSTALL_DIR"
        for attempt in {1..20}; do
            log "  Attempt $attempt/20 — waiting for panel container..."
            if docker compose -f deploy/compose.yml exec -T panel panel admin create \
                --username "$ADMIN_USER" --password "$ADMIN_PASS" --sudo 2>&1 | tee /tmp/admin_create.log; then
                if grep -qi "created\|success\|already" /tmp/admin_create.log 2>/dev/null; then
                    created=1
                    break
                fi
            fi
            sleep 3
        done
    else
        # Native: try multiple methods
        
        # Method 1: API call (if panel is running)
        for attempt in {1..20}; do
            if curl -sf http://127.0.0.1:8080/api/health &>/dev/null; then
                log "  Panel is up — creating admin via API..."
                local http_code
                http_code=$(curl -s -o /tmp/admin_response.txt -w "%{http_code}" \
                    -X POST http://127.0.0.1:8080/api/admin/init \
                    -H "Content-Type: application/json" \
                    -d "{\"username\":\"$ADMIN_USER\",\"password\":\"$ADMIN_PASS\"}")
                
                if [[ "$http_code" == "201" ]]; then
                    created=1
                    break
                elif [[ "$http_code" == "403" ]]; then
                    log "  Admin already exists."
                    created=1
                    break
                fi
                # API failed, try CLI
                break
            fi
            log "  Attempt $attempt/20 — waiting for panel..."
            sleep 3
        done
        
        # Method 2: CLI binary with explicit env
        if [[ "$created" -eq 0 ]]; then
            log "  Trying CLI method..."
            export VORTEX_DATABASE_URL="postgres://vortex:vortex@127.0.0.1:5432/vortex?sslmode=disable"
            export VORTEX_JWT_SECRET=$(grep '^VORTEX_JWT_SECRET=' /etc/vortexui/panel.env 2>/dev/null | cut -d= -f2-)
            if [[ -z "$VORTEX_JWT_SECRET" ]]; then
                export VORTEX_JWT_SECRET=$(openssl rand -hex 32)
            fi
            export VORTEX_REDIS_URL="redis://127.0.0.1:6379/0"
            export VORTEX_TLS_CERT="/opt/vortexui/deploy/certs/panel.crt"
            export VORTEX_TLS_KEY="/opt/vortexui/deploy/certs/panel.key"
            export VORTEX_TLS_CA="/opt/vortexui/deploy/certs/ca.crt"
            
            if /usr/local/bin/vortex-panel admin create --username "$ADMIN_USER" --password "$ADMIN_PASS" --sudo 2>&1; then
                created=1
            fi
        fi
    fi
    
    rm -f /tmp/admin_create.log /tmp/admin_response.txt
    
    if [[ "$created" -eq 1 ]]; then
        echo ""
        echo -e "  ${GREEN}✓ Admin account '${ADMIN_USER}' created successfully!${NC}"
        echo ""
    else
        echo ""
        echo -e "  ${YELLOW}⚠ Could not create admin automatically.${NC}"
        echo -e "  ${YELLOW}  After installation, run:${NC}"
        echo ""
        if [[ "$MODE" == "docker" ]]; then
            echo -e "  cd ${INSTALL_DIR} && docker compose -f deploy/compose.yml exec panel panel admin create --username ${ADMIN_USER} --password YOUR_PASS --sudo"
        else
            echo -e "  source /etc/vortexui/panel.env && vortex-panel admin create --username ${ADMIN_USER} --password YOUR_PASS --sudo"
        fi
        echo ""
    fi
}

print_docker_success() {
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
    if [[ -n "${USER_DOMAIN:-}" ]]; then
        if [[ "${WEB_PORT:-443}" != "80" && "${WEB_PORT:-443}" != "443" ]]; then
            echo -e "  ${BLUE}Panel URL:${NC}    https://${USER_DOMAIN}:${WEB_PORT}"
        else
            echo -e "  ${BLUE}Panel URL:${NC}    https://${USER_DOMAIN}"
        fi
    else
        echo -e "  ${BLUE}Panel URL:${NC}    http://${PUBLIC_IP}:${WEB_PORT:-80}"
    fi
    echo -e "  ${BLUE}Install Dir:${NC}  ${INSTALL_DIR}"
    echo -e "  ${BLUE}Config:${NC}       ${INSTALL_DIR}/.env"
    echo -e "  ${BLUE}Logs:${NC}         docker compose -f ${INSTALL_DIR}/deploy/compose.yml logs -f"
    echo ""
    if [[ -n "${ADMIN_USER:-}" ]]; then
    echo -e "  ${YELLOW}Admin:${NC}       ${ADMIN_USER}"
    echo ""
    fi
    echo -e "  ${CYAN}Commands:${NC}"
    echo -e "    Stop:     docker compose -f ${INSTALL_DIR}/deploy/compose.yml down"
    echo -e "    Start:    docker compose -f ${INSTALL_DIR}/deploy/compose.yml up -d"
    echo -e "    Update:   sudo ./setup.sh --docker"
    echo -e "    Doctor:   docker compose -f ${INSTALL_DIR}/deploy/compose.yml exec panel ./panel doctor"
    echo ""
}

# --- Systemd Mode ---

deploy_systemd() {
    # Clone repo if fresh install, or pull if exists
    clone_or_pull
    generate_secrets
    generate_certs

    cd "$INSTALL_DIR"

    echo "==> VortexUI v${VERSION} deploy (systemd)"

    # Database Migrations
    if [[ "$SKIP_MIGRATE" -eq 0 ]]; then
        echo "==> running database migrations"
        if command -v goose >/dev/null 2>&1 && [[ -n "${VORTEX_DATABASE_URL:-}" ]]; then
            goose -dir migrations postgres "$VORTEX_DATABASE_URL" up
        else
            echo "    (skipped: goose not found or VORTEX_DATABASE_URL not set)"
            echo "    Migrations will auto-apply on panel startup."
        fi
    fi

    # Frontend
    if [[ "$SKIP_FRONTEND" -eq 0 ]]; then
        echo "==> building frontend"

        # Install Node.js if npm not available
        if ! command -v npm >/dev/null 2>&1; then
            echo "    Node.js not found, installing..."
            curl -fsSL https://deb.nodesource.com/setup_20.x | bash - >/dev/null 2>&1
            apt-get install -y -qq nodejs >/dev/null 2>&1
        fi

        cd "$INSTALL_DIR/web"
        npm install --prefer-offline
        npm run build

        echo "==> deploying static files to $WEB_ROOT"
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
        if ! command -v go >/dev/null 2>&1 && ! [[ -x /usr/local/go/bin/go ]]; then
            echo "    Go not found, installing..."
            GO_VERSION="1.23.4"
            GOARCH=$(uname -m); case "$GOARCH" in x86_64) GOARCH="amd64";; aarch64) GOARCH="arm64";; esac
            curl -sL "https://go.dev/dl/go${GO_VERSION}.linux-${GOARCH}.tar.gz" | tar -C /usr/local -xzf -
        fi
        export PATH="$PATH:/usr/local/go/bin"

        CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" \
            -o /usr/local/bin/vortex-panel ./cmd/panel
    fi

    # Install proxy core binaries (xray + sing-box)
    if [[ ! -x /usr/local/bin/xray ]]; then
        echo "==> installing xray-core"
        local XARCH=$(uname -m); case "$XARCH" in x86_64) XARCH="64";; aarch64) XARCH="arm64-v8a";; esac
        curl -sL "https://github.com/XTLS/Xray-core/releases/latest/download/Xray-linux-${XARCH}.zip" -o /tmp/xray.zip
        unzip -o /tmp/xray.zip xray -d /usr/local/bin/ >/dev/null 2>&1
        chmod +x /usr/local/bin/xray
        rm -f /tmp/xray.zip
    fi

    if [[ ! -x /usr/local/bin/sing-box ]]; then
        echo "==> installing sing-box"
        local SARCH=$(uname -m); case "$SARCH" in x86_64) SARCH="amd64";; aarch64) SARCH="arm64";; esac
        local SB_VER=$(curl -sL "https://api.github.com/repos/SagerNet/sing-box/releases/latest" | grep -oP '"tag_name": "v\K[^"]+' || echo "1.10.7")
        curl -sL "https://github.com/SagerNet/sing-box/releases/download/v${SB_VER}/sing-box-${SB_VER}-linux-${SARCH}.tar.gz" -o /tmp/singbox.tar.gz
        tar -xzf /tmp/singbox.tar.gz -C /tmp/ 2>/dev/null
        cp /tmp/sing-box-*/sing-box /usr/local/bin/sing-box 2>/dev/null || find /tmp -name "sing-box" -type f -exec cp {} /usr/local/bin/sing-box \;
        chmod +x /usr/local/bin/sing-box
        rm -rf /tmp/singbox.tar.gz /tmp/sing-box-*
    fi

    # Ensure service file exists
    if [[ ! -f "/etc/systemd/system/${SERVICE}.service" ]]; then
        echo "==> creating systemd service"
        cat > "/etc/systemd/system/${SERVICE}.service" <<SVCEOF
[Unit]
Description=VortexUI Panel
After=network.target docker.service
Wants=network-online.target

[Service]
Type=simple
EnvironmentFile=-/etc/vortexui/panel.env
ExecStart=/usr/local/bin/vortex-panel
WorkingDirectory=/opt/vortexui
Restart=on-failure
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
SVCEOF
        systemctl daemon-reload
    fi

    # Create panel.env if not exists
    if [[ ! -f /etc/vortexui/panel.env ]]; then
        mkdir -p /etc/vortexui
        PUBLIC_IP=$(curl -sf https://api.ipify.org || hostname -I | awk '{print $1}')
        cat > /etc/vortexui/panel.env <<ENVEOF
VORTEX_HTTP_ADDR=:8080
VORTEX_GRPC_ADDR=:50051
VORTEX_DATABASE_URL=postgres://vortex:vortex@127.0.0.1:5432/vortex?sslmode=disable
VORTEX_REDIS_URL=redis://127.0.0.1:6379/0
VORTEX_JWT_SECRET=$(openssl rand -hex 32)
VORTEX_JWT_TTL=1h
VORTEX_TLS_CERT=/opt/vortexui/deploy/certs/panel.crt
VORTEX_TLS_KEY=/opt/vortexui/deploy/certs/panel.key
VORTEX_TLS_CA=/opt/vortexui/deploy/certs/ca.crt
VORTEX_LOCAL_NODE=true
VORTEX_LOCAL_NODE_NAME=local
VORTEX_LOCAL_NODE_HOST=${PUBLIC_IP}
VORTEX_CORE=xray
VORTEX_ENABLED_CORES=xray,singbox
VORTEX_CORE_BIN=/usr/local/bin/xray
VORTEX_CORE_CONFIG=/etc/vortex/local-core.json
VORTEX_CORE_API_PORT=10085
ENVEOF
        mkdir -p /etc/vortex
        echo "==> panel.env created at /etc/vortexui/panel.env"
    fi

    # Ensure PostgreSQL and Redis are running
    echo "==> setting up database dependencies"
    if command -v docker &>/dev/null && [[ -f "$INSTALL_DIR/docker-compose.yml" ]]; then
        # Use Docker for DB/Redis (preferred - includes TimescaleDB)
        docker compose -f "$INSTALL_DIR/docker-compose.yml" up -d 2>/dev/null || true
        sleep 5
    else
        # Install system PostgreSQL + Redis if not present
        if ! command -v psql &>/dev/null; then
            log "Installing PostgreSQL..."
            apt-get update -qq
            apt-get install -y -qq postgresql postgresql-contrib >/dev/null 2>&1
            systemctl enable --now postgresql
        fi
        if ! command -v redis-server &>/dev/null; then
            log "Installing Redis..."
            apt-get install -y -qq redis-server >/dev/null 2>&1
            systemctl enable --now redis-server
        fi

        # Create vortex database and user if not exists
        if ! sudo -u postgres psql -lqt 2>/dev/null | cut -d \| -f 1 | grep -qw vortex; then
            log "Creating PostgreSQL database..."
            sudo -u postgres psql -c "CREATE USER vortex WITH PASSWORD 'vortex';" 2>/dev/null || true
            sudo -u postgres psql -c "CREATE DATABASE vortex OWNER vortex;" 2>/dev/null || true
            sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE vortex TO vortex;" 2>/dev/null || true
        fi
        sleep 2
    fi

    # Verify PostgreSQL is accepting connections
    log "Verifying database connection..."
    for i in {1..10}; do
        if pg_isready -h 127.0.0.1 -p 5432 -U vortex &>/dev/null || sudo -u postgres psql -c "SELECT 1;" &>/dev/null; then
            pass "PostgreSQL ready"
            break
        fi
        sleep 2
    done

    # Unmask service if masked, then restart
    echo "==> restarting $SERVICE"
    if systemctl is-enabled "$SERVICE" 2>/dev/null | grep -q "masked"; then
        log "Service $SERVICE is masked — unmasking..."
        systemctl unmask "$SERVICE"
    fi
    systemctl daemon-reload
    systemctl enable "$SERVICE" 2>/dev/null || true
    systemctl restart "$SERVICE"

    # Wait for panel health endpoint (up to 60 seconds)
    echo "==> waiting for panel to start..."
    for i in {1..30}; do
        if curl -sf http://127.0.0.1:8080/api/health &>/dev/null; then
            log "Panel is ready!"
            break
        fi
        if [[ $i -eq 30 ]]; then
            warn "Panel did not start within 60 seconds. Check: journalctl -u $SERVICE"
        fi
        sleep 2
    done

    # Caddy
    if systemctl is-active --quiet caddy 2>/dev/null; then
        echo "==> reloading caddy"
        systemctl reload caddy
    fi
}

print_systemd_success() {
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║   VortexUI v${VERSION} deployed (systemd)!          ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════╝${NC}"
    echo ""
    if [[ "$SKIP_FRONTEND" -eq 0 ]]; then
        echo -e "  ${BLUE}Frontend assets:${NC}"
        ls -la "$WEB_ROOT"/assets/ 2>/dev/null | head -5
    fi
    echo -e "  ${BLUE}Panel binary:${NC} /usr/local/bin/vortex-panel (v${VERSION})"
    echo -e "  ${BLUE}Service:${NC}      systemctl status $SERVICE"
    echo -e "  ${BLUE}Update:${NC}       sudo ./setup.sh --systemd"
    if [[ -n "${USER_DOMAIN:-}" ]]; then
        if [[ "${WEB_PORT:-443}" != "80" && "${WEB_PORT:-443}" != "443" ]]; then
            echo -e "  ${BLUE}Panel URL:${NC}    https://${USER_DOMAIN}:${WEB_PORT}"
        else
            echo -e "  ${BLUE}Panel URL:${NC}    https://${USER_DOMAIN}"
        fi
    else
        local PUBLIC_IP
        PUBLIC_IP=$(curl -sf https://api.ipify.org || hostname -I | awk '{print $1}')
        echo -e "  ${BLUE}Panel URL:${NC}    http://${PUBLIC_IP}:${WEB_PORT:-80}"
    fi
    if [[ -n "${ADMIN_USER:-}" ]]; then
        echo -e "  ${YELLOW}Admin:${NC}        ${ADMIN_USER}"
    fi
    echo ""
}

# --- Interactive Wizard ---

ask_install_type() {
    echo ""
    echo -e "  ${CYAN}What would you like to install?${NC}"
    echo -e "   ${BLUE}1)${NC} Panel  (control plane + web UI + local node)"
    echo -e "   ${BLUE}2)${NC} Node   (remote node agent only)"
    echo ""
    read -r -p "  Choose [1/2]: " INSTALL_TYPE
    case "$INSTALL_TYPE" in
        2) INSTALL_TYPE="node" ;;
        *) INSTALL_TYPE="panel" ;;
    esac
}

ask_deploy_method() {
    echo ""
    echo -e "  ${CYAN}Deploy method:${NC}"
    echo -e "   ${BLUE}1)${NC} Docker Compose  ${GREEN}(recommended)${NC}"
    echo -e "   ${BLUE}2)${NC} Native (systemd + build from source)"
    echo ""
    read -r -p "  Choose [1/2]: " DEPLOY_METHOD
    case "$DEPLOY_METHOD" in
        2) MODE="systemd" ;;
        *) MODE="docker" ;;
    esac
}

ask_domain() {
    echo ""
    echo -e "  ${CYAN}Domain for SSL (optional):${NC}"
    echo -e "  ${BLUE}Enter your domain (e.g. panel.example.com) for automatic HTTPS.${NC}"
    echo -e "  ${BLUE}Leave empty to use IP address with HTTP only.${NC}"
    echo ""
    read -r -p "  Domain: " USER_DOMAIN
    USER_DOMAIN="${USER_DOMAIN:-}"
}

ask_web_port() {
    echo ""
    echo -e "  ${CYAN}Web panel port (default 80):${NC}"
    echo -e "  ${BLUE}Change if port 80 is used by VPN configs. Common alternatives: 8080, 2086, 2095${NC}"
    echo ""
    read -r -p "  Port [80]: " WEB_PORT
    WEB_PORT="${WEB_PORT:-80}"
}

setup_ssl() {
    if [[ -z "$USER_DOMAIN" ]]; then
        log "No domain specified — using HTTP on port 8080"
        return
    fi

    log "Setting up SSL for $USER_DOMAIN..."

    # Install Caddy if not present
    if ! command -v caddy &>/dev/null; then
        log "Installing Caddy..."
        apt-get install -y -qq debian-keyring debian-archive-keyring apt-transport-https curl 2>/dev/null || true
        curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg 2>/dev/null
        curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list >/dev/null
        apt-get update -qq && apt-get install -y -qq caddy
    fi

    # Determine the site address for Caddy
    local SITE_ADDR="${USER_DOMAIN}"
    if [[ "${WEB_PORT:-80}" != "80" && "${WEB_PORT:-443}" != "443" ]]; then
        SITE_ADDR="${USER_DOMAIN}:${WEB_PORT}"
    fi

    # Write Caddyfile
    mkdir -p /etc/caddy
    cat > /etc/caddy/Caddyfile <<CADDYEOF
${SITE_ADDR} {
    encode gzip
    handle /api/* {
        reverse_proxy 127.0.0.1:8080
    }
    handle /sub/* {
        reverse_proxy 127.0.0.1:8080
    }
    handle /health {
        reverse_proxy 127.0.0.1:8080
    }
    handle {
        root * /var/www/vortexui
        try_files {path} /index.html
        file_server
    }
}
CADDYEOF

    systemctl enable caddy 2>/dev/null || true
    systemctl restart caddy
    log "SSL configured for ${USER_DOMAIN} (auto Let's Encrypt)"
}

install_node() {
    log "Installing VortexUI Node Agent..."
    
    echo ""
    echo -e "  ${CYAN}Panel address (where this node connects to):${NC}"
    read -r -p "  Panel host (e.g. panel.example.com or IP): " PANEL_HOST
    
    echo ""
    echo -e "  ${CYAN}Node enrollment token:${NC}"
    echo -e "  ${BLUE}Get this from Panel → Nodes → Node Enrollment Bundle${NC}"
    read -r -p "  Paste enrollment bundle (base64): " ENROLL_BUNDLE
    
    # Create directories
    mkdir -p /etc/vortexui/certs
    
    # Decode enrollment bundle (ca.crt + node.crt + node.key)
    if [[ -n "$ENROLL_BUNDLE" ]]; then
        echo "$ENROLL_BUNDLE" | base64 -d | tar -xzf - -C /etc/vortexui/certs/
        log "Certificates extracted to /etc/vortexui/certs/"
    fi
    
    # Download node binary
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64)  ARCH_DL="amd64" ;;
        aarch64) ARCH_DL="arm64" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac
    
    log "Downloading node agent..."
    local rel
    rel=$(curl -fsSL https://api.github.com/repos/iPmartNetwork/VortexUI/releases/latest 2>/dev/null \
        | grep -oE '"tag_name": *"v[0-9.]+"' | head -1 | grep -oE 'v[0-9.]+' || echo "v${VERSION}")
    
    if curl -fL -o /usr/local/bin/vortex-node \
        "https://github.com/iPmartNetwork/VortexUI/releases/download/${rel}/vortexui-node-linux-${ARCH_DL}" 2>/dev/null; then
        chmod +x /usr/local/bin/vortex-node
        log "Node binary installed"
    else
        # Fallback: build from source
        if command -v go &>/dev/null || [[ -x /usr/local/go/bin/go ]]; then
            export PATH="$PATH:/usr/local/go/bin"
            log "Building node from source..."
            git clone --depth 1 "$REPO_URL" /tmp/vortexui-src 2>/dev/null || true
            (cd /tmp/vortexui-src && go build -o /usr/local/bin/vortex-node ./cmd/node)
            rm -rf /tmp/vortexui-src
        else
            error "Cannot install node: no release binary and Go is not installed"
        fi
    fi
    
    # Write node.env
    cat > /etc/vortexui/node.env <<NEOF
VORTEX_PANEL_ADDR=${PANEL_HOST}:50051
VORTEX_TLS_CERT=/etc/vortexui/certs/node.crt
VORTEX_TLS_KEY=/etc/vortexui/certs/node.key
VORTEX_TLS_CA=/etc/vortexui/certs/ca.crt
VORTEX_CORE=xray
VORTEX_CORE_BIN=/usr/local/bin/xray
NEOF
    
    # Create systemd service
    cat > /etc/systemd/system/vortexui-node.service <<SEOF
[Unit]
Description=VortexUI Node Agent
After=network.target

[Service]
Type=simple
EnvironmentFile=/etc/vortexui/node.env
ExecStart=/usr/local/bin/vortex-node
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
SEOF
    
    systemctl daemon-reload
    systemctl enable --now vortexui-node
    
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║     VortexUI Node Agent installed!              ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "  ${BLUE}Panel:${NC}    ${PANEL_HOST}"
    echo -e "  ${BLUE}Service:${NC}  systemctl status vortexui-node"
    echo -e "  ${BLUE}Logs:${NC}     journalctl -u vortexui-node -f"
    echo ""
}

# --- Main ---

parse_args "$@"
header
check_root

# If mode already set via flags, skip wizard
if [[ -z "$MODE" ]]; then
    ask_install_type
    
    if [[ "$INSTALL_TYPE" == "node" ]]; then
        install_node
        exit 0
    fi
    
    ask_deploy_method
fi

ask_domain
ask_web_port

# Pre-flight doctor check
doctor_check "Pre-flight"

case "$MODE" in
    docker)
        deploy_docker
        setup_ssl
        doctor_check "Post-deploy"
        create_admin_account
        print_docker_success
        if [[ -n "$USER_DOMAIN" ]]; then
            echo -e "  ${GREEN}HTTPS:${NC}    https://${USER_DOMAIN}"
        fi
        ;;
    systemd)
        deploy_systemd
        setup_ssl
        doctor_check "Post-deploy"
        create_admin_account
        print_systemd_success
        if [[ -n "$USER_DOMAIN" ]]; then
            echo -e "  ${GREEN}HTTPS:${NC}    https://${USER_DOMAIN}"
        fi
        ;;
    *)
        error "Unknown mode: $MODE"
        ;;
esac
