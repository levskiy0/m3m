#!/bin/bash
set -e

# Directory where script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
M3M_BASE="$SCRIPT_DIR/.m3m"
M3M_DATA="$M3M_BASE/data"
M3M_REPO="https://github.com/levskiy0/m3m.git"
M3M_IMAGE="m3m:local"
M3M_CONTAINER="m3m"
M3M_CONFIG="$M3M_BASE/config"
M3M_VERSION_FILE="$M3M_BASE/version"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[M3M]${NC} $1"; }
warn() { echo -e "${YELLOW}[M3M]${NC} $1"; }
error() { echo -e "${RED}[M3M]${NC} $1"; exit 1; }

# Load config if exists
load_config() {
    if [ -f "$M3M_CONFIG" ]; then
        source "$M3M_CONFIG"
    fi
    # Defaults
    M3M_PORT="${M3M_PORT:-8080}"
    M3M_JWT_SECRET="${M3M_JWT_SECRET:-}"
    M3M_SERVER_URI="${M3M_SERVER_URI:-http://localhost:$M3M_PORT}"
}

cmd_install() {
    local version="${1:-latest}"

    log "Installing M3M to $M3M_BASE..."

    if [ -d "$M3M_BASE/src" ]; then
        warn "Already installed. Use './m3m.sh update' to update."
        return
    fi

    mkdir -p "$M3M_BASE"/{src,plugins}
    mkdir -p "$M3M_DATA"

    log "Cloning repository..."
    git clone "$M3M_REPO" "$M3M_BASE/src"

    cd "$M3M_BASE/src"

    # Checkout version
    if [ "$version" = "latest" ]; then
        # Get latest release tag
        version=$(git describe --tags --abbrev=0 2>/dev/null || echo "main")
    fi

    if [ "$version" != "main" ]; then
        log "Checking out version: $version"
        git checkout "$version"
    fi

    echo "$version" > "$M3M_VERSION_FILE"

    log "Building Docker image..."
    docker build -t "$M3M_IMAGE" .

    # Create default config
    if [ ! -f "$M3M_CONFIG" ]; then
        # Generate random JWT secret
        JWT_SECRET=$(openssl rand -hex 32 2>/dev/null || head -c 64 /dev/urandom | base64 | tr -d '\n' | head -c 64)
        cat > "$M3M_CONFIG" << EOF
# M3M Configuration
M3M_PORT=8080
M3M_JWT_SECRET=$JWT_SECRET
M3M_SERVER_URI=http://localhost:8080
EOF
        log "Created config: $M3M_CONFIG"
    fi

    log "Done! Run './m3m.sh start' to start the server."
    log "Config file: $M3M_CONFIG"
}

cmd_update() {
    local version="${1:-latest}"

    log "Updating M3M..."

    [ -d "$M3M_BASE/src" ] || error "M3M not installed. Run './m3m.sh install' first."

    cd "$M3M_BASE/src"
    git fetch --tags

    # Determine target version
    if [ "$version" = "latest" ]; then
        version=$(git describe --tags --abbrev=0 2>/dev/null || echo "main")
    fi

    current_version=$(cat "$M3M_VERSION_FILE" 2>/dev/null || echo "unknown")

    if [ "$version" = "$current_version" ]; then
        log "Already on version $version"
        return
    fi

    log "Updating from $current_version to $version..."

    if [ "$version" = "main" ]; then
        git checkout main
        git pull
    else
        git checkout "$version"
    fi

    echo "$version" > "$M3M_VERSION_FILE"

    cmd_rebuild
}

cmd_rebuild() {
    log "Rebuilding M3M with plugins..."

    [ -d "$M3M_BASE/src" ] || error "M3M not installed. Run './m3m.sh install' first."

    # Copy plugins to source
    if [ -d "$M3M_BASE/plugins" ] && [ "$(ls -A $M3M_BASE/plugins 2>/dev/null)" ]; then
        log "Copying plugins..."
        cp -r "$M3M_BASE/plugins"/* "$M3M_BASE/src/plugins/" 2>/dev/null || true
    fi

    log "Building Docker image..."
    cd "$M3M_BASE/src"
    docker build -t "$M3M_IMAGE" .

    # Restart if running
    if docker ps -q -f name="$M3M_CONTAINER" | grep -q .; then
        log "Restarting container..."
        cmd_restart
    fi

    log "Done!"
}

cmd_start() {
    log "Starting M3M..."

    load_config

    [ -f "$M3M_CONFIG" ] || error "Config not found. Run './m3m.sh install' first."
    [ -n "$M3M_JWT_SECRET" ] || error "M3M_JWT_SECRET is not set in $M3M_CONFIG"

    if docker ps -q -f name="$M3M_CONTAINER" | grep -q .; then
        warn "Already running."
        return
    fi

    # Remove stopped container if exists
    docker rm "$M3M_CONTAINER" 2>/dev/null || true

    docker run -d \
        --name "$M3M_CONTAINER" \
        --restart unless-stopped \
        -p "$M3M_PORT:8080" \
        -v "$M3M_DATA:/app/data" \
        -e "M3M_JWT_SECRET=$M3M_JWT_SECRET" \
        -e "M3M_SERVER_URI=$M3M_SERVER_URI" \
        "$M3M_IMAGE"

    log "Started! $M3M_SERVER_URI"
}

cmd_stop() {
    log "Stopping M3M..."
    docker stop "$M3M_CONTAINER" 2>/dev/null || warn "Not running."
}

cmd_restart() {
    cmd_stop
    sleep 1
    cmd_start
}

cmd_logs() {
    docker logs -f "$M3M_CONTAINER"
}

cmd_status() {
    if docker ps -q -f name="$M3M_CONTAINER" | grep -q .; then
        log "Running"
        docker ps -f name="$M3M_CONTAINER" --format "table {{.Status}}\t{{.Ports}}"
    else
        warn "Not running"
    fi
}

cmd_admin() {
    [ -z "$1" ] || [ -z "$2" ] && error "Usage: ./m3m.sh admin <email> <password>"

    docker exec "$M3M_CONTAINER" /app/m3m new-admin "$1" "$2"
}

cmd_config() {
    [ -f "$M3M_CONFIG" ] || error "Config not found. Run './m3m.sh install' first."

    if [ -n "$1" ]; then
        # Set config value: ./m3m.sh config KEY=VALUE
        echo "$1" >> "$M3M_CONFIG"
        log "Updated config"
    else
        # Show config
        log "Config file: $M3M_CONFIG"
        echo "---"
        cat "$M3M_CONFIG"
        echo "---"
        log "Edit with: nano $M3M_CONFIG"
    fi
}

cmd_version() {
    local current="unknown"
    [ -f "$M3M_VERSION_FILE" ] && current=$(cat "$M3M_VERSION_FILE")

    log "Installed version: $current"

    if [ -d "$M3M_BASE/src" ]; then
        cd "$M3M_BASE/src"
        git fetch --tags 2>/dev/null
        local latest=$(git describe --tags --abbrev=0 2>/dev/null || echo "none")
        log "Latest release: $latest"

        if [ "$current" != "$latest" ] && [ "$latest" != "none" ]; then
            warn "Update available! Run './m3m.sh update' or './m3m.sh update $latest'"
        fi
    fi
}

cmd_uninstall() {
    warn "This will remove M3M and all data!"
    read -p "Are you sure? (y/N) " -n 1 -r
    echo
    [[ $REPLY =~ ^[Yy]$ ]] || exit 0

    cmd_stop 2>/dev/null || true
    docker rm "$M3M_CONTAINER" 2>/dev/null || true
    docker rmi "$M3M_IMAGE" 2>/dev/null || true
    rm -rf "$M3M_BASE/src"

    log "Uninstalled. Data preserved in $M3M_DATA"
}

cmd_help() {
    cat << EOF
M3M - Mini Services Manager

Usage: ./m3m.sh <command> [args]

Commands:
  install [version]   Install M3M (default: latest release, or 'main' for dev)
  update [version]    Update to version (default: latest release)
  rebuild             Rebuild image (after adding plugins)
  start               Start the container
  stop                Stop the container
  restart             Restart the container
  logs                Show container logs
  status              Show container status
  version             Show installed and latest versions
  admin <email> <pw>  Create admin user
  config [KEY=VALUE]  Show or edit config
  uninstall           Remove M3M (keeps data)

Examples:
  ./m3m.sh install              # Install latest release
  ./m3m.sh install v1.0.0       # Install specific version
  ./m3m.sh install main         # Install development version
  ./m3m.sh update               # Update to latest release
  ./m3m.sh update v1.1.0        # Update to specific version

Directory structure (.m3m/):
  data/       Persistent data (mounted to container)
  plugins/    Plugin sources (copy here, then 'rebuild')
  config      Configuration file
  src/        Repository clone

Config variables:
  M3M_PORT           Server port (default: 8080)
  M3M_JWT_SECRET     JWT signing secret (auto-generated)
  M3M_SERVER_URI     Public server URI

EOF
}

# Main
case "${1:-help}" in
    install)   cmd_install "$2" ;;
    update)    cmd_update "$2" ;;
    rebuild)   cmd_rebuild ;;
    start)     cmd_start ;;
    stop)      cmd_stop ;;
    restart)   cmd_restart ;;
    logs)      cmd_logs ;;
    status)    cmd_status ;;
    version)   cmd_version ;;
    admin)     cmd_admin "$2" "$3" ;;
    config)    cmd_config "$2" ;;
    uninstall) cmd_uninstall ;;
    help|*)    cmd_help ;;
esac
