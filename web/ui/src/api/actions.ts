import { api } from './client';
import type {
  Action,
  ActionRuntimeState,
  CreateActionRequest,
  UpdateActionRequest,
} from '@/types';

export const actionsApi = {
  list: async (projectId: string): Promise<Action[]> => {
    return api.get<Action[]>(`/api/projects/${projectId}/actions`);
  },

  create: async (
    projectId: string,
    data: CreateActionRequest
  ): Promise<Action> => {
    return api.post<Action>(`/api/projects/${projectId}/actions`, data);
  },

  update: async (
    projectId: string,
    actionId: string,
    data: UpdateActionRequest
  ): Promise<Action> => {
    return api.put<Action>(
      `/api/projects/${projectId}/actions/${actionId}`,
      data
    );
  },

  delete: async (projectId: string, actionId: string): Promise<void> => {
    return api.delete(`/api/projects/${projectId}/actions/${actionId}`);
  },

  getStates: async (projectId: string): Promise<ActionRuntimeState[]> => {
    return api.get<ActionRuntimeState[]>(
      `/api/projects/${projectId}/actions/states`
    );
  },

  trigger: async (projectSlug: string, actionSlug: string): Promise<void> => {
    return api.post(`/r/${projectSlug}/actions/${actionSlug}`);
  },
};
