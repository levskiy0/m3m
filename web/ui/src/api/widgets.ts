import { api } from './client';
import type { Widget, CreateWidgetRequest, UpdateWidgetRequest, ReorderWidgetsRequest } from '@/types';

export const widgetsApi = {
  list: async (projectId: string): Promise<Widget[]> => {
    return api.get<Widget[]>(`/api/projects/${projectId}/widgets`);
  },

  create: async (projectId: string, data: CreateWidgetRequest): Promise<Widget> => {
    return api.post<Widget>(`/api/projects/${projectId}/widgets`, data);
  },

  update: async (
    projectId: string,
    widgetId: string,
    data: UpdateWidgetRequest
  ): Promise<Widget> => {
    return api.put<Widget>(`/api/projects/${projectId}/widgets/${widgetId}`, data);
  },

  delete: async (projectId: string, widgetId: string): Promise<void> => {
    return api.delete(`/api/projects/${projectId}/widgets/${widgetId}`);
  },

  reorder: async (projectId: string, data: ReorderWidgetsRequest): Promise<void> => {
    return api.post(`/api/projects/${projectId}/widgets/reorder`, data);
  },
};
