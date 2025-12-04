import { api } from './client';
import type {
  User,
  CreateUserRequest,
  UpdateUserRequest,
  UpdateMeRequest,
  ChangePasswordRequest,
} from '@/types';

export const usersApi = {
  list: async (): Promise<User[]> => {
    return api.get<User[]>('/api/users');
  },

  get: async (id: string): Promise<User> => {
    return api.get<User>(`/api/users/${id}`);
  },

  create: async (data: CreateUserRequest): Promise<User> => {
    return api.post<User>('/api/users', data);
  },

  update: async (id: string, data: UpdateUserRequest): Promise<User> => {
    return api.put<User>(`/api/users/${id}`, data);
  },

  delete: async (id: string): Promise<void> => {
    return api.delete(`/api/users/${id}`);
  },

  block: async (id: string): Promise<User> => {
    return api.post<User>(`/api/users/${id}/block`);
  },

  unblock: async (id: string): Promise<User> => {
    return api.post<User>(`/api/users/${id}/unblock`);
  },

  updateMe: async (data: UpdateMeRequest): Promise<User> => {
    return api.put<User>('/api/users/me', data);
  },

  changePassword: async (data: ChangePasswordRequest): Promise<void> => {
    return api.put('/api/users/me/password', data);
  },

  updateAvatar: async (file: File): Promise<User> => {
    return api.upload<User>('/api/users/me/avatar', file, 'avatar');
  },
};
