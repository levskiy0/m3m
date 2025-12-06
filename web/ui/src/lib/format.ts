export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

export function formatNumber(value: number): string {
  if (value >= 1000000) return `${(value / 1000000).toFixed(1)}M`;
  if (value >= 1000) return `${(value / 1000).toFixed(1)}K`;
  return value.toLocaleString();
}

/**
 * Format uptime in seconds to human-readable string (e.g., "2d 5h" or "45m")
 */
export function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);

  if (days > 0) {
    return `${days}d ${hours}h`;
  }
  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  }
  return `${minutes}m`;
}

/**
 * Format a date to relative time (e.g., "2 hours ago")
 */
export function formatRelativeTime(date: Date | string): string {
  const now = new Date();
  const then = new Date(date);
  const diff = now.getTime() - then.getTime();
  const seconds = Math.floor(diff / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) return `${days}d ago`;
  if (hours > 0) return `${hours}h ago`;
  if (minutes > 0) return `${minutes}m ago`;
  return 'just now';
}

/**
 * Calculate trend percentage between two periods
 */
export function calculateTrend(
  recentValues: number[],
  olderValues: number[]
): number | null {
  if (recentValues.length === 0 || olderValues.length === 0) return null;

  const recentSum = recentValues.reduce((a, b) => a + b, 0);
  const olderSum = olderValues.reduce((a, b) => a + b, 0);

  if (olderSum === 0) return null;

  return ((recentSum - olderSum) / olderSum) * 100;
}

/**
 * Format environment value for display
 */
export function formatEnvValue(
  value: string,
  type: string,
  options: { masked?: boolean; maxLength?: number } = {}
): string {
  const { masked = false, maxLength = 50 } = options;

  if (masked) {
    return '••••••••';
  }

  if (type === 'boolean') {
    return value === 'true' ? 'True' : 'False';
  }

  if (type === 'json') {
    try {
      return JSON.stringify(JSON.parse(value), null, 2);
    } catch {
      return value;
    }
  }

  return value.length > maxLength ? value.slice(0, maxLength) + '...' : value;
}
