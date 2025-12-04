import type { EnvType, FieldType, GoalType, ReleaseTag } from '@/types';

export const PRESET_COLORS = [
  { name: 'slate', value: '#64748b' },
  { name: 'gray', value: '#6b7280' },
  { name: 'zinc', value: '#71717a' },
  { name: 'red', value: '#ef4444' },
  { name: 'orange', value: '#f97316' },
  { name: 'amber', value: '#f59e0b' },
  { name: 'yellow', value: '#eab308' },
  { name: 'lime', value: '#84cc16' },
  { name: 'green', value: '#22c55e' },
  { name: 'emerald', value: '#10b981' },
  { name: 'teal', value: '#14b8a6' },
  { name: 'cyan', value: '#06b6d4' },
  { name: 'sky', value: '#0ea5e9' },
  { name: 'blue', value: '#3b82f6' },
  { name: 'indigo', value: '#6366f1' },
  { name: 'violet', value: '#8b5cf6' },
  { name: 'purple', value: '#a855f7' },
  { name: 'fuchsia', value: '#d946ef' },
  { name: 'pink', value: '#ec4899' },
  { name: 'rose', value: '#f43f5e' },
] as const;

export const FIELD_TYPES: { value: FieldType; label: string }[] = [
  { value: 'string', label: 'String' },
  { value: 'text', label: 'Text' },
  { value: 'number', label: 'Number' },
  { value: 'float', label: 'Float' },
  { value: 'bool', label: 'Boolean' },
  { value: 'document', label: 'Document (JSON)' },
  { value: 'file', label: 'File' },
  { value: 'ref', label: 'Reference' },
  { value: 'date', label: 'Date' },
  { value: 'datetime', label: 'DateTime' },
];

export const ENV_TYPES: { value: EnvType; label: string }[] = [
  { value: 'string', label: 'String' },
  { value: 'text', label: 'Text' },
  { value: 'json', label: 'JSON' },
  { value: 'integer', label: 'Integer' },
  { value: 'float', label: 'Float' },
  { value: 'boolean', label: 'Boolean' },
];

export const GOAL_TYPES: { value: GoalType; label: string }[] = [
  { value: 'counter', label: 'Counter' },
  { value: 'daily_counter', label: 'Daily Counter' },
];

export const RELEASE_TAGS: { value: ReleaseTag; label: string }[] = [
  { value: 'stable', label: 'Stable' },
  { value: 'hot-fix', label: 'Hot Fix' },
  { value: 'night-build', label: 'Night Build' },
  { value: 'develop', label: 'Develop' },
];

export const DEFAULT_SERVICE_CODE = `// M3M Service Template

service.boot(() => {
  logger.info('Service booting...');
});

service.start(() => {
  logger.info('Service started!');

  // Example: HTTP route
  router.get('/health', (ctx) => {
    return router.response(200, { status: 'ok' });
  });

  // Example: Scheduled task
  schedule.every('1h', () => {
    logger.info('Hourly task executed');
  });
});

service.shutdown(() => {
  logger.info('Service shutting down...');
});
`;

export const FILE_ICONS: Record<string, string> = {
  // Images
  'image/jpeg': 'image',
  'image/png': 'image',
  'image/gif': 'image',
  'image/webp': 'image',
  'image/svg+xml': 'image',
  // Documents
  'application/pdf': 'file-text',
  'application/msword': 'file-text',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document': 'file-text',
  // Code
  'application/json': 'file-code',
  'application/javascript': 'file-code',
  'text/javascript': 'file-code',
  'text/html': 'file-code',
  'text/css': 'file-code',
  'text/yaml': 'file-code',
  'text/x-yaml': 'file-code',
  // Text
  'text/plain': 'file-text',
  'text/markdown': 'file-text',
  // Archives
  'application/zip': 'file-archive',
  'application/x-tar': 'file-archive',
  'application/gzip': 'file-archive',
  // Default
  default: 'file',
};

export const EDITABLE_MIME_TYPES = [
  'application/json',
  'application/javascript',
  'text/javascript',
  'text/html',
  'text/css',
  'text/yaml',
  'text/x-yaml',
  'text/plain',
  'text/markdown',
];

export const IMAGE_MIME_TYPES = [
  'image/jpeg',
  'image/png',
  'image/gif',
  'image/webp',
  'image/svg+xml',
];
