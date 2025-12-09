# Build stage
FROM golang:1.24 AS builder

RUN apt-get update && apt-get install -y git make nodejs npm && rm -rf /var/lib/apt/lists/*

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN cd web/ui && npm install && npm run build

ARG VERSION=dev

# Build main binary
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags "-X github.com/levskiy0/m3m/internal/app.Version=${VERSION}" \
    -o m3m ./cmd/m3m

# Build plugins
RUN mkdir -p /build/built-plugins && \
    touch /build/built-plugins/.keep && \
    for dir in plugins/*/; do \
        if [ -f "$dir/plugin.go" ]; then \
            plugin_name=$(basename "$dir"); \
            echo "Building plugin: $plugin_name"; \
            cd "$dir" && go build -buildmode=plugin -o "/build/built-plugins/${plugin_name}.so" . && cd /build; \
        fi \
    done

# Runtime stage - all-in-one with MongoDB
FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

# Install MongoDB 7.0
RUN apt-get update && apt-get install -y gnupg curl wget ca-certificates \
    && curl -fsSL https://www.mongodb.org/static/pgp/server-7.0.asc | gpg --dearmor -o /usr/share/keyrings/mongodb-server-7.0.gpg \
    && echo "deb [ arch=amd64,arm64 signed-by=/usr/share/keyrings/mongodb-server-7.0.gpg ] https://repo.mongodb.org/apt/ubuntu jammy/mongodb-org/7.0 multiverse" > /etc/apt/sources.list.d/mongodb-org-7.0.list \
    && apt-get update \
    && apt-get install -y mongodb-org \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Create directories
RUN mkdir -p /app/data/storage /app/data/logs /app/data/mongodb /app/data/plugins

# Copy binary, config, entrypoint, and built plugins
COPY --from=builder /build/m3m /app/m3m
COPY --from=builder /build/docker-config.yaml /app/config.yaml
COPY --from=builder /build/docker-entrypoint.sh /app/docker-entrypoint.sh
COPY --from=builder /build/built-plugins/ /app/data/plugins/

RUN chmod +x /app/docker-entrypoint.sh

# M3M_JWT_SECRET must be provided at runtime via -e or docker-compose
# Other settings are in /app/config.yaml and can be overridden with M3M_* env vars

EXPOSE 8080

VOLUME ["/app/data"]

HEALTHCHECK --interval=30s --timeout=3s --start-period=15s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/docker-entrypoint.sh"]
