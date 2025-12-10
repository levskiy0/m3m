import { useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Database, Settings, Table, Trash2, MoreHorizontal } from 'lucide-react';
import { toast } from 'sonner';

import { modelsApi } from '@/api';
import { queryKeys } from '@/lib/query-keys';
import { useAutoSlug, useDeleteDialog, useTitle } from '@/hooks';
import type { CreateModelRequest, ModelField } from '@/types';
import { Button } from '@/components/ui/button';
import { LoadingButton } from '@/components/ui/loading-button';
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
import { Input } from '@/components/ui/input';
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { EmptyState } from '@/components/shared/empty-state';
import { PageHeader } from '@/components/shared/page-header';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';

export function ModelsPage() {
  useTitle('Models');
  const { projectId } = useParams<{ projectId: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [createOpen, setCreateOpen] = useState(false);
  const deleteDialog = useDeleteDialog<string>();
  const { name, slug, setName, setSlug, reset: resetSlug } = useAutoSlug({ separator: '_' });

  const { data: models = [], isLoading } = useQuery({
    queryKey: queryKeys.models.all(projectId!),
    queryFn: () => modelsApi.list(projectId!),
    enabled: !!projectId,
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateModelRequest) => modelsApi.create(projectId!, data),
    onSuccess: (model) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.models.all(projectId!) });
      setCreateOpen(false);
      resetSlug();
      toast.success('Model created');
      navigate(`/projects/${projectId}/models/${model.id}/schema`);
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to create model');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (modelId: string) => modelsApi.delete(projectId!, modelId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.models.all(projectId!) });
      deleteDialog.close();
      toast.success('Model deleted');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to delete model');
    },
  });

  const handleCreate = () => {
    const defaultFields: ModelField[] = [
      { key: 'name', type: 'string', required: true },
    ];
    createMutation.mutate({ name, slug, fields: defaultFields });
  };

  if (isLoading) {
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
      <PageHeader
        title="Models"
        description="Define data schemas and manage records"
        action={
          <Button onClick={() => setCreateOpen(true)}>
            <Plus className="mr-2 size-4" />
            New Model
          </Button>
        }
      />

      {models.length === 0 ? (
        <EmptyState
          icon={<Database className="size-12" />}
          title="No models yet"
          description="Create a model to define your data schema"
          action={
            <Button onClick={() => setCreateOpen(true)}>
              <Plus className="mr-2 size-4" />
              Create Model
            </Button>
          }
        />
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {models.map((model) => {
            const hasSchema = model.fields.length > 1;
            const targetPath = hasSchema
              ? `/projects/${projectId}/models/${model.id}/data`
              : `/projects/${projectId}/models/${model.id}/schema`;

            return (
              <Card
                key={model.id}
                className="group cursor-pointer transition-colors hover:bg-muted/50"
                onClick={() => navigate(targetPath)}
              >
                <CardHeader className="pb-2">
                  <div className="flex items-start justify-between">
                    <div>
                      <CardTitle className="text-lg">{model.name}</CardTitle>
                      <CardDescription className="font-mono text-xs">
                        {model.slug}
                      </CardDescription>
                    </div>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
                        <Button variant="ghost" size="icon" className="size-8">
                          <MoreHorizontal className="size-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem asChild>
                          <Link to={`/projects/${projectId}/models/${model.id}/schema`}>
                            <Settings className="mr-2 size-4" />
                            Edit Schema
                          </Link>
                        </DropdownMenuItem>
                        <DropdownMenuItem asChild>
                          <Link to={`/projects/${projectId}/models/${model.id}/data`}>
                            <Table className="mr-2 size-4" />
                            View Data
                          </Link>
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          className="text-destructive"
                          onClick={(e) => {
                            e.stopPropagation();
                            deleteDialog.open(model.id);
                          }}
                        >
                          <Trash2 className="mr-2 size-4" />
                          Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                </CardHeader>
                <CardContent>
                  <Badge variant="secondary">
                    {model.fields.length} fields
                  </Badge>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Model</DialogTitle>
            <DialogDescription>
              Define a new data model for your project
            </DialogDescription>
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel>Name</FieldLabel>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="User Profile"
              />
            </Field>
            <Field>
              <FieldLabel>Slug</FieldLabel>
              <Input
                value={slug}
                onChange={(e) => setSlug(e.target.value)}
                placeholder="user_profile"
              />
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

      <ConfirmDialog
        open={deleteDialog.isOpen}
        onOpenChange={(open) => !open && deleteDialog.close()}
        title="Delete Model"
        description="Are you sure you want to delete this model? All data will be lost."
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={() => deleteDialog.itemToDelete && deleteMutation.mutate(deleteDialog.itemToDelete)}
        isLoading={deleteMutation.isPending}
      />
    </div>
  );
}
