import { useTitle } from '@/hooks';

export function LifecyclePage() {
  useTitle('Life Cycle - Docs');

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight mb-2">Life Cycle</h1>
        <p className="text-muted-foreground text-lg">
          Understanding how to develop, debug, and deploy your services.
        </p>
      </div>

      {/* Development vs Release */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Development Branches vs Releases</h2>
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
        <h2 className="text-xl font-semibold">Running and Debugging</h2>

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
        <h2 className="text-xl font-semibold">Debugging Tips</h2>

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
