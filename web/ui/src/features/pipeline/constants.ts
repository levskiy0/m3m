import type { ReleaseTag } from '@/types';

export const RELEASE_TAGS: { value: ReleaseTag; label: string }[] = [
  { value: 'stable', label: 'Stable' },
  { value: 'hot-fix', label: 'Hot Fix' },
  { value: 'night-build', label: 'Night Build' },
  { value: 'develop', label: 'Develop' },
];

export const DEFAULT_SERVICE_CODE = `// M3M Service Template

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
`;
