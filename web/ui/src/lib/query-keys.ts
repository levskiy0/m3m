/**
 * Centralized query key factory for React Query.
 * Ensures consistent query keys across the application.
 */
export const queryKeys = {
  // Projects
  projects: {
    all: ['projects'] as const,
    detail: (id: string) => ['project', id] as const,
    status: (id: string) => ['project-status', id] as const,
    releases: (id: string) => ['project-releases', id] as const,
  },

  // Pipeline
  pipeline: {
    branches: (projectId: string) => ['branches', projectId] as const,
    releases: (projectId: string) => ['releases', projectId] as const,
    branchCode: (projectId: string, branchId: string) =>
      ['branch-code', projectId, branchId] as const,
    releaseCode: (projectId: string, releaseId: string) =>
      ['release-code', projectId, releaseId] as const,
  },

  // Models
  models: {
    all: (projectId: string) => ['models', projectId] as const,
    detail: (projectId: string, modelId: string) =>
      ['model', projectId, modelId] as const,
    data: (projectId: string, modelId: string) =>
      ['model-data', projectId, modelId] as const,
  },

  // Goals
  goals: {
    project: (projectId: string) => ['project-goals', projectId] as const,
    global: ['global-goals'] as const,
    stats: (projectId: string, goalIds: string[], startDate?: string, endDate?: string) =>
      ['project-goal-stats', projectId, goalIds, startDate, endDate] as const,
  },

  // Environment
  environment: {
    all: (projectId: string) => ['environment', projectId] as const,
  },

  // Storage
  storage: {
    files: (projectId: string, path: string) =>
      ['storage', projectId, path] as const,
  },

  // Users
  users: {
    all: ['users'] as const,
    profile: ['profile'] as const,
  },

  // Logs
  logs: {
    project: (projectId: string) => ['logs', projectId] as const,
  },

  // Widgets
  widgets: {
    all: (projectId: string) => ['widgets', projectId] as const,
  },
} as const;
