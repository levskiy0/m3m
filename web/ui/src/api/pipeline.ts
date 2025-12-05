import { api } from './client';
import type {
  Branch,
  BranchSummary,
  Release,
  ReleaseSummary,
  CreateBranchRequest,
  UpdateBranchRequest,
  ResetBranchRequest,
  CreateReleaseRequest,
} from '@/types';

export const pipelineApi = {
  // Branches
  listBranches: async (projectId: string): Promise<BranchSummary[]> => {
    return api.get<BranchSummary[]>(`/api/projects/${projectId}/pipeline/branches`);
  },

  getBranch: async (projectId: string, name: string): Promise<Branch> => {
    return api.get<Branch>(`/api/projects/${projectId}/pipeline/branches/${name}`);
  },

  createBranch: async (projectId: string, data: CreateBranchRequest): Promise<Branch> => {
    return api.post<Branch>(`/api/projects/${projectId}/pipeline/branches`, data);
  },

  updateBranch: async (
    projectId: string,
    name: string,
    data: UpdateBranchRequest
  ): Promise<Branch> => {
    return api.put<Branch>(`/api/projects/${projectId}/pipeline/branches/${name}`, data);
  },

  resetBranch: async (
    projectId: string,
    name: string,
    data: ResetBranchRequest
  ): Promise<Branch> => {
    return api.post<Branch>(
      `/api/projects/${projectId}/pipeline/branches/${name}/reset`,
      data
    );
  },

  deleteBranch: async (projectId: string, name: string): Promise<void> => {
    return api.delete(`/api/projects/${projectId}/pipeline/branches/${name}`);
  },

  // Releases
  listReleases: async (projectId: string): Promise<ReleaseSummary[]> => {
    return api.get<ReleaseSummary[]>(`/api/projects/${projectId}/pipeline/releases`);
  },

  createRelease: async (projectId: string, data: CreateReleaseRequest): Promise<Release> => {
    return api.post<Release>(`/api/projects/${projectId}/pipeline/releases`, data);
  },

  deleteRelease: async (projectId: string, version: string): Promise<void> => {
    return api.delete(`/api/projects/${projectId}/pipeline/releases/${version}`);
  },

  activateRelease: async (projectId: string, version: string): Promise<Release> => {
    return api.post<Release>(
      `/api/projects/${projectId}/pipeline/releases/${version}/activate`
    );
  },
};
