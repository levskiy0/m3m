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
  Cpu,
  Plus,
  Trash2,
  LayoutGrid,
  BarChart3,
  Minus,
} from 'lucide-react';
import { toast } from 'sonner';

import { projectsApi, runtimeApi, pipelineApi, goalsApi, widgetsApi } from '@/api';
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
import { EditorTabs, EditorTab } from '@/components/ui/editor-tabs';
import { StatusBadge } from '@/components/shared/status-badge';
import { Skeleton } from '@/components/ui/skeleton';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Sparkline } from '@/components/shared/sparkline';
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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Label } from '@/components/ui/label';
import { cn } from '@/lib/utils';
import type { LogEntry, Goal, GoalStats, Widget, WidgetVariant, CreateWidgetRequest } from '@/types';
import type { StartOptions } from '@/api/runtime';

type LogLevel = 'all' | 'debug' | 'info' | 'warn' | 'error';
type OverviewTab = 'instance' | 'logs';

const LOG_LEVEL_COLORS: Record<string, string> = {
  debug: 'text-gray-400',
  info: 'text-blue-400',
  warn: 'text-amber-400',
  error: 'text-red-400',
};

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
  const initialTab = (location.state as { tab?: string } | null)?.tab as OverviewTab || 'instance';
  const [activeTab, setActiveTab] = useState<OverviewTab>(initialTab);

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

  const { data: widgets = [] } = useQuery({
    queryKey: ['project-widgets', projectId],
    queryFn: () => widgetsApi.list(projectId!),
    enabled: !!projectId,
  });

  // Add Widget Dialog state
  const [addWidgetOpen, setAddWidgetOpen] = useState(false);
  const [selectedGoalId, setSelectedGoalId] = useState<string>('');
  const [selectedVariant, setSelectedVariant] = useState<WidgetVariant>('mini');

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

  const createWidgetMutation = useMutation({
    mutationFn: (data: CreateWidgetRequest) => widgetsApi.create(projectId!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-widgets', projectId] });
      setAddWidgetOpen(false);
      setSelectedGoalId('');
      setSelectedVariant('mini');
      toast.success('Widget added');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to add widget');
    },
  });

  const deleteWidgetMutation = useMutation({
    mutationFn: (widgetId: string) => widgetsApi.delete(projectId!, widgetId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-widgets', projectId] });
      toast.success('Widget removed');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to remove widget');
    },
  });

  const handleAddWidget = () => {
    if (!selectedGoalId) {
      toast.error('Please select a goal');
      return;
    }
    createWidgetMutation.mutate({
      goalId: selectedGoalId,
      variant: selectedVariant,
    });
  };

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

  const publicUrl = `${config.apiURL}/r/${project.slug}`;
  const isRunning = project.status === 'running';
  const isPending = startMutation.isPending || stopMutation.isPending || restartMutation.isPending;


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
              {isRunning && project.runningSource && (
                project.runningSource.startsWith('debug:') ? (
                  <Badge className="bg-amber-500 hover:bg-amber-600 text-white border-transparent">
                    <span className="mr-1.5 size-1.5 rounded-full bg-amber-200 animate-pulse" />
                    {project.runningSource.replace('debug:', '')}
                  </Badge>
                ) : project.runningSource.startsWith('release:') ? (
                  <Badge className="bg-green-500 hover:bg-green-600 text-white border-transparent">
                    <span className="mr-1.5 size-1.5 rounded-full bg-green-200 animate-pulse" />
                    v{project.runningSource.replace('release:', '')}
                  </Badge>
                ) : null
              )}
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
      <div className="w-full">
        <EditorTabs>
          <EditorTab
            active={activeTab === 'instance'}
            onClick={() => setActiveTab('instance')}
            icon={<Activity className="size-4" />}
          >
            Instance
          </EditorTab>
          <EditorTab
            active={activeTab === 'logs'}
            onClick={() => setActiveTab('logs')}
            icon={<Code className="size-4" />}
            badge={
              logs.filter(l => l.level === 'error').length > 0 ? (
                <Badge variant="destructive" className="ml-1 h-5 px-1.5 text-xs">
                  {logs.filter(l => l.level === 'error').length}
                </Badge>
              ) : undefined
            }
          >
            Logs
          </EditorTab>
        </EditorTabs>

        {/* Instance Tab */}
        {activeTab === 'instance' && (
          <div className="space-y-6 pt-4">
          {/* Runtime Stats - Row 1 */}
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            {/* Uptime */}
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardDescription className="text-sm font-medium">Uptime</CardDescription>
                <Clock className="size-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {isRunning && stats?.uptime_formatted ? stats.uptime_formatted : '--'}
                </div>
                <p className="text-xs text-muted-foreground mt-1">
                  {stats?.started_at
                    ? `Since ${new Date(stats.started_at).toLocaleString()}`
                    : 'Service not running'}
                </p>
              </CardContent>
            </Card>

            {/* Requests */}
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardDescription className="text-sm font-medium">Requests</CardDescription>
                <Zap className="size-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="flex items-end justify-between gap-4">
                  <div>
                    <div className="text-2xl font-bold">
                      {isRunning && stats?.total_requests != null
                        ? stats.total_requests.toLocaleString()
                        : '--'}
                    </div>
                    <p className="text-xs text-muted-foreground mt-1">
                      {isRunning && stats?.routes_count
                        ? `${stats.routes_count} route${stats.routes_count !== 1 ? 's' : ''}`
                        : 'No routes'}
                    </p>
                  </div>
                  <Sparkline
                    data={stats?.history?.requests || []}
                    width={80}
                    height={32}
                    color="hsl(var(--primary))"
                    strokeWidth={2}
                    fillOpacity={0.15}
                  />
                </div>
              </CardContent>
            </Card>

            {/* Scheduled Jobs */}
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardDescription className="text-sm font-medium">Scheduled Jobs</CardDescription>
                <Clock className="size-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="flex items-end justify-between gap-4">
                  <div>
                    <div className="text-2xl font-bold">
                      {isRunning && stats?.scheduled_jobs != null ? stats.scheduled_jobs : '--'}
                    </div>
                    <p className="text-xs text-muted-foreground mt-1">
                      {isRunning && stats
                        ? stats.scheduler_active ? 'Scheduler active' : 'Scheduler inactive'
                        : 'No data'}
                    </p>
                  </div>
                  <Sparkline
                    data={stats?.history?.jobs || []}
                    width={80}
                    height={32}
                    color="hsl(var(--chart-3))"
                    strokeWidth={2}
                    fillOpacity={0.15}
                  />
                </div>
              </CardContent>
            </Card>

            {/* Storage */}
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardDescription className="text-sm font-medium">Storage</CardDescription>
                <HardDrive className="size-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {stats?.storage_bytes != null ? formatBytes(stats.storage_bytes) : '--'}
                </div>
                <p className="text-xs text-muted-foreground mt-1">
                  Project files
                </p>
              </CardContent>
            </Card>
          </div>

          {/* Runtime Stats - Row 2 */}
          <div className="grid gap-4 md:grid-cols-3">
            {/* Database */}
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardDescription className="text-sm font-medium">Database</CardDescription>
                <Database className="size-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {stats?.database_bytes != null ? formatBytes(stats.database_bytes) : '--'}
                </div>
                <p className="text-xs text-muted-foreground mt-1">
                  Collections data
                </p>
              </CardContent>
            </Card>

            {/* Memory */}
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardDescription className="text-sm font-medium">Memory</CardDescription>
                <MemoryStick className="size-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="flex items-end justify-between gap-4">
                  <div>
                    <div className="text-2xl font-bold">
                      {isRunning && stats?.memory?.alloc != null ? formatBytes(stats.memory.alloc) : '--'}
                    </div>
                    <p className="text-xs text-muted-foreground mt-1">Current usage</p>
                  </div>
                  <Sparkline
                    data={stats?.history?.memory || []}
                    width={80}
                    height={32}
                    color="hsl(var(--chart-2))"
                    strokeWidth={2}
                    fillOpacity={0.15}
                  />
                </div>
              </CardContent>
            </Card>

            {/* CPU */}
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardDescription className="text-sm font-medium">CPU</CardDescription>
                <Cpu className="size-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="flex items-end justify-between gap-4">
                  <div>
                    <div className="text-2xl font-bold">
                      {stats?.cpu_percent != null ? `${stats.cpu_percent.toFixed(1)}%` : '--'}
                    </div>
                    <p className="text-xs text-muted-foreground mt-1">Process usage</p>
                  </div>
                  <Sparkline
                    data={stats?.history?.cpu || []}
                    width={80}
                    height={32}
                    color="hsl(var(--chart-4))"
                    strokeWidth={2}
                    fillOpacity={0.15}
                  />
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Project Info Cards */}
          <div className="grid gap-4 md:grid-cols-2">
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

          {/* Widgets Section */}
          <div>
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold">Widgets</h2>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setAddWidgetOpen(true)}
                disabled={goals.length === 0}
              >
                <Plus className="mr-1.5 size-4" />
                Add Widget
              </Button>
            </div>
            {widgets.length > 0 ? (
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {widgets.map((widget) => {
                  const goal = goals.find(g => g.id === widget.goalId);
                  const stat = goal ? goalStatsMap.get(goal.id) : undefined;
                  if (!goal) return null;
                  return (
                    <WidgetCard
                      key={widget.id}
                      widget={widget}
                      goal={goal}
                      stats={stat}
                      onDelete={() => deleteWidgetMutation.mutate(widget.id)}
                    />
                  );
                })}
              </div>
            ) : (
              <Card>
                <CardContent className="flex flex-col items-center justify-center py-8">
                  <LayoutGrid className="size-10 text-muted-foreground mb-3" />
                  <p className="text-muted-foreground text-center text-sm">
                    {goals.length === 0
                      ? 'Create goals first to add widgets.'
                      : 'No widgets added. Add widgets to display goal metrics.'}
                  </p>
                  {goals.length === 0 && (
                    <Button variant="outline" size="sm" className="mt-3" asChild>
                      <Link to={`/projects/${projectId}/goals`}>
                        Create Goals
                      </Link>
                    </Button>
                  )}
                </CardContent>
              </Card>
            )}
          </div>
          </div>
        )}

        {/* Logs Tab */}
        {activeTab === 'logs' && (
          <Card className="flex flex-col gap-0 rounded-t-none py-0 overflow-hidden" style={{ height: 'calc(100vh - 280px)' }}>
            {/* Logs Header */}
            <div className="flex items-center justify-between flex-shrink-0 px-4 py-3 border-b">
              <div className="flex items-center gap-4">
                <span className="text-sm font-medium">
                  {filteredLogs.length} log entries
                </span>
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
            <ScrollArea
              ref={scrollRef}
              className="flex-1 bg-zinc-950 font-mono text-xs"
            >
              {filteredLogs.length === 0 ? (
                <div className="flex items-center justify-center h-full text-muted-foreground py-12">
                  No logs available
                </div>
              ) : (
                <div className="space-y-0.5 p-4">
                  {filteredLogs.map((log, index) => (
                    <div key={index} className="flex gap-2 text-gray-300">
                      <span className="text-gray-500 shrink-0">
                        {new Date(log.timestamp).toLocaleTimeString()}
                      </span>
                      <span
                        className={cn(
                          'shrink-0 uppercase w-12',
                          LOG_LEVEL_COLORS[log.level]
                        )}
                      >
                        [{log.level}]
                      </span>
                      <span className="text-gray-200 break-all">{log.message}</span>
                    </div>
                  ))}
                </div>
              )}
            </ScrollArea>
          </Card>
        )}
      </div>

      {/* Add Widget Dialog */}
      <Dialog open={addWidgetOpen} onOpenChange={setAddWidgetOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Add Widget</DialogTitle>
            <DialogDescription>
              Select a goal and display variant to add a widget to your dashboard.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            {/* Goal Selection */}
            <div className="space-y-2">
              <Label>Goal</Label>
              <Select value={selectedGoalId} onValueChange={setSelectedGoalId}>
                <SelectTrigger>
                  <SelectValue placeholder="Select a goal" />
                </SelectTrigger>
                <SelectContent>
                  {goals.map((goal) => (
                    <SelectItem key={goal.id} value={goal.id}>
                      <div className="flex items-center gap-2">
                        <span
                          className="size-2.5 rounded-full"
                          style={{ backgroundColor: goal.color || '#6b7280' }}
                        />
                        {goal.name}
                        <Badge variant="secondary" className="ml-auto text-xs">
                          {goal.type === 'daily_counter' ? 'Daily' : 'Counter'}
                        </Badge>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Variant Selection */}
            <div className="space-y-2">
              <Label>Display Variant</Label>
              <RadioGroup
                value={selectedVariant}
                onValueChange={(v) => setSelectedVariant(v as WidgetVariant)}
                className="grid grid-cols-3 gap-2"
              >
                <Label
                  htmlFor="variant-mini"
                  className={cn(
                    "flex flex-col items-center justify-center rounded-md border-2 p-3 cursor-pointer hover:bg-accent",
                    selectedVariant === 'mini' ? "border-primary" : "border-muted"
                  )}
                >
                  <RadioGroupItem value="mini" id="variant-mini" className="sr-only" />
                  <Minus className="size-5 mb-1" />
                  <span className="text-xs">Mini</span>
                </Label>
                <Label
                  htmlFor="variant-detailed"
                  className={cn(
                    "flex flex-col items-center justify-center rounded-md border-2 p-3 cursor-pointer hover:bg-accent",
                    selectedVariant === 'detailed' ? "border-primary" : "border-muted"
                  )}
                >
                  <RadioGroupItem value="detailed" id="variant-detailed" className="sr-only" />
                  <BarChart3 className="size-5 mb-1" />
                  <span className="text-xs">Detailed</span>
                </Label>
                <Label
                  htmlFor="variant-simple"
                  className={cn(
                    "flex flex-col items-center justify-center rounded-md border-2 p-3 cursor-pointer hover:bg-accent",
                    selectedVariant === 'simple' ? "border-primary" : "border-muted"
                  )}
                >
                  <RadioGroupItem value="simple" id="variant-simple" className="sr-only" />
                  <Target className="size-5 mb-1" />
                  <span className="text-xs">Simple</span>
                </Label>
              </RadioGroup>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setAddWidgetOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleAddWidget}
              disabled={!selectedGoalId || createWidgetMutation.isPending}
            >
              Add Widget
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

// Widget Card Component - renders goal widget based on variant
interface WidgetCardProps {
  widget: Widget;
  goal: Goal;
  stats?: GoalStats;
  onDelete: () => void;
}

function WidgetCard({ widget, goal, stats, onDelete }: WidgetCardProps) {
  const sparklineData = stats?.dailyStats?.slice(-14).map((d) => d.value) || [];
  const isDailyCounter = goal.type === 'daily_counter';

  // Render based on variant
  if (widget.variant === 'mini') {
    // Mini variant - compact with sparkline for daily counters
    return (
      <Card className="group relative">
        <Button
          variant="ghost"
          size="icon"
          className="absolute top-2 right-2 size-6 opacity-0 group-hover:opacity-100 transition-opacity"
          onClick={onDelete}
        >
          <Trash2 className="size-3.5 text-muted-foreground hover:text-destructive" />
        </Button>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <div className="flex items-center gap-2">
            <span
              className="size-2.5 rounded-full"
              style={{ backgroundColor: goal.color || '#6b7280' }}
            />
            <CardDescription className="text-sm font-medium">{goal.name}</CardDescription>
          </div>
          <Target className="size-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          {isDailyCounter && sparklineData.length > 0 ? (
            <div className="flex items-end justify-between gap-4">
              <div>
                <div className="text-2xl font-bold">
                  {stats?.value?.toLocaleString() ?? 0}
                </div>
                <p className="text-xs text-muted-foreground mt-1">
                  {goal.description || 'Daily counter'}
                </p>
              </div>
              <Sparkline
                data={sparklineData}
                width={80}
                height={32}
                color={goal.color || '#6b7280'}
                strokeWidth={2}
                fillOpacity={0.15}
              />
            </div>
          ) : (
            <div>
              <div className="text-2xl font-bold">
                {stats?.value?.toLocaleString() ?? 0}
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                {goal.description || 'Total count'}
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    );
  }

  if (widget.variant === 'detailed') {
    // Detailed variant - larger card with more info
    return (
      <Card className="group relative">
        <Button
          variant="ghost"
          size="icon"
          className="absolute top-2 right-2 size-6 opacity-0 group-hover:opacity-100 transition-opacity"
          onClick={onDelete}
        >
          <Trash2 className="size-3.5 text-muted-foreground hover:text-destructive" />
        </Button>
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
              {isDailyCounter ? 'Daily' : 'Counter'}
            </Badge>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-end justify-between gap-4">
            <div>
              <div className="text-2xl font-bold">
                {stats?.value?.toLocaleString() ?? 0}
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                {goal.description || (isDailyCounter ? 'Daily counter' : 'Total count')}
              </p>
            </div>
            {isDailyCounter && sparklineData.length > 0 && (
              <Sparkline
                data={sparklineData}
                width={80}
                height={32}
                color={goal.color || '#6b7280'}
                strokeWidth={2}
                fillOpacity={0.15}
              />
            )}
          </div>
        </CardContent>
      </Card>
    );
  }

  // Simple variant - just the number
  return (
    <Card className="group relative">
      <Button
        variant="ghost"
        size="icon"
        className="absolute top-2 right-2 size-6 opacity-0 group-hover:opacity-100 transition-opacity"
        onClick={onDelete}
      >
        <Trash2 className="size-3.5 text-muted-foreground hover:text-destructive" />
      </Button>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <div className="flex items-center gap-2">
          <span
            className="size-2.5 rounded-full"
            style={{ backgroundColor: goal.color || '#6b7280' }}
          />
          <CardDescription className="text-sm font-medium">{goal.name}</CardDescription>
        </div>
        <Target className="size-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">
          {stats?.value?.toLocaleString() ?? 0}
        </div>
        <p className="text-xs text-muted-foreground mt-1">
          {goal.description || (isDailyCounter ? 'Daily counter' : 'Total count')}
        </p>
      </CardContent>
    </Card>
  );
}

