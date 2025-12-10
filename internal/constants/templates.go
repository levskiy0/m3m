package constants

// DefaultServiceCode is the template code for new services
const DefaultServiceCode = `// M3M Service Template

$service.boot(() => {
  $logger.info('Service booting...');
});

$service.start(() => {
  $logger.info('Service started!');

  // Example: HTTP route
  $router.get('/health', (ctx) => {
    return ctx.response(200, { status: 'ok' });
  });

  // Example: Scheduled task
  $schedule.every('1h', () => {
    $logger.info('Hourly task executed');
  });
});

$service.shutdown(() => {
  $logger.info('Service shutting down...');
});
`
