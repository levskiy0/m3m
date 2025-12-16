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
  CreatePipelineFileRequest,
  UpdatePipelineFileRequest,
  RenamePipelineFileRequest,
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

  // Files
  createFile: async (
    projectId: string,
    branchId: string,
    data: CreatePipelineFileRequest
  ): Promise<Branch> => {
    return api.post<Branch>(
      `/api/projects/${projectId}/pipeline/branches/${branchId}/files`,
      data
    );
  },

  updateFile: async (
    projectId: string,
    branchId: string,
    fileName: string,
    data: UpdatePipelineFileRequest
  ): Promise<void> => {
    return api.put(
      `/api/projects/${projectId}/pipeline/branches/${branchId}/files/${encodeURIComponent(fileName)}`,
      data
    );
  },

  deleteFile: async (
    projectId: string,
    branchId: string,
    fileName: string
  ): Promise<Branch> => {
    return api.delete<Branch>(
      `/api/projects/${projectId}/pipeline/branches/${branchId}/files/${encodeURIComponent(fileName)}`
    );
  },

  renameFile: async (
    projectId: string,
    branchId: string,
    fileName: string,
    data: RenamePipelineFileRequest
  ): Promise<Branch> => {
    return api.post<Branch>(
      `/api/projects/${projectId}/pipeline/branches/${branchId}/files/${encodeURIComponent(fileName)}/rename`,
      data
    );
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
