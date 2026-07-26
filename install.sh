#!/usr/bin/env bash
# VortexUI — One-line installer (interactive wizard)
# Usage: bash <(curl -sL https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)

set -euo pipefail

# If setup.sh exists next to this script, use it directly
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}" 2>/dev/null)" 2>/dev/null && pwd)" 2>/dev/null || SCRIPT_DIR=""
if [[ -n "$SCRIPT_DIR" && -f "$SCRIPT_DIR/setup.sh" ]]; then
    exec "$SCRIPT_DIR/setup.sh" "$@"
fi

# Otherwise download and run setup.sh (curl | bash case)
SETUP_URL="https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/setup.sh"
SETUP_SCRIPT=$(mktemp)
curl -fsSL "$SETUP_URL" -o "$SETUP_SCRIPT" || { echo "Failed to download setup.sh"; exit 1; }
chmod +x "$SETUP_SCRIPT"
exec bash "$SETUP_SCRIPT" "$@"
