# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

M3M (Mini Services Manager) is a platform for creating and managing JavaScript mini-services/workers. It provides a Go backend with a JavaScript runtime (GOJA) that allows users to write and deploy small services with built-in modules for routing, scheduling, storage, database, and more.

## Build & Development Commands

```bash
# Full build (frontend + backend)
make build

# Backend only
make build-backend

# Run with build
make run

# Development mode (backend, no build)
make dev

# Frontend
make web-install    # Install npm dependencies
make web-build      # Build frontend
make web-dev        # Run Vite dev server

# Tests
make test           # go test -v ./...

# Create admin user
make new-admin EMAIL=admin@example.com PASSWORD=yourpassword

# Build plugin
make build-plugin PLUGIN=telegram
```

## Architecture

### Backend (Go)

The backend uses dependency injection via `uber-go/fx`. The application is structured in layers:

- **cmd/m3m/main.go**: CLI entry point using Cobra (serve, new-admin, version commands)
- **internal/app/app.go**: Application bootstrap, DI wiring, route registration
- **internal/config/**: YAML configuration via Viper
- **internal/domain/**: Domain models (Project, Pipeline, Goal, Environment, Model, User)
- **internal/repository/**: MongoDB repositories
- **internal/service/**: Business logic layer
- **internal/handler/**: HTTP handlers (Gin framework)
- **internal/middleware/**: Auth (JWT) and CORS middleware
- **internal/runtime/**: JavaScript runtime manager using GOJA
- **internal/plugin/**: Plugin loader for extending runtime

### JavaScript Runtime

The runtime (`internal/runtime/runtime.go`) provides services with lifecycle management and built-in modules:

**Lifecycle hooks (service module):**
- `service.boot(callback)` - initialization phase
- `service.start(callback)` - when service is ready
- `service.shutdown(callback)` - graceful shutdown

**Built-in modules:** logger, router, schedule, env, storage, database, goals, http, crypto, encoding, utils, delayed

Type definitions for Monaco IntelliSense are in `internal/runtime/modules/types.go`.

### Frontend (React + Vite)

Located in `web/ui/`:
- React 19 with TypeScript
- Tailwind CSS v4 with @tailwindcss/vite
- Radix UI primitives with shadcn/ui components
- Built assets are embedded into Go binary via `web/static.go`

**UI style reference:** See `ui-style.md` for component patterns using shadcn/ui Field components.

### Key Domain Concepts

- **Project**: A deployable JavaScript service with API key, status (running/stopped), auto-start capability
- **Pipeline**: Branches and releases for code versioning (Branch contains editable code, Release is immutable)
- **Goal**: Metrics tracking (counter or daily_counter types)
- **Environment**: Key-value environment variables per project
- **Model**: Schema definitions for the database module

### Configuration

Default config path: `config.yaml`. Key settings:
- Server: host, port, URI
- MongoDB: uri, database
- JWT: secret, expiration
- Storage/Plugins/Logs: paths
- Runtime: worker_pool_size, timeout

## Code Patterns

- Repositories return domain types, services handle business logic
- Handlers use Gin context with JSON binding
- All routes under `/api/` are protected except auth endpoints
- Frontend embeds into binary when built (SPA fallback in `registerUIRoutes`)
