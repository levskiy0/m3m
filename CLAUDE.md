# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

M3M (Mini Services Manager) is a platform for creating and managing JavaScript mini-services/workers. It provides a Go backend with a JavaScript runtime (GOJA) that allows users to write and deploy small services with built-in modules for routing, scheduling, storage, database, and more.

## Build & Development Commands

```bash
# Initialize project (create directories, tidy deps)
make init

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
cd web/ui && npm run lint      # Lint frontend code
cd web/ui && npm run lint:fix  # Fix linting issues
cd web/ui && npm run format    # Format code with Prettier

# Tests
make test                              # Run all tests
go test -v ./internal/service/...      # Run tests for a specific package
go test -v -run TestFunctionName ./... # Run a single test by name

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
- **internal/domain/**: Domain models (Project, Pipeline, Goal, Environment, Model, User, Widget)
- **internal/repository/**: MongoDB repositories (one per domain entity)
- **internal/service/**: Business logic layer with validation (model_validation.go, model_schema_validation.go)
- **internal/handler/**: HTTP handlers (Gin framework)
- **internal/middleware/**: Auth (JWT) and CORS middleware
- **internal/runtime/**: JavaScript runtime manager using GOJA
- **internal/websocket/**: WebSocket support for real-time logs
- **internal/plugin/**: Plugin loader for extending runtime

#### Backend Handler → Service → Repository Flow

```
Handler (HTTP) → Service (Business Logic) → Repository (MongoDB)
     ↓
  runtime_handler.go  →  RuntimeManager  →  Project/Pipeline repos
  project_handler.go  →  ProjectService  →  ProjectRepository
  model_handler.go    →  ModelService    →  ModelRepository
  storage_handler.go  →  StorageService  →  (file system)
```

#### Key Handlers
- **auth_handler.go**: Login/logout endpoints
- **project_handler.go**: CRUD for projects, start/stop control
- **pipeline_handler.go**: Branch and release management
- **model_handler.go**: Schema definitions CRUD
- **goal_handler.go**: Metrics tracking endpoints
- **runtime_handler.go**: Execute code, manage running services, logs
- **storage_handler.go**: File storage operations
- **websocket_handler.go**: Real-time log streaming

### JavaScript Runtime

The runtime (`internal/runtime/runtime.go`) provides services with lifecycle management and built-in modules:

**Lifecycle hooks (service module):**
- `service.boot(callback)` - initialization phase
- `service.start(callback)` - when service is ready
- `service.shutdown(callback)` - graceful shutdown

**Built-in modules** (accessed with `$` prefix in JS code, e.g., `$logger`, `$router`):
- Core: logger, router, schedule, env, service
- Data: storage, database, goals
- Network: http, smtp
- Utilities: crypto, encoding, utils, delayed, validator
- Media: image, draw

Type definitions for Monaco IntelliSense are in `internal/runtime/modules/types.go`. Schema validation logic is in `schema.go`.

### Frontend (React + Vite)

Located in `web/ui/`:
- React 19 with TypeScript
- Tailwind CSS v3 with tailwindcss-animate
- Radix UI primitives with shadcn/ui components (new-york style)
- Routing: react-router-dom v7
- State: zustand for global state, @tanstack/react-query for server state
- Forms: react-hook-form with zod validation
- Built assets are embedded into Go binary via `web/static.go`

#### Frontend Structure

```
src/
├── api/                    # API client functions (one file per domain)
│   ├── client.ts           # Axios instance, auth interceptors
│   ├── auth.ts, projects.ts, models.ts, pipeline.ts, goals.ts, etc.
├── components/
│   ├── ui/                 # shadcn/ui base components
│   ├── layout/             # App layout, sidebar (app-layout.tsx, app-sidebar.tsx)
│   └── shared/             # Reusable components (code-editor, logs-viewer, widget-card, etc.)
├── features/               # Feature modules (domain-based)
│   ├── auth/               # Login page
│   ├── projects/           # Project list, CRUD
│   ├── pipeline/           # Branch/release management, code editor
│   ├── models/             # Schema definitions
│   │   ├── schema/         # Schema editor components
│   │   ├── model-data/     # Data browser components
│   │   └── lib/            # Schema utilities
│   ├── goals/              # Metrics tracking UI
│   ├── environment/        # Environment variables management
│   ├── storage/            # File storage browser
│   ├── users/              # User management
│   ├── docs/               # Documentation viewer
│   └── modules/            # Runtime modules info
├── hooks/                  # Custom React hooks
│   ├── use-crud-mutation.ts    # Generic CRUD mutation helper
│   ├── use-form-dialog.ts      # Dialog state management
│   ├── use-project-runtime.ts  # Runtime state (logs, status)
│   └── use-websocket.ts        # WebSocket connection hook
├── providers/              # React context providers
│   ├── auth-provider.tsx   # Auth state context
│   ├── query-provider.tsx  # React Query client
│   └── theme-provider.tsx  # Dark/light theme
├── lib/                    # Utilities
│   ├── query-keys.ts       # React Query key factory
│   ├── websocket.ts        # WebSocket client class
│   └── utils.ts            # Helper functions (cn, formatters)
├── routes/                 # Route definitions (index.tsx)
└── types/                  # TypeScript types (mirrors backend domain)
```

#### Key Patterns

- **API layer**: Each `api/*.ts` file exports functions that call backend endpoints. Uses axios with JWT interceptors in `client.ts`.
- **React Query**: All server state uses `@tanstack/react-query`. Query keys defined in `lib/query-keys.ts`.
- **Feature isolation**: Each feature folder is self-contained with its own components, hooks, and sub-features.
- **Form handling**: Uses `react-hook-form` + `zod` schemas. shadcn/ui Field components for consistent form layouts.
- **Real-time logs**: WebSocket via `use-websocket.ts` hook connecting to `/api/ws/logs/{projectId}`.

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
- Access current user in handlers: `middleware.GetCurrentUser(c)` or `middleware.GetCurrentUserID(c)`
