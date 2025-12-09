import { useEffect, useCallback, useRef } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { wsClient, type WSEventHandlers } from '@/lib/websocket';
import { queryKeys } from '@/lib/query-keys';
import type { RuntimeStats } from '@/types';

interface UseWebSocketOptions {
  projectId?: string;
  enabled?: boolean;
  onMonitor?: (data: RuntimeStats) => void;
  onLog?: () => void;
  onRunning?: (running: boolean) => void;
  onGoals?: (data: unknown) => void;
}

export function useWebSocket(options: UseWebSocketOptions = {}) {
  const {
    projectId,
    enabled = true,
    onMonitor,
    onLog,
    onRunning,
    onGoals,
  } = options;

  const queryClient = useQueryClient();
  const handlersRef = useRef<WSEventHandlers>({});

  // Update handlers when callbacks change
  useEffect(() => {
    handlersRef.current = {
      onMonitor: (pid, data) => {
        if (projectId && pid === projectId) {
          // Update query cache with new monitor data
          queryClient.setQueryData(
            queryKeys.projects.status(projectId),
            data
          );
          onMonitor?.(data as RuntimeStats);
        }
      },
      onLog: (pid, data) => {
        if (projectId && pid === projectId && data.hasNewLogs) {
          onLog?.();
        }
      },
      onRunning: (pid, data) => {
        if (projectId && pid === projectId) {
          const newStatus = data.running ? 'running' : 'stopped';

          // Update status query cache
          queryClient.setQueryData(
            queryKeys.projects.status(projectId),
            (old: Record<string, unknown> | undefined) => {
              return {
                ...old,
                status: newStatus,
              };
            }
          );

          // Update project detail cache
          queryClient.setQueryData(
            queryKeys.projects.detail(projectId),
            (old: Record<string, unknown> | undefined) => {
              if (!old) return old;
              return { ...old, status: newStatus };
            }
          );

          // Also invalidate to get fresh data
          queryClient.invalidateQueries({
            queryKey: queryKeys.projects.status(projectId),
          });
          queryClient.invalidateQueries({
            queryKey: queryKeys.projects.detail(projectId),
          });

          onRunning?.(data.running);
        }
      },
      onGoals: (pid, data) => {
        if (projectId && pid === projectId) {
          onGoals?.(data);
        }
      },
    };

    wsClient.setHandlers(handlersRef.current);
  }, [projectId, queryClient, onMonitor, onLog, onRunning, onGoals]);

  // Connect and subscribe
  useEffect(() => {
    if (!enabled) return;

    wsClient.connect();

    if (projectId) {
      wsClient.subscribe(projectId);
    }

    return () => {
      if (projectId) {
        wsClient.unsubscribe(projectId);
      }
    };
  }, [projectId, enabled]);

  const subscribe = useCallback((pid: string) => {
    wsClient.subscribe(pid);
  }, []);

  const unsubscribe = useCallback((pid: string) => {
    wsClient.unsubscribe(pid);
  }, []);

  return {
    subscribe,
    unsubscribe,
    isConnected: wsClient.isConnected(),
  };
}

// Hook for managing WebSocket connection at app level
export function useWebSocketConnection() {
  useEffect(() => {
    wsClient.connect();

    return () => {
      wsClient.disconnect();
    };
  }, []);
}
