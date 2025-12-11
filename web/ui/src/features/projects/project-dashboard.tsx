import { useState, useMemo, useEffect } from 'react';
import { useParams, Link, useLocation } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Play,
  Square,
  RotateCcw,
  Code,
  Target,
  ExternalLink,
  Activity,
  AlertTriangle,
  ChevronRight,
  ChevronDown,
  Bug,
  Tag,
  Plus,
  LayoutGrid,
  BarChart3,
  Minus, ScrollText,
} from 'lucide-react';
import { toast } from 'sonner';
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core';
import type { DragEndEvent } from '@dnd-kit/core';
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  rectSortingStrategy,
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';

import { projectsApi, pipelineApi, goalsApi, widgetsApi, actionsApi } from '@/api';
import { config } from '@/lib/config';
import { queryKeys } from '@/lib/query-keys';
import { cn } from '@/lib/utils';
import { useProjectRuntime, useTitle, useWebSocket } from '@/hooks';
import type { LogEntry, GoalStats, WidgetVariant, WidgetType, CreateWidgetRequest, UpdateWidgetRequest, Widget, Goal, RuntimeStats, ActionRuntimeState, ActionState } from '@/types';
import { ActionsDropdown } from '@/components/shared/actions-dropdown';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { EditorTabs, EditorTab } from '@/components/ui/editor-tabs';
import { StatusBadge } from '@/components/shared/status-badge';
import { Skeleton } from '@/components/ui/skeleton';
import { LogsViewer } from '@/components/shared/logs-viewer';
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

const WIDGET_TYPES: { value: WidgetType; label: string; description: string }[] = [
  { value: 'goal', label: 'Goal', description: 'Display goal metrics' },
  { value: 'memory', label: 'Memory', description: 'Current memory usage' },
  { value: 'requests', label: 'Requests', description: 'Total request count' },
  { value: 'cpu', label: 'CPU', description: 'CPU utilization' },
  { value: 'storage', label: 'Storage', description: 'Storage usage' },
  { value: 'database', label: 'Database', description: 'Database size' },
  { value: 'uptime', label: 'Uptime', description: 'Service uptime' },
  { value: 'jobs', label: 'Jobs', description: 'Scheduled jobs count' },
];

export function ProjectDashboard() {
  const { projectId } = useParams<{ projectId: string }>();
  const location = useLocation();
  const queryClient = useQueryClient();

  const initialTab = (location.state as { tab?: string } | null)?.tab as OverviewTab || 'instance';
  const [activeTab, setActiveTab] = useState<OverviewTab>(initialTab);

  const { data: project, isLoading: projectLoading } = useQuery({
    queryKey: queryKeys.projects.detail(projectId!),
    queryFn: () => projectsApi.get(projectId!),
    enabled: !!projectId,
  });

  useTitle(project?.name);

  const isRunning = project?.status === 'running';
  const isDebugMode = isRunning && project?.runningSource?.startsWith('debug:');
  const publicUrl = project ? `${config.apiURL}/r/${project.slug}` : '';

  // Confirm dialogs for restart/stop in release mode
  const [stopConfirmOpen, setStopConfirmOpen] = useState(false);
  const [restartConfirmOpen, setRestartConfirmOpen] = useState(false);

  // Runtime hook - handles start/stop/restart, logs, monitor
  // Now uses WebSocket for real-time updates instead of polling
  const runtime = useProjectRuntime({
    projectId: projectId!,
    projectSlug: project?.slug,
    enabled: !!projectId,
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
    // Goals are updated via WebSocket every 30 seconds
    staleTime: 30000,
  });

  const { data: widgets = [] } = useQuery({
    queryKey: queryKeys.widgets.all(projectId!),
    queryFn: () => widgetsApi.list(projectId!),
    enabled: !!projectId,
  });

  // Actions
  const { data: actions = [] } = useQuery({
    queryKey: queryKeys.actions.all(projectId!),
    queryFn: () => actionsApi.list(projectId!),
    enabled: !!projectId,
  });

  const [actionStates, setActionStates] = useState<Map<string, ActionState>>(new Map());

  // WebSocket for action states (only when running)
  useWebSocket({
    projectId,
    enabled: !!projectId && isRunning,
    onActions: (data: ActionRuntimeState[]) => {
      const newStates = new Map<string, ActionState>();
      data.forEach((item) => {
        newStates.set(item.slug, item.state);
      });
      setActionStates(newStates);
    },
  });

  // Widget Dialog state
  const [widgetDialogOpen, setWidgetDialogOpen] = useState(false);
  const [editingWidget, setEditingWidget] = useState<Widget | null>(null);
  const [selectedWidgetType, setSelectedWidgetType] = useState<WidgetType>('goal');
  const [selectedGoalId, setSelectedGoalId] = useState<string>('');
  const [selectedVariant, setSelectedVariant] = useState<WidgetVariant>('mini');
  const [selectedGridSpan, setSelectedGridSpan] = useState(1);

  // Sort widgets by order
  const sortedWidgets = useMemo(() => {
    return [...widgets].sort((a, b) => (a.order || 0) - (b.order || 0));
  }, [widgets]);

  // DnD sensors
  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  const handleWidgetDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;

    if (over && active.id !== over.id) {
      const oldIndex = sortedWidgets.findIndex((w) => w.id === active.id);
      const newIndex = sortedWidgets.findIndex((w) => w.id === over.id);

      const newOrder = arrayMove(sortedWidgets, oldIndex, newIndex);

      // Update order for reordered widgets
      widgetsApi.reorder(projectId!, { widgetIds: newOrder.map(w => w.id) }).then(() => {
        queryClient.invalidateQueries({ queryKey: queryKeys.widgets.all(projectId!) });
      });
    }
  };

  const logs: LogEntry[] = Array.isArray(runtime.logs) ? runtime.logs : [];
  const stats = runtime.monitor;

  // Fetch logs when Logs tab is opened
  useEffect(() => {
    if (activeTab === 'logs') {
      runtime.refetchLogs();
    }
  }, [activeTab, runtime.refetchLogs]);

  const createWidgetMutation = useMutation({
    mutationFn: (data: CreateWidgetRequest) => widgetsApi.create(projectId!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.widgets.all(projectId!) });
      closeWidgetDialog();
      toast.success('Widget added');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to add widget');
    },
  });

  const updateWidgetMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateWidgetRequest }) =>
      widgetsApi.update(projectId!, id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.widgets.all(projectId!) });
      closeWidgetDialog();
      toast.success('Widget updated');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to update widget');
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

  const openAddWidget = () => {
    setEditingWidget(null);
    setSelectedWidgetType('goal');
    setSelectedGoalId('');
    setSelectedVariant('mini');
    setSelectedGridSpan(1);
    setWidgetDialogOpen(true);
  };

  const openEditWidget = (widget: Widget) => {
    setEditingWidget(widget);
    setSelectedWidgetType(widget.type || 'goal');
    setSelectedGoalId(widget.goalId || '');
    setSelectedVariant(widget.variant);
    setSelectedGridSpan(widget.gridSpan || 1);
    setWidgetDialogOpen(true);
  };

  const closeWidgetDialog = () => {
    setWidgetDialogOpen(false);
    setEditingWidget(null);
    setSelectedWidgetType('goal');
    setSelectedGoalId('');
    setSelectedVariant('mini');
    setSelectedGridSpan(1);
  };

  const handleSaveWidget = () => {
    if (editingWidget) {
      // Update existing widget
      updateWidgetMutation.mutate({
        id: editingWidget.id,
        data: {
          variant: selectedVariant,
          gridSpan: selectedGridSpan,
        },
      });
    } else {
      // Create new widget
      if (selectedWidgetType === 'goal' && !selectedGoalId) {
        toast.error('Please select a goal');
        return;
      }
      createWidgetMutation.mutate({
        type: selectedWidgetType,
        goalId: selectedWidgetType === 'goal' ? selectedGoalId : undefined,
        variant: selectedVariant,
        gridSpan: selectedGridSpan,
      });
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
              <a
                href={publicUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="text-muted-foreground text-sm hover:text-primary hover:underline flex items-center gap-1"
              >
                <code>{project.slug}</code>
                <ExternalLink className="size-3" />
              </a>
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
                onClick={() => {
                  if (isDebugMode) {
                    runtime.restart();
                  } else {
                    setRestartConfirmOpen(true);
                  }
                }}
                disabled={runtime.isPending}
                className="gap-2"
              >
                <RotateCcw className={cn("size-4", runtime.isRestarting && "animate-spin")} />
                Restart
              </Button>
              <Button
                variant="destructive"
                size="sm"
                onClick={() => {
                  if (isDebugMode) {
                    runtime.stop();
                  } else {
                    setStopConfirmOpen(true);
                  }
                }}
                disabled={runtime.isPending}
                className="gap-2"
              >
                <Square className="size-4" />
                Stop
              </Button>
              {actions.length > 0 && project?.slug && (
                <ActionsDropdown
                  projectSlug={project.slug}
                  actions={actions}
                  actionStates={actionStates}
                />
              )}
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
            icon={<ScrollText className="size-4" />}
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
              {/* Widgets Section */}
              <div>
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-lg font-semibold">Widgets</h2>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={openAddWidget}
                  >
                    <Plus className="mr-1.5 size-4" />
                    Add Widget
                  </Button>
                </div>
                {sortedWidgets.length > 0 ? (
                  <DndContext
                    sensors={sensors}
                    collisionDetection={closestCenter}
                    onDragEnd={handleWidgetDragEnd}
                  >
                    <SortableContext items={sortedWidgets.map(w => w.id)} strategy={rectSortingStrategy}>
                      <div className="grid gap-4 grid-cols-5">
                        {sortedWidgets.map((widget) => {
                          const goal = widget.type === 'goal' && widget.goalId
                            ? goals.find(g => g.id === widget.goalId)
                            : undefined;
                          const stat = goal ? goalStatsMap.get(goal.id) : undefined;

                          // For goal widgets, skip if goal not found
                          if (widget.type === 'goal' && !goal) return null;

                          return (
                            <SortableWidgetCard
                              key={widget.id}
                              widget={widget}
                              goal={goal}
                              stats={stat}
                              runtimeStats={stats}
                              isRunning={isRunning}
                              onEdit={() => openEditWidget(widget)}
                              onDelete={() => deleteWidgetMutation.mutate(widget.id)}
                            />
                          );
                        })}
                      </div>
                    </SortableContext>
                  </DndContext>
                ) : (
                  <Card>
                    <CardContent className="flex flex-col items-center justify-center py-8">
                      <LayoutGrid className="size-10 text-muted-foreground mb-3" />
                      <p className="text-muted-foreground text-center text-sm">
                        No widgets added. Add widgets to display metrics.
                      </p>
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

      {/* Widget Dialog (Add/Edit) */}
      <Dialog open={widgetDialogOpen} onOpenChange={(open) => !open && closeWidgetDialog()}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{editingWidget ? 'Edit Widget' : 'Add Widget'}</DialogTitle>
            <DialogDescription>
              {editingWidget
                ? 'Update the widget settings.'
                : 'Choose a widget type and display settings.'}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            {/* Widget Type Selection (only for new widgets) */}
            {!editingWidget && (
              <div className="space-y-2">
                <Label>Widget Type</Label>
                <Select value={selectedWidgetType} onValueChange={(v) => setSelectedWidgetType(v as WidgetType)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {WIDGET_TYPES.map((type) => (
                      <SelectItem key={type.value} value={type.value}>
                        <div className="flex flex-col">
                          <span>{type.label}</span>
                          <span className="text-xs text-muted-foreground">{type.description}</span>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}

            {/* Goal Selection (only for goal type) */}
            {selectedWidgetType === 'goal' && !editingWidget && (
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
            )}

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

            <div className="space-y-2">
              <Label>Grid Span (columns)</Label>
              <Select value={String(selectedGridSpan)} onValueChange={(v) => setSelectedGridSpan(Number(v))}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {[1, 2, 3, 4, 5].map((span) => (
                    <SelectItem key={span} value={String(span)}>
                      {span} {span === 1 ? 'column' : 'columns'}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={closeWidgetDialog}>
              Cancel
            </Button>
            <Button
              onClick={handleSaveWidget}
              disabled={
                (selectedWidgetType === 'goal' && !editingWidget && !selectedGoalId) ||
                createWidgetMutation.isPending ||
                updateWidgetMutation.isPending
              }
            >
              {editingWidget ? 'Save' : 'Add Widget'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Confirm Stop Dialog (for release mode) */}
      <ConfirmDialog
        open={stopConfirmOpen}
        onOpenChange={setStopConfirmOpen}
        title="Stop Service?"
        description="Service is running in production mode. Stopping will make it unavailable for users."
        confirmLabel="Stop"
        variant="destructive"
        onConfirm={() => {
          runtime.stop();
          setStopConfirmOpen(false);
        }}
        isLoading={runtime.isStopping}
      />

      {/* Confirm Restart Dialog (for release mode) */}
      <ConfirmDialog
        open={restartConfirmOpen}
        onOpenChange={setRestartConfirmOpen}
        title="Restart Service?"
        description="Service is running in production mode. Restarting will cause brief downtime."
        confirmLabel="Restart"
        onConfirm={() => {
          runtime.restart();
          setRestartConfirmOpen(false);
        }}
        isLoading={runtime.isRestarting}
      />
    </div>
  );
}

function SortableWidgetCard({
  widget,
  goal,
  stats,
  runtimeStats,
  isRunning,
  onEdit,
  onDelete,
}: {
  widget: Widget;
  goal?: Goal;
  stats?: GoalStats;
  runtimeStats?: RuntimeStats;
  isRunning?: boolean;
  onEdit: () => void;
  onDelete: () => void;
}) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: widget.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  return (
    <WidgetCard
      ref={setNodeRef}
      widget={widget}
      goal={goal}
      stats={stats}
      runtimeStats={runtimeStats}
      isRunning={isRunning}
      onEdit={onEdit}
      onDelete={onDelete}
      dragHandleProps={{ ...attributes, ...listeners }}
      isDragging={isDragging}
      style={style}
    />
  );
}
