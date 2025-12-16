import { useTitle } from '@/hooks';
import { FileCode, Package, Play, RefreshCw, Bug } from 'lucide-react';

export function LifecyclePage() {
  useTitle('Development Guide - Docs');

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight mb-2">Development Guide</h1>
        <p className="text-muted-foreground text-lg">
          Understanding how to develop, debug, and deploy your services.
        </p>
      </div>

      {/* Service Lifecycle */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <RefreshCw className="h-5 w-5 text-primary" />
          <h2 className="text-xl font-semibold">Service Lifecycle</h2>
        </div>
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

      {/* Multi-file Support */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <FileCode className="h-5 w-5 text-primary" />
          <h2 className="text-xl font-semibold">Multi-file Support</h2>
        </div>
        <p className="text-muted-foreground mb-4">
          Split your code into multiple files using <code className="font-mono text-sm bg-muted px-1.5 py-0.5 rounded">$require</code> and{' '}
          <code className="font-mono text-sm bg-muted px-1.5 py-0.5 rounded">$exports</code>.
        </p>

        <div className="space-y-4">
          {/* $exports */}
          <div className="border rounded-lg p-4">
            <h3 className="font-medium font-mono mb-2">$exports(object)</h3>
            <p className="text-sm text-muted-foreground mb-3">
              Export values from the current file for use by other files.
            </p>
            <div className="space-y-2">
              <p className="text-xs font-medium text-muted-foreground">utils.js</p>
              <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
                <code>{`// Helper functions
function formatDate(date) {
  return new Date(date).toISOString().split('T')[0];
}

function generateId() {
  return $utils.uuid();
}

const VERSION = '1.0.0';

// Export multiple values
$exports({
  formatDate,
  generateId,
  VERSION
});`}</code>
              </pre>
            </div>
          </div>

          {/* $require */}
          <div className="border rounded-lg p-4">
            <h3 className="font-medium font-mono mb-2">$require(name)</h3>
            <p className="text-sm text-muted-foreground mb-3">
              Import exports from another file. Files are executed once and cached.
            </p>
            <div className="space-y-2">
              <p className="text-xs font-medium text-muted-foreground">main.js</p>
              <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
                <code>{`// Import specific exports
const { formatDate, generateId, VERSION } = $require('utils');

$service.boot(() => {
  $logger.info('App version:', VERSION);

  $router.post('/items', (ctx) => {
    const item = {
      id: generateId(),
      createdAt: formatDate(new Date()),
      ...ctx.body
    };
    return $database.collection('items').insert(item);
  });
});`}</code>
              </pre>
            </div>
          </div>

          {/* Example: Project Structure */}
          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-2">Example: Project Structure</h3>
            <p className="text-sm text-muted-foreground mb-3">
              Organize your code into logical modules:
            </p>
            <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto mb-3">
              <code>{`Pipeline files:
├── main.js       # Entry point with $service.boot
├── routes.js     # API route handlers
├── db.js         # Database helpers
└── utils.js      # Utility functions`}</code>
            </pre>

            <div className="space-y-3">
              <div>
                <p className="text-xs font-medium text-muted-foreground mb-1">db.js</p>
                <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
                  <code>{`const users = $database.collection('users');
const posts = $database.collection('posts');

$exports({ users, posts });`}</code>
                </pre>
              </div>

              <div>
                <p className="text-xs font-medium text-muted-foreground mb-1">routes.js</p>
                <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
                  <code>{`const { users, posts } = $require('db');

function setupRoutes() {
  $router.get('/users', () => users.find({}));
  $router.get('/posts', () => posts.find({}));
}

$exports({ setupRoutes });`}</code>
                </pre>
              </div>

              <div>
                <p className="text-xs font-medium text-muted-foreground mb-1">main.js</p>
                <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
                  <code>{`const { setupRoutes } = $require('routes');

$service.boot(() => {
  setupRoutes();
  $logger.info('Routes configured');
});`}</code>
                </pre>
              </div>
            </div>
          </div>

          {/* Important Notes */}
          <div className="bg-muted/50 rounded-lg p-4">
            <h4 className="font-medium mb-2">Important Notes</h4>
            <ul className="text-sm text-muted-foreground space-y-1">
              <li className="flex items-start gap-2">
                <span className="text-primary mt-0.5">•</span>
                <span>Files are executed <strong>once</strong> and cached — subsequent <code className="font-mono bg-muted px-1 rounded">$require</code> calls return the cached exports</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-primary mt-0.5">•</span>
                <span>Use file name without extension: <code className="font-mono bg-muted px-1 rounded">$require('utils')</code> not <code className="font-mono bg-muted px-1 rounded">$require('utils.js')</code></span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-primary mt-0.5">•</span>
                <span>Circular dependencies are detected and will throw an error</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-primary mt-0.5">•</span>
                <span>All global modules (<code className="font-mono bg-muted px-1 rounded">$logger</code>, <code className="font-mono bg-muted px-1 rounded">$database</code>, etc.) are available in all files</span>
              </li>
            </ul>
          </div>
        </div>
      </section>

      {/* Development vs Release */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <Package className="h-5 w-5 text-primary" />
          <h2 className="text-xl font-semibold">Development Branches vs Releases</h2>
        </div>
        <p className="text-muted-foreground">
          Each project has a Pipeline with two types of code versions:
        </p>

        <div className="space-y-3">
          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-1">Development Branch (develop)</h3>
            <p className="text-sm text-muted-foreground mb-2">
              This is where you write and test your code. Development branches:
            </p>
            <ul className="text-sm text-muted-foreground list-disc list-inside space-y-1">
              <li>Can be edited at any time</li>
              <li>Are used for testing and debugging</li>
              <li>
                <strong>Do not auto-start on server restart</strong> — you need to manually start them
              </li>
              <li>Perfect for active development and experimentation</li>
            </ul>
          </div>

          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-1">Releases</h3>
            <p className="text-sm text-muted-foreground mb-2">
              Immutable snapshots of your code for production use. Releases:
            </p>
            <ul className="text-sm text-muted-foreground list-disc list-inside space-y-1">
              <li>Cannot be edited after creation</li>
              <li>Are stable versions of your service</li>
              <li>
                <strong>Auto-start on server restart</strong> if the project has auto-start enabled
              </li>
              <li>Should be used for production workloads</li>
            </ul>
          </div>
        </div>
      </section>

      {/* Where to Write Code */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Where to Write Code</h2>
        <p className="text-muted-foreground mb-4">
          Navigate to your project and open the <strong>Pipeline</strong> tab. Here you'll find:
        </p>

        <div className="border rounded-lg p-4 space-y-3">
          <div>
            <h3 className="font-medium mb-1">Code Editor</h3>
            <p className="text-sm text-muted-foreground">
              The built-in Monaco editor (same as VS Code) provides syntax highlighting,
              autocomplete, and IntelliSense for all built-in modules (
              <code className="font-mono text-sm bg-muted px-1.5 py-0.5 rounded">$logger</code>,{' '}
              <code className="font-mono text-sm bg-muted px-1.5 py-0.5 rounded">$router</code>,
              etc.).
            </p>
          </div>
          <div>
            <h3 className="font-medium mb-1">File Tabs</h3>
            <p className="text-sm text-muted-foreground">
              Create multiple files for your service. Click <strong>+</strong> to add a new file.
              Use <code className="font-mono text-sm bg-muted px-1.5 py-0.5 rounded">$require</code> to import between files.
            </p>
          </div>
          <div>
            <h3 className="font-medium mb-1">Branch Selection</h3>
            <p className="text-sm text-muted-foreground">
              Use the dropdown to switch between branches. The{' '}
              <code className="font-mono text-sm bg-muted px-1.5 py-0.5 rounded">develop</code>{' '}
              branch is created automatically with each project.
            </p>
          </div>
        </div>
      </section>

      {/* How to Run and Debug */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <Play className="h-5 w-5 text-primary" />
          <h2 className="text-xl font-semibold">Running and Debugging</h2>
        </div>

        <div className="space-y-3">
          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-1">1. Save Your Code</h3>
            <p className="text-sm text-muted-foreground">
              Make changes in the editor and click <strong>Save</strong> (or use{' '}
              <kbd className="px-1.5 py-0.5 bg-muted rounded text-xs">Cmd/Ctrl + S</kbd>).
            </p>
          </div>

          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-1">2. Start the Service</h3>
            <p className="text-sm text-muted-foreground">
              Click the <strong>Start</strong> button in the Pipeline page. The service will begin
              running with your current code.
            </p>
          </div>

          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-1">3. Check Logs</h3>
            <p className="text-sm text-muted-foreground">
              The logs panel shows real-time output from your service. Use{' '}
              <code className="font-mono text-sm bg-muted px-1.5 py-0.5 rounded">
                $logger.info()
              </code>
              ,{' '}
              <code className="font-mono text-sm bg-muted px-1.5 py-0.5 rounded">
                $logger.error()
              </code>
              , etc. for debugging.
            </p>
          </div>

          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-1">4. Stop and Iterate</h3>
            <p className="text-sm text-muted-foreground">
              Click <strong>Stop</strong> to stop the service, make changes, and start again. The
              cycle is quick — no Docker builds or deployments.
            </p>
          </div>
        </div>
      </section>

      {/* Creating a Release */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Creating a Release</h2>
        <p className="text-muted-foreground mb-4">
          When your service is ready for production:
        </p>

        <div className="space-y-3">
          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-1">1. Create Release</h3>
            <p className="text-sm text-muted-foreground">
              In the Pipeline page, click <strong>Create Release</strong>. This creates an immutable
              snapshot of your current code.
            </p>
          </div>

          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-1">2. Select Active Release</h3>
            <p className="text-sm text-muted-foreground">
              Choose which release to run. The selected release will be used when the service
              starts.
            </p>
          </div>

          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-1">3. Start</h3>
            <p className="text-sm text-muted-foreground">
              In the project Overview, click <strong>Start</strong> and choose release for running.
            </p>
          </div>
        </div>
      </section>

      {/* Debugging Tips */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <Bug className="h-5 w-5 text-primary" />
          <h2 className="text-xl font-semibold">Debugging Tips</h2>
        </div>

        <pre className="bg-muted rounded-lg p-4 text-sm overflow-x-auto">
          <code>{`// Log information at different levels
$logger.debug('Detailed debug info');
$logger.info('General information');
$logger.warn('Warning message');
$logger.error('Error occurred', err);

// Log objects
$logger.info('User data:', { id: 1, name: 'John' });

// Track execution flow
$service.boot(() => {
  $logger.info('Boot phase started');

  $router.get('/test', (ctx) => {
    $logger.info('Request received:', ctx.params);
    return { ok: true };
  });

  $logger.info('Boot phase completed');
});

$service.start(() => {
  $logger.info('Service is now running');
});`}</code>
        </pre>
      </section>

      {/* Built-in Modules Overview */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Built-in Modules</h2>
        <p className="text-sm text-muted-foreground mb-4">
          All modules are available globally with <code className="font-mono bg-muted px-1 rounded">$</code> prefix:
        </p>

        <div className="grid sm:grid-cols-2 gap-3">
          <div className="border rounded-lg p-3">
            <p className="text-sm font-medium mb-1">Core</p>
            <p className="text-xs text-muted-foreground">
              <code className="font-mono bg-muted px-1 rounded">$service</code>{' '}
              <code className="font-mono bg-muted px-1 rounded">$router</code>{' '}
              <code className="font-mono bg-muted px-1 rounded">$schedule</code>{' '}
              <code className="font-mono bg-muted px-1 rounded">$logger</code>{' '}
              <code className="font-mono bg-muted px-1 rounded">$env</code>
            </p>
          </div>
          <div className="border rounded-lg p-3">
            <p className="text-sm font-medium mb-1">Data</p>
            <p className="text-xs text-muted-foreground">
              <code className="font-mono bg-muted px-1 rounded">$database</code>{' '}
              <code className="font-mono bg-muted px-1 rounded">$storage</code>{' '}
              <code className="font-mono bg-muted px-1 rounded">$goals</code>
            </p>
          </div>
          <div className="border rounded-lg p-3">
            <p className="text-sm font-medium mb-1">Network</p>
            <p className="text-xs text-muted-foreground">
              <code className="font-mono bg-muted px-1 rounded">$http</code>{' '}
              <code className="font-mono bg-muted px-1 rounded">$mail</code>
            </p>
          </div>
          <div className="border rounded-lg p-3">
            <p className="text-sm font-medium mb-1">Utils</p>
            <p className="text-xs text-muted-foreground">
              <code className="font-mono bg-muted px-1 rounded">$crypto</code>{' '}
              <code className="font-mono bg-muted px-1 rounded">$encoding</code>{' '}
              <code className="font-mono bg-muted px-1 rounded">$utils</code>{' '}
              <code className="font-mono bg-muted px-1 rounded">$validator</code>{' '}
              <code className="font-mono bg-muted px-1 rounded">$delayed</code>
            </p>
          </div>
          <div className="border rounded-lg p-3">
            <p className="text-sm font-medium mb-1">Media</p>
            <p className="text-xs text-muted-foreground">
              <code className="font-mono bg-muted px-1 rounded">$image</code>{' '}
              <code className="font-mono bg-muted px-1 rounded">$draw</code>
            </p>
          </div>
          <div className="border rounded-lg p-3">
            <p className="text-sm font-medium mb-1">UI</p>
            <p className="text-xs text-muted-foreground">
              <code className="font-mono bg-muted px-1 rounded">$ui</code>{' '}
              <code className="font-mono bg-muted px-1 rounded">$hook</code>
            </p>
          </div>
        </div>

        <p className="text-sm text-muted-foreground">
          See the <strong>API Reference</strong> section in the sidebar for detailed documentation on each module.
        </p>
      </section>

      {/* Important Notes */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Important Notes</h2>

        <div className="space-y-3">
          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-1">Development branches don't auto-start</h3>
            <p className="text-sm text-muted-foreground">
              After server restart, you need to manually navigate to the Pipeline page and start the
              service.
            </p>
          </div>

          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-1">Only releases auto-start</h3>
            <p className="text-sm text-muted-foreground">
              If you need your service to persist through restarts, create a release and enable
              auto-start in project settings.
            </p>
          </div>

          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-1">Code changes require restart</h3>
            <p className="text-sm text-muted-foreground">
              After saving code, stop and start the service to apply changes.
            </p>
          </div>
        </div>
      </section>
    </div>
  );
}
