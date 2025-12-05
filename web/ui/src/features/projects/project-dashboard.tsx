import { useState, useEffect, useRef } from 'react';
import { useParams, Link, useLocation } from 'react-router-dom';
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
  Settings,
  Key,
  ExternalLink,
  Clock,
  Activity,
  AlertTriangle,
  Zap,
  MemoryStick,
  Copy,
  Check,
  ArrowDown,
  Download,
  ChevronRight,
  ChevronDown,
  Bug,
  Tag,
} from 'lucide-react';
import { toast } from 'sonner';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip as RechartsTooltip,
  ResponsiveContainer,
} from 'recharts';

import { projectsApi, runtimeApi, pipelineApi, goalsApi } from '@/api';
import { config } from '@/lib/config';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { StatusBadge } from '@/components/shared/status-badge';
import { Skeleton } from '@/components/ui/skeleton';
import { ScrollArea } from '@/components/ui/scroll-area';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { cn } from '@/lib/utils';
import type { LogEntry, Goal, GoalStats } from '@/types';
import type { StartOptions } from '@/api/runtime';

type LogLevel = 'all' | 'debug' | 'info' | 'warn' | 'error';

const LOG_LEVEL_COLORS: Record<string, string> = {
  debug: 'text-gray-400',
  info: 'text-blue-400',
  warn: 'text-amber-400',
  error: 'text-red-400',
};

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;

  if (days > 0) {
    return `${days}d ${hours}h ${minutes}m`;
  }
  if (hours > 0) {
    return `${hours}h ${minutes}m ${secs}s`;
  }
  if (minutes > 0) {
    return `${minutes}m ${secs}s`;
  }
  return `${secs}s`;
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

export function ProjectDashboard() {
  const { projectId } = useParams<{ projectId: string }>();
  const location = useLocation();
  const queryClient = useQueryClient();
  const [copied, setCopied] = useState(false);
  const [levelFilter, setLevelFilter] = useState<LogLevel>('all');
  const [autoScroll, setAutoScroll] = useState(true);
  const scrollRef = useRef<HTMLDivElement>(null);

  // Get initial tab from location state
  const initialTab = (location.state as { tab?: string } | null)?.tab || 'instance';

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

  const { data: branches = [] } = useQuery({
    queryKey: ['branches', projectId],
    queryFn: () => pipelineApi.listBranches(projectId!),
    enabled: !!projectId,
  });

  const { data: status } = useQuery({
    queryKey: ['runtime-status', projectId],
    queryFn: () => runtimeApi.status(projectId!),
    enabled: !!projectId,
    refetchInterval: project?.status === 'running' ? 3000 : false,
  });

  const { data: stats } = useQuery({
    queryKey: ['runtime-stats', projectId],
    queryFn: () => runtimeApi.monitor(projectId!),
    enabled: !!projectId && project?.status === 'running',
    refetchInterval: project?.status === 'running' ? 5000 : false,
  });

  const { data: logsData = [] } = useQuery({
    queryKey: ['logs', projectId],
    queryFn: () => runtimeApi.logs(projectId!),
    enabled: !!projectId,
    refetchInterval: 3000,
  });

  const { data: goals = [] } = useQuery({
    queryKey: ['project-goals', projectId],
    queryFn: () => goalsApi.listProject(projectId!),
    enabled: !!projectId,
  });

  const { data: goalStats = [] } = useQuery({
    queryKey: ['project-goal-stats', projectId, goals.map(g => g.id)],
    queryFn: () => goalsApi.getStats({ goalIds: goals.map(g => g.id) }),
    enabled: !!projectId && goals.length > 0,
    refetchInterval: 10000,
  });

  const logs: LogEntry[] = Array.isArray(logsData) ? logsData : [];

  const filteredLogs = logs.filter((log) =>
    levelFilter === 'all' ? true : log.level === levelFilter
  );

  useEffect(() => {
    if (autoScroll && scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  const startMutation = useMutation({
    mutationFn: (options?: StartOptions) => runtimeApi.start(projectId!, options),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project', projectId] });
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      queryClient.invalidateQueries({ queryKey: ['runtime-status', projectId] });
      queryClient.invalidateQueries({ queryKey: ['runtime-stats', projectId] });
      toast.success('Service started');
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
      toast.success('Service stopped');
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
      queryClient.invalidateQueries({ queryKey: ['runtime-stats', projectId] });
      toast.success('Service restarted');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to restart');
    },
  });

  const copyApiKey = async () => {
    if (project?.apiKey) {
      await navigator.clipboard.writeText(project.apiKey);
      setCopied(true);
      toast.success('API key copied');
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleDownloadLogs = async () => {
    try {
      const blob = await runtimeApi.downloadLogs(projectId!);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${project?.slug || projectId}-logs.txt`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      console.error('Failed to download logs:', err);
    }
  };

  if (projectLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-10 w-64" />
        <Skeleton className="h-[400px]" />
      </div>
    );
  }

  if (!project) {
    return <div>Project not found</div>;
  }

  const activeRelease = releases.find((r) => r.isActive);
  const publicUrl = `${config.apiURL}/r/${project.slug}`;
  const isRunning = project.status === 'running';
  const isPending = startMutation.isPending || stopMutation.isPending || restartMutation.isPending;

  const quickLinks = [
    { icon: Code, label: 'Pipeline', description: 'Edit code & releases', href: `/projects/${projectId}/pipeline` },
    { icon: HardDrive, label: 'Storage', description: 'Manage files', href: `/projects/${projectId}/storage` },
    { icon: Database, label: 'Models', description: 'Database schemas', href: `/projects/${projectId}/models` },
    { icon: Target, label: 'Goals', description: 'Track metrics', href: `/projects/${projectId}/goals` },
    { icon: Variable, label: 'Environment', description: 'Config variables', href: `/projects/${projectId}/environment` },
    { icon: Settings, label: 'Settings', description: 'Project settings', href: `/projects/${projectId}/settings` },
  ];

  // Create goal stats map for quick lookup
  const goalStatsMap = new Map<string, GoalStats>();
  goalStats.forEach((s) => goalStatsMap.set(s.goalID, s));

  return (
    <div className="space-y-6">
      {/* Header with Service Controls */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-3">
          <div
            className="size-10 rounded-lg flex items-center justify-center"
            style={{ backgroundColor: project.color || '#6b7280' }}
          >
            <Activity className="size-5 text-white" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight">{project.name}</h1>
            <div className="flex items-center gap-2 mt-0.5">
              <code className="text-muted-foreground text-sm">{project.slug}</code>
              <StatusBadge status={project.status} />
            </div>
          </div>
        </div>

        {/* Service Control Panel */}
        <div className="flex items-center gap-2 bg-muted/50 rounded-lg p-2">
          {isRunning ? (
            <>
              <Button
                variant="outline"
                size="sm"
                onClick={() => restartMutation.mutate()}
                disabled={isPending}
                className="gap-2"
              >
                <RotateCcw className={cn("size-4", restartMutation.isPending && "animate-spin")} />
                Restart
              </Button>
              <Button
                variant="destructive"
                size="sm"
                onClick={() => stopMutation.mutate()}
                disabled={isPending}
                className="gap-2"
              >
                <Square className="size-4" />
                Stop
              </Button>
            </>
          ) : (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  size="sm"
                  disabled={isPending || (releases.length === 0 && branches.length === 0)}
                  className="gap-2 bg-green-600 hover:bg-green-700"
                >
                  <Play className="size-4" />
                  Start Service
                  <ChevronDown className="size-4 ml-1" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-56">
                {releases.length > 0 && (
                  <>
                    <DropdownMenuLabel className="flex items-center gap-2">
                      <Tag className="size-3" />
                      Releases
                    </DropdownMenuLabel>
                    {releases.slice(0, 5).map((release) => (
                      <DropdownMenuItem
                        key={release.id}
                        onClick={() => startMutation.mutate({ version: release.version })}
                      >
                        <span className="font-mono">{release.version}</span>
                        {release.tag && (
                          <Badge variant="secondary" className="ml-auto text-xs capitalize">
                            {release.tag}
                          </Badge>
                        )}
                      </DropdownMenuItem>
                    ))}
                    {releases.length > 5 && (
                      <DropdownMenuItem asChild>
                        <Link to={`/projects/${projectId}/pipeline`} className="text-muted-foreground">
                          View all releases...
                        </Link>
                      </DropdownMenuItem>
                    )}
                  </>
                )}
                {branches.length > 0 && (
                  <>
                    {releases.length > 0 && <DropdownMenuSeparator />}
                    <DropdownMenuLabel className="flex items-center gap-2">
                      <Bug className="size-3" />
                      Debug (Branches)
                    </DropdownMenuLabel>
                    {branches.map((branch) => (
                      <DropdownMenuItem
                        key={branch.id}
                        onClick={() => startMutation.mutate({ branch: branch.name })}
                      >
                        <Code className="size-4 mr-2" />
                        {branch.name}
                      </DropdownMenuItem>
                    ))}
                  </>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
      </div>

      {/* Warning if no releases and no branches */}
      {releases.length === 0 && branches.length === 0 && (
        <Card className="border-amber-500/50 bg-amber-500/5">
          <CardContent className="flex items-center gap-4 p-4">
            <div className="size-10 rounded-full bg-amber-500/20 flex items-center justify-center shrink-0">
              <AlertTriangle className="size-5 text-amber-500" />
            </div>
            <div className="flex-1">
              <p className="font-medium">No code available</p>
              <p className="text-sm text-muted-foreground">
                Create a release or write code in the Pipeline to start your service.
              </p>
            </div>
            <Button asChild size="sm">
              <Link to={`/projects/${projectId}/pipeline`}>
                Go to Pipeline
                <ChevronRight className="ml-1 size-4" />
              </Link>
            </Button>
          </CardContent>
        </Card>
      )}

      {/* Main Content Tabs */}
      <Tabs defaultValue={initialTab} className="space-y-4">
        <TabsList>
          <TabsTrigger value="instance" className="gap-2">
            <Activity className="size-4" />
            Instance
          </TabsTrigger>
          <TabsTrigger value="logs" className="gap-2">
            <Code className="size-4" />
            Logs
            {logs.filter(l => l.level === 'error').length > 0 && (
              <Badge variant="destructive" className="ml-1 h-5 px-1.5 text-xs">
                {logs.filter(l => l.level === 'error').length}
              </Badge>
            )}
          </TabsTrigger>
          <TabsTrigger value="stats" className="gap-2">
            <Target className="size-4" />
            Stats
          </TabsTrigger>
        </TabsList>

        {/* Instance Tab */}
        <TabsContent value="instance" className="space-y-6">
          {/* Runtime Stats */}
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardDescription className="text-sm font-medium">Uptime</CardDescription>
                <Clock className="size-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {isRunning && status?.uptime ? formatUptime(status.uptime) : '--'}
                </div>
                <p className="text-xs text-muted-foreground mt-1">
                  {status?.startedAt
                    ? `Since ${new Date(status.startedAt).toLocaleString()}`
                    : 'Service not running'}
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardDescription className="text-sm font-medium">Requests</CardDescription>
                <Zap className="size-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {isRunning && stats?.requests != null ? stats.requests.toLocaleString() : '--'}
                </div>
                <p className="text-xs text-muted-foreground mt-1">
                  {isRunning && stats?.avgResponseTime != null
                    ? `Avg ${stats.avgResponseTime.toFixed(0)}ms response`
                    : 'No data'}
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardDescription className="text-sm font-medium">Errors</CardDescription>
                <AlertTriangle className="size-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className={cn(
                  "text-2xl font-bold",
                  isRunning && stats?.errors != null && stats.errors > 0 && "text-red-500"
                )}>
                  {isRunning && stats?.errors != null ? stats.errors.toLocaleString() : '--'}
                </div>
                <p className="text-xs text-muted-foreground mt-1">
                  {isRunning && stats?.requests != null && stats.requests > 0
                    ? `${(((stats.errors ?? 0) / stats.requests) * 100).toFixed(1)}% error rate`
                    : 'No errors'}
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardDescription className="text-sm font-medium">Memory</CardDescription>
                <MemoryStick className="size-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {isRunning && stats?.memoryUsage != null ? formatBytes(stats.memoryUsage) : '--'}
                </div>
                <p className="text-xs text-muted-foreground mt-1">Current usage</p>
              </CardContent>
            </Card>
          </div>

          {/* Project Info Cards */}
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            <Card className={cn(isRunning && project.runningSource?.startsWith('debug:') && "border-amber-500/50")}>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">
                  {isRunning ? 'Running' : 'Active Release'}
                </CardTitle>
              </CardHeader>
              <CardContent>
                {isRunning && project.runningSource ? (
                  <div className="flex items-center justify-between">
                    <div>
                      {project.runningSource.startsWith('debug:') ? (
                        <>
                          <div className="flex items-center gap-2">
                            <Bug className="size-5 text-amber-500" />
                            <p className="text-2xl font-bold text-amber-500">Debug</p>
                          </div>
                          <Badge variant="outline" className="mt-1 border-amber-500/50 text-amber-500">
                            {project.runningSource.replace('debug:', '')}
                          </Badge>
                        </>
                      ) : (
                        <>
                          <p className="text-2xl font-bold">
                            {project.runningSource.replace('release:', '')}
                          </p>
                          {activeRelease?.tag && (
                            <Badge variant="secondary" className="mt-1 capitalize">
                              {activeRelease.tag}
                            </Badge>
                          )}
                        </>
                      )}
                    </div>
                    <Button variant="ghost" size="sm" asChild>
                      <Link to={`/projects/${projectId}/pipeline`}>
                        View
                        <ChevronRight className="ml-1 size-4" />
                      </Link>
                    </Button>
                  </div>
                ) : activeRelease ? (
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-2xl font-bold">{activeRelease.version}</p>
                      {activeRelease.tag && (
                        <Badge variant="secondary" className="mt-1 capitalize">
                          {activeRelease.tag}
                        </Badge>
                      )}
                    </div>
                    <Button variant="ghost" size="sm" asChild>
                      <Link to={`/projects/${projectId}/pipeline`}>
                        View
                        <ChevronRight className="ml-1 size-4" />
                      </Link>
                    </Button>
                  </div>
                ) : (
                  <p className="text-muted-foreground">No active release</p>
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Public URL</CardTitle>
              </CardHeader>
              <CardContent>
                <a
                  href={publicUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-2 text-primary hover:underline"
                >
                  <span className="font-mono text-sm truncate">/r/{project.slug}</span>
                  <ExternalLink className="size-4 shrink-0" />
                </a>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">API Key</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex items-center gap-2">
                  <Key className="size-4 text-muted-foreground shrink-0" />
                  <code className="text-sm truncate flex-1 font-mono">
                    {project.apiKey?.slice(0, 12)}...
                  </code>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="size-8 shrink-0"
                    onClick={copyApiKey}
                  >
                    {copied ? (
                      <Check className="size-4 text-green-500" />
                    ) : (
                      <Copy className="size-4" />
                    )}
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Quick Links */}
          <div>
            <h2 className="text-lg font-semibold mb-4">Quick Access</h2>
            <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
              {quickLinks.map((link) => (
                <Link key={link.href} to={link.href}>
                  <Card className="transition-all hover:bg-muted/50 hover:border-primary/30">
                    <CardContent className="flex items-center gap-4 p-4">
                      <div className="size-10 rounded-lg bg-muted flex items-center justify-center shrink-0">
                        <link.icon className="size-5 text-muted-foreground" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="font-medium">{link.label}</p>
                        <p className="text-sm text-muted-foreground truncate">
                          {link.description}
                        </p>
                      </div>
                      <ChevronRight className="size-4 text-muted-foreground" />
                    </CardContent>
                  </Card>
                </Link>
              ))}
            </div>
          </div>
        </TabsContent>

        {/* Logs Tab */}
        <TabsContent value="logs" className="space-y-4">
          <Card>
            <CardHeader className="pb-2">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <CardTitle className="text-sm font-medium">
                    {filteredLogs.length} log entries
                  </CardTitle>
                  <Select
                    value={levelFilter}
                    onValueChange={(v) => setLevelFilter(v as LogLevel)}
                  >
                    <SelectTrigger className="w-32 h-8">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All Levels</SelectItem>
                      <SelectItem value="debug">Debug</SelectItem>
                      <SelectItem value="info">Info</SelectItem>
                      <SelectItem value="warn">Warning</SelectItem>
                      <SelectItem value="error">Error</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setAutoScroll(!autoScroll)}
                    className={cn("h-8", autoScroll && "bg-muted")}
                  >
                    Auto-scroll: {autoScroll ? 'On' : 'Off'}
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8"
                    onClick={() => {
                      if (scrollRef.current) {
                        scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
                      }
                    }}
                  >
                    <ArrowDown className="size-4" />
                  </Button>
                  <Button variant="outline" size="sm" onClick={handleDownloadLogs} className="h-8">
                    <Download className="mr-2 size-4" />
                    Download
                  </Button>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <ScrollArea
                ref={scrollRef}
                className="h-[500px] rounded-md border bg-zinc-950 p-4 font-mono text-sm"
              >
                {filteredLogs.length === 0 ? (
                  <div className="flex items-center justify-center h-full text-muted-foreground">
                    No logs available
                  </div>
                ) : (
                  <div className="space-y-1">
                    {filteredLogs.map((log, index) => (
                      <div key={index} className="flex gap-3 text-gray-300 leading-relaxed">
                        <span className="text-gray-500 shrink-0 text-xs">
                          {new Date(log.timestamp).toLocaleTimeString()}
                        </span>
                        <span
                          className={cn(
                            'shrink-0 uppercase w-14 text-xs font-medium',
                            LOG_LEVEL_COLORS[log.level]
                          )}
                        >
                          [{log.level}]
                        </span>
                        <span className="break-all text-gray-200">{log.message}</span>
                      </div>
                    ))}
                  </div>
                )}
              </ScrollArea>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Stats Tab */}
        <TabsContent value="stats" className="space-y-6">
          {/* Goals Section */}
          {goals.length > 0 ? (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold">Goals & Metrics</h2>
                <Button variant="outline" size="sm" asChild>
                  <Link to={`/projects/${projectId}/goals`}>
                    Manage Goals
                    <ChevronRight className="ml-1 size-4" />
                  </Link>
                </Button>
              </div>
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {goals.map((goal) => {
                  const stat = goalStatsMap.get(goal.id);
                  return (
                    <GoalCard key={goal.id} goal={goal} stats={stat} />
                  );
                })}
              </div>
            </div>
          ) : (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12">
                <Target className="size-12 text-muted-foreground mb-4" />
                <h3 className="text-lg font-medium mb-2">No goals configured</h3>
                <p className="text-muted-foreground text-center mb-4">
                  Create goals to track metrics like page views, signups, and more.
                </p>
                <Button asChild>
                  <Link to={`/projects/${projectId}/goals`}>
                    Create Goals
                  </Link>
                </Button>
              </CardContent>
            </Card>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}

// Goal Card Component with mini chart
function GoalCard({ goal, stats }: { goal: Goal; stats?: GoalStats }) {
  const chartData = stats?.dailyStats?.slice(-7).map((d) => ({
    date: new Date(d.date).toLocaleDateString('en', { weekday: 'short' }),
    value: d.value,
  })) || [];

  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span
              className="size-3 rounded-full"
              style={{ backgroundColor: goal.color || '#6b7280' }}
            />
            <CardTitle className="text-sm font-medium">{goal.name}</CardTitle>
          </div>
          <Badge variant="secondary" className="text-xs">
            {goal.type === 'counter' ? 'Counter' : 'Daily'}
          </Badge>
        </div>
      </CardHeader>
      <CardContent>
        <div className="text-3xl font-bold mb-2">
          {stats?.value?.toLocaleString() ?? 0}
        </div>
        {chartData.length > 0 && (
          <div className="h-16 mt-2">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={chartData}>
                <defs>
                  <linearGradient id={`gradient-${goal.id}`} x1="0" y1="0" x2="0" y2="1">
                    <stop
                      offset="0%"
                      stopColor={goal.color || '#6b7280'}
                      stopOpacity={0.3}
                    />
                    <stop
                      offset="100%"
                      stopColor={goal.color || '#6b7280'}
                      stopOpacity={0}
                    />
                  </linearGradient>
                </defs>
                <XAxis dataKey="date" hide />
                <YAxis hide />
                <RechartsTooltip
                  contentStyle={{
                    backgroundColor: 'hsl(var(--popover))',
                    border: '1px solid hsl(var(--border))',
                    borderRadius: '6px',
                    fontSize: '12px',
                  }}
                />
                <Area
                  type="monotone"
                  dataKey="value"
                  stroke={goal.color || '#6b7280'}
                  fill={`url(#gradient-${goal.id})`}
                  strokeWidth={2}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        )}
        {goal.description && (
          <p className="text-xs text-muted-foreground mt-2 line-clamp-1">
            {goal.description}
          </p>
        )}
      </CardContent>
    </Card>
  );
}
