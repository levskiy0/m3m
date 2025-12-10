import { api } from './client';

export const templatesApi = {
  getServiceTemplate: async (): Promise<string> => {
    const response = await api.get<{ code: string }>('/templates/service');
    return response.code;
  },
};
