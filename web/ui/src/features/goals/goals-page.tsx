import { useState, useMemo } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Plus,
  Target,
  Trash2,
  Edit,
  MoreHorizontal,
  TrendingUp,
  TrendingDown,
  Download,
  Calendar,
  GripVertical,
  RotateCcw,
} from 'lucide-react';
import { toast } from 'sonner';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip as RechartsTooltip,
  ResponsiveContainer,
  CartesianGrid,
} from 'recharts';
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
import {
  format,
  subDays,
  subMonths,
  subQuarters,
  subYears,
  startOfDay,
  endOfDay,
} from 'date-fns';

import { goalsApi } from '@/api';
import { GOAL_TYPES } from '@/lib/constants';
import { queryKeys } from '@/lib/query-keys';
import { formatNumber, calculateTrend } from '@/lib/format';
import { cn } from '@/lib/utils';
import { useAutoSlug, useFormDialog, useDeleteDialog } from '@/hooks';
import type { Goal, CreateGoalRequest, UpdateGoalRequest, GoalType, GoalStats } from '@/types';
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';
import { Field, FieldGroup, FieldLabel, FieldDescription } from '@/components/ui/field';
import { ColorPicker } from '@/components/shared/color-picker';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { EmptyState } from '@/components/shared/empty-state';
import { PageHeader } from '@/components/shared/page-header';
import { Skeleton } from '@/components/ui/skeleton';
import { Calendar as CalendarComponent } from '@/components/ui/calendar';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';

type DatePreset = 'week' | 'month' | 'quarter' | 'year' | 'custom';

const DATE_PRESETS: { value: DatePreset; label: string }[] = [
  { value: 'week', label: 'Week' },
  { value: 'month', label: 'Month' },
  { value: 'quarter', label: 'Quarter' },
  { value: 'year', label: 'Year' },
  { value: 'custom', label: 'Custom' },
];

const GRID_SPAN_OPTIONS = [
  { value: 1, label: '1 col' },
  { value: 2, label: '2 cols' },
  { value: 3, label: '3 cols' },
  { value: 4, label: '4 cols' },
  { value: 5, label: '5 cols' },
];

function getDateRange(preset: DatePreset): { from: Date; to: Date } {
  const now = new Date();
  const to = endOfDay(now);

  switch (preset) {
    case 'week':
      return { from: startOfDay(subDays(now, 7)), to };
    case 'month':
      return { from: startOfDay(subMonths(now, 1)), to };
    case 'quarter':
      return { from: startOfDay(subQuarters(now, 1)), to };
    case 'year':
      return { from: startOfDay(subYears(now, 1)), to };
    default:
      return { from: startOfDay(subDays(now, 7)), to };
  }
}

export function GoalsPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const queryClient = useQueryClient();

  const formDialog = useFormDialog<Goal>();
  const deleteDialog = useDeleteDialog<Goal>();
  const resetDialog = useDeleteDialog<Goal>(); // reusing the same hook for reset confirmation

  // Form state
  const [color, setColor] = useState<string | undefined>();
  const [type, setType] = useState<GoalType>('counter');
  const [description, setDescription] = useState('');
  const [gridSpan, setGridSpan] = useState(1);
  const [showTotal, setShowTotal] = useState(false);
  const { name, slug, setName, setSlug, reset: resetSlug } = useAutoSlug();

  // Date range state
  const [datePreset, setDatePreset] = useState<DatePreset>('week');
  const [customDateRange, setCustomDateRange] = useState<{ from?: Date; to?: Date }>({});
  const [calendarOpen, setCalendarOpen] = useState(false);

  // Export dialog state
  const [exportOpen, setExportOpen] = useState(false);
  const [selectedGoalIds, setSelectedGoalIds] = useState<string[]>([]);

  const dateRange = useMemo(() => {
    if (datePreset === 'custom' && customDateRange.from && customDateRange.to) {
      return customDateRange as { from: Date; to: Date };
    }
    return getDateRange(datePreset);
  }, [datePreset, customDateRange]);

  const { data: goals = [], isLoading } = useQuery({
    queryKey: queryKeys.goals.project(projectId!),
    queryFn: () => goalsApi.listProject(projectId!),
    enabled: !!projectId,
  });

  // Sort goals by order
  const sortedGoals = useMemo(() => {
    return [...goals].sort((a, b) => (a.order || 0) - (b.order || 0));
  }, [goals]);

  const { data: goalStats = [] } = useQuery({
    queryKey: queryKeys.goals.stats(projectId!, goals.map(g => g.id), format(dateRange.from, 'yyyy-MM-dd'), format(dateRange.to, 'yyyy-MM-dd')),
    queryFn: () => goalsApi.getStats({
      goalIds: goals.map(g => g.id),
      startDate: format(dateRange.from, 'yyyy-MM-dd'),
      endDate: format(dateRange.to, 'yyyy-MM-dd'),
    }),
    enabled: !!projectId && goals.length > 0,
    refetchInterval: 10000,
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateGoalRequest) => goalsApi.createProject(projectId!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.goals.project(projectId!) });
      formDialog.close();
      resetForm();
      toast.success('Goal created');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to create goal');
    },
  });

  const updateMutation = useMutation({
    mutationFn: (data: UpdateGoalRequest) =>
      goalsApi.updateProject(projectId!, formDialog.selectedItem!.id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.goals.project(projectId!) });
      formDialog.close();
      resetForm();
      toast.success('Goal updated');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to update goal');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => goalsApi.deleteProject(projectId!, deleteDialog.itemToDelete!.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.goals.project(projectId!) });
      deleteDialog.close();
      toast.success('Goal deleted');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to delete goal');
    },
  });

  const resetMutation = useMutation({
    mutationFn: (goalId: string) => goalsApi.resetProject(projectId!, goalId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.goals.stats(projectId!, goals.map(g => g.id), format(dateRange.from, 'yyyy-MM-dd'), format(dateRange.to, 'yyyy-MM-dd')) });
      resetDialog.close();
      toast.success('Goal statistics reset');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to reset goal');
    },
  });

  // DnD sensors
  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;

    if (over && active.id !== over.id) {
      const oldIndex = sortedGoals.findIndex((g) => g.id === active.id);
      const newIndex = sortedGoals.findIndex((g) => g.id === over.id);

      const newOrder = arrayMove(sortedGoals, oldIndex, newIndex);

      // Update order for each goal
      newOrder.forEach((goal, index) => {
        if (goal.order !== index) {
          goalsApi.updateProject(projectId!, goal.id, { order: index }).then(() => {
            queryClient.invalidateQueries({ queryKey: queryKeys.goals.project(projectId!) });
          });
        }
      });
    }
  };

  const resetForm = () => {
    resetSlug();
    setColor(undefined);
    setType('counter');
    setDescription('');
    setGridSpan(1);
    setShowTotal(false);
  };

  const handleEdit = (goal: Goal) => {
    setName(goal.name);
    setSlug(goal.slug);
    setColor(goal.color);
    setType(goal.type);
    setDescription(goal.description || '');
    setGridSpan(goal.gridSpan || 1);
    setShowTotal(goal.showTotal || false);
    formDialog.openEdit(goal);
  };

  const handleSubmit = () => {
    if (formDialog.mode === 'create') {
      createMutation.mutate({ name, slug, color, type, description });
    } else {
      updateMutation.mutate({ name, color, description, gridSpan, showTotal });
    }
  };

  const handleExportCSV = () => {
    const goalsToExport = selectedGoalIds.length > 0
      ? goals.filter(g => selectedGoalIds.includes(g.id))
      : goals;

    if (goalsToExport.length === 0) {
      toast.error('No goals to export');
      return;
    }

    // Build CSV content
    const rows: string[] = ['Date,Goal,Value'];

    goalsToExport.forEach(goal => {
      const stats = statsMap.get(goal.id);
      if (stats?.dailyStats) {
        stats.dailyStats.forEach(d => {
          rows.push(`${d.date},"${goal.name}",${d.value}`);
        });
      }
    });

    const csvContent = rows.join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `goals_${format(dateRange.from, 'yyyy-MM-dd')}_${format(dateRange.to, 'yyyy-MM-dd')}.csv`;
    link.click();
    URL.revokeObjectURL(url);

    setExportOpen(false);
    setSelectedGoalIds([]);
    toast.success('Data exported');
  };

  const statsMap = new Map<string, GoalStats>();
  goalStats.forEach((s) => statsMap.set(s.goalID, s));

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-48" />
        <div className="grid gap-4 grid-cols-5">
          {[1, 2, 3, 4].map((i) => (
            <Skeleton key={i} className="h-48" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Goals"
        description="Track metrics and counters for this project"
        action={
          <div className="flex items-center gap-2">
            {/* Date Range Selector */}
            <div className="flex items-center gap-1 bg-muted rounded-lg p-1">
              {DATE_PRESETS.filter(p => p.value !== 'custom').map((preset) => (
                <Button
                  key={preset.value}
                  variant={datePreset === preset.value ? 'default' : 'ghost'}
                  size="sm"
                  onClick={() => setDatePreset(preset.value)}
                  className="h-7 px-2 text-xs"
                >
                  {preset.label}
                </Button>
              ))}
              <Popover open={calendarOpen} onOpenChange={setCalendarOpen}>
                <PopoverTrigger asChild>
                  <Button
                    variant={datePreset === 'custom' ? 'default' : 'ghost'}
                    size="sm"
                    className="h-7 px-2 text-xs gap-1"
                  >
                    <Calendar className="size-3" />
                    {datePreset === 'custom' && customDateRange.from && customDateRange.to
                      ? `${format(customDateRange.from, 'MMM d')} - ${format(customDateRange.to, 'MMM d')}`
                      : 'Custom'}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="end">
                  <CalendarComponent
                    mode="range"
                    selected={{
                      from: customDateRange.from,
                      to: customDateRange.to,
                    }}
                    onSelect={(range) => {
                      setCustomDateRange({ from: range?.from, to: range?.to });
                      if (range?.from && range?.to) {
                        setDatePreset('custom');
                        setCalendarOpen(false);
                      }
                    }}
                    numberOfMonths={2}
                  />
                </PopoverContent>
              </Popover>
            </div>

            {/* Export Button */}
            <Button variant="outline" size="sm" onClick={() => setExportOpen(true)}>
              <Download className="mr-1.5 size-4" />
              Export
            </Button>

            <Button onClick={() => { resetForm(); formDialog.open(); }}>
              <Plus className="mr-2 size-4" />
              New Goal
            </Button>
          </div>
        }
      />

      {sortedGoals.length === 0 ? (
        <EmptyState
          icon={<Target className="size-12" />}
          title="No goals yet"
          description="Create goals to track metrics in your service"
          action={
            <Button onClick={() => { resetForm(); formDialog.open(); }}>
              <Plus className="mr-2 size-4" />
              Create Goal
            </Button>
          }
        />
      ) : (
        <DndContext
          sensors={sensors}
          collisionDetection={closestCenter}
          onDragEnd={handleDragEnd}
        >
          <SortableContext items={sortedGoals.map(g => g.id)} strategy={rectSortingStrategy}>
            <div className="grid gap-4 grid-cols-5">
              {sortedGoals.map((goal) => (
                <SortableGoalCard
                  key={goal.id}
                  goal={goal}
                  stats={statsMap.get(goal.id)}
                  dateRange={dateRange}
                  onEdit={() => handleEdit(goal)}
                  onDelete={() => deleteDialog.open(goal)}
                  onReset={() => resetDialog.open(goal)}
                />
              ))}
            </div>
          </SortableContext>
        </DndContext>
      )}

      {/* Create/Edit Dialog */}
      <Dialog open={formDialog.isOpen} onOpenChange={(open) => !open && formDialog.close()}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{formDialog.mode === 'create' ? 'Create Goal' : 'Edit Goal'}</DialogTitle>
            {formDialog.mode === 'create' && (
              <DialogDescription>
                Define a new metric to track in your service
              </DialogDescription>
            )}
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel>Name</FieldLabel>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Page Views"
              />
            </Field>
            <Field>
              <FieldLabel>Slug</FieldLabel>
              <Input
                value={slug}
                onChange={(e) => setSlug(e.target.value)}
                placeholder="page-views"
                disabled={formDialog.mode === 'edit'}
              />
              {formDialog.mode === 'create' && (
                <FieldDescription>
                  Use in runtime: goals.increment("{slug}")
                </FieldDescription>
              )}
            </Field>
            <Field>
              <FieldLabel>Type</FieldLabel>
              {formDialog.mode === 'create' ? (
                <Select value={type} onValueChange={(v) => setType(v as GoalType)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {GOAL_TYPES.map((t) => (
                      <SelectItem key={t.value} value={t.value}>
                        {t.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              ) : (
                <Input
                  value={type === 'counter' ? 'Counter' : 'Daily Counter'}
                  disabled
                />
              )}
            </Field>
            <Field>
              <FieldLabel>Color</FieldLabel>
              <ColorPicker value={color} onChange={setColor} />
            </Field>
            <Field>
              <FieldLabel>Grid Span</FieldLabel>
              <Select value={String(gridSpan)} onValueChange={(v) => setGridSpan(Number(v))}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {GRID_SPAN_OPTIONS.map((opt) => (
                    <SelectItem key={opt.value} value={String(opt.value)}>
                      {opt.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FieldDescription>
                Number of columns this card occupies (grid has 5 columns)
              </FieldDescription>
            </Field>
            {(type === 'daily_counter' || formDialog.selectedItem?.type === 'daily_counter') && (
              <Field>
                <div className="flex items-center justify-between">
                  <FieldLabel>Show All-Time Total</FieldLabel>
                  <Switch
                    checked={showTotal}
                    onCheckedChange={setShowTotal}
                  />
                </div>
                <FieldDescription>
                  Display the cumulative total alongside daily values
                </FieldDescription>
              </Field>
            )}
            <Field>
              <FieldLabel>Description</FieldLabel>
              <Textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Optional description..."
                rows={3}
              />
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => formDialog.close()}>
              Cancel
            </Button>
            <LoadingButton
              onClick={handleSubmit}
              disabled={!name || !slug}
              loading={createMutation.isPending || updateMutation.isPending}
            >
              {formDialog.mode === 'create' ? 'Create' : 'Save'}
            </LoadingButton>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Export Dialog */}
      <Dialog open={exportOpen} onOpenChange={setExportOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Export Goals Data</DialogTitle>
            <DialogDescription>
              Select goals to export. Data will be exported for the selected date range ({format(dateRange.from, 'MMM d')} - {format(dateRange.to, 'MMM d, yyyy')}).
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-3 py-4">
            <div className="flex items-center space-x-2">
              <Checkbox
                id="select-all"
                checked={selectedGoalIds.length === goals.length}
                onCheckedChange={(checked) => {
                  if (checked) {
                    setSelectedGoalIds(goals.map(g => g.id));
                  } else {
                    setSelectedGoalIds([]);
                  }
                }}
              />
              <Label htmlFor="select-all" className="font-medium">Select All</Label>
            </div>
            <div className="border-t pt-3 space-y-2">
              {goals.map((goal) => (
                <div key={goal.id} className="flex items-center space-x-2">
                  <Checkbox
                    id={goal.id}
                    checked={selectedGoalIds.includes(goal.id)}
                    onCheckedChange={(checked) => {
                      if (checked) {
                        setSelectedGoalIds([...selectedGoalIds, goal.id]);
                      } else {
                        setSelectedGoalIds(selectedGoalIds.filter(id => id !== goal.id));
                      }
                    }}
                  />
                  <Label htmlFor={goal.id} className="flex items-center gap-2">
                    <span
                      className="size-2.5 rounded-full"
                      style={{ backgroundColor: goal.color || '#6b7280' }}
                    />
                    {goal.name}
                  </Label>
                </div>
              ))}
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setExportOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleExportCSV}>
              <Download className="mr-1.5 size-4" />
              Export CSV
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={deleteDialog.isOpen}
        onOpenChange={(open) => !open && deleteDialog.close()}
        title="Delete Goal"
        description={`Are you sure you want to delete "${deleteDialog.itemToDelete?.name}"? All statistics will be lost.`}
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={() => deleteMutation.mutate()}
        isLoading={deleteMutation.isPending}
      />

      <ConfirmDialog
        open={resetDialog.isOpen}
        onOpenChange={(open) => !open && resetDialog.close()}
        title="Reset Goal Statistics"
        description={`Are you sure you want to reset all statistics for "${resetDialog.itemToDelete?.name}"? This will delete all recorded data and cannot be undone.`}
        confirmLabel="Reset"
        variant="destructive"
        onConfirm={() => resetDialog.itemToDelete && resetMutation.mutate(resetDialog.itemToDelete.id)}
        isLoading={resetMutation.isPending}
      />
    </div>
  );
}

function SortableGoalCard({
  goal,
  stats,
  dateRange,
  onEdit,
  onDelete,
  onReset,
}: {
  goal: Goal;
  stats?: GoalStats;
  dateRange: { from: Date; to: Date };
  onEdit: () => void;
  onDelete: () => void;
  onReset: () => void;
}) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: goal.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    gridColumn: `span ${goal.gridSpan || 1}`,
    opacity: isDragging ? 0.5 : 1,
  };

  const chartData = stats?.dailyStats?.map((d) => ({
    date: new Date(d.date).toLocaleDateString('en', { month: 'short', day: 'numeric' }),
    value: d.value,
  })) || [];

  const trend = stats?.dailyStats && stats.dailyStats.length >= 2
    ? calculateTrend(
        stats.dailyStats.slice(-7).map(d => d.value),
        stats.dailyStats.slice(-14, -7).map(d => d.value)
      )
    : null;

  const isDailyCounter = goal.type === 'daily_counter';

  return (
    <Card ref={setNodeRef} style={style} className={cn("group", isDragging && 'z-50')}>
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div
              className="size-10 rounded-lg flex items-center justify-center"
              style={{ backgroundColor: goal.color || '#6b7280' }}
            >
              <Target className="size-5 text-white" />
            </div>
            <div>
              <CardTitle className="text-lg">{goal.name}</CardTitle>
              <CardDescription className="font-mono text-xs">
                {goal.slug}
              </CardDescription>
            </div>
          </div>
          <div className="flex items-center gap-1">
            <button
              className="cursor-grab active:cursor-grabbing text-muted-foreground hover:text-foreground opacity-0 group-hover:opacity-100 transition-opacity p-1"
              {...attributes}
              {...listeners}
            >
              <GripVertical className="size-4" />
            </button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon" className="size-8">
                  <MoreHorizontal className="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={onEdit}>
                  <Edit className="mr-2 size-4" />
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem onClick={onReset}>
                  <RotateCcw className="mr-2 size-4" />
                  Reset Stats
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem className="text-destructive" onClick={onDelete}>
                  <Trash2 className="mr-2 size-4" />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {chartData.length > 0 && (
          <div className="h-32">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={chartData}>
                <defs>
                  <linearGradient id={`gradient-${goal.id}`} x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor={goal.color || '#6b7280'} stopOpacity={0.3} />
                    <stop offset="100%" stopColor={goal.color || '#6b7280'} stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                <XAxis
                  dataKey="date"
                  tick={{ fontSize: 10 }}
                  tickLine={false}
                  axisLine={false}
                  className="text-muted-foreground"
                />
                <YAxis
                  tick={{ fontSize: 10 }}
                  tickLine={false}
                  axisLine={false}
                  width={30}
                  className="text-muted-foreground"
                />
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

        <div className="flex-1">
          <div className="flex items-baseline gap-2">
            {isDailyCounter && goal.showTotal && stats?.totalValue != null ? (
              <>
                <p className="text-3xl font-bold">{formatNumber(stats.totalValue)}</p>
                <p className="text-lg text-muted-foreground">
                  / {formatNumber(stats?.value ?? 0)} today
                </p>
              </>
            ) : (
              <p className="text-3xl font-bold">{formatNumber(stats?.value ?? 0)}</p>
            )}
          </div>
          <div className="flex items-center gap-2 mt-1">
            {trend !== null && (
              <div className={cn(
                "flex items-center gap-1 text-xs font-medium",
                trend > 0 ? "text-green-500" : trend < 0 ? "text-red-500" : "text-muted-foreground"
              )}>
                {trend > 0 ? <TrendingUp className="size-3" /> : trend < 0 ? <TrendingDown className="size-3" /> : null}
                {trend > 0 ? '+' : ''}{trend.toFixed(0)}%
              </div>
            )}
            <span className="text-xs text-muted-foreground">
              {format(dateRange.from, 'MMM d')} - {format(dateRange.to, 'MMM d')}
            </span>
          </div>
        </div>

        {goal.description && (
          <p className="text-sm text-muted-foreground">{goal.description}</p>
        )}

        <div className="rounded-md bg-muted/50 p-2">
          <code className="text-xs text-muted-foreground">
            $goals.increment("{goal.slug}")
          </code>
        </div>
      </CardContent>
    </Card>
  );
}
