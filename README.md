<p align="left">
  <img src="web/ui/public/favicon.svg" width="80" alt="M3M Logo" />
</p>

<h1 align="left">M3M — Mini Services Manager</h1>

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

## Why M3M?

I built M3M because I was tired of writing `Dockerfile` and `docker-compose.yml` for simple integrations and cron scripts.

Every 50-line webhook handler or simple microservice required **500MB of base images** (Node.js, Python, PHP), a stack of external dependencies, and annoying deployment rituals.

I wanted a tool where I could just write JavaScript, hit **Run**, and have persistent storage, scheduling, and HTTP routing **out of the box**.

* No heavy containers.
* No orchestration overhead.
* No NPM dependency hell for micro-tasks.

---

## Features

| |                                                                                             |
|---|---------------------------------------------------------------------------------------------|
| **Zero Config** | Single binary deployment. No NPM or Composer install, no Docker required.                   |
| **Batteries Included** | Built-in Router, Scheduler, Key-Value DB, Logger, HTTP client, and more.                    |
| **Isolated Runtime** | Powered by [Goja](https://github.com/dop251/goja) (pure Go JS engine). Safe and lightweight. |
| **Plugin System** | Support for `.so` plugins for high-performance.                                             |
| **Stateful** | Persistent storage built-in. Your data survives restarts.                                   |
| **Web UI** | Monaco-based code editor, real-time logs, project management — all in the browser.          |

---

## Show Me The Code

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

## Performance

Runs comfortably on a **$5/mo VPS** with 512MB RAM alongside 20+ active services. The Go binary + embedded UI weighs ~30MB. No Node.js runtime, no container layers.

---

## Installation

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
