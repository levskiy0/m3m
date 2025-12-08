import { api } from './client';

export interface VersionInfo {
  version: string;
  name: string;
}

export const versionApi = {
  get: async (): Promise<VersionInfo> => {
    const { data } = await api.get<VersionInfo>('/version');
    return data;
  },
};
