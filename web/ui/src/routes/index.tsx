import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom';
import { useAuth } from '@/providers/auth-provider';
import { AppLayout } from '@/components/layout/app-layout';
import { LoginPage } from '@/features/auth/login-page';
import { ProjectsPage } from '@/features/projects/projects-page';
import { ProjectDashboard } from '@/features/projects/project-dashboard';
import { ProjectSettings } from '@/features/projects/project-settings';
import { PipelinePage } from '@/features/pipeline/pipeline-page';
import { StoragePage } from '@/features/storage/storage-page';
import { ModelsPage } from '@/features/models/models-page';
import { ModelSchemaPage } from '@/features/models/model-schema-page';
import { ModelDataPage } from '@/features/models/model-data-page';
import { GoalsPage } from '@/features/goals/goals-page';
import { ModulesPage } from '@/features/modules/modules-page';
import { DocsLayout } from '@/features/docs/docs-layout';
import { GettingStartedPage } from '@/features/docs/getting-started';
import { LifecyclePage } from '@/features/docs/lifecycle';
import { ModulePage } from '@/features/docs/module-page';
import { DatabaseGuidePage } from '@/features/docs/database-guide';
import { EnvironmentPage } from '@/features/environment/environment-page';
import { UsersPage } from '@/features/users/users-page';
import { ProfilePage } from '@/features/users/profile-page';
import { LogsPage } from '@/features/projects/logs-page';
import { Skeleton } from '@/components/ui/skeleton';

function ProtectedRoute() {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="space-y-4 w-64">
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-3/4" />
          <Skeleton className="h-4 w-1/2" />
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <Outlet />;
}

function PublicRoute() {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="space-y-4 w-64">
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-3/4" />
        </div>
      </div>
    );
  }

  if (isAuthenticated) {
    return <Navigate to="/projects" replace />;
  }

  return <Outlet />;
}

export function AppRouter() {
  return (
    <BrowserRouter>
      <Routes>
        {/* Public routes */}
        <Route element={<PublicRoute />}>
          <Route path="/login" element={<LoginPage />} />
        </Route>

        {/* Protected routes */}
        <Route element={<ProtectedRoute />}>
          <Route element={<AppLayout />}>
            {/* Dashboard redirect */}
            <Route path="/" element={<Navigate to="/projects" replace />} />

            {/* Projects */}
            <Route path="/projects" element={<ProjectsPage />} />
            <Route path="/projects/:projectId" element={<ProjectDashboard />} />
            <Route path="/projects/:projectId/settings" element={<ProjectSettings />} />
            <Route path="/projects/:projectId/pipeline" element={<PipelinePage />} />
            <Route path="/projects/:projectId/storage" element={<StoragePage />} />
            <Route path="/projects/:projectId/models" element={<ModelsPage />} />
            <Route
              path="/projects/:projectId/models/:modelId/schema"
              element={<ModelSchemaPage />}
            />
            <Route
              path="/projects/:projectId/models/:modelId/data"
              element={<ModelDataPage />}
            />
            <Route path="/projects/:projectId/goals" element={<GoalsPage />} />
            <Route path="/projects/:projectId/environment" element={<EnvironmentPage />} />
            <Route path="/projects/:projectId/logs" element={<LogsPage />} />

            {/* Global */}
            <Route path="/modules" element={<ModulesPage />} />

            {/* Documentation */}
            <Route path="/docs" element={<DocsLayout />}>
              <Route index element={<Navigate to="/docs/getting-started" replace />} />
              <Route path="getting-started" element={<GettingStartedPage />} />
              <Route path="lifecycle" element={<LifecyclePage />} />
              <Route path="database" element={<DatabaseGuidePage />} />
              <Route path="api/:moduleId" element={<ModulePage />} />
            </Route>

            {/* Settings */}
            <Route path="/settings/users" element={<UsersPage />} />
            <Route path="/settings/profile" element={<ProfilePage />} />
          </Route>
        </Route>

        {/* Catch all */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
