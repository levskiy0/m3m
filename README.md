<p align="left">
  <img src="web/ui/public/logo.svg" width="180" alt="M3M Logo" />
</p>

<h1 align="left">Mini Services Manager</h1>

<p align="left">
  <strong>Run JavaScript services with zero infrastructure overhead</strong>
</p>

<p align="left">
  <img alt="Go" src="https://img.shields.io/badge/Go-1.23-00ADD8?logo=go&logoColor=white" />
  <img alt="License" src="https://img.shields.io/badge/License-MIT-blue" />
  <img alt="Platform" src="https://img.shields.io/badge/Platform-Linux%20%7C%20Mac%20%7C%20Windows-lightgrey" />
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

### Docker (Recommended)

The fastest way to get started. Requires Docker and Docker Compose.

```bash
# Quick start with Docker Compose
curl -O https://raw.githubusercontent.com/levskiy0/m3m/main/docker-compose.yml

# Start M3M with MongoDB
docker compose up -d

# Create admin user
docker compose exec m3m /app/m3m new-admin admin@example.com yourpassword

# Open http://localhost:3000 in your browser
```

**Using a specific version:**

```bash
# Edit docker-compose.yml and change the image tag:
# image: ghcr.io/levskiy0/m3m:v1.0.0

# Or use environment variable
M3M_VERSION=v1.0.0 docker compose up -d
```

**Environment variables:**

| Variable | Default | Description |
|----------|---------|-------------|
| `M3M_PORT` | `3000` | Host port to expose |
| `M3M_URI` | `http://localhost:3000` | Public URL |
| `M3M_JWT_SECRET` | `change-me-in-production` | JWT signing secret |
| `M3M_DB_NAME` | `m3m` | MongoDB database name |

**Production example:**

```bash
M3M_PORT=8080 \
M3M_URI=https://m3m.example.com \
M3M_JWT_SECRET=$(openssl rand -hex 32) \
docker compose up -d
```

### Build Docker Image Locally

```bash
# Clone repository
git clone https://github.com/levskiy0/m3m.git
cd m3m

# Build image
docker build -t m3m:local .

# Or with specific version tag
docker build --build-arg VERSION=v1.0.0 -t m3m:v1.0.0 .

# Run with local build (edit docker-compose.yml)
# Uncomment the 'build:' section and comment out 'image:'
docker compose up -d
```

### From Source

```bash
# Clone
git clone https://github.com/levskiy0/m3m.git
cd m3m

# Build (requires Go 1.23+, Node.js 20+)
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
