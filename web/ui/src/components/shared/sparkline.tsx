import { useMemo } from 'react';
import { cn } from '@/lib/utils';

interface SparklineProps {
  data: number[];
  width?: number;
  height?: number;
  strokeWidth?: number;
  color?: string;
  fillOpacity?: number;
  className?: string;
}

export function Sparkline({
  data,
  width = 100,
  height = 24,
  strokeWidth = 1.5,
  color = 'currentColor',
  fillOpacity = 0.1,
  className,
}: SparklineProps) {
  const path = useMemo(() => {
    if (!data || data.length === 0) return '';

    const min = Math.min(...data);
    const max = Math.max(...data);
    const range = max - min || 1;

    const points = data.map((value, index) => {
      const x = (index / (data.length - 1)) * width;
      const y = height - ((value - min) / range) * (height - 4) - 2;
      return `${x},${y}`;
    });

    return `M${points.join(' L')}`;
  }, [data, width, height]);

  const areaPath = useMemo(() => {
    if (!data || data.length === 0) return '';

    const min = Math.min(...data);
    const max = Math.max(...data);
    const range = max - min || 1;

    const points = data.map((value, index) => {
      const x = (index / (data.length - 1)) * width;
      const y = height - ((value - min) / range) * (height - 4) - 2;
      return `${x},${y}`;
    });

    return `M0,${height} L${points.join(' L')} L${width},${height} Z`;
  }, [data, width, height]);

  if (!data || data.length === 0) {
    return (
      <svg
        width={width}
        height={height}
        className={cn('text-muted-foreground/30', className)}
      >
        <line
          x1={0}
          y1={height / 2}
          x2={width}
          y2={height / 2}
          stroke="currentColor"
          strokeWidth={1}
          strokeDasharray="2,2"
        />
      </svg>
    );
  }

  // Show single point as a dot with dashed line
  if (data.length === 1) {
    return (
      <svg
        width={width}
        height={height}
        className={className}
        style={{ color }}
      >
        <line
          x1={0}
          y1={height / 2}
          x2={width}
          y2={height / 2}
          stroke="currentColor"
          strokeWidth={1}
          strokeDasharray="3,3"
          opacity={0.3}
        />
        <circle
          cx={width}
          cy={height / 2}
          r={3}
          fill="currentColor"
        />
      </svg>
    );
  }

  return (
    <svg
      width={width}
      height={height}
      className={className}
      style={{ color }}
    >
      {/* Fill area */}
      <path
        d={areaPath}
        fill="currentColor"
        fillOpacity={fillOpacity}
      />
      {/* Line */}
      <path
        d={path}
        fill="none"
        stroke="currentColor"
        strokeWidth={strokeWidth}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

interface SparklineCardProps {
  label: string;
  value: string | number;
  subtext?: string;
  data: number[];
  color?: string;
  icon?: React.ReactNode;
  className?: string;
}

export function SparklineCard({
  label,
  value,
  subtext,
  data,
  color = 'hsl(var(--primary))',
  icon,
  className,
}: SparklineCardProps) {
  return (
    <div className={cn('flex items-center justify-between gap-4', className)}>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          {icon}
          <span>{label}</span>
        </div>
        <div className="text-2xl font-bold mt-1">{value}</div>
        {subtext && (
          <div className="text-xs text-muted-foreground mt-0.5">{subtext}</div>
        )}
      </div>
      <Sparkline
        data={data}
        width={80}
        height={32}
        color={color}
        fillOpacity={0.15}
        strokeWidth={2}
      />
    </div>
  );
}
