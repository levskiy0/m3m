import type { LucideIcon } from 'lucide-react';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
} from '@/components/ui/card';
import { Sparkline } from '@/components/shared/sparkline';

interface MetricCardProps {
  label: string;
  value: string | number;
  subtext?: string;
  icon: LucideIcon;
  sparklineData?: number[];
  sparklineColor?: string;
}

export function MetricCard({
  label,
  value,
  subtext,
  icon: Icon,
  sparklineData,
  sparklineColor = 'hsl(var(--primary))',
}: MetricCardProps) {
  const hasSparkline = sparklineData && sparklineData.length > 0;

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardDescription className="text-sm font-medium">{label}</CardDescription>
        <Icon className="size-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        {hasSparkline ? (
          <div className="flex items-end justify-between gap-4">
            <div>
              <div className="text-2xl font-bold">{value}</div>
              {subtext && (
                <p className="text-xs text-muted-foreground mt-1">{subtext}</p>
              )}
            </div>
            <Sparkline
              data={sparklineData}
              width={80}
              height={32}
              color={sparklineColor}
              strokeWidth={2}
              fillOpacity={0.15}
            />
          </div>
        ) : (
          <>
            <div className="text-2xl font-bold">{value}</div>
            {subtext && (
              <p className="text-xs text-muted-foreground mt-1">{subtext}</p>
            )}
          </>
        )}
      </CardContent>
    </Card>
  );
}
