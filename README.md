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

### Quick Install (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/levskiy0/m3m/main/m3m.sh -o m3m.sh
chmod +x m3m.sh
./m3m.sh install
```

This will:
- Clone the repository to `~/.m3m/src`
- Build Docker image locally
- Create config file at `~/.m3m/.env`

**Important:** Edit `~/.m3m/.env` and set a secure `M3M_JWT_SECRET`.

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
./m3m.sh install     # Clone repo and build Docker image
./m3m.sh start       # Start the container
./m3m.sh stop        # Stop the container
./m3m.sh restart     # Restart the container
./m3m.sh logs        # Show container logs
./m3m.sh status      # Show container status
./m3m.sh admin       # Create admin: ./m3m.sh admin <email> <password>
./m3m.sh update      # Pull latest changes and rebuild
./m3m.sh rebuild     # Rebuild image (after adding plugins)
./m3m.sh uninstall   # Remove M3M completely
```

### Directory Structure

```
~/.m3m/
├── src/        # Repository clone
├── data/       # Persistent data (mounted to container)
│   ├── storage/    # File storage for services
│   ├── plugins/    # Compiled plugins (.so files)
│   ├── logs/       # Application logs
│   └── mongodb/    # MongoDB database files
├── plugins/    # Plugin sources (copy here, then rebuild)
└── .env        # Environment config
```

### Adding Plugins

```bash
# Copy plugin source to plugins directory
cp -r my-telegram-plugin ~/.m3m/plugins/

# Rebuild image (plugins are compiled during build)
./m3m.sh rebuild
```

### Environment Variables

Edit `~/.m3m/.env`:

| Variable | Default | Description |
|----------|---------|-------------|
| `M3M_JWT_SECRET` | - | **Required.** JWT signing secret |
| `M3M_SERVER_URI` | `http://localhost:8080` | Public server URI |
| `M3M_SERVER_PORT` | `8080` | Server port |
| `M3M_JWT_EXPIRATION` | `168h` | JWT token expiration |
| `M3M_LOGGING_LEVEL` | `info` | Log level (debug/info/warn/error) |

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

### Configuration

Default config file: `config.yaml`

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  uri: "http://localhost:8080"

mongodb:
  uri: "mongodb://localhost:27017"
  database: "m3m"

jwt:
  secret: "change-me-in-production"
  expiration: 168h

storage:
  path: "./storage"

logs:
  path: "./logs"
```

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
