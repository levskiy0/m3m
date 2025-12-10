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

  upload: async (
    projectId: string,
    path: string,
    file: File,
    onProgress?: (progress: number) => void
  ): Promise<StorageItem> => {
    return new Promise((resolve, reject) => {
      const formData = new FormData();
      formData.append('file', file);
      formData.append('path', path);

      const xhr = new XMLHttpRequest();

      xhr.upload.addEventListener('progress', (event) => {
        if (event.lengthComputable && onProgress) {
          const progress = Math.round((event.loaded / event.total) * 100);
          onProgress(progress);
        }
      });

      xhr.addEventListener('load', () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            const result = JSON.parse(xhr.responseText);
            resolve(result);
          } catch {
            reject(new Error('Invalid response'));
          }
        } else {
          reject(new Error('Upload failed'));
        }
      });

      xhr.addEventListener('error', () => {
        reject(new Error('Upload failed'));
      });

      xhr.open('POST', `${api['baseURL']}/api/projects/${projectId}/storage/upload`);

      const token = localStorage.getItem('m3m_token');
      if (token) {
        xhr.setRequestHeader('Authorization', `Bearer ${token}`);
      }

      xhr.send(formData);
    });
  },

  download: async (projectId: string, path: string): Promise<Blob> => {
    return api.download(`/api/projects/${projectId}/storage/download/${path}`);
  },

  rename: async (projectId: string, data: RenameRequest): Promise<void> => {
    // Backend expects {old_path, new_path}
    const dir = data.path.substring(0, data.path.lastIndexOf('/')) || '';
    const newPath = dir ? `${dir}/${data.newName}` : data.newName;
    return api.put(`/api/projects/${projectId}/storage/rename`, {
      old_path: data.path,
      new_path: newPath,
    });
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

  move: async (projectId: string, sourcePath: string, targetDir: string): Promise<void> => {
    // Move is rename with different directory
    const fileName = sourcePath.split('/').pop() || '';
    const newPath = targetDir ? `${targetDir}/${fileName}` : fileName;
    return api.put(`/api/projects/${projectId}/storage/rename`, {
      old_path: sourcePath,
      new_path: newPath,
    });
  },
};
