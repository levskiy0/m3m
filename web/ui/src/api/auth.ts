import { api, setToken, removeToken } from './client';
import type { LoginRequest, LoginResponse, User } from '@/types';

export const authApi = {
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    const response = await api.post<LoginResponse>('/api/auth/login', data);
    setToken(response.token);
    return response;
  },

  logout: async (): Promise<void> => {
    try {
      await api.post('/api/auth/logout');
    } finally {
      removeToken();
    }
  },

  me: async (): Promise<User> => {
    return api.get<User>('/api/users/me');
  },
};
