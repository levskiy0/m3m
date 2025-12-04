import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Play,
  Square,
  RotateCcw,
  Code,
  HardDrive,
  Database,
  Target,
  Variable,
  ScrollText,
  Settings,
  Key,
  ExternalLink,
} from 'lucide-react';
import { toast } from 'sonner';

import { projectsApi, runtimeApi, pipelineApi } from '@/api';
import { config } from '@/lib/config';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
} from '@/components/ui/card';
import { StatusBadge } from '@/components/shared/status-badge';
import { Skeleton } from '@/components/ui/skeleton';

export function ProjectDashboard() {
  const { projectId } = useParams<{ projectId: string }>();
  const queryClient = useQueryClient();

  const { data: project, isLoading: projectLoading } = useQuery({
    queryKey: ['project', projectId],
    queryFn: () => projectsApi.get(projectId!),
    enabled: !!projectId,
  });

  const { data: releases = [] } = useQuery({
    queryKey: ['releases', projectId],
    queryFn: () => pipelineApi.listReleases(projectId!),
    enabled: !!projectId,
  });

  const { data: status } = useQuery({
    queryKey: ['runtime-status', projectId],
    queryFn: () => runtimeApi.status(projectId!),
    enabled: !!projectId,
    refetchInterval: project?.status === 'running' ? 5000 : false,
  });

  const startMutation = useMutation({
    mutationFn: () => runtimeApi.start(projectId!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project', projectId] });
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      queryClient.invalidateQueries({ queryKey: ['runtime-status', projectId] });
      toast.success('Project started');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to start');
    },
  });

  const stopMutation = useMutation({
    mutationFn: () => runtimeApi.stop(projectId!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project', projectId] });
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      queryClient.invalidateQueries({ queryKey: ['runtime-status', projectId] });
      toast.success('Project stopped');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to stop');
    },
  });

  const restartMutation = useMutation({
    mutationFn: () => runtimeApi.restart(projectId!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project', projectId] });
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      queryClient.invalidateQueries({ queryKey: ['runtime-status', projectId] });
      toast.success('Project restarted');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to restart');
    },
  });

  if (projectLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-48" />
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {[1, 2, 3, 4].map((i) => (
            <Skeleton key={i} className="h-32" />
          ))}
        </div>
      </div>
    );
  }

  if (!project) {
    return <div>Project not found</div>;
  }

  const activeRelease = releases.find((r) => r.isActive);
  const publicUrl = `${config.apiURL}/r/${project.slug}`;

  const quickLinks = [
    { icon: Code, label: 'Pipeline', href: `/projects/${projectId}/pipeline` },
    { icon: HardDrive, label: 'Storage', href: `/projects/${projectId}/storage` },
    { icon: Database, label: 'Models', href: `/projects/${projectId}/models` },
    { icon: Target, label: 'Goals', href: `/projects/${projectId}/goals` },
    { icon: Variable, label: 'Environment', href: `/projects/${projectId}/environment` },
    { icon: ScrollText, label: 'Logs', href: `/projects/${projectId}/logs` },
    { icon: Settings, label: 'Settings', href: `/projects/${projectId}/settings` },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span
            className="size-4 rounded-full"
            style={{ backgroundColor: project.color || '#6b7280' }}
          />
          <div>
            <h1 className="text-2xl font-bold tracking-tight">{project.name}</h1>
            <p className="text-muted-foreground font-mono text-sm">
              {project.slug}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {project.status === 'running' ? (
            <>
              <Button
                variant="outline"
                onClick={() => restartMutation.mutate()}
                disabled={restartMutation.isPending}
              >
                <RotateCcw className="mr-2 size-4" />
                Restart
              </Button>
              <Button
                variant="destructive"
                onClick={() => stopMutation.mutate()}
                disabled={stopMutation.isPending}
              >
                <Square className="mr-2 size-4" />
                Stop
              </Button>
            </>
          ) : (
            <Button
              onClick={() => startMutation.mutate()}
              disabled={startMutation.isPending || releases.length === 0}
            >
              <Play className="mr-2 size-4" />
              Start
            </Button>
          )}
        </div>
      </div>

      {/* Status Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Status</CardDescription>
          </CardHeader>
          <CardContent>
            <StatusBadge status={project.status} />
            {status?.uptime && (
              <p className="mt-2 text-xs text-muted-foreground">
                Uptime: {Math.floor(status.uptime / 60)}m {status.uptime % 60}s
              </p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Active Release</CardDescription>
          </CardHeader>
          <CardContent>
            {activeRelease ? (
              <div>
                <p className="text-2xl font-bold">{activeRelease.version}</p>
                {activeRelease.tag && (
                  <p className="text-xs text-muted-foreground capitalize">
                    {activeRelease.tag}
                  </p>
                )}
              </div>
            ) : (
              <p className="text-muted-foreground">No release</p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Public URL</CardDescription>
          </CardHeader>
          <CardContent>
            <a
              href={publicUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-1 text-sm text-primary hover:underline"
            >
              <span className="truncate">/r/{project.slug}</span>
              <ExternalLink className="size-3" />
            </a>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription>API Key</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2">
              <Key className="size-4 text-muted-foreground" />
              <code className="text-xs truncate">
                {project.apiKey?.slice(0, 8)}...
              </code>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Quick Links */}
      <div>
        <h2 className="text-lg font-semibold mb-4">Quick Access</h2>
        <div className="grid gap-4 md:grid-cols-3 lg:grid-cols-4">
          {quickLinks.map((link) => (
            <Link key={link.href} to={link.href}>
              <Card className="transition-colors hover:bg-muted/50">
                <CardContent className="flex items-center gap-3 p-4">
                  <link.icon className="size-5 text-muted-foreground" />
                  <span className="font-medium">{link.label}</span>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      </div>

      {/* Warning if no releases */}
      {releases.length === 0 && (
        <Card className="border-amber-500/50 bg-amber-500/5">
          <CardContent className="flex items-center gap-4 p-4">
            <div className="size-10 rounded-full bg-amber-500/20 flex items-center justify-center">
              <Code className="size-5 text-amber-500" />
            </div>
            <div>
              <p className="font-medium">No releases available</p>
              <p className="text-sm text-muted-foreground">
                Create a release in the Pipeline to start your service.
              </p>
            </div>
            <Button asChild className="ml-auto">
              <Link to={`/projects/${projectId}/pipeline`}>
                Go to Pipeline
              </Link>
            </Button>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
