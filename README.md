<p align="left">
  <img src="web/ui/public/logo.svg" width="180" alt="M3M Logo" />
</p>

<h1 align="left">Mini Services Manager</h1>

<p align="left">
  <strong>Run JavaScript services with zero infrastructure overhead</strong>
</p>

<p align="left">
  <img alt="Go" src="https://img.shields.io/badge/Go-1.24-00ADD8?logo=go&logoColor=white" />
  <img alt="License" src="https://img.shields.io/badge/License-MIT-blue" />
  <img alt="Platform" src="https://img.shields.io/badge/Platform-Linux%20%7C%20Mac-lightgrey" />
  <img alt="Status" src="https://img.shields.io/badge/Status-Development-orange" />
</p>

---

> **Note:** M3M is currently in active development. Some features may be incomplete, unstable, or subject to change. Use in production at your own risk.

---

## Table of Contents

- [Why M3M?](#why-m3m)
  - [The Problem: Infrastructure Overkill](#the-problem-infrastructure-overkill)
  - [The Solution: Unified Runtime](#the-solution-unified-runtime)
  - [The Philosophy](#the-philosophy)
- [Example](#example)
- [Service Lifecycle](#service-lifecycle)
  - [Boot Phase](#1-boot-phase)
  - [Start Phase](#2-start-phase)
  - [Shutdown Phase](#3-shutdown-phase)
- [Performance](#performance)
- [Installation](#installation)
  - [Quick Install](#quick-install-recommended-docker--mongodb)
  - [Start & Setup](#start--setup)
  - [All Commands](#all-commands)
  - [Directory Structure](#directory-structure)
  - [Adding Plugins](#adding-plugins)
  - [Configuration](#configuration)
- [From Source (Manual)](#from-source-manual)
  - [Standalone Binary](#standalone-binary)
  - [Configuration](#configuration-1)
  - [Database Drivers](#database-drivers)
- [CLI Commands](#cli-commands)
- [Accessing Services](#accessing-services)
- [Development](#development)
- [License](#license)

---

## Why M3M?

### The Problem: Infrastructure Overkill

We live in an era where deploying a simple 50-line integration or a webhook handler requires a Dockerfile, a docker-compose.yml, 500MB of Node.js base images, and a bunch of external dependencies.

Your VPS is bloated with layers, your RAM is consumed by identical Node processes, and updating a single line of code feels like a ritual.

### The Solution: Unified Runtime

M3M was born to kill this complexity. It is a single, lightweight binary that acts as a **private ecosystem** for your scripts.

Instead of managing separate containers for every task, you have one environment that provides everything out of the box.

### The Philosophy

M3M is for developers who want to write logic in the browser (or their editor), hit **Save**, and walk away. It's for those who value their time and their server resources.

- No more NPM dependency hell for micro-tasks
- No more Docker storage anxiety
- No more cloud bills for things that should run on your $5 VPS

---

## Example

```javascript
$service.boot(() => {
  const users = $database.collection('users');

  // HTTP Endpoint
  $router.get('/users', (ctx) => {
    return { users: users.find({}) };
  });

  $router.post('/users', (ctx) => {
    const user = users.insert(ctx.body);
    return { user };
  });

  // Cron Task (Every hour)
  $schedule.every('1h', () => {
    $logger.info('Total users:', users.count({}));
  });
});

$service.start(() => {
  $logger.info('Service is ready!');
});
```

**Built-in modules** (accessed with `$` prefix):
- **Core:** `$service`, `$router`, `$schedule`, `$logger`, `$env`
- **Data:** `$database`, `$storage`, `$goals`
- **Network:** `$http`, `$smtp`
- **Utils:** `$crypto`, `$encoding`, `$utils`, `$delayed`, `$validator`
- **Media:** `$image`, `$draw`

---

## Service Lifecycle

Every service has three lifecycle phases managed by the `$service` module:

### 1. Boot Phase

Initialize your service, set up routes, and configure modules.

```javascript
$service.boot(() => {
  // Set up routes
  $router.get('/hello', (ctx) => {
    return { message: 'Hello World!' };
  });

  // Configure scheduler
  $schedule.every('1h', () => {
    $logger.info('Hourly task running');
  });
});
```

### 2. Start Phase

Called when service is ready. Good for initial data loading.

```javascript
$service.start(() => {
  $logger.info('Service started!');
  // Load initial data, etc.
});
```

### 3. Shutdown Phase

Called when service is stopping. Clean up resources here.

```javascript
$service.shutdown(() => {
  $logger.info('Service stopping...');
  // Close connections, save state, etc.
});
```

---

## Performance

Runs comfortably on a **$5/mo VPS** with 512MB RAM alongside 20+ active services. The Go binary + embedded UI weighs ~30MB. No Node.js runtime, no container layers.

---

## Installation

### Quick Install (Recommended: Docker + MongoDB)

```bash
# Download script
curl -fsSL https://raw.githubusercontent.com/levskiy0/m3m/main/m3m.sh -o m3m.sh
chmod +x m3m.sh

# Install latest release
./m3m.sh install

# Or install specific version
./m3m.sh install v1.0.0

# Or install development version
./m3m.sh install main
```

This will:
- Clone the repository to `.m3m/src`
- Build Docker image locally
- Generate config file with random JWT secret

### Start & Setup

```bash
# Start server
./m3m.sh start

# Create admin user
./m3m.sh admin admin@example.com yourpassword

# View logs
./m3m.sh logs
```

Open http://localhost:8080 and login.

### All Commands

```bash
./m3m.sh install [version]    # Install (latest, v1.0.0, or main)
./m3m.sh update [version]     # Update to version
./m3m.sh rebuild              # Rebuild image (after adding plugins)
./m3m.sh start                # Start the container
./m3m.sh stop                 # Stop the container
./m3m.sh restart              # Restart the container
./m3m.sh logs                 # Show container logs
./m3m.sh status               # Show container status
./m3m.sh version              # Show installed/latest versions
./m3m.sh admin <email> <pw>   # Create admin user
./m3m.sh config               # Show config
./m3m.sh uninstall            # Remove M3M (keeps data)
```

### Directory Structure

```
.m3m/
├── config      # Configuration file
├── version     # Installed version
├── src/        # Repository clone
├── plugins/    # Plugin sources (copy here, then rebuild)
└── data/       # Persistent data (mounted to container)
    ├── storage/    # File storage for services
    ├── plugins/    # Compiled plugins (.so files)
    ├── logs/       # Application logs
    └── mongodb/    # MongoDB database files
```

### Adding Plugins

```bash
# Copy plugin source to plugins directory
cp -r my-telegram-plugin .m3m/plugins/

# Rebuild image (plugins are compiled during build)
./m3m.sh rebuild
```

### Configuration

Edit `.m3m/config`:

```bash
# M3M Configuration
M3M_PORT=8080
M3M_JWT_SECRET=<auto-generated>
M3M_SERVER_URI=http://localhost:8080
```

| Variable | Default | Description |
|----------|---------|-------------|
| `M3M_PORT` | `8080` | Server port |
| `M3M_JWT_SECRET` | auto | JWT signing secret (auto-generated) |
| `M3M_SERVER_URI` | `http://localhost:8080` | Public server URI |

---

## From Source (Manual)

```bash
# Clone
git clone https://github.com/levskiy0/m3m.git
cd m3m

# Build (requires Go 1.24+, Node.js 20+)
make build

# Create first admin user
./build/m3m new-admin admin@example.com yourpassword

# Run
./build/m3m serve
```

### Standalone Binary

You can also just download the binary and run — **zero configuration required**:

```bash
# Download binary from releases
./m3m new-admin admin@example.com yourpassword
./m3m serve
```

On first run, M3M automatically creates `config.yaml` with:
- **SQLite database** (embedded, no external server needed)
- **Random JWT secret** (secure by default)

> **Note:** Docker installation via `m3m.sh` is recommended for production as it includes automatic updates, backup management, and proper process supervision.

### Configuration

Default config file: `config.yaml` (auto-created on first run)

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  uri: "http://127.0.0.1:8080"

database:
  driver: "sqlite"  # "mongodb" or "sqlite"

# MongoDB - external server required
mongodb:
  uri: "mongodb://localhost:27017"
  database: "m3m"

# SQLite - embedded, no external dependencies (default)
sqlite:
  path: "./data"
  database: "m3m"

jwt:
  secret: "change-me-in-production"
  expiration: 168h

storage:
  path: "./storage"

logging:
  level: "info"
  path: "./logs"
```

#### Database Drivers

| Driver | Config | Requirements |
|--------|--------|--------------|
| `sqlite` | `driver: "sqlite"` | **None** — embedded in binary |
| `mongodb` | `driver: "mongodb"` | External MongoDB server |

**SQLite mode** uses embedded [FerretDB](https://www.ferretdb.com/) with SQLite backend. All MongoDB query syntax (`$eq`, `$gt`, `$in`, etc.) works identically in both modes — switch anytime without code changes.

---

## CLI Commands

```bash
# Start server
m3m serve

# Start with custom config
m3m serve -c /path/to/config.yaml

# Create root admin
m3m new-admin admin@example.com yourpassword

# Check version
m3m version
```

---

## Accessing Services

Once running, your service endpoints are available at:

```
GET  /r/{project-slug}/your-route
POST /r/{project-slug}/your-route
```

---

## Development

```bash
make init          # Initialize project
make dev           # Run backend in dev mode
make web-dev       # Run frontend dev server
make test          # Run tests
```

---

## License

MIT
