import { api } from './client';
import type { Environment, CreateEnvRequest, UpdateEnvRequest } from '@/types';

export const environmentApi = {
  list: async (projectId: string): Promise<Environment[]> => {
    return api.get<Environment[]>(`/api/projects/${projectId}/env`);
  },

  create: async (projectId: string, data: CreateEnvRequest): Promise<Environment> => {
    return api.post<Environment>(`/api/projects/${projectId}/env`, data);
  },

  update: async (
    projectId: string,
    key: string,
    data: UpdateEnvRequest
  ): Promise<Environment> => {
    return api.put<Environment>(`/api/projects/${projectId}/env/${key}`, data);
  },

  delete: async (projectId: string, key: string): Promise<void> => {
    return api.delete(`/api/projects/${projectId}/env/${key}`);
  },
};
