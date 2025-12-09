# Build stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git make nodejs npm

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN cd web/ui && npm ci && npm run build

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X m3m/internal/app.Version=${VERSION}" \
    -o m3m ./cmd/m3m

# Runtime stage - all-in-one with MongoDB
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata mongodb bash

WORKDIR /app

# Create directories
RUN mkdir -p /app/data/storage /app/data/logs /app/data/mongodb /app/plugins

# Copy binary
COPY --from=builder /build/m3m /app/m3m

# Copy entrypoint script
COPY --from=builder /build/docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

# Environment
ENV M3M_SERVER_HOST=0.0.0.0
ENV M3M_SERVER_PORT=8080
ENV M3M_MONGODB_URI=mongodb://127.0.0.1:27017
ENV M3M_MONGODB_DATABASE=m3m
ENV M3M_STORAGE_PATH=/app/data/storage
ENV M3M_LOGGING_PATH=/app/data/logs
ENV M3M_PLUGINS_PATH=/app/plugins
ENV M3M_JWT_SECRET=change-me-in-production
ENV M3M_JWT_EXPIRATION=168h

EXPOSE 8080

VOLUME ["/app/data"]

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/docker-entrypoint.sh"]
