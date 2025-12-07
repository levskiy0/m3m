import { useState } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Variable, Trash2, Edit, Eye, EyeOff, Copy } from 'lucide-react';
import { toast } from 'sonner';

import { environmentApi } from '@/api';
import { ENV_TYPES } from '@/lib/constants';
import { queryKeys } from '@/lib/query-keys';
import { formatEnvValue } from '@/lib/format';
import { copyToClipboard } from '@/lib/utils';
import { useFormDialog, useDeleteDialog } from '@/hooks';
import type { Environment, CreateEnvRequest, UpdateEnvRequest, EnvType } from '@/types';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';
import { Field, FieldGroup, FieldLabel, FieldDescription } from '@/components/ui/field';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { EmptyState } from '@/components/shared/empty-state';
import { PageHeader } from '@/components/shared/page-header';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';

function TypedInput({
  type,
  value,
  onChange,
}: {
  type: EnvType;
  value: string;
  onChange: (value: string) => void;
}) {
  switch (type) {
    case 'text':
    case 'json':
      return (
        <Textarea
          value={value}
          onChange={(e) => onChange(e.target.value)}
          rows={4}
          className={type === 'json' ? 'font-mono text-sm' : ''}
          placeholder={type === 'json' ? '{}' : ''}
        />
      );
    case 'integer':
      return (
        <Input
          type="number"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          step="1"
        />
      );
    case 'float':
      return (
        <Input
          type="number"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          step="0.01"
        />
      );
    case 'boolean':
      return (
        <div className="flex items-center gap-2">
          <Switch
            checked={value === 'true'}
            onCheckedChange={(checked) => onChange(checked ? 'true' : 'false')}
          />
          <span className="text-sm text-muted-foreground">
            {value === 'true' ? 'True' : 'False'}
          </span>
        </div>
      );
    default:
      return (
        <Input value={value} onChange={(e) => onChange(e.target.value)} />
      );
  }
}

export function EnvironmentPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const queryClient = useQueryClient();

  const formDialog = useFormDialog<Environment>();
  const deleteDialog = useDeleteDialog<Environment>();
  const [showValues, setShowValues] = useState<Record<string, boolean>>({});

  const [key, setKey] = useState('');
  const [type, setType] = useState<EnvType>('string');
  const [value, setValue] = useState('');

  const { data: envVars = [], isLoading } = useQuery({
    queryKey: queryKeys.environment.all(projectId!),
    queryFn: () => environmentApi.list(projectId!),
    enabled: !!projectId,
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateEnvRequest) => environmentApi.create(projectId!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.environment.all(projectId!) });
      formDialog.close();
      resetForm();
      toast.success('Variable created');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to create variable');
    },
  });

  const updateMutation = useMutation({
    mutationFn: (data: UpdateEnvRequest) =>
      environmentApi.update(projectId!, formDialog.selectedItem!.key, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.environment.all(projectId!) });
      formDialog.close();
      resetForm();
      toast.success('Variable updated');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to update variable');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => environmentApi.delete(projectId!, deleteDialog.itemToDelete!.key),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.environment.all(projectId!) });
      deleteDialog.close();
      toast.success('Variable deleted');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to delete variable');
    },
  });

  const resetForm = () => {
    setKey('');
    setType('string');
    setValue('');
  };

  const handleEdit = (env: Environment) => {
    setKey(env.key);
    setType(env.type);
    setValue(env.value);
    formDialog.openEdit(env);
  };

  const handleSubmit = () => {
    if (formDialog.mode === 'create') {
      createMutation.mutate({ key, type, value });
    } else {
      updateMutation.mutate({ type, value });
    }
  };

  const toggleShowValue = (envKey: string) => {
    setShowValues((prev) => ({ ...prev, [envKey]: !prev[envKey] }));
  };

  const handleCopy = async (val: string) => {
    const success = await copyToClipboard(val);
    if (success) {
      toast.success('Copied to clipboard');
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-64" />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <PageHeader
        title="Environment"
        description="Configure environment variables for your service"
        action={
          <Button onClick={() => { resetForm(); formDialog.open(); }}>
            <Plus className="mr-2 size-4" />
            Add Variable
          </Button>
        }
      />

      {envVars.length === 0 ? (
        <EmptyState
          icon={<Variable className="size-12" />}
          title="No environment variables"
          description="Add variables to configure your service at runtime"
          action={
            <Button onClick={() => { resetForm(); formDialog.open(); }}>
              <Plus className="mr-2 size-4" />
              Add Variable
            </Button>
          }
        />
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Key</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Value</TableHead>
                  <TableHead className="w-32"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {envVars.map((env) => (
                  <TableRow key={env.key}>
                    <TableCell>
                      <code className="text-sm font-medium">{env.key}</code>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{env.type}</Badge>
                    </TableCell>
                    <TableCell>
                      <code className="text-sm text-muted-foreground">
                        {formatEnvValue(env.value, env.type, { masked: !showValues[env.key] })}
                      </code>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <Button
                          variant="ghost"
                          size="icon"
                          className="size-8"
                          onClick={() => toggleShowValue(env.key)}
                        >
                          {showValues[env.key] ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="size-8"
                          onClick={() => handleCopy(env.value)}
                        >
                          <Copy className="size-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="size-8"
                          onClick={() => handleEdit(env)}
                        >
                          <Edit className="size-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="size-8"
                          onClick={() => deleteDialog.open(env)}
                        >
                          <Trash2 className="size-4 text-destructive" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      <Dialog open={formDialog.isOpen} onOpenChange={(open) => !open && formDialog.close()}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {formDialog.mode === 'create' ? 'Add Variable' : 'Edit Variable'}
            </DialogTitle>
            {formDialog.mode === 'create' && (
              <DialogDescription>Create a new environment variable</DialogDescription>
            )}
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel>Key</FieldLabel>
              <Input
                value={key}
                onChange={(e) =>
                  setKey(e.target.value.toUpperCase().replace(/[^A-Z0-9_]/g, '_'))
                }
                placeholder="API_KEY"
                disabled={formDialog.mode === 'edit'}
              />
              {formDialog.mode === 'create' && (
                <FieldDescription>Use uppercase letters and underscores</FieldDescription>
              )}
            </Field>
            <Field>
              <FieldLabel>Type</FieldLabel>
              <Select value={type} onValueChange={(v) => setType(v as EnvType)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {ENV_TYPES.map((t) => (
                    <SelectItem key={t.value} value={t.value}>
                      {t.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
            <Field>
              <FieldLabel>Value</FieldLabel>
              <TypedInput type={type} value={value} onChange={setValue} />
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => formDialog.close()}>
              Cancel
            </Button>
            <LoadingButton
              onClick={handleSubmit}
              disabled={!key}
              loading={createMutation.isPending || updateMutation.isPending}
            >
              {formDialog.mode === 'create' ? 'Create' : 'Save'}
            </LoadingButton>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={deleteDialog.isOpen}
        onOpenChange={(open) => !open && deleteDialog.close()}
        title="Delete Variable"
        description={`Are you sure you want to delete "${deleteDialog.itemToDelete?.key}"?`}
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={() => deleteMutation.mutate()}
        isLoading={deleteMutation.isPending}
      />
    </div>
  );
}
