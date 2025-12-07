import { useState } from 'react';
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
  ChevronRight,
  ChevronDown,
  Bug,
  Tag,
  Cpu,
  Plus,
  LayoutGrid,
  BarChart3,
  Minus,
} from 'lucide-react';
import { toast } from 'sonner';

import { projectsApi, pipelineApi, goalsApi, widgetsApi } from '@/api';
import { config } from '@/lib/config';
import { formatBytes } from '@/lib/format';
import { queryKeys } from '@/lib/query-keys';
import { cn, copyToClipboard } from '@/lib/utils';
import { useProjectRuntime } from '@/hooks';
import type { LogEntry, GoalStats, WidgetVariant, CreateWidgetRequest } from '@/types';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { EditorTabs, EditorTab } from '@/components/ui/editor-tabs';
import { StatusBadge } from '@/components/shared/status-badge';
import { Skeleton } from '@/components/ui/skeleton';
import { LogsViewer } from '@/components/shared/logs-viewer';
import { MetricCard } from '@/components/shared/metric-card';
import { WidgetCard } from '@/components/shared/widget-card';
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

type OverviewTab = 'instance' | 'logs';

export function ProjectDashboard() {
  const { projectId } = useParams<{ projectId: string }>();
  const location = useLocation();
  const queryClient = useQueryClient();
  const [copied, setCopied] = useState(false);

  const initialTab = (location.state as { tab?: string } | null)?.tab as OverviewTab || 'instance';
  const [activeTab, setActiveTab] = useState<OverviewTab>(initialTab);

  const { data: project, isLoading: projectLoading } = useQuery({
    queryKey: queryKeys.projects.detail(projectId!),
    queryFn: () => projectsApi.get(projectId!),
    enabled: !!projectId,
  });

  const isRunning = project?.status === 'running';

  // Runtime hook - handles start/stop/restart, logs, monitor
  const runtime = useProjectRuntime({
    projectId: projectId!,
    projectSlug: project?.slug,
    enabled: !!projectId,
    refetchLogsInterval: 3000,
    refetchStatusInterval: isRunning ? 5000 : false,
  });

  const { data: releases = [] } = useQuery({
    queryKey: queryKeys.pipeline.releases(projectId!),
    queryFn: () => pipelineApi.listReleases(projectId!),
    enabled: !!projectId,
  });

  const { data: branches = [] } = useQuery({
    queryKey: queryKeys.pipeline.branches(projectId!),
    queryFn: () => pipelineApi.listBranches(projectId!),
    enabled: !!projectId,
  });

  const { data: goals = [] } = useQuery({
    queryKey: queryKeys.goals.project(projectId!),
    queryFn: () => goalsApi.listProject(projectId!),
    enabled: !!projectId,
  });

  const { data: goalStats = [] } = useQuery({
    queryKey: queryKeys.goals.stats(projectId!, goals.map(g => g.id)),
    queryFn: () => goalsApi.getStats({ goalIds: goals.map(g => g.id) }),
    enabled: !!projectId && goals.length > 0,
    refetchInterval: 10000,
  });

  const { data: widgets = [] } = useQuery({
    queryKey: queryKeys.widgets.all(projectId!),
    queryFn: () => widgetsApi.list(projectId!),
    enabled: !!projectId,
  });

  // Widget Dialog state
  const [addWidgetOpen, setAddWidgetOpen] = useState(false);
  const [selectedGoalId, setSelectedGoalId] = useState<string>('');
  const [selectedVariant, setSelectedVariant] = useState<WidgetVariant>('mini');

  const logs: LogEntry[] = Array.isArray(runtime.logs) ? runtime.logs : [];
  const stats = runtime.monitor;

  const createWidgetMutation = useMutation({
    mutationFn: (data: CreateWidgetRequest) => widgetsApi.create(projectId!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.widgets.all(projectId!) });
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
      queryClient.invalidateQueries({ queryKey: queryKeys.widgets.all(projectId!) });
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
    if (project?.api_key) {
      const success = await copyToClipboard(project.api_key);
      if (success) {
        setCopied(true);
        toast.success('API key copied');
        setTimeout(() => setCopied(false), 2000);
      }
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
        <div className="flex items-center gap-2 bg-muted/50 rounded-xl p-2">
          {isRunning ? (
            <>
              <Button
                variant="outline"
                size="sm"
                onClick={() => runtime.restart()}
                disabled={runtime.isPending}
                className="gap-2"
              >
                <RotateCcw className={cn("size-4", runtime.isRestarting && "animate-spin")} />
                Restart
              </Button>
              <Button
                variant="destructive"
                size="sm"
                onClick={() => runtime.stop()}
                disabled={runtime.isPending}
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
                  disabled={runtime.isPending || (releases.length === 0 && branches.length === 0)}
                  className="gap-2 bg-green-600 hover:bg-green-700 text-white"
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
                        onClick={() => runtime.start({ version: release.version })}
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
                        onClick={() => runtime.start({ branch: branch.name })}
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
            className={activeTab === 'instance' ? 'bg-background' : undefined}
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
          <div className="bg-background rounded-b-xl">
            <div className="border-b" />
            <div className="space-y-6 px-0 py-4">
              {/* Runtime Stats - Row 1 */}
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <MetricCard
                  label="Uptime"
                  value={isRunning && stats?.uptime_formatted ? stats.uptime_formatted : '--'}
                  subtext={stats?.started_at
                    ? `Since ${new Date(stats.started_at).toLocaleString()}`
                    : 'Service not running'}
                  icon={Clock}
                />
                <MetricCard
                  label="Requests"
                  value={isRunning && stats?.total_requests != null
                    ? stats.total_requests.toLocaleString()
                    : '--'}
                  subtext={isRunning && stats?.routes_count
                    ? `${stats.routes_count} route${stats.routes_count !== 1 ? 's' : ''}`
                    : 'No routes'}
                  icon={Zap}
                  sparklineData={stats?.history?.requests}
                  sparklineColor="hsl(var(--primary))"
                />
                <MetricCard
                  label="Scheduled Jobs"
                  value={isRunning && stats?.scheduled_jobs != null ? stats.scheduled_jobs : '--'}
                  subtext={isRunning && stats
                    ? stats.scheduler_active ? 'Scheduler active' : 'Scheduler inactive'
                    : 'No data'}
                  icon={Clock}
                  sparklineData={stats?.history?.jobs}
                  sparklineColor="hsl(var(--chart-3))"
                />
                <MetricCard
                  label="Storage"
                  value={stats?.storage_bytes != null ? formatBytes(stats.storage_bytes) : '--'}
                  subtext="Project files"
                  icon={HardDrive}
                />
              </div>

              {/* Runtime Stats - Row 2 */}
              <div className="grid gap-4 md:grid-cols-3">
                <MetricCard
                  label="Database"
                  value={stats?.database_bytes != null ? formatBytes(stats.database_bytes) : '--'}
                  subtext="Collections data"
                  icon={Database}
                />
                <MetricCard
                  label="Memory"
                  value={isRunning && stats?.memory?.alloc != null ? formatBytes(stats.memory.alloc) : '--'}
                  subtext="Current usage"
                  icon={MemoryStick}
                  sparklineData={stats?.history?.memory}
                  sparklineColor="hsl(var(--chart-2))"
                />
                <MetricCard
                  label="CPU"
                  value={stats?.cpu_percent != null ? `${stats.cpu_percent.toFixed(1)}%` : '--'}
                  subtext="Process usage"
                  icon={Cpu}
                  sparklineData={stats?.history?.cpu}
                  sparklineColor="hsl(var(--chart-4))"
                />
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
                        {project.api_key?.slice(0, 12)}...
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
          </div>
        )}

        {/* Logs Tab */}
        {activeTab === 'logs' && (
          <Card className="flex flex-col gap-0 rounded-t-none py-0 overflow-hidden">
            <LogsViewer
              height={"calc(100vh - 205px)"}
              logs={logs}
              onDownload={runtime.downloadLogs}
            />
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
