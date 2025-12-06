import { api } from './client';
import type {
  Model,
  ModelData,
  CreateModelRequest,
  UpdateModelRequest,
  QueryDataRequest,
  PaginatedResponse,
} from '@/types';

export const modelsApi = {
  // Models
  list: async (projectId: string): Promise<Model[]> => {
    return api.get<Model[]>(`/api/projects/${projectId}/models`);
  },

  get: async (projectId: string, modelId: string): Promise<Model> => {
    return api.get<Model>(`/api/projects/${projectId}/models/${modelId}`);
  },

  create: async (projectId: string, data: CreateModelRequest): Promise<Model> => {
    return api.post<Model>(`/api/projects/${projectId}/models`, data);
  },

  update: async (
    projectId: string,
    modelId: string,
    data: UpdateModelRequest
  ): Promise<Model> => {
    return api.put<Model>(`/api/projects/${projectId}/models/${modelId}`, data);
  },

  delete: async (projectId: string, modelId: string): Promise<void> => {
    return api.delete(`/api/projects/${projectId}/models/${modelId}`);
  },

  // Data
  listData: async (
    projectId: string,
    modelId: string,
    params?: { page?: number; limit?: number; sort?: string; order?: string }
  ): Promise<PaginatedResponse<ModelData>> => {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.set('page', params.page.toString());
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.sort) searchParams.set('sort', params.sort);
    if (params?.order) searchParams.set('order', params.order);
    const query = searchParams.toString();
    return api.get<PaginatedResponse<ModelData>>(
      `/api/projects/${projectId}/models/${modelId}/data${query ? `?${query}` : ''}`
    );
  },

  queryData: async (
    projectId: string,
    modelId: string,
    query: QueryDataRequest
  ): Promise<PaginatedResponse<ModelData>> => {
    return api.post<PaginatedResponse<ModelData>>(
      `/api/projects/${projectId}/models/${modelId}/data/query`,
      query
    );
  },

  getData: async (
    projectId: string,
    modelId: string,
    dataId: string
  ): Promise<ModelData> => {
    return api.get<ModelData>(
      `/api/projects/${projectId}/models/${modelId}/data/${dataId}`
    );
  },

  createData: async (
    projectId: string,
    modelId: string,
    data: Record<string, unknown>
  ): Promise<ModelData> => {
    return api.post<ModelData>(
      `/api/projects/${projectId}/models/${modelId}/data`,
      data
    );
  },

  updateData: async (
    projectId: string,
    modelId: string,
    dataId: string,
    data: Record<string, unknown>
  ): Promise<ModelData> => {
    return api.put<ModelData>(
      `/api/projects/${projectId}/models/${modelId}/data/${dataId}`,
      data
    );
  },

  deleteData: async (
    projectId: string,
    modelId: string,
    dataId: string
  ): Promise<void> => {
    return api.delete(`/api/projects/${projectId}/models/${modelId}/data/${dataId}`);
  },
};
