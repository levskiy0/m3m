import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import type { ProjectStatus } from '@/types';

interface StatusBadgeProps {
  status: ProjectStatus;
  className?: string;
}

export function StatusBadge({ status, className }: StatusBadgeProps) {
  return (
    <Badge
      variant={status === 'running' ? 'default' : 'secondary'}
      className={cn(
        status === 'running' && 'bg-green-500 hover:bg-green-600',
        className
      )}
    >
      <span
        className={cn(
          'mr-1.5 size-1.5 rounded-full',
          status === 'running' ? 'bg-green-200 animate-pulse' : 'bg-muted-foreground'
        )}
      />
      {status === 'running' ? 'Running' : 'Stopped'}
    </Badge>
  );
}
