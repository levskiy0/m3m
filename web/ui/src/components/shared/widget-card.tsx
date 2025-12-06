import { Target, Trash2 } from 'lucide-react';

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
import type { Goal, GoalStats, Widget } from '@/types';

interface WidgetCardProps {
  widget: Widget;
  goal: Goal;
  stats?: GoalStats;
  onDelete: () => void;
}

export function WidgetCard({ widget, goal, stats, onDelete }: WidgetCardProps) {
  const sparklineData = stats?.dailyStats?.slice(-14).map((d) => d.value) || [];
  const isDailyCounter = goal.type === 'daily_counter';

  // Render based on variant
  if (widget.variant === 'mini') {
    // Mini variant - compact with sparkline for daily counters
    return (
      <Card className="group relative">
        <Button
          variant="ghost"
          size="icon"
          className="absolute top-2 right-2 size-6 opacity-0 group-hover:opacity-100 transition-opacity"
          onClick={onDelete}
        >
          <Trash2 className="size-3.5 text-muted-foreground hover:text-destructive" />
        </Button>
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
                <div className="text-2xl font-bold">
                  {stats?.value?.toLocaleString() ?? 0}
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

  if (widget.variant === 'detailed') {
    // Detailed variant - larger card with more info
    return (
      <Card className="group relative">
        <Button
          variant="ghost"
          size="icon"
          className="absolute top-2 right-2 size-6 opacity-0 group-hover:opacity-100 transition-opacity"
          onClick={onDelete}
        >
          <Trash2 className="size-3.5 text-muted-foreground hover:text-destructive" />
        </Button>
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
              <div className="text-2xl font-bold">
                {stats?.value?.toLocaleString() ?? 0}
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

  // Simple variant - just the number
  return (
    <Card className="group relative">
      <Button
        variant="ghost"
        size="icon"
        className="absolute top-2 right-2 size-6 opacity-0 group-hover:opacity-100 transition-opacity"
        onClick={onDelete}
      >
        <Trash2 className="size-3.5 text-muted-foreground hover:text-destructive" />
      </Button>
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
        <div className="text-2xl font-bold">
          {stats?.value?.toLocaleString() ?? 0}
        </div>
        <p className="text-xs text-muted-foreground mt-1">
          {goal.description || (isDailyCounter ? 'Daily counter' : 'Total count')}
        </p>
      </CardContent>
    </Card>
  );
}
