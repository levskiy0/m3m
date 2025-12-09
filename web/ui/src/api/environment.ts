import { api } from './client';
import type { Environment, BulkUpdateEnvRequest } from '@/types';

export const environmentApi = {
  list: async (projectId: string): Promise<Environment[]> => {
    return api.get<Environment[]>(`/api/projects/${projectId}/env`);
  },

  bulkUpdate: async (
    projectId: string,
    data: BulkUpdateEnvRequest
  ): Promise<Environment[]> => {
    return api.put<Environment[]>(`/api/projects/${projectId}/env`, data);
  },

  delete: async (projectId: string, key: string): Promise<void> => {
    return api.delete(`/api/projects/${projectId}/env/${key}`);
  },
};
