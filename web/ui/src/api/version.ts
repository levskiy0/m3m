import { api } from './client';

export interface VersionInfo {
  version: string;
  name: string;
}

export const versionApi = {
  get: async (): Promise<VersionInfo> => {
    return api.get<VersionInfo>('/api/version');
  },
};
