import { useState, useEffect, useRef } from 'react';
import { ArrowDown, Download } from 'lucide-react';

import { Button } from '@/components/ui/button';
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
import type { LogEntry } from '@/types';

type LogLevel = 'all' | 'debug' | 'info' | 'warn' | 'error';

interface LogsViewerProps {
  logs: LogEntry[];
  height?: string;
  limit?: number;
  emptyMessage?: string;
  onDownload?: () => void;
  className?: string;
}

export function LogsViewer({
  logs,
  height = 'calc(100vh - 280px)',
  limit,
  emptyMessage = 'No logs available',
  onDownload,
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

  const scrollToBottom = () => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  };

  return (
    <div className={cn('flex flex-col overflow-hidden', className)} style={{ height }}>
      {/* Header */}
      <div className="flex items-center justify-between flex-shrink-0 px-4 py-3 border-b bg-background">
        <div className="flex items-center gap-4">
          <span className="text-sm font-medium">
            {displayLogs.length} log entries
          </span>
          <Select
            value={levelFilter}
            onValueChange={(v) => setLevelFilter(v as LogLevel)}
          >
            <SelectTrigger className="w-32 h-8">
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
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setAutoScroll(!autoScroll)}
            className={cn('h-8', autoScroll && 'bg-muted')}
          >
            Auto-scroll: {autoScroll ? 'On' : 'Off'}
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={scrollToBottom}
          >
            <ArrowDown className="size-4" />
          </Button>
          {onDownload && (
            <Button variant="outline" size="sm" onClick={onDownload} className="h-8">
              <Download className="mr-2 size-4" />
              Download
            </Button>
          )}
        </div>
      </div>

      {/* Logs Content */}
      <ScrollArea ref={scrollRef} className="flex-1 bg-zinc-950 font-mono text-xs">
        {displayLogs.length === 0 ? (
          <div className="flex items-center justify-center h-full text-muted-foreground py-12">
            {logs.length === 0 ? emptyMessage : 'No logs match the filter'}
          </div>
        ) : (
          <div className="space-y-0.5 p-4">
            {displayLogs.map((log, index) => (
              <div key={index} className="flex gap-2 text-gray-300">
                <span className="text-gray-500 shrink-0">
                  {new Date(log.timestamp).toLocaleTimeString()}
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
