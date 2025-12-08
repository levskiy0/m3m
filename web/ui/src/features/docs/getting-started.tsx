import { useTitle } from '@/hooks';
import { AlertTriangle } from 'lucide-react';

export function GettingStartedPage() {
  useTitle('Getting Started - Docs');

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight mb-2">Getting Started</h1>
        <p className="text-muted-foreground text-lg">
          Learn how to create and deploy JavaScript services with M3M.
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
          <h3 className="font-medium">The Philosophy: "Code First, Infrastructure Never"</h3>
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

      {/* CLI Commands */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">CLI Commands</h2>
        <p className="text-muted-foreground mb-4">
          M3M provides a command-line interface for managing the server and users:
        </p>

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

      {/* Service Lifecycle */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Service Lifecycle</h2>
        <p className="text-muted-foreground mb-4">
          Every service has three lifecycle phases managed by the <code className="font-mono text-sm bg-muted px-1.5 py-0.5 rounded">$service</code> module:
        </p>

        <div className="space-y-3">
          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-1">1. Boot Phase</h3>
            <p className="text-sm text-muted-foreground mb-2">
              Initialize your service, set up routes, and configure modules.
            </p>
            <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
              <code>{`$service.boot(() => {
  // Set up routes
  $router.get('/hello', (ctx) => {
    return { message: 'Hello World!' };
  });

  // Configure scheduler
  $schedule.every('1h', () => {
    $logger.info('Hourly task running');
  });
});`}</code>
            </pre>
          </div>

          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-1">2. Start Phase</h3>
            <p className="text-sm text-muted-foreground mb-2">
              Called when service is ready. Good for initial data loading.
            </p>
            <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
              <code>{`$service.start(() => {
  $logger.info('Service started!');
  // Load initial data, etc.
});`}</code>
            </pre>
          </div>

          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-1">3. Shutdown Phase</h3>
            <p className="text-sm text-muted-foreground mb-2">
              Called when service is stopping. Clean up resources here.
            </p>
            <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
              <code>{`$service.shutdown(() => {
  $logger.info('Service stopping...');
  // Close connections, save state, etc.
});`}</code>
            </pre>
          </div>
        </div>
      </section>

      {/* Basic Example */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Complete Example</h2>
        <p className="text-muted-foreground mb-4">
          Here's a complete service with API endpoints, scheduled tasks, and database usage:
        </p>

        <pre className="bg-muted rounded-lg p-4 text-sm overflow-x-auto">
          <code>{`// Boot phase - configure everything
$service.boot(() => {
  // Get database collections
  const users = $database.collection('users');

  // API Routes
  $router.get('/users', (ctx) => {
    const allUsers = users.find({});
    return { users: allUsers };
  });

  $router.post('/users', (ctx) => {
    const user = users.insert(ctx.body);
    return { user };
  });

  $router.get('/users/:id', (ctx) => {
    const user = users.findOne({ _id: ctx.params.id });
    if (!user) {
      return ctx.response(404, { error: 'User not found' });
    }
    return { user };
  });

  // Scheduled task - runs every hour
  $schedule.every('1h', () => {
    const count = users.count({});
    $logger.info('Total users:', count);
  });

  // Daily cleanup at midnight
  $schedule.cron('0 0 * * *', () => {
    $logger.info('Daily cleanup running');
  });
});

// Start phase
$service.start(() => {
  $logger.info('Service is ready!');
});`}</code>
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
