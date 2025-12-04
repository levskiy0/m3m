import { useEffect, useRef, useState } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Download, ArrowDown } from 'lucide-react';

import { runtimeApi } from '@/api';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { cn } from '@/lib/utils';

type LogLevel = 'all' | 'debug' | 'info' | 'warn' | 'error';

const LOG_LEVEL_COLORS: Record<string, string> = {
  debug: 'text-muted-foreground',
  info: 'text-blue-500',
  warn: 'text-amber-500',
  error: 'text-red-500',
};

export function LogsPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const scrollRef = useRef<HTMLDivElement>(null);
  const [autoScroll, setAutoScroll] = useState(true);
  const [levelFilter, setLevelFilter] = useState<LogLevel>('all');

  const { data: logs = [], isLoading } = useQuery({
    queryKey: ['logs', projectId],
    queryFn: () => runtimeApi.logs(projectId!),
    enabled: !!projectId,
    refetchInterval: 3000, // Poll every 3 seconds
  });

  const filteredLogs = logs.filter((log) =>
    levelFilter === 'all' ? true : log.level === levelFilter
  );

  useEffect(() => {
    if (autoScroll && scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  const handleDownload = async () => {
    try {
      const blob = await runtimeApi.downloadLogs(projectId!);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${projectId}-logs.txt`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      console.error('Failed to download logs:', err);
    }
  };

  const scrollToBottom = () => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Logs</h1>
          <p className="text-muted-foreground">
            View runtime logs for this project
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Select
            value={levelFilter}
            onValueChange={(v) => setLevelFilter(v as LogLevel)}
          >
            <SelectTrigger className="w-32">
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
          <Button variant="outline" onClick={handleDownload}>
            <Download className="mr-2 size-4" />
            Download
          </Button>
        </div>
      </div>

      <Card>
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm font-medium">
              {filteredLogs.length} log entries
            </CardTitle>
            <div className="flex items-center gap-2">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setAutoScroll(!autoScroll)}
                className={cn(autoScroll && 'bg-muted')}
              >
                Auto-scroll: {autoScroll ? 'On' : 'Off'}
              </Button>
              <Button variant="ghost" size="icon" onClick={scrollToBottom}>
                <ArrowDown className="size-4" />
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <ScrollArea
            ref={scrollRef}
            className="h-[600px] rounded-md border bg-black p-4 font-mono text-sm"
          >
            {isLoading ? (
              <div className="space-y-2">
                {[1, 2, 3, 4, 5].map((i) => (
                  <Skeleton key={i} className="h-4 bg-gray-800" />
                ))}
              </div>
            ) : filteredLogs.length === 0 ? (
              <div className="flex items-center justify-center h-full text-muted-foreground">
                No logs available
              </div>
            ) : (
              <div className="space-y-1">
                {filteredLogs.map((log, index) => (
                  <div key={index} className="flex gap-2 text-gray-300">
                    <span className="text-gray-500 shrink-0">
                      [{new Date(log.timestamp).toLocaleString()}]
                    </span>
                    <span
                      className={cn(
                        'shrink-0 uppercase w-12',
                        LOG_LEVEL_COLORS[log.level]
                      )}
                    >
                      [{log.level}]
                    </span>
                    <span className="break-all">{log.message}</span>
                  </div>
                ))}
              </div>
            )}
          </ScrollArea>
        </CardContent>
      </Card>
    </div>
  );
}
