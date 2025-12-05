import { api } from './client';
import type {
  StorageItem,
  CreateDirRequest,
  RenameRequest,
  CreateFileRequest,
} from '@/types';

export const storageApi = {
  list: async (projectId: string, path: string = ''): Promise<StorageItem[]> => {
    const searchParams = new URLSearchParams();
    if (path) {
      searchParams.set('path', path);
    }
    const query = searchParams.toString();
    return api.get<StorageItem[]>(
      `/api/projects/${projectId}/storage${query ? `?${query}` : ''}`
    );
  },

  mkdir: async (projectId: string, data: CreateDirRequest): Promise<void> => {
    // Backend expects {path: "full/path"}, so combine path and name
    const fullPath = data.path ? `${data.path}/${data.name}` : data.name;
    return api.post(`/api/projects/${projectId}/storage/mkdir`, { path: fullPath });
  },

  upload: async (projectId: string, path: string, file: File): Promise<StorageItem> => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('path', path);

    const token = localStorage.getItem('m3m_token');
    const headers: HeadersInit = {};
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(
      `${api['baseURL']}/api/projects/${projectId}/storage/upload`,
      {
        method: 'POST',
        headers,
        body: formData,
      }
    );

    if (!response.ok) {
      throw new Error('Upload failed');
    }

    return response.json();
  },

  download: async (projectId: string, path: string): Promise<Blob> => {
    return api.download(`/api/projects/${projectId}/storage/download/${path}`);
  },

  rename: async (projectId: string, data: RenameRequest): Promise<void> => {
    return api.put(`/api/projects/${projectId}/storage/rename`, data);
  },

  delete: async (projectId: string, path: string): Promise<void> => {
    return api.delete(`/api/projects/${projectId}/storage/${path}`);
  },

  createFile: async (projectId: string, data: CreateFileRequest): Promise<StorageItem> => {
    // Backend expects {path: "full/path", content: "..."}, so combine path and name
    const fullPath = data.path ? `${data.path}/${data.name}` : data.name;
    return api.post<StorageItem>(`/api/projects/${projectId}/storage/file`, {
      path: fullPath,
      content: data.content,
    });
  },

  updateFile: async (projectId: string, path: string, content: string): Promise<void> => {
    // Backend reads raw body, not JSON
    return api.putText(`/api/projects/${projectId}/storage/file/${path}`, content);
  },

  getThumbnail: async (projectId: string, path: string): Promise<Blob> => {
    return api.download(`/api/projects/${projectId}/storage/thumbnail/${path}`);
  },
};
