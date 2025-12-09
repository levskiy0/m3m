#!/bin/bash
set -e

MONGO_DATA_DIR="/app/data/mongodb"
MONGO_LOG="/app/data/logs/mongodb.log"

# Ensure directories exist with correct permissions
mkdir -p "$MONGO_DATA_DIR" /app/data/logs /app/data/storage /app/data/plugins
chmod -R 755 /app/data

# Copy built-in plugins if not already present
if [ -d "/app/built-plugins" ]; then
    for plugin in /app/built-plugins/*.so; do
        [ -f "$plugin" ] || continue
        plugin_name=$(basename "$plugin")
        if [ ! -f "/app/data/plugins/$plugin_name" ]; then
            echo "Installing plugin: $plugin_name"
            cp "$plugin" /app/data/plugins/
        fi
    done
fi

echo "Starting MongoDB..."
echo "Data dir: $MONGO_DATA_DIR"

# Start MongoDB without fork first to see errors, then background it
mongod --dbpath "$MONGO_DATA_DIR" --bind_ip 127.0.0.1 --logpath "$MONGO_LOG" --logappend &
MONGO_PID=$!

# Wait for MongoDB
echo "Waiting for MongoDB (PID: $MONGO_PID)..."
for i in {1..30}; do
    if mongosh --quiet --eval "db.adminCommand('ping')" >/dev/null 2>&1; then
        echo "MongoDB is ready"
        break
    fi
    if ! kill -0 $MONGO_PID 2>/dev/null; then
        echo "MongoDB failed to start. Log:"
        cat "$MONGO_LOG" 2>/dev/null || echo "No log available"
        exit 1
    fi
    sleep 1
done

# Final check
if ! mongosh --quiet --eval "db.adminCommand('ping')" >/dev/null 2>&1; then
    echo "MongoDB failed to start after 30 seconds. Log:"
    cat "$MONGO_LOG" 2>/dev/null || echo "No log available"
    exit 1
fi

echo "Starting M3M..."
exec /app/m3m serve "$@"
