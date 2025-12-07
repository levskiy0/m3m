import { forwardRef } from 'react';
import {
  Target,
  Trash2,
  GripVertical,
  Edit,
  MemoryStick,
  Zap,
  Cpu,
  HardDrive,
  Database,
  Clock,
  CalendarClock,
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Sparkline } from '@/components/shared/sparkline';
import { cn } from '@/lib/utils';
import { formatBytes } from '@/lib/format';
import type { Goal, GoalStats, Widget, WidgetType, RuntimeStats } from '@/types';

interface WidgetCardProps {
  widget: Widget;
  // For goal widgets
  goal?: Goal;
  stats?: GoalStats;
  // For monitoring widgets
  runtimeStats?: RuntimeStats;
  isRunning?: boolean;
  // Actions
  onEdit?: () => void;
  onDelete: () => void;
  dragHandleProps?: React.HTMLAttributes<HTMLButtonElement>;
  isDragging?: boolean;
  style?: React.CSSProperties;
}

const WIDGET_CONFIG: Record<WidgetType, {
  icon: typeof Target;
  label: string;
  color: string;
}> = {
  goal: { icon: Target, label: 'Goal', color: '#6b7280' },
  memory: { icon: MemoryStick, label: 'Memory', color: '#8b5cf6' },
  requests: { icon: Zap, label: 'Requests', color: '#3b82f6' },
  cpu: { icon: Cpu, label: 'CPU', color: '#f59e0b' },
  storage: { icon: HardDrive, label: 'Storage', color: '#10b981' },
  database: { icon: Database, label: 'Database', color: '#ec4899' },
  uptime: { icon: Clock, label: 'Uptime', color: '#06b6d4' },
  jobs: { icon: CalendarClock, label: 'Jobs', color: '#84cc16' },
};

export const WidgetCard = forwardRef<HTMLDivElement, WidgetCardProps>(
  ({ widget, goal, stats, runtimeStats, isRunning, onEdit, onDelete, dragHandleProps, isDragging, style }, ref) => {
    const config = WIDGET_CONFIG[widget.type] || WIDGET_CONFIG.goal;
    const Icon = config.icon;

    const cardStyle = {
      ...style,
      gridColumn: `span ${widget.gridSpan || 1}`,
      opacity: isDragging ? 0.5 : 1,
    };

    // Render goal widget
    if (widget.type === 'goal' && goal) {
      return (
        <GoalWidgetCard
          ref={ref}
          widget={widget}
          goal={goal}
          stats={stats}
          onEdit={onEdit}
          onDelete={onDelete}
          dragHandleProps={dragHandleProps}
          isDragging={isDragging}
          style={cardStyle}
        />
      );
    }

    // Render monitoring widget
    return (
      <MonitoringWidgetCard
        ref={ref}
        widget={widget}
        runtimeStats={runtimeStats}
        isRunning={isRunning}
        config={config}
        Icon={Icon}
        onEdit={onEdit}
        onDelete={onDelete}
        dragHandleProps={dragHandleProps}
        style={cardStyle}
      />
    );
  }
);

WidgetCard.displayName = 'WidgetCard';

// Goal Widget Card
const GoalWidgetCard = forwardRef<HTMLDivElement, {
  widget: Widget;
  goal: Goal;
  stats?: GoalStats;
  onEdit?: () => void;
  onDelete: () => void;
  dragHandleProps?: React.HTMLAttributes<HTMLButtonElement>;
  isDragging?: boolean;
  style?: React.CSSProperties;
}>(({ widget, goal, stats, onEdit, onDelete, dragHandleProps, isDragging, style }, ref) => {
  const sparklineData = stats?.dailyStats?.slice(-14).map((d) => d.value) || [];
  const isDailyCounter = goal.type === 'daily_counter';

  // Mini variant
  if (widget.variant === 'mini') {
    return (
      <Card ref={ref} style={style} className={cn("group relative", isDragging && "z-50")}>
        <div className="absolute top-2 right-2 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
          {dragHandleProps && (
            <button
              className="cursor-grab active:cursor-grabbing text-muted-foreground hover:text-foreground p-1"
              {...dragHandleProps}
            >
              <GripVertical className="size-3.5" />
            </button>
          )}
          {onEdit && (
            <Button variant="ghost" size="icon" className="size-6" onClick={onEdit}>
              <Edit className="size-3.5 text-muted-foreground hover:text-foreground" />
            </Button>
          )}
          <Button variant="ghost" size="icon" className="size-6" onClick={onDelete}>
            <Trash2 className="size-3.5 text-muted-foreground hover:text-destructive" />
          </Button>
        </div>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <div className="flex items-center gap-2">
            <span
              className="size-2.5 rounded-full"
              style={{ backgroundColor: goal.color || '#6b7280' }}
            />
            <CardDescription className="text-sm font-medium">{goal.name}</CardDescription>
          </div>
          <Target className="size-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          {isDailyCounter && sparklineData.length > 0 ? (
            <div className="flex items-end justify-between gap-4">
              <div>
                <div className="flex items-baseline gap-2">
                  <span className="text-2xl font-bold">
                    {stats?.value?.toLocaleString() ?? 0}
                  </span>
                  {goal.showTotal && stats?.totalValue != null && (
                    <span className="text-sm text-muted-foreground">
                      / {stats.totalValue.toLocaleString()}
                    </span>
                  )}
                </div>
                <p className="text-xs text-muted-foreground mt-1">
                  {goal.description || 'Daily counter'}
                </p>
              </div>
              <Sparkline
                data={sparklineData}
                width={80}
                height={32}
                color={goal.color || '#6b7280'}
                strokeWidth={2}
                fillOpacity={0.15}
              />
            </div>
          ) : (
            <div>
              <div className="text-2xl font-bold">
                {stats?.value?.toLocaleString() ?? 0}
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                {goal.description || 'Total count'}
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    );
  }

  // Detailed variant
  if (widget.variant === 'detailed') {
    return (
      <Card ref={ref} style={style} className={cn("group relative", isDragging && "z-50")}>
        <div className="absolute top-2 right-2 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
          {dragHandleProps && (
            <button
              className="cursor-grab active:cursor-grabbing text-muted-foreground hover:text-foreground p-1"
              {...dragHandleProps}
            >
              <GripVertical className="size-3.5" />
            </button>
          )}
          {onEdit && (
            <Button variant="ghost" size="icon" className="size-6" onClick={onEdit}>
              <Edit className="size-3.5 text-muted-foreground hover:text-foreground" />
            </Button>
          )}
          <Button variant="ghost" size="icon" className="size-6" onClick={onDelete}>
            <Trash2 className="size-3.5 text-muted-foreground hover:text-destructive" />
          </Button>
        </div>
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span
                className="size-3 rounded-full"
                style={{ backgroundColor: goal.color || '#6b7280' }}
              />
              <CardTitle className="text-sm font-medium">{goal.name}</CardTitle>
            </div>
            <Badge variant="secondary" className="text-xs">
              {isDailyCounter ? 'Daily' : 'Counter'}
            </Badge>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-end justify-between gap-4">
            <div>
              <div className="flex items-baseline gap-2">
                <span className="text-2xl font-bold">
                  {stats?.value?.toLocaleString() ?? 0}
                </span>
                {isDailyCounter && goal.showTotal && stats?.totalValue != null && (
                  <span className="text-sm text-muted-foreground">
                    / {stats.totalValue.toLocaleString()}
                  </span>
                )}
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                {goal.description || (isDailyCounter ? 'Daily counter' : 'Total count')}
              </p>
            </div>
            {isDailyCounter && sparklineData.length > 0 && (
              <Sparkline
                data={sparklineData}
                width={80}
                height={32}
                color={goal.color || '#6b7280'}
                strokeWidth={2}
                fillOpacity={0.15}
              />
            )}
          </div>
        </CardContent>
      </Card>
    );
  }

  // Simple variant
  return (
    <Card ref={ref} style={style} className={cn("group relative", isDragging && "z-50")}>
      <div className="absolute top-2 right-2 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
        {dragHandleProps && (
          <button
            className="cursor-grab active:cursor-grabbing text-muted-foreground hover:text-foreground p-1"
            {...dragHandleProps}
          >
            <GripVertical className="size-3.5" />
          </button>
        )}
        {onEdit && (
          <Button variant="ghost" size="icon" className="size-6" onClick={onEdit}>
            <Edit className="size-3.5 text-muted-foreground hover:text-foreground" />
          </Button>
        )}
        <Button variant="ghost" size="icon" className="size-6" onClick={onDelete}>
          <Trash2 className="size-3.5 text-muted-foreground hover:text-destructive" />
        </Button>
      </div>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <div className="flex items-center gap-2">
          <span
            className="size-2.5 rounded-full"
            style={{ backgroundColor: goal.color || '#6b7280' }}
          />
          <CardDescription className="text-sm font-medium">{goal.name}</CardDescription>
        </div>
        <Target className="size-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="flex items-baseline gap-2">
          <span className="text-2xl font-bold">
            {stats?.value?.toLocaleString() ?? 0}
          </span>
          {isDailyCounter && goal.showTotal && stats?.totalValue != null && (
            <span className="text-sm text-muted-foreground">
              / {stats.totalValue.toLocaleString()}
            </span>
          )}
        </div>
        <p className="text-xs text-muted-foreground mt-1">
          {goal.description || (isDailyCounter ? 'Daily counter' : 'Total count')}
        </p>
      </CardContent>
    </Card>
  );
});

GoalWidgetCard.displayName = 'GoalWidgetCard';

// Monitoring Widget Card
const MonitoringWidgetCard = forwardRef<HTMLDivElement, {
  widget: Widget;
  runtimeStats?: RuntimeStats;
  isRunning?: boolean;
  config: { label: string; color: string };
  Icon: typeof Target;
  onEdit?: () => void;
  onDelete: () => void;
  dragHandleProps?: React.HTMLAttributes<HTMLButtonElement>;
  style?: React.CSSProperties;
}>(({ widget, runtimeStats, isRunning, config, Icon, onEdit, onDelete, dragHandleProps, style }, ref) => {
  const getValue = (): string => {
    // Storage and Database don't require runtime to be running
    if (widget.type === 'storage') {
      return runtimeStats?.storage_bytes != null ? formatBytes(runtimeStats.storage_bytes) : '--';
    }
    if (widget.type === 'database') {
      return runtimeStats?.database_bytes != null ? formatBytes(runtimeStats.database_bytes) : '--';
    }

    // Other metrics require running service
    if (!isRunning || !runtimeStats) return '--';

    switch (widget.type) {
      case 'memory':
        return runtimeStats.memory?.alloc != null ? formatBytes(runtimeStats.memory.alloc) : '--';
      case 'requests':
        return runtimeStats.total_requests?.toLocaleString() ?? '--';
      case 'cpu':
        return runtimeStats.cpu_percent != null ? `${runtimeStats.cpu_percent.toFixed(1)}%` : '--';
      case 'uptime':
        return runtimeStats.uptime_formatted ?? '--';
      case 'jobs':
        return runtimeStats.scheduled_jobs?.toString() ?? '--';
      default:
        return '--';
    }
  };

  const getSubtext = (): string => {
    switch (widget.type) {
      case 'memory':
        return 'Current usage';
      case 'requests':
        return runtimeStats?.routes_count
          ? `${runtimeStats.routes_count} route${runtimeStats.routes_count !== 1 ? 's' : ''}`
          : 'No routes';
      case 'cpu':
        return 'Process usage';
      case 'storage':
        return 'Project files';
      case 'database':
        return 'Collections data';
      case 'uptime':
        return runtimeStats?.started_at
          ? `Since ${new Date(runtimeStats.started_at).toLocaleString()}`
          : 'Service not running';
      case 'jobs':
        return runtimeStats?.scheduler_active ? 'Scheduler active' : 'Scheduler inactive';
      default:
        return '';
    }
  };

  const getSparklineData = (): number[] | undefined => {
    if (!runtimeStats?.history) return undefined;

    switch (widget.type) {
      case 'memory':
        return runtimeStats.history.memory;
      case 'requests':
        return runtimeStats.history.requests;
      case 'cpu':
        return runtimeStats.history.cpu;
      case 'jobs':
        return runtimeStats.history.jobs;
      default:
        return undefined;
    }
  };

  const sparklineData = getSparklineData();

  return (
    <Card ref={ref} style={style} className="group relative">
      <div className="absolute top-2 right-2 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
        {dragHandleProps && (
          <button
            className="cursor-grab active:cursor-grabbing text-muted-foreground hover:text-foreground p-1"
            {...dragHandleProps}
          >
            <GripVertical className="size-3.5" />
          </button>
        )}
        {onEdit && (
          <Button variant="ghost" size="icon" className="size-6" onClick={onEdit}>
            <Edit className="size-3.5 text-muted-foreground hover:text-foreground" />
          </Button>
        )}
        <Button variant="ghost" size="icon" className="size-6" onClick={onDelete}>
          <Trash2 className="size-3.5 text-muted-foreground hover:text-destructive" />
        </Button>
      </div>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{config.label}</CardTitle>
        <Icon className="size-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="flex items-end justify-between gap-4">
          <div>
            <div className="text-2xl font-bold">{getValue()}</div>
            <p className="text-xs text-muted-foreground mt-1">{getSubtext()}</p>
          </div>
          {sparklineData && sparklineData.length > 0 && (
            <Sparkline
              data={sparklineData}
              width={80}
              height={32}
              color={config.color}
              strokeWidth={2}
              fillOpacity={0.15}
            />
          )}
        </div>
      </CardContent>
    </Card>
  );
});

MonitoringWidgetCard.displayName = 'MonitoringWidgetCard';
