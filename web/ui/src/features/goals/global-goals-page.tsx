import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Target, Trash2, Edit, MoreHorizontal } from 'lucide-react';
import { toast } from 'sonner';

import { goalsApi, projectsApi } from '@/api';
import { useAuth } from '@/providers/auth-provider';
import { useTitle } from '@/hooks';
import { GOAL_TYPES } from '@/features/goals/constants';
import type { Goal, CreateGoalRequest, UpdateGoalRequest, GoalType } from '@/types';
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
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Checkbox } from '@/components/ui/checkbox';
import { Field, FieldGroup, FieldLabel, FieldDescription } from '@/components/ui/field';
import { ColorPicker } from '@/components/shared/color-picker';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { EmptyState } from '@/components/shared/empty-state';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';

function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '');
}

export function GlobalGoalsPage() {
  useTitle('Goals');
  const { user } = useAuth();
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
  const [projectAccess, setProjectAccess] = useState<string[]>([]);

  const isAdmin = user?.permissions?.manageUsers || user?.isRoot;

  const { data: goals = [], isLoading: goalsLoading } = useQuery({
    queryKey: ['global-goals'],
    queryFn: goalsApi.listGlobal,
  });

  const { data: projects = [] } = useQuery({
    queryKey: ['projects'],
    queryFn: projectsApi.list,
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateGoalRequest) => goalsApi.createGlobal(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['global-goals'] });
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
      goalsApi.update(selectedGoal!.id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['global-goals'] });
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
    mutationFn: () => goalsApi.delete(selectedGoal!.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['global-goals'] });
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
    setProjectAccess([]);
  };

  const handleNameChange = (value: string) => {
    setName(value);
    if (!slugEdited) {
      setSlug(slugify(value));
    }
  };

  const handleCreate = () => {
    createMutation.mutate({ name, slug, color, type, description, projectAccess });
  };

  const handleEdit = (goal: Goal) => {
    setSelectedGoal(goal);
    setName(goal.name);
    setSlug(goal.slug);
    setColor(goal.color);
    setType(goal.type);
    setDescription(goal.description || '');
    setProjectAccess(goal.projectAccess || []);
    setEditOpen(true);
  };

  const handleUpdate = () => {
    updateMutation.mutate({ name, color, description, projectAccess });
  };

  const toggleProjectAccess = (projectId: string) => {
    setProjectAccess((prev) =>
      prev.includes(projectId)
        ? prev.filter((id) => id !== projectId)
        : [...prev, projectId]
    );
  };

  if (goalsLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-48" />
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-32" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Global Goals</h1>
          <p className="text-muted-foreground">
            System-wide metrics available across projects
          </p>
        </div>
        {isAdmin && (
          <Button onClick={() => setCreateOpen(true)}>
            <Plus className="mr-2 size-4" />
            New Goal
          </Button>
        )}
      </div>

      {goals.length === 0 ? (
        <EmptyState
          icon={<Target className="size-12" />}
          title="No global goals"
          description="Global goals can be accessed by multiple projects"
          action={
            isAdmin && (
              <Button onClick={() => setCreateOpen(true)}>
                <Plus className="mr-2 size-4" />
                Create Goal
              </Button>
            )
          }
        />
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {goals.map((goal) => (
            <Card key={goal.id}>
              <CardHeader className="pb-2">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-2">
                    <span
                      className="size-3 rounded-full"
                      style={{ backgroundColor: goal.color || '#6b7280' }}
                    />
                    <CardTitle className="text-lg">{goal.name}</CardTitle>
                  </div>
                  {isAdmin && (
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="size-8">
                          <MoreHorizontal className="size-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => handleEdit(goal)}>
                          <Edit className="mr-2 size-4" />
                          Edit
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          className="text-destructive"
                          onClick={() => {
                            setSelectedGoal(goal);
                            setDeleteOpen(true);
                          }}
                        >
                          <Trash2 className="mr-2 size-4" />
                          Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  )}
                </div>
                <CardDescription className="font-mono text-xs">
                  {goal.slug}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex items-center gap-2 flex-wrap">
                  <Badge variant="secondary">
                    {goal.type === 'counter' ? 'Counter' : 'Daily Counter'}
                  </Badge>
                  <Badge variant="outline">
                    {goal.projectAccess?.length || 0} projects
                  </Badge>
                </div>
                {goal.description && (
                  <p className="mt-2 text-sm text-muted-foreground line-clamp-2">
                    {goal.description}
                  </p>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Create Dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Create Global Goal</DialogTitle>
            <DialogDescription>
              Create a metric that can be shared across projects
            </DialogDescription>
          </DialogHeader>
          <FieldGroup>
            <div className="grid grid-cols-2 gap-4">
              <Field>
                <FieldLabel>Name</FieldLabel>
                <Input
                  value={name}
                  onChange={(e) => handleNameChange(e.target.value)}
                  placeholder="Total Signups"
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
                  placeholder="total-signups"
                />
              </Field>
            </div>
            <div className="grid grid-cols-2 gap-4">
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
            </div>
            <Field>
              <FieldLabel>Description</FieldLabel>
              <Textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Optional description..."
                rows={2}
              />
            </Field>
            <Field>
              <FieldLabel>Project Access</FieldLabel>
              <FieldDescription>
                Select which projects can write to this goal
              </FieldDescription>
              <div className="mt-2 space-y-2 max-h-48 overflow-y-auto border rounded-md p-3">
                {projects.map((project) => (
                  <div key={project.id} className="flex items-center gap-2">
                    <Checkbox
                      id={project.id}
                      checked={projectAccess.includes(project.id)}
                      onCheckedChange={() => toggleProjectAccess(project.id)}
                    />
                    <label
                      htmlFor={project.id}
                      className="text-sm flex items-center gap-2 cursor-pointer"
                    >
                      <span
                        className="size-2 rounded-full"
                        style={{ backgroundColor: project.color || '#6b7280' }}
                      />
                      {project.name}
                    </label>
                  </div>
                ))}
              </div>
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>
              Cancel
            </Button>
            <LoadingButton
              onClick={handleCreate}
              disabled={!name || !slug}
              loading={createMutation.isPending}
            >
              Create
            </LoadingButton>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Dialog */}
      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Edit Goal</DialogTitle>
          </DialogHeader>
          <FieldGroup>
            <div className="grid grid-cols-2 gap-4">
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
              </Field>
            </div>
            <div className="grid grid-cols-2 gap-4">
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
            </div>
            <Field>
              <FieldLabel>Description</FieldLabel>
              <Textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={2}
              />
            </Field>
            <Field>
              <FieldLabel>Project Access</FieldLabel>
              <div className="mt-2 space-y-2 max-h-48 overflow-y-auto border rounded-md p-3">
                {projects.map((project) => (
                  <div key={project.id} className="flex items-center gap-2">
                    <Checkbox
                      id={`edit-${project.id}`}
                      checked={projectAccess.includes(project.id)}
                      onCheckedChange={() => toggleProjectAccess(project.id)}
                    />
                    <label
                      htmlFor={`edit-${project.id}`}
                      className="text-sm flex items-center gap-2 cursor-pointer"
                    >
                      <span
                        className="size-2 rounded-full"
                        style={{ backgroundColor: project.color || '#6b7280' }}
                      />
                      {project.name}
                    </label>
                  </div>
                ))}
              </div>
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)}>
              Cancel
            </Button>
            <LoadingButton
              onClick={handleUpdate}
              loading={updateMutation.isPending}
            >
              Save
            </LoadingButton>
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
