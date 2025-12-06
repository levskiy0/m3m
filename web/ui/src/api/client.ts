import { config } from '@/lib/config';
import type { ApiError } from '@/types';

const TOKEN_KEY = 'm3m_token';

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function removeToken(): void {
  localStorage.removeItem(TOKEN_KEY);
}

export class ApiClient {
  private baseURL: string;

  constructor(baseURL: string = config.apiURL) {
    this.baseURL = baseURL;
  }

  private getAuthHeaders(): HeadersInit {
    const token = getToken();
    const headers: Record<string, string> = {};
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    return headers;
  }

  private async handleErrorResponse(response: Response): Promise<never> {
    if (response.status === 401) {
      removeToken();
      window.location.href = '/login';
      throw new Error('Unauthorized');
    }

    const error: ApiError = await response.json().catch(() => ({
      error: 'Unknown error',
    }));

    throw new Error(error.message || error.error || 'Request failed');
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...this.getAuthHeaders(),
      ...options.headers,
    };

    const response = await fetch(`${this.baseURL}${endpoint}`, {
      ...options,
      headers,
    });

    if (!response.ok) {
      return this.handleErrorResponse(response);
    }

    if (response.status === 204) {
      return undefined as T;
    }

    return response.json();
  }

  async get<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'GET' });
  }

  async getText(endpoint: string): Promise<string> {
    const response = await fetch(`${this.baseURL}${endpoint}`, {
      method: 'GET',
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      return this.handleErrorResponse(response);
    }

    return response.text();
  }

  async post<T>(endpoint: string, data?: unknown): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async put<T>(endpoint: string, data?: unknown): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async putText(endpoint: string, text: string): Promise<void> {
    const response = await fetch(`${this.baseURL}${endpoint}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'text/plain',
        ...this.getAuthHeaders(),
      },
      body: text,
    });

    if (!response.ok) {
      return this.handleErrorResponse(response);
    }
  }

  async delete<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'DELETE' });
  }

  async upload<T>(endpoint: string, file: File, fieldName: string = 'file'): Promise<T> {
    const formData = new FormData();
    formData.append(fieldName, file);

    const response = await fetch(`${this.baseURL}${endpoint}`, {
      method: 'POST',
      headers: this.getAuthHeaders(),
      body: formData,
    });

    if (!response.ok) {
      return this.handleErrorResponse(response);
    }

    return response.json();
  }

  async download(endpoint: string): Promise<Blob> {
    const response = await fetch(`${this.baseURL}${endpoint}`, {
      method: 'GET',
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Download failed');
    }

    return response.blob();
  }

  getDownloadURL(endpoint: string): string {
    const token = getToken();
    return `${this.baseURL}${endpoint}${token ? `?token=${token}` : ''}`;
  }
}

export const api = new ApiClient();
