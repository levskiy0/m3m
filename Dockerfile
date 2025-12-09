# =============================================================================
# M3M - Mini Services Manager
# Multi-stage Dockerfile
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Build Frontend
# -----------------------------------------------------------------------------
FROM node:20-alpine AS frontend-builder

WORKDIR /app/web/ui

# Copy package files first for better caching
COPY web/ui/package*.json ./
RUN npm ci

# Copy frontend source and build
COPY web/ui/ ./
RUN npm run build

# -----------------------------------------------------------------------------
# Stage 2: Build Backend and Plugins
# -----------------------------------------------------------------------------
FROM golang:1.24-alpine AS backend-builder

# Install build dependencies (gcc required for CGO/plugins)
RUN apk add --no-cache git make gcc musl-dev

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Copy built frontend from previous stage
COPY --from=frontend-builder /app/web/ui/dist ./web/ui/dist

# Build plugins and backend using Makefile
ARG VERSION=dev
RUN make build-plugins && \
    CGO_ENABLED=1 make build-backend VERSION=${VERSION}

# -----------------------------------------------------------------------------
# Stage 3: Final Runtime Image
# -----------------------------------------------------------------------------
FROM alpine:3.19

# Install CA certificates and libc for CGO binaries
RUN apk add --no-cache ca-certificates tzdata libc6-compat

# Create non-root user
RUN addgroup -S m3m && adduser -S m3m -G m3m

WORKDIR /app

# Copy binary and plugins from builder
COPY --from=backend-builder /app/build/m3m /app/m3m
COPY --from=backend-builder /app/build/plugins/*.so /app/plugins/

# Create necessary directories
RUN mkdir -p /app/storage /app/logs /app/data && \
    chown -R m3m:m3m /app

# Copy default config (will be overridden by volume mount in production)
COPY --chown=m3m:m3m config.yaml /app/config.yaml

# Switch to non-root user
USER m3m

# Expose default port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

# Default command
ENTRYPOINT ["/app/m3m"]
CMD ["serve"]
