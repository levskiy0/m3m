import { useTitle } from '@/hooks';

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

      {/* Overview */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">What is M3M?</h2>
        <p className="text-muted-foreground">
          M3M was created to simplify deploying personal services and small automations.
          When you need to quickly build a small service that handles webhooks with custom logic,
          collects statistics, and stores data — without the overhead of configuring Docker,
          managing resources, and orchestrating deployments.
        </p>
        <p className="text-muted-foreground">
          Write your service in JavaScript, deploy it on your own hosting — and you're done.
          Need to update or remove it? Just as quick. Need statistics or database records?
          They're all built-in.
        </p>
        <p className="text-muted-foreground">
          Each service runs in an isolated JavaScript runtime with access to built-in modules
          for routing, scheduling, storage, database operations, goal tracking, and more.
        </p>
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
