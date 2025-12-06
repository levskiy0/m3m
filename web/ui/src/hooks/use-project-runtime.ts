import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

import { runtimeApi } from '@/api';
import type { StartOptions } from '@/api/runtime';
import { queryKeys } from '@/lib/query-keys';
import { downloadBlob } from '@/lib/utils';

interface UseProjectRuntimeOptions {
  projectId: string;
  projectSlug?: string;
  enabled?: boolean;
  refetchLogsInterval?: number | false;
  refetchStatusInterval?: number | false;
}

/**
 * Hook for managing project runtime operations (start, stop, restart, logs, status).
 * Consolidates all runtime-related logic in one place.
 */
export function useProjectRuntime({
  projectId,
  projectSlug,
  enabled = true,
  refetchLogsInterval = false,
  refetchStatusInterval = false,
}: UseProjectRuntimeOptions) {
  const queryClient = useQueryClient();

  // Queries
  const statusQuery = useQuery({
    queryKey: queryKeys.projects.status(projectId),
    queryFn: () => runtimeApi.status(projectId),
    enabled,
    refetchInterval: refetchStatusInterval,
  });

  const monitorQuery = useQuery({
    queryKey: ['monitor', projectId],
    queryFn: () => runtimeApi.monitor(projectId),
    enabled,
    refetchInterval: refetchStatusInterval,
  });

  const logsQuery = useQuery({
    queryKey: queryKeys.logs.project(projectId),
    queryFn: () => runtimeApi.logs(projectId),
    enabled,
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
