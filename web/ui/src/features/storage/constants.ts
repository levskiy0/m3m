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
