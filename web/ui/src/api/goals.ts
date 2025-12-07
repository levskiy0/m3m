import { api } from './client';
import type { Goal, CreateGoalRequest, UpdateGoalRequest, GoalStats } from '@/types';

interface GetStatsParams {
  goalIds?: string[];
  startDate?: string;
  endDate?: string;
}

export const goalsApi = {
  // Global goals
  listGlobal: async (): Promise<Goal[]> => {
    return api.get<Goal[]>('/api/goals');
  },

  get: async (id: string): Promise<Goal> => {
    return api.get<Goal>(`/api/goals/${id}`);
  },

  createGlobal: async (data: CreateGoalRequest): Promise<Goal> => {
    return api.post<Goal>('/api/goals', data);
  },

  update: async (id: string, data: UpdateGoalRequest): Promise<Goal> => {
    return api.put<Goal>(`/api/goals/${id}`, data);
  },

  delete: async (id: string): Promise<void> => {
    return api.delete(`/api/goals/${id}`);
  },

  getStats: async (params?: GetStatsParams): Promise<GoalStats[]> => {
    const searchParams = new URLSearchParams();
    if (params?.goalIds?.length) {
      searchParams.set('goalIds', params.goalIds.join(','));
    }
    if (params?.startDate) {
      searchParams.set('startDate', params.startDate);
    }
    if (params?.endDate) {
      searchParams.set('endDate', params.endDate);
    }
    const query = searchParams.toString();
    return api.get<GoalStats[]>(`/api/goals/stats${query ? `?${query}` : ''}`);
  },

  // Project goals
  listProject: async (projectId: string): Promise<Goal[]> => {
    return api.get<Goal[]>(`/api/projects/${projectId}/goals`);
  },

  createProject: async (projectId: string, data: CreateGoalRequest): Promise<Goal> => {
    return api.post<Goal>(`/api/projects/${projectId}/goals`, data);
  },

  updateProject: async (
    projectId: string,
    goalId: string,
    data: UpdateGoalRequest
  ): Promise<Goal> => {
    return api.put<Goal>(`/api/projects/${projectId}/goals/${goalId}`, data);
  },

  deleteProject: async (projectId: string, goalId: string): Promise<void> => {
    return api.delete(`/api/projects/${projectId}/goals/${goalId}`);
  },

  resetProject: async (projectId: string, goalId: string): Promise<void> => {
    return api.post(`/api/projects/${projectId}/goals/${goalId}/reset`, {});
  },
};
