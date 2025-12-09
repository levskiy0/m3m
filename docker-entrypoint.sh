#!/bin/bash
set -e

MONGO_DATA_DIR="/app/data/mongodb"

echo "Starting MongoDB..."
mongod --dbpath "$MONGO_DATA_DIR" --bind_ip 127.0.0.1 --fork --logpath /app/data/logs/mongodb.log

# Wait for MongoDB
echo "Waiting for MongoDB..."
for i in {1..30}; do
    if mongosh --quiet --eval "db.adminCommand('ping')" >/dev/null 2>&1; then
        echo "MongoDB is ready"
        break
    fi
    sleep 1
done

echo "Starting M3M..."
exec /app/m3m serve "$@"
