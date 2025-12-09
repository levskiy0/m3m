import { api } from './client';
import type { Project, CreateProjectRequest, UpdateProjectRequest } from '@/types';

export const projectsApi = {
  list: async (): Promise<Project[]> => {
    return api.get<Project[]>('/api/projects');
  },

  get: async (id: string): Promise<Project> => {
    return api.get<Project>(`/api/projects/${id}`);
  },

  create: async (data: CreateProjectRequest): Promise<Project> => {
    return api.post<Project>('/api/projects', data);
  },

  update: async (id: string, data: UpdateProjectRequest): Promise<Project> => {
    return api.put<Project>(`/api/projects/${id}`, data);
  },

  delete: async (id: string): Promise<void> => {
    return api.delete(`/api/projects/${id}`);
  },

  regenerateKey: async (id: string): Promise<Project> => {
    return api.post<Project>(`/api/projects/${id}/regenerate-key`);
  },

  addMember: async (id: string, userId: string): Promise<Project> => {
    return api.post<Project>(`/api/projects/${id}/members`, { userId });
  },

  removeMember: async (id: string, userId: string): Promise<Project> => {
    return api.delete<Project>(`/api/projects/${id}/members/${userId}`);
  },
};
