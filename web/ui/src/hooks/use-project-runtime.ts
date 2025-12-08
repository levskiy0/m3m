import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useCallback, useRef } from 'react';
import { toast } from 'sonner';

import { runtimeApi } from '@/api';
import type { StartOptions } from '@/api/runtime';
import { queryKeys } from '@/lib/query-keys';
import { downloadBlob } from '@/lib/utils';
import { useWebSocket } from './use-websocket';
import type { RuntimeStats } from '@/types';

interface UseProjectRuntimeOptions {
  projectId: string;
  projectSlug?: string;
  enabled?: boolean;
  /** @deprecated Use WebSocket instead. Only used as fallback. */
  refetchLogsInterval?: number | false;
  /** @deprecated Use WebSocket instead. Only used as fallback. */
  refetchStatusInterval?: number | false;
}

/**
 * Hook for managing project runtime operations (start, stop, restart, logs, status).
 * Uses WebSocket for real-time updates instead of polling.
 */
export function useProjectRuntime({
  projectId,
  projectSlug,
  enabled = true,
  refetchLogsInterval = false,
  refetchStatusInterval = false,
}: UseProjectRuntimeOptions) {
  const queryClient = useQueryClient();
  const lastLogFetchRef = useRef<number>(0);

  // Handle WebSocket events
  const handleMonitorUpdate = useCallback((data: RuntimeStats) => {
    queryClient.setQueryData(['monitor', projectId], data);
  }, [queryClient, projectId]);

  const handleLogUpdate = useCallback(() => {
    // Throttle log fetches - only fetch if last fetch was > 3 seconds ago
    const now = Date.now();
    if (now - lastLogFetchRef.current > 3000) {
      lastLogFetchRef.current = now;
      queryClient.invalidateQueries({ queryKey: queryKeys.logs.project(projectId) });
    }
  }, [queryClient, projectId]);

  const handleRunningChange = useCallback((running: boolean) => {
    // Update status in cache
    queryClient.setQueryData(queryKeys.projects.status(projectId), (old: unknown) => ({
      ...(old as object),
      status: running ? 'running' : 'stopped',
    }));
    // Also refresh full data
    queryClient.invalidateQueries({ queryKey: queryKeys.projects.status(projectId) });
    queryClient.invalidateQueries({ queryKey: ['monitor', projectId] });
  }, [queryClient, projectId]);

  // Subscribe to WebSocket events
  useWebSocket({
    projectId,
    enabled,
    onMonitor: handleMonitorUpdate,
    onLog: handleLogUpdate,
    onRunning: handleRunningChange,
  });

  // Queries - fetch initial data, then rely on WebSocket for updates
  const statusQuery = useQuery({
    queryKey: queryKeys.projects.status(projectId),
    queryFn: () => runtimeApi.status(projectId),
    enabled,
    // Fallback polling only if WebSocket is not available
    refetchInterval: refetchStatusInterval,
    staleTime: 30000, // Consider data fresh for 30 seconds (WS will update it)
  });

  const monitorQuery = useQuery({
    queryKey: ['monitor', projectId],
    queryFn: () => runtimeApi.monitor(projectId),
    enabled,
    // Fallback polling only if WebSocket is not available
    refetchInterval: refetchStatusInterval,
    staleTime: 30000,
  });

  const logsQuery = useQuery({
    queryKey: queryKeys.logs.project(projectId),
    queryFn: () => runtimeApi.logs(projectId),
    enabled,
    // Fallback polling only if WebSocket is not available
    refetchInterval: refetchLogsInterval,
  });

  // Mutations
  const startMutation = useMutation({
    mutationFn: (options?: StartOptions) => runtimeApi.start(projectId, options),
    onSuccess: () => {
      invalidateAll();
      toast.success('Service started');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to start service');
    },
  });

  const stopMutation = useMutation({
    mutationFn: () => runtimeApi.stop(projectId),
    onSuccess: () => {
      invalidateAll();
      toast.success('Service stopped');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to stop service');
    },
  });

  const restartMutation = useMutation({
    mutationFn: () => runtimeApi.restart(projectId),
    onSuccess: () => {
      invalidateAll();
      toast.success('Service restarted');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to restart service');
    },
  });

  // Actions
  const invalidateAll = () => {
    queryClient.invalidateQueries({ queryKey: queryKeys.projects.all });
    queryClient.invalidateQueries({ queryKey: queryKeys.projects.detail(projectId) });
    queryClient.invalidateQueries({ queryKey: queryKeys.projects.status(projectId) });
    queryClient.invalidateQueries({ queryKey: queryKeys.logs.project(projectId) });
    queryClient.invalidateQueries({ queryKey: ['monitor', projectId] });
  };

  const downloadLogs = async () => {
    try {
      const blob = await runtimeApi.downloadLogs(projectId);
      const filename = `${projectSlug || projectId}-logs.txt`;
      await downloadBlob(blob, filename);
    } catch (err) {
      console.error('Failed to download logs:', err);
      toast.error('Failed to download logs');
    }
  };

  return {
    // Queries
    status: statusQuery.data,
    statusLoading: statusQuery.isLoading,
    monitor: monitorQuery.data,
    monitorLoading: monitorQuery.isLoading,
    logs: logsQuery.data,
    logsLoading: logsQuery.isLoading,
    refetchLogs: logsQuery.refetch,

    // Mutations
    start: startMutation.mutate,
    stop: stopMutation.mutate,
    restart: restartMutation.mutate,
    isStarting: startMutation.isPending,
    isStopping: stopMutation.isPending,
    isRestarting: restartMutation.isPending,
    isPending: startMutation.isPending || stopMutation.isPending || restartMutation.isPending,

    // Actions
    downloadLogs,
    invalidateAll,
  };
}
