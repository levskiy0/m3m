import { useState, useEffect, useRef } from 'react';
import { ArrowDownToLine, Download, RefreshCw } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { ScrollArea } from '@/components/ui/scroll-area';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { cn } from '@/lib/utils';
import { LOG_LEVEL_COLORS } from '@/lib/constants';
import { formatTime } from '@/lib/format';
import type { LogEntry } from '@/types';

type LogLevel = 'all' | 'debug' | 'info' | 'warn' | 'error';

interface LogsViewerProps {
  logs: LogEntry[];
  height?: string;
  limit?: number;
  emptyMessage?: string;
  onDownload?: () => void;
  onRefresh?: () => void;
  className?: string;
}

export function LogsViewer({
  logs,
  height = 'calc(100vh - 280px)',
  limit,
  emptyMessage = 'No logs available',
  onDownload,
  onRefresh,
  className,
}: LogsViewerProps) {
  const [levelFilter, setLevelFilter] = useState<LogLevel>('all');
  const [autoScroll, setAutoScroll] = useState(true);
  const scrollRef = useRef<HTMLDivElement>(null);

  const filteredLogs = logs.filter((log) =>
    levelFilter === 'all' ? true : log.level === levelFilter
  );

  const displayLogs = limit ? filteredLogs.slice(-limit) : filteredLogs;

  useEffect(() => {
    if (autoScroll && scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  return (
    <div className={cn('flex flex-col overflow-hidden bg-muted/30', className)} style={{ height }}>
      {/* Header */}
      <div className="flex items-center justify-between flex-shrink-0 h-[41px] px-3 border-b">
        <Select
          value={levelFilter}
          onValueChange={(v) => setLevelFilter(v as LogLevel)}
        >
          <SelectTrigger className="w-28 h-6 px-2 py-0 text-xs">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Levels</SelectItem>
            <SelectItem value="debug">Debug</SelectItem>
            <SelectItem value="info">Info</SelectItem>
            <SelectItem value="warn">Warning</SelectItem>
            <SelectItem value="error">Error</SelectItem>
          </SelectContent>
        </Select>
        <div className="flex items-center gap-0.5">
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setAutoScroll(!autoScroll)}
                className={cn('size-6', autoScroll && 'bg-accent')}
              >
                <ArrowDownToLine className="size-3.5" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Auto-scroll: {autoScroll ? 'On' : 'Off'}</TooltipContent>
          </Tooltip>
          {onRefresh && (
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="ghost" size="icon" onClick={onRefresh} className="size-6">
                  <RefreshCw className="size-3.5" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Refresh</TooltipContent>
            </Tooltip>
          )}
          {onDownload && (
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="ghost" size="icon" onClick={onDownload} className="size-6">
                  <Download className="size-3.5" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Download</TooltipContent>
            </Tooltip>
          )}
        </div>
      </div>

      {/* Logs Content */}
      <ScrollArea viewportRef={scrollRef} className="flex-1 bg-zinc-950 font-mono text-xs h-full">
        {displayLogs.length === 0 ? (
          <div className="flex items-center justify-center h-full text-muted-foreground py-12">
            {logs.length === 0 ? emptyMessage : 'No logs match the filter'}
          </div>
        ) : (
          <div className="space-y-0.5 p-4 mb-20">
            {displayLogs.map((log, index) => (
              <div key={index} className="flex gap-2 text-gray-300">
                <span className="text-gray-500 shrink-0">
                  {formatTime(log.timestamp)}
                </span>
                <span
                  className={cn(
                    'shrink-0 uppercase w-12',
                    LOG_LEVEL_COLORS[log.level]
                  )}
                >
                  [{log.level}]
                </span>
                <span className="text-gray-200 break-all">{log.message}</span>
              </div>
            ))}
          </div>
        )}
      </ScrollArea>
    </div>
  );
}
