import { useState, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Plus,
  FolderOpen,
  Play,
  Square,
  MoreHorizontal,
  Settings,
  Activity,
  Clock,
  ExternalLink,
  RotateCcw,
} from 'lucide-react';
import { toast } from 'sonner';

import { projectsApi, runtimeApi, pipelineApi } from '@/api';
import { config } from '@/lib/config';
import { queryKeys } from '@/lib/query-keys';
import { formatUptime } from '@/lib/format';
import { useAuth } from '@/providers/auth-provider';
import { useAutoSlug, useTitle } from '@/hooks';
import type { Project, CreateProjectRequest } from '@/types';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { LoadingButton } from '@/components/ui/loading-button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Field, FieldGroup, FieldLabel, FieldError } from '@/components/ui/field';
import { ColorPicker } from '@/components/shared/color-picker';
import { StatusBadge } from '@/components/shared/status-badge';
import { EmptyState } from '@/components/shared/empty-state';
import { PageHeader } from '@/components/shared/page-header';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';

export function ProjectsPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { user } = useAuth();

  const [createOpen, setCreateOpen] = useState(false);
  const [color, setColor] = useState<string | undefined>();
  const [error, setError] = useState('');
  const { name, slug, setName, setSlug, reset: resetSlug } = useAutoSlug();

  useEffect(() => {
    if (location.state?.openCreate) {
      setCreateOpen(true);
      navigate(location.pathname, { replace: true, state: {} });
    }
  }, [location.state, location.pathname, navigate]);

  const { data: projects = [], isLoading } = useQuery({
    queryKey: queryKeys.projects.all,
    queryFn: projectsApi.list,
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateProjectRequest) => projectsApi.create(data),
    onSuccess: (project) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.projects.all });
      setCreateOpen(false);
      resetForm();
      toast.success('Project created successfully');
      navigate(`/projects/${project.id}`);
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Failed to create project');
    },
  });

  const startMutation = useMutation({
    mutationFn: (projectId: string) => runtimeApi.start(projectId),
    onSuccess: (_, projectId) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.projects.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.projects.status(projectId) });
      toast.success('Service started');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to start service');
    },
  });

  const stopMutation = useMutation({
    mutationFn: (projectId: string) => runtimeApi.stop(projectId),
    onSuccess: (_, projectId) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.projects.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.projects.status(projectId) });
      toast.success('Service stopped');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to stop service');
    },
  });

  const restartMutation = useMutation({
    mutationFn: (projectId: string) => runtimeApi.restart(projectId),
    onSuccess: (_, projectId) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.projects.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.projects.status(projectId) });
      toast.success('Service restarted');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to restart service');
    },
  });

  const resetForm = () => {
    resetSlug();
    setColor(undefined);
    setError('');
  };

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    createMutation.mutate({ name, slug, color });
  };

  const canCreateProjects = user?.permissions?.createProjects || user?.isRoot;
  const runningCount = projects.filter((p) => p.status === 'running').length;
  const stoppedCount = projects.filter((p) => p.status === 'stopped').length;

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="h-8 w-48 bg-muted rounded animate-pulse" />
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <Card key={i} className="animate-pulse">
              <CardHeader>
                <div className="h-5 w-24 bg-muted rounded" />
                <div className="h-4 w-32 bg-muted rounded mt-2" />
              </CardHeader>
              <CardContent>
                <div className="h-8 w-20 bg-muted rounded" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Projects"
        description="Manage your mini-services and workers"
        action={
          canCreateProjects && (
            <Button onClick={() => setCreateOpen(true)}>
              <Plus className="mr-2 size-4" />
              New Project
            </Button>
          )
        }
      >
        {projects.length > 0 && (
          <div className="flex items-center gap-3 text-sm text-muted-foreground">
            <span className="flex items-center gap-1.5">
              <span className="size-2 rounded-full bg-green-500 animate-pulse" />
              {runningCount} running
            </span>
            <span className="flex items-center gap-1.5">
              <span className="size-2 rounded-full bg-gray-400" />
              {stoppedCount} stopped
            </span>
          </div>
        )}
      </PageHeader>

      {projects.length === 0 ? (
        <EmptyState
          icon={<FolderOpen className="size-12" />}
          title="No projects yet"
          description="Create your first project to get started with mini-services"
          action={
            canCreateProjects && (
              <Button onClick={() => setCreateOpen(true)}>
                <Plus className="mr-2 size-4" />
                Create Project
              </Button>
            )
          }
        />
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {projects.map((project) => (
            <ProjectCard
              key={project.id}
              project={project}
              onStart={() => startMutation.mutate(project.id)}
              onStop={() => stopMutation.mutate(project.id)}
              onRestart={() => restartMutation.mutate(project.id)}
              isPending={startMutation.isPending || stopMutation.isPending || restartMutation.isPending}
            />
          ))}
        </div>
      )}

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <form onSubmit={handleCreate}>
            <DialogHeader>
              <DialogTitle>Create Project</DialogTitle>
              <DialogDescription>
                Create a new mini-service project
              </DialogDescription>
            </DialogHeader>
            <div className="py-4">
              <FieldGroup>
                <Field>
                  <FieldLabel htmlFor="name">Name</FieldLabel>
                  <Input
                    id="name"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="My Project"
                    required
                  />
                </Field>
                <Field>
                  <FieldLabel htmlFor="slug">Slug</FieldLabel>
                  <Input
                    id="slug"
                    value={slug}
                    onChange={(e) => setSlug(e.target.value)}
                    placeholder="my-project"
                    required
                  />
                </Field>
                <Field>
                  <FieldLabel>Color</FieldLabel>
                  <ColorPicker value={color} onChange={setColor} />
                </Field>
                {error && <FieldError>{error}</FieldError>}
              </FieldGroup>
            </div>
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setCreateOpen(false);
                  resetForm();
                }}
              >
                Cancel
              </Button>
              <LoadingButton type="submit" loading={createMutation.isPending}>
                Create
              </LoadingButton>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}

interface ProjectCardProps {
  project: Project;
  onStart: () => void;
  onStop: () => void;
  onRestart: () => void;
  isPending: boolean;
}

function ProjectCard({
  project,
  onStart,
  onStop,
  onRestart,
  isPending,
}: ProjectCardProps) {
  const navigate = useNavigate();
  const isRunning = project.status === 'running';

  const { data: status } = useQuery({
    queryKey: queryKeys.projects.status(project.id),
    queryFn: () => runtimeApi.status(project.id),
    enabled: isRunning,
    refetchInterval: isRunning ? 10000 : false,
  });

  const { data: releases = [] } = useQuery({
    queryKey: queryKeys.projects.releases(project.id),
    queryFn: () => pipelineApi.listReleases(project.id),
    staleTime: 60000,
  });

  const activeRelease = releases.find((r) => r.isActive);
  const publicUrl = `${config.apiURL}/r/${project.slug}`;

  return (
    <Card
      className={cn(
        "cursor-pointer transition-all hover:shadow-md",
        isRunning ? "border-green-500/30 hover:border-green-500/50" : "hover:bg-muted/50"
      )}
      onClick={() => navigate(`/projects/${project.id}`)}
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div
              className="size-10 rounded-lg flex items-center justify-center shrink-0"
              style={{ backgroundColor: project.color || '#6b7280' }}
            >
              <Activity className="size-5 text-white" />
            </div>
            <div className="min-w-0">
              <CardTitle className="text-lg truncate">{project.name}</CardTitle>
              <CardDescription className="font-mono text-xs truncate">
                {project.slug}
              </CardDescription>
            </div>
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
              <Button variant="ghost" size="icon" className="size-8 shrink-0">
                <MoreHorizontal className="size-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {isRunning ? (
                <>
                  <DropdownMenuItem
                    onClick={(e) => {
                      e.stopPropagation();
                      onRestart();
                    }}
                    disabled={isPending}
                  >
                    <RotateCcw className="mr-2 size-4" />
                    Restart
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={(e) => {
                      e.stopPropagation();
                      onStop();
                    }}
                    disabled={isPending}
                    className="text-destructive"
                  >
                    <Square className="mr-2 size-4" />
                    Stop
                  </DropdownMenuItem>
                </>
              ) : (
                <DropdownMenuItem
                  onClick={(e) => {
                    e.stopPropagation();
                    onStart();
                  }}
                  disabled={isPending || releases.length === 0}
                >
                  <Play className="mr-2 size-4" />
                  Start
                </DropdownMenuItem>
              )}
              <DropdownMenuSeparator />
              <DropdownMenuItem
                onClick={(e) => {
                  e.stopPropagation();
                  window.open(publicUrl, '_blank');
                }}
              >
                <ExternalLink className="mr-2 size-4" />
                Open URL
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={(e) => {
                  e.stopPropagation();
                  navigate(`/projects/${project.id}/settings`);
                }}
              >
                <Settings className="mr-2 size-4" />
                Settings
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="flex items-center justify-between">
          <StatusBadge status={project.status} />
          {activeRelease && (
            <Badge variant="outline" className="font-mono text-xs">
              {activeRelease.version}
            </Badge>
          )}
        </div>

        {isRunning && status?.uptime && (
          <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
            <Clock className="size-3" />
            <span>Uptime: {formatUptime(status.uptime)}</span>
          </div>
        )}

        {releases.length === 0 && (
          <p className="text-xs text-amber-500">
            No releases - create one to start
          </p>
        )}
      </CardContent>
    </Card>
  );
}
