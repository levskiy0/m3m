import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Trash2, Edit, MoreHorizontal, Zap, Circle } from 'lucide-react';
import { toast } from 'sonner';

import { actionsApi } from '@/api/actions';
import { queryKeys } from '@/lib/query-keys';
import { useAutoSlug, useFormDialog, useDeleteDialog } from '@/hooks';
import type { Action, CreateActionRequest, UpdateActionRequest } from '@/types';
import { Button } from '@/components/ui/button';
import { CardContent } from '@/components/ui/card';
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
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Field, FieldLabel, FieldDescription, FieldGroup } from '@/components/ui/field';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { EmptyState } from '@/components/shared/empty-state';
import { Skeleton } from '@/components/ui/skeleton';

const ACTION_COLORS = [
  { value: 'none', label: 'Default' },
  { value: 'blue', label: 'Blue' },
  { value: 'green', label: 'Green' },
  { value: 'yellow', label: 'Yellow' },
  { value: 'red', label: 'Red' },
  { value: 'purple', label: 'Purple' },
  { value: 'orange', label: 'Orange' },
];

const colorClasses: Record<string, string> = {
  blue: 'text-blue-500',
  green: 'text-green-500',
  yellow: 'text-yellow-500',
  red: 'text-red-500',
  purple: 'text-purple-500',
  orange: 'text-orange-500',
};

interface ActionsListProps {
  projectId: string;
}

export function ActionsList({ projectId }: ActionsListProps) {
  const queryClient = useQueryClient();

  const formDialog = useFormDialog<Action>();
  const deleteDialog = useDeleteDialog<Action>();

  const { name, slug, setName, setSlug, reset: resetSlug } = useAutoSlug();
  const [group, setGroup] = useState('');
  const [showInMenu, setShowInMenu] = useState(true);
  const [color, setColor] = useState('none');

  const { data: actions = [], isLoading } = useQuery({
    queryKey: queryKeys.actions.all(projectId),
    queryFn: () => actionsApi.list(projectId),
    enabled: !!projectId,
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateActionRequest) => actionsApi.create(projectId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.actions.all(projectId) });
      toast.success('Action created');
      formDialog.close();
      resetForm();
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to create action');
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateActionRequest }) =>
      actionsApi.update(projectId, id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.actions.all(projectId) });
      toast.success('Action updated');
      formDialog.close();
      resetForm();
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to update action');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => actionsApi.delete(projectId, id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.actions.all(projectId) });
      toast.success('Action deleted');
      deleteDialog.close();
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to delete action');
    },
  });

  const resetForm = () => {
    resetSlug();
    setGroup('');
    setShowInMenu(true);
    setColor('none');
  };

  const openCreateDialog = () => {
    resetForm();
    formDialog.open();
  };

  const openEditDialog = (action: Action) => {
    setName(action.name);
    setSlug(action.slug);
    setGroup(action.group || '');
    setShowInMenu(action.show_in_menu);
    setColor(action.color || 'none');
    formDialog.openEdit(action);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    const colorValue = color === 'none' ? undefined : color;

    if (formDialog.mode === 'edit' && formDialog.selectedItem) {
      updateMutation.mutate({
        id: formDialog.selectedItem.id,
        data: {
          name,
          group: group || undefined,
          show_in_menu: showInMenu,
          color: colorValue,
        },
      });
    } else {
      createMutation.mutate({
        name,
        slug,
        group: group || undefined,
        show_in_menu: showInMenu,
        color: colorValue,
      });
    }
  };

  if (isLoading) {
    return (
      <>
        <div className="flex items-center justify-between flex-shrink-0 px-4 py-3 border-b">
          <Skeleton className="h-4 w-24" />
          <Skeleton className="h-8 w-28" />
        </div>
        <CardContent className="flex-1 overflow-auto p-6">
          <Skeleton className="h-32 w-full" />
        </CardContent>
      </>
    );
  }

  return (
    <>
      {/* Actions Header */}
      <div className="flex items-center justify-between flex-shrink-0 px-4 py-3 border-b">
        <span className="text-sm text-muted-foreground">
          {actions.length} action{actions.length !== 1 ? 's' : ''}
        </span>
        <Button
          variant="outline"
          size="sm"
          onClick={openCreateDialog}
        >
          <Zap className="mr-2 size-4" />
          Add Action
        </Button>
      </div>

      {/* Actions Content */}
      <CardContent className="flex-1 overflow-auto p-6">
        {actions.length === 0 ? (
          <EmptyState
            title="No actions"
            description="Create an action to trigger from the UI when your project is running"
            className="py-8"
          />
        ) : (
          <div className="space-y-2">
            {actions.map((action) => (
              <div
                key={action.id}
                className="flex items-center justify-between p-3 rounded-lg bg-background border"
              >
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <Zap className={`size-4 ${action.color ? colorClasses[action.color] : 'text-muted-foreground'}`} />
                    <span className="font-medium">{action.name}</span>
                    <code className="rounded bg-muted px-2 py-0.5 text-xs text-muted-foreground">
                      {action.slug}
                    </code>
                    {action.group && (
                      <span className="rounded bg-muted px-2 py-0.5 text-xs text-muted-foreground">
                        {action.group}
                      </span>
                    )}
                    {!action.show_in_menu && (
                      <span className="rounded bg-yellow-100 dark:bg-yellow-900/30 px-2 py-0.5 text-xs text-yellow-700 dark:text-yellow-400">
                        Hidden
                      </span>
                    )}
                  </div>
                </div>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon" className="size-8">
                      <MoreHorizontal className="size-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={() => openEditDialog(action)}>
                      <Edit className="mr-2 size-4" />
                      Edit
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem
                      className="text-destructive"
                      onClick={() => deleteDialog.open(action)}
                    >
                      <Trash2 className="mr-2 size-4" />
                      Delete
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            ))}
          </div>
        )}
      </CardContent>

      {/* Create/Edit Dialog */}
      <Dialog open={formDialog.isOpen} onOpenChange={formDialog.close}>
        <DialogContent>
          <form onSubmit={handleSubmit}>
            <DialogHeader>
              <DialogTitle>
                {formDialog.mode === 'edit' ? 'Edit Action' : 'Create Action'}
              </DialogTitle>
              <DialogDescription>
                {formDialog.mode === 'edit'
                  ? 'Update the action details.'
                  : 'Create a new action that can be triggered from the UI.'}
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <Field>
                <FieldLabel>Name</FieldLabel>
                <Input
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="My Action"
                  required
                />
                <FieldDescription>
                  The display name for this action.
                </FieldDescription>
              </Field>
              {formDialog.mode !== 'edit' && (
                <Field>
                  <FieldLabel>Slug</FieldLabel>
                  <Input
                    value={slug}
                    onChange={(e) => setSlug(e.target.value)}
                    placeholder="my-action"
                    required
                    pattern="^[a-z0-9-]+$"
                  />
                  <FieldDescription>
                    Unique identifier for this action (lowercase, hyphens only).
                  </FieldDescription>
                </Field>
              )}
              <Field>
                <FieldLabel>Group</FieldLabel>
                <Input
                  value={group}
                  onChange={(e) => setGroup(e.target.value)}
                  placeholder="e.g. Admin, Debug"
                />
                <FieldDescription>
                  Optional group name to organize actions in the menu.
                </FieldDescription>
              </Field>
              <FieldGroup className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label htmlFor="show-in-menu">Show in Menu</Label>
                  <FieldDescription>
                    Display this action in the actions dropdown.
                  </FieldDescription>
                </div>
                <Switch
                  id="show-in-menu"
                  checked={showInMenu}
                  onCheckedChange={setShowInMenu}
                />
              </FieldGroup>
              <Field>
                <FieldLabel>Color</FieldLabel>
                <Select value={color} onValueChange={setColor}>
                  <SelectTrigger>
                    <SelectValue placeholder="Default" />
                  </SelectTrigger>
                  <SelectContent>
                    {ACTION_COLORS.map((c) => (
                      <SelectItem key={c.value} value={c.value}>
                        <div className="flex items-center gap-2">
                          {c.value !== 'none' && (
                            <Circle
                              className={`size-3 fill-current ${colorClasses[c.value] || ''}`}
                            />
                          )}
                          {c.label}
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <FieldDescription>
                  Color indicator for the action button.
                </FieldDescription>
              </Field>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={formDialog.close}>
                Cancel
              </Button>
              <LoadingButton
                type="submit"
                loading={createMutation.isPending || updateMutation.isPending}
              >
                {formDialog.mode === 'edit' ? 'Save' : 'Create'}
              </LoadingButton>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation */}
      <ConfirmDialog
        open={deleteDialog.isOpen}
        onOpenChange={deleteDialog.close}
        title="Delete Action"
        description={`Are you sure you want to delete "${deleteDialog.itemToDelete?.name}"? This action cannot be undone.`}
        onConfirm={() => deleteDialog.itemToDelete && deleteMutation.mutate(deleteDialog.itemToDelete.id)}
        isLoading={deleteMutation.isPending}
        variant="destructive"
      />
    </>
  );
}
