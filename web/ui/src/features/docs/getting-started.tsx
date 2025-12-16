import { useTitle } from '@/hooks';
import { AlertTriangle, Terminal, Server, Database, Folder } from 'lucide-react';

export function GettingStartedPage() {
  useTitle('Getting Started - Docs');

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight mb-2">Getting Started</h1>
        <p className="text-muted-foreground text-lg">
          Learn how to install, configure and run M3M.
        </p>
      </div>

      {/* Development Warning Banner */}
      <div className="flex items-start gap-3 p-4 bg-orange-500/10 border border-orange-500/30 rounded-lg">
        <AlertTriangle className="h-5 w-5 text-orange-500 shrink-0 mt-0.5" />
        <div className="text-sm">
          <p className="font-medium text-orange-600 dark:text-orange-400">
            Project Under Development
          </p>
          <p className="text-muted-foreground mt-1">
            M3M is currently in active development. Some features may be incomplete, unstable, or
            subject to change. Use in production at your own risk.
          </p>
        </div>
      </div>

      {/* Why M3M? */}
      <section className="space-y-6">
        <h2 className="text-xl font-semibold">Why M3M?</h2>

        {/* The Problem */}
        <div className="space-y-3">
          <h3 className="font-medium text-red-500">The Problem: Infrastructure Overkill</h3>
          <p className="text-muted-foreground">
            We live in an era where deploying a simple 50-line integration or a webhook handler
            requires a Dockerfile, a docker-compose.yml, 500MB of Node.js base images, and a bunch
            of external dependencies.
          </p>
          <p className="text-muted-foreground">
            Your VPS is bloated with layers, your RAM is consumed by identical Node processes,
            and updating a single line of code feels like a ritual.
          </p>
        </div>

        {/* The Solution */}
        <div className="space-y-3">
          <h3 className="font-medium text-green-600 dark:text-green-400">
            The Solution: Unified Runtime
          </h3>
          <p className="text-muted-foreground">
            M3M was born to kill this complexity. It is a single, lightweight binary that acts
            as a <strong>private ecosystem</strong> for your scripts.
          </p>
          <p className="text-muted-foreground">
            Instead of managing separate containers for every task, you have one environment
            that provides everything out of the box:
          </p>
          <div className="grid sm:grid-cols-2 gap-3">
            <div className="border rounded-lg p-3">
              <p className="text-sm">
                <span className="text-muted-foreground">Need an API?</span>
                <br />
                <code className="font-mono text-sm bg-muted px-1.5 py-0.5 rounded">$router</code>
              </p>
            </div>
            <div className="border rounded-lg p-3">
              <p className="text-sm">
                <span className="text-muted-foreground">Need a cron job?</span>
                <br />
                <code className="font-mono text-sm bg-muted px-1.5 py-0.5 rounded">$schedule</code>
              </p>
            </div>
            <div className="border rounded-lg p-3">
              <p className="text-sm">
                <span className="text-muted-foreground">Need to save data?</span>
                <br />
                <code className="font-mono text-sm bg-muted px-1.5 py-0.5 rounded">$database</code>{' '}
                <span className="text-xs text-muted-foreground">(zero-config)</span>
              </p>
            </div>
            <div className="border rounded-lg p-3">
              <p className="text-sm">
                <span className="text-muted-foreground">Need background tasks?</span>
                <br />
                <code className="font-mono text-sm bg-muted px-1.5 py-0.5 rounded">$delayed</code>
              </p>
            </div>
          </div>
        </div>

        {/* The Philosophy */}
        <div className="space-y-3">
          <h3 className="font-medium">The Philosophy</h3>
          <p className="text-muted-foreground">
            M3M is for developers who want to write logic in the browser (or their editor),
            hit <strong>Save</strong>, and walk away. It's for those who value their time
            and their server resources.
          </p>
          <ul className="space-y-2 text-sm text-muted-foreground">
            <li className="flex items-start gap-2">
              <span className="text-green-500 mt-0.5">✓</span>
              <span>No more NPM dependency hell for micro-tasks</span>
            </li>
            <li className="flex items-start gap-2">
              <span className="text-green-500 mt-0.5">✓</span>
              <span>No more Docker storage anxiety</span>
            </li>
            <li className="flex items-start gap-2">
              <span className="text-green-500 mt-0.5">✓</span>
              <span>No more cloud bills for things that should run on your $5 VPS</span>
            </li>
          </ul>
        </div>
      </section>

      {/* Installation */}
      <section className="space-y-6">
        <div className="flex items-center gap-2">
          <Server className="h-5 w-5 text-primary" />
          <h2 className="text-xl font-semibold">Installation</h2>
        </div>

        {/* Docker Install */}
        <div className="space-y-4">
          <h3 className="font-medium">Quick Install (Docker + MongoDB)</h3>
          <p className="text-sm text-muted-foreground">
            Recommended for production. Includes MongoDB, automatic updates, and process supervision.
          </p>

          <div className="border rounded-lg p-4 space-y-3">
            <div>
              <p className="text-sm font-medium mb-2">1. Download and install</p>
              <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
                <code>{`# Download script
curl -fsSL https://raw.githubusercontent.com/levskiy0/m3m/main/m3m.sh -o m3m.sh
chmod +x m3m.sh

# Install latest release
./m3m.sh install

# Or install specific version
./m3m.sh install v1.0.0`}</code>
              </pre>
            </div>

            <div>
              <p className="text-sm font-medium mb-2">2. Start and create admin user</p>
              <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
                <code>{`# Start server
./m3m.sh start

# Create admin user
./m3m.sh admin admin@example.com yourpassword

# View logs
./m3m.sh logs`}</code>
              </pre>
            </div>

            <p className="text-sm text-muted-foreground">
              Open <code className="font-mono bg-muted px-1.5 py-0.5 rounded">http://localhost:8080</code> and login.
            </p>
          </div>
        </div>

        {/* Binary Install */}
        <div className="space-y-4">
          <h3 className="font-medium">Standalone Binary</h3>
          <p className="text-sm text-muted-foreground">
            Zero configuration required. Uses embedded SQLite database.
          </p>

          <div className="border rounded-lg p-4 space-y-3">
            <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
              <code>{`# Download binary from releases, then:
./m3m new-admin admin@example.com yourpassword
./m3m serve`}</code>
            </pre>
            <p className="text-sm text-muted-foreground">
              On first run, M3M automatically creates <code className="font-mono bg-muted px-1.5 py-0.5 rounded">config.yaml</code> with
              SQLite database and random JWT secret.
            </p>
          </div>
        </div>

        {/* Build from Source */}
        <div className="space-y-4">
          <h3 className="font-medium">Build from Source</h3>
          <p className="text-sm text-muted-foreground">
            Requires Go 1.24+ and Node.js 20+.
          </p>

          <div className="border rounded-lg p-4">
            <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
              <code>{`git clone https://github.com/levskiy0/m3m.git
cd m3m
make build
./build/m3m new-admin admin@example.com yourpassword
./build/m3m serve`}</code>
            </pre>
          </div>
        </div>
      </section>

      {/* Directory Structure */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <Folder className="h-5 w-5 text-primary" />
          <h2 className="text-xl font-semibold">Directory Structure</h2>
        </div>
        <p className="text-sm text-muted-foreground">
          After Docker installation, you'll have the following structure:
        </p>
        <pre className="bg-muted rounded-lg p-4 text-sm overflow-x-auto">
          <code>{`.m3m/
├── config      # Configuration file
├── version     # Installed version
├── src/        # Repository clone
├── plugins/    # Plugin sources (copy here, then rebuild)
└── data/       # Persistent data (mounted to container)
    ├── storage/    # File storage for services
    ├── plugins/    # Compiled plugins (.so files)
    ├── logs/       # Application logs
    └── mongodb/    # MongoDB database files`}</code>
        </pre>
      </section>

      {/* Database Drivers */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <Database className="h-5 w-5 text-primary" />
          <h2 className="text-xl font-semibold">Database Drivers</h2>
        </div>
        <p className="text-sm text-muted-foreground">
          M3M supports two database backends. Both use identical MongoDB query syntax.
        </p>

        <div className="border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-muted">
              <tr>
                <th className="px-4 py-2 text-left font-medium">Driver</th>
                <th className="px-4 py-2 text-left font-medium">Config</th>
                <th className="px-4 py-2 text-left font-medium">Requirements</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              <tr>
                <td className="px-4 py-2 font-mono">sqlite</td>
                <td className="px-4 py-2 text-muted-foreground">
                  <code className="bg-muted px-1 rounded">driver: "sqlite"</code>
                </td>
                <td className="px-4 py-2 text-muted-foreground">None — embedded in binary</td>
              </tr>
              <tr>
                <td className="px-4 py-2 font-mono">mongodb</td>
                <td className="px-4 py-2 text-muted-foreground">
                  <code className="bg-muted px-1 rounded">driver: "mongodb"</code>
                </td>
                <td className="px-4 py-2 text-muted-foreground">External MongoDB server</td>
              </tr>
            </tbody>
          </table>
        </div>

        <p className="text-sm text-muted-foreground">
          <strong>SQLite mode</strong> uses embedded FerretDB with SQLite backend.
          All MongoDB query syntax (<code className="font-mono bg-muted px-1 rounded">$eq</code>,
          <code className="font-mono bg-muted px-1 rounded">$gt</code>,
          <code className="font-mono bg-muted px-1 rounded">$in</code>, etc.) works identically in both modes —
          switch anytime without code changes.
        </p>
      </section>

      {/* CLI Commands */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <Terminal className="h-5 w-5 text-primary" />
          <h2 className="text-xl font-semibold">CLI Commands</h2>
        </div>

        <div className="space-y-3">
          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-1">Start the Server</h3>
            <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
              <code>{`m3m serve`}</code>
            </pre>
            <p className="text-sm text-muted-foreground mt-2">
              Starts the M3M server using the default config file (<code className="font-mono text-sm bg-muted px-1.5 py-0.5 rounded">config.yaml</code>).
            </p>
          </div>

          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-1">Custom Config File</h3>
            <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
              <code>{`m3m serve -c /path/to/config.yaml
# or
m3m serve --config /path/to/config.yaml`}</code>
            </pre>
            <p className="text-sm text-muted-foreground mt-2">
              Use a custom configuration file path.
            </p>
          </div>

          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-1">Create Root Admin</h3>
            <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
              <code>{`m3m new-admin admin@example.com yourpassword`}</code>
            </pre>
            <p className="text-sm text-muted-foreground mt-2">
              Creates the first root administrator user. This user has full access to all projects and can manage other users.
            </p>
          </div>

          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-1">Check Version</h3>
            <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
              <code>{`m3m version`}</code>
            </pre>
            <p className="text-sm text-muted-foreground mt-2">
              Displays the current M3M version.
            </p>
          </div>
        </div>
      </section>

      {/* Docker Commands */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Docker Commands (m3m.sh)</h2>
        <p className="text-sm text-muted-foreground mb-4">
          If you installed via Docker, use these commands:
        </p>

        <pre className="bg-muted rounded-lg p-4 text-sm overflow-x-auto">
          <code>{`./m3m.sh install [version]    # Install (latest, v1.0.0, or main)
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
./m3m.sh uninstall            # Remove M3M (keeps data)`}</code>
        </pre>
      </section>

      {/* Configuration */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Configuration</h2>
        <p className="text-sm text-muted-foreground mb-4">
          Default config file: <code className="font-mono bg-muted px-1.5 py-0.5 rounded">config.yaml</code> (auto-created on first run)
        </p>

        <pre className="bg-muted rounded-lg p-4 text-sm overflow-x-auto">
          <code>{`server:
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
  path: "./logs"`}</code>
        </pre>
      </section>

      {/* Accessing Your Service */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Accessing Your Service</h2>
        <p className="text-muted-foreground mb-4">
          Once your service is running, API endpoints are available at:
        </p>
        <pre className="bg-muted rounded-lg p-4 text-sm">
          <code>{`GET  /r/{project-slug}/your-route
POST /r/{project-slug}/your-route`}</code>
        </pre>
        <p className="text-sm text-muted-foreground">
          The project slug is set in project settings.
        </p>
      </section>
    </div>
  );
}
