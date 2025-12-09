#!/bin/bash
set -e

M3M_DIR="${M3M_DIR:-$HOME/.m3m}"
M3M_REPO="https://github.com/levskiy0/m3m.git"
M3M_IMAGE="m3m:local"
M3M_CONTAINER="m3m"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[M3M]${NC} $1"; }
warn() { echo -e "${YELLOW}[M3M]${NC} $1"; }
error() { echo -e "${RED}[M3M]${NC} $1"; exit 1; }

cmd_install() {
    log "Installing M3M to $M3M_DIR..."

    if [ -d "$M3M_DIR/src" ]; then
        warn "Already installed. Use 'm3m.sh update' to update."
        return
    fi

    mkdir -p "$M3M_DIR"/{src,data,plugins}

    log "Cloning repository..."
    git clone "$M3M_REPO" "$M3M_DIR/src"

    log "Building Docker image..."
    cd "$M3M_DIR/src"
    docker build -t "$M3M_IMAGE" .

    # Create default config
    if [ ! -f "$M3M_DIR/.env" ]; then
        cat > "$M3M_DIR/.env" << 'EOF'
M3M_JWT_SECRET=change-me-to-random-string
M3M_SERVER_URI=http://localhost:8080
EOF
        warn "Created $M3M_DIR/.env - please edit M3M_JWT_SECRET!"
    fi

    log "Done! Run 'm3m.sh start' to start the server."
}

cmd_update() {
    log "Updating M3M..."

    [ -d "$M3M_DIR/src" ] || error "M3M not installed. Run 'm3m.sh install' first."

    cd "$M3M_DIR/src"
    git pull

    cmd_rebuild
}

cmd_rebuild() {
    log "Rebuilding M3M with plugins..."

    [ -d "$M3M_DIR/src" ] || error "M3M not installed. Run 'm3m.sh install' first."

    # Copy plugins to source
    if [ -d "$M3M_DIR/plugins" ] && [ "$(ls -A $M3M_DIR/plugins 2>/dev/null)" ]; then
        log "Copying plugins..."
        cp -r "$M3M_DIR/plugins"/* "$M3M_DIR/src/plugins/" 2>/dev/null || true
    fi

    log "Building Docker image..."
    cd "$M3M_DIR/src"
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

    [ -f "$M3M_DIR/.env" ] || error "Config not found. Run 'm3m.sh install' first."

    if docker ps -q -f name="$M3M_CONTAINER" | grep -q .; then
        warn "Already running."
        return
    fi

    # Remove stopped container if exists
    docker rm "$M3M_CONTAINER" 2>/dev/null || true

    docker run -d \
        --name "$M3M_CONTAINER" \
        --restart unless-stopped \
        -p 8080:8080 \
        -v "$M3M_DIR/data:/app/data" \
        --env-file "$M3M_DIR/.env" \
        "$M3M_IMAGE"

    log "Started! http://localhost:8080"
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
    [ -z "$1" ] || [ -z "$2" ] && error "Usage: m3m.sh admin <email> <password>"

    docker exec "$M3M_CONTAINER" /app/m3m new-admin "$1" "$2"
}

cmd_uninstall() {
    warn "This will remove M3M and all data!"
    read -p "Are you sure? (y/N) " -n 1 -r
    echo
    [[ $REPLY =~ ^[Yy]$ ]] || exit 0

    cmd_stop 2>/dev/null || true
    docker rm "$M3M_CONTAINER" 2>/dev/null || true
    docker rmi "$M3M_IMAGE" 2>/dev/null || true
    rm -rf "$M3M_DIR"

    log "Uninstalled."
}

cmd_help() {
    cat << EOF
M3M - Mini Services Manager

Usage: m3m.sh <command> [args]

Commands:
  install     Clone repo and build Docker image
  update      Pull latest changes and rebuild
  rebuild     Rebuild image (after adding plugins)
  start       Start the container
  stop        Stop the container
  restart     Restart the container
  logs        Show container logs
  status      Show container status
  admin       Create admin user: m3m.sh admin <email> <password>
  uninstall   Remove M3M completely

Directories:
  $M3M_DIR/data      Persistent data (mounted to container)
  $M3M_DIR/plugins   Put plugin sources here, then run 'rebuild'
  $M3M_DIR/.env      Environment config

EOF
}

# Main
case "${1:-help}" in
    install)   cmd_install ;;
    update)    cmd_update ;;
    rebuild)   cmd_rebuild ;;
    start)     cmd_start ;;
    stop)      cmd_stop ;;
    restart)   cmd_restart ;;
    logs)      cmd_logs ;;
    status)    cmd_status ;;
    admin)     cmd_admin "$2" "$3" ;;
    uninstall) cmd_uninstall ;;
    help|*)    cmd_help ;;
esac
