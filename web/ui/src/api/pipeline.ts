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

  getBranch: async (projectId: string, branchId: string): Promise<Branch> => {
    return api.get<Branch>(`/api/projects/${projectId}/pipeline/branches/${branchId}`);
  },

  createBranch: async (projectId: string, data: CreateBranchRequest): Promise<Branch> => {
    return api.post<Branch>(`/api/projects/${projectId}/pipeline/branches`, data);
  },

  updateBranch: async (
    projectId: string,
    branchId: string,
    data: UpdateBranchRequest
  ): Promise<Branch> => {
    return api.put<Branch>(`/api/projects/${projectId}/pipeline/branches/${branchId}`, data);
  },

  resetBranch: async (
    projectId: string,
    branchId: string,
    data: ResetBranchRequest
  ): Promise<Branch> => {
    return api.post<Branch>(
      `/api/projects/${projectId}/pipeline/branches/${branchId}/reset`,
      data
    );
  },

  deleteBranch: async (projectId: string, branchId: string): Promise<void> => {
    return api.delete(`/api/projects/${projectId}/pipeline/branches/${branchId}`);
  },

  // Releases
  listReleases: async (projectId: string): Promise<ReleaseSummary[]> => {
    return api.get<ReleaseSummary[]>(`/api/projects/${projectId}/pipeline/releases`);
  },

  createRelease: async (projectId: string, data: CreateReleaseRequest): Promise<Release> => {
    return api.post<Release>(`/api/projects/${projectId}/pipeline/releases`, data);
  },

  deleteRelease: async (projectId: string, releaseId: string): Promise<void> => {
    return api.delete(`/api/projects/${projectId}/pipeline/releases/${releaseId}`);
  },

  activateRelease: async (projectId: string, releaseId: string): Promise<Release> => {
    return api.post<Release>(
      `/api/projects/${projectId}/pipeline/releases/${releaseId}/activate`
    );
  },
};
