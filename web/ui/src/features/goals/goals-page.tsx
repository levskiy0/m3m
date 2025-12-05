import { useState } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Target, Trash2, Edit, MoreHorizontal, TrendingUp, TrendingDown } from 'lucide-react';
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

import { goalsApi } from '@/api';
import { GOAL_TYPES } from '@/lib/constants';
import type { Goal, CreateGoalRequest, UpdateGoalRequest, GoalType, GoalStats } from '@/types';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
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
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Field, FieldGroup, FieldLabel, FieldDescription } from '@/components/ui/field';
import { ColorPicker } from '@/components/shared/color-picker';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { EmptyState } from '@/components/shared/empty-state';
import { Sparkline } from '@/components/shared/sparkline';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';

function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '');
}

function formatNumber(value: number): string {
  if (value >= 1000000) {
    return `${(value / 1000000).toFixed(1)}M`;
  }
  if (value >= 1000) {
    return `${(value / 1000).toFixed(1)}K`;
  }
  return value.toLocaleString();
}

export function GoalsPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const queryClient = useQueryClient();

  const [createOpen, setCreateOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [selectedGoal, setSelectedGoal] = useState<Goal | null>(null);

  // Form state
  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');
  const [slugEdited, setSlugEdited] = useState(false);
  const [color, setColor] = useState<string | undefined>();
  const [type, setType] = useState<GoalType>('counter');
  const [description, setDescription] = useState('');

  const { data: goals = [], isLoading } = useQuery({
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

  const createMutation = useMutation({
    mutationFn: (data: CreateGoalRequest) =>
      goalsApi.createProject(projectId!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-goals', projectId] });
      setCreateOpen(false);
      resetForm();
      toast.success('Goal created');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to create goal');
    },
  });

  const updateMutation = useMutation({
    mutationFn: (data: UpdateGoalRequest) =>
      goalsApi.updateProject(projectId!, selectedGoal!.id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-goals', projectId] });
      setEditOpen(false);
      setSelectedGoal(null);
      resetForm();
      toast.success('Goal updated');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to update goal');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => goalsApi.deleteProject(projectId!, selectedGoal!.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-goals', projectId] });
      setDeleteOpen(false);
      setSelectedGoal(null);
      toast.success('Goal deleted');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to delete goal');
    },
  });

  const resetForm = () => {
    setName('');
    setSlug('');
    setSlugEdited(false);
    setColor(undefined);
    setType('counter');
    setDescription('');
  };

  const handleNameChange = (value: string) => {
    setName(value);
    if (!slugEdited) {
      setSlug(slugify(value));
    }
  };

  const handleCreate = () => {
    createMutation.mutate({ name, slug, color, type, description });
  };

  const handleEdit = (goal: Goal) => {
    setSelectedGoal(goal);
    setName(goal.name);
    setSlug(goal.slug);
    setColor(goal.color);
    setType(goal.type);
    setDescription(goal.description || '');
    setEditOpen(true);
  };

  const handleUpdate = () => {
    updateMutation.mutate({ name, color, description });
  };

  // Create stats map for quick lookup
  const statsMap = new Map<string, GoalStats>();
  goalStats.forEach((s) => statsMap.set(s.goalID, s));

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-48" />
        <div className="grid gap-4 md:grid-cols-2">
          {[1, 2, 3, 4].map((i) => (
            <Skeleton key={i} className="h-48" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Goals</h1>
          <p className="text-muted-foreground">
            Track metrics and counters for this project
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-2 size-4" />
          New Goal
        </Button>
      </div>

      {goals.length === 0 ? (
        <EmptyState
          icon={<Target className="size-12" />}
          title="No goals yet"
          description="Create goals to track metrics in your service"
          action={
            <Button onClick={() => setCreateOpen(true)}>
              <Plus className="mr-2 size-4" />
              Create Goal
            </Button>
          }
        />
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {goals.map((goal) => {
            const stats = statsMap.get(goal.id);
            return (
              <GoalCard
                key={goal.id}
                goal={goal}
                stats={stats}
                onEdit={() => handleEdit(goal)}
                onDelete={() => {
                  setSelectedGoal(goal);
                  setDeleteOpen(true);
                }}
              />
            );
          })}
        </div>
      )}

      {/* Create Dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Goal</DialogTitle>
            <DialogDescription>
              Define a new metric to track in your service
            </DialogDescription>
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel>Name</FieldLabel>
              <Input
                value={name}
                onChange={(e) => handleNameChange(e.target.value)}
                placeholder="Page Views"
              />
            </Field>
            <Field>
              <FieldLabel>Slug</FieldLabel>
              <Input
                value={slug}
                onChange={(e) => {
                  setSlug(slugify(e.target.value));
                  setSlugEdited(true);
                }}
                placeholder="page-views"
              />
              <FieldDescription>
                Use in runtime: goals.increment("{slug}")
              </FieldDescription>
            </Field>
            <Field>
              <FieldLabel>Type</FieldLabel>
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
            </Field>
            <Field>
              <FieldLabel>Color</FieldLabel>
              <ColorPicker value={color} onChange={setColor} />
            </Field>
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
            <Button variant="outline" onClick={() => setCreateOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleCreate}
              disabled={!name || !slug || createMutation.isPending}
            >
              {createMutation.isPending ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Dialog */}
      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Goal</DialogTitle>
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel>Name</FieldLabel>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </Field>
            <Field>
              <FieldLabel>Slug</FieldLabel>
              <Input value={slug} disabled />
              <FieldDescription>Slug cannot be changed</FieldDescription>
            </Field>
            <Field>
              <FieldLabel>Type</FieldLabel>
              <Input
                value={type === 'counter' ? 'Counter' : 'Daily Counter'}
                disabled
              />
            </Field>
            <Field>
              <FieldLabel>Color</FieldLabel>
              <ColorPicker value={color} onChange={setColor} />
            </Field>
            <Field>
              <FieldLabel>Description</FieldLabel>
              <Textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={3}
              />
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleUpdate}
              disabled={updateMutation.isPending}
            >
              {updateMutation.isPending ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirm */}
      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title="Delete Goal"
        description={`Are you sure you want to delete "${selectedGoal?.name}"? All statistics will be lost.`}
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={() => deleteMutation.mutate()}
        isLoading={deleteMutation.isPending}
      />
    </div>
  );
}

// Enhanced Goal Card with chart
function GoalCard({
  goal,
  stats,
  onEdit,
  onDelete,
}: {
  goal: Goal;
  stats?: GoalStats;
  onEdit: () => void;
  onDelete: () => void;
}) {
  const chartData = stats?.dailyStats?.slice(-14).map((d) => ({
    date: new Date(d.date).toLocaleDateString('en', { month: 'short', day: 'numeric' }),
    value: d.value,
  })) || [];

  // Calculate trend
  const trend = stats?.dailyStats && stats.dailyStats.length >= 2
    ? (() => {
        const recent = stats.dailyStats.slice(-7);
        const older = stats.dailyStats.slice(-14, -7);
        if (recent.length === 0 || older.length === 0) return null;
        const recentSum = recent.reduce((a, b) => a + b.value, 0);
        const olderSum = older.reduce((a, b) => a + b.value, 0);
        if (olderSum === 0) return null;
        return ((recentSum - olderSum) / olderSum) * 100;
      })()
    : null;

  return (
    <Card>
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
              <DropdownMenuSeparator />
              <DropdownMenuItem
                className="text-destructive"
                onClick={onDelete}
              >
                <Trash2 className="mr-2 size-4" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Value, sparkline and trend */}
        <div className="flex items-end justify-between gap-4">
          <div className="flex-1">
            <p className="text-3xl font-bold">
              {formatNumber(stats?.value ?? 0)}
            </p>
            <div className="flex items-center gap-2 mt-1">
              <Badge variant="secondary">
                {goal.type === 'counter' ? 'Total Counter' : 'Daily Counter'}
              </Badge>
              {trend !== null && (
                <div className={cn(
                  "flex items-center gap-1 text-xs font-medium",
                  trend > 0 ? "text-green-500" : trend < 0 ? "text-red-500" : "text-muted-foreground"
                )}>
                  {trend > 0 ? (
                    <TrendingUp className="size-3" />
                  ) : trend < 0 ? (
                    <TrendingDown className="size-3" />
                  ) : null}
                  {trend > 0 ? '+' : ''}{trend.toFixed(0)}%
                </div>
              )}
            </div>
          </div>
          {/* 7-day sparkline for quick overview */}
          {goal.type === 'daily_counter' && stats?.dailyStats && stats.dailyStats.length > 1 && (
            <Sparkline
              data={stats.dailyStats.slice(-7).map(d => d.value)}
              width={80}
              height={36}
              color={goal.color || '#6b7280'}
              strokeWidth={2}
              fillOpacity={0.2}
            />
          )}
        </div>

        {/* Chart */}
        {chartData.length > 0 && (
          <div className="h-32">
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

        {goal.description && (
          <p className="text-sm text-muted-foreground">
            {goal.description}
          </p>
        )}

        {/* Usage hint */}
        <div className="rounded-md bg-muted/50 p-2">
          <code className="text-xs text-muted-foreground">
            goals.increment("{goal.slug}")
          </code>
        </div>
      </CardContent>
    </Card>
  );
}
