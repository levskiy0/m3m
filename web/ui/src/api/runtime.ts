import { api } from './client';
import type { RuntimeStatus, RuntimeStats, LogEntry } from '@/types';

interface Plugin {
  name: string;
  description?: string;
  version?: string;
}

export interface StartOptions {
  version?: string; // Release version to run
  branch?: string;  // Branch name to run (debug mode)
}

export const runtimeApi = {
  start: async (projectId: string, options?: StartOptions): Promise<{ runningSource?: string }> => {
    return api.post(`/api/projects/${projectId}/start`, options);
  },

  stop: async (projectId: string): Promise<void> => {
    return api.post(`/api/projects/${projectId}/stop`);
  },

  restart: async (projectId: string): Promise<void> => {
    return api.post(`/api/projects/${projectId}/restart`);
  },

  status: async (projectId: string): Promise<RuntimeStatus> => {
    return api.get<RuntimeStatus>(`/api/projects/${projectId}/status`);
  },

  monitor: async (projectId: string): Promise<RuntimeStats> => {
    return api.get<RuntimeStats>(`/api/projects/${projectId}/monitor`);
  },

  logs: async (projectId: string): Promise<LogEntry[]> => {
    return api.get<LogEntry[]>(`/api/projects/${projectId}/logs`);
  },

  downloadLogs: async (projectId: string): Promise<Blob> => {
    return api.download(`/api/projects/${projectId}/logs/download`);
  },

  getTypes: async (): Promise<string> => {
    return api.getText('/api/runtime/types');
  },

  listPlugins: async (): Promise<Plugin[]> => {
    return api.get<Plugin[]>('/plugins');
  },
};
