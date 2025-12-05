import { useState } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Variable, Trash2, Edit, Eye, EyeOff, Copy } from 'lucide-react';
import { toast } from 'sonner';

import { environmentApi } from '@/api';
import { ENV_TYPES } from '@/lib/constants';
import type { Environment, CreateEnvRequest, UpdateEnvRequest, EnvType } from '@/types';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
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
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';

function getValueInput(
  type: EnvType,
  value: string,
  onChange: (value: string) => void
) {
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
        <Input
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
      );
  }
}

export function EnvironmentPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const queryClient = useQueryClient();

  const [createOpen, setCreateOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [selectedEnv, setSelectedEnv] = useState<Environment | null>(null);
  const [showValues, setShowValues] = useState<Record<string, boolean>>({});

  // Form state
  const [key, setKey] = useState('');
  const [type, setType] = useState<EnvType>('string');
  const [value, setValue] = useState('');

  const { data: envVars = [], isLoading } = useQuery({
    queryKey: ['environment', projectId],
    queryFn: () => environmentApi.list(projectId!),
    enabled: !!projectId,
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateEnvRequest) =>
      environmentApi.create(projectId!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['environment', projectId] });
      setCreateOpen(false);
      resetForm();
      toast.success('Variable created');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to create variable');
    },
  });

  const updateMutation = useMutation({
    mutationFn: (data: UpdateEnvRequest) =>
      environmentApi.update(projectId!, selectedEnv!.key, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['environment', projectId] });
      setEditOpen(false);
      setSelectedEnv(null);
      resetForm();
      toast.success('Variable updated');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to update variable');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => environmentApi.delete(projectId!, selectedEnv!.key),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['environment', projectId] });
      setDeleteOpen(false);
      setSelectedEnv(null);
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

  const handleCreate = () => {
    createMutation.mutate({ key, type, value });
  };

  const handleEdit = (env: Environment) => {
    setSelectedEnv(env);
    setKey(env.key);
    setType(env.type);
    setValue(env.value);
    setEditOpen(true);
  };

  const handleUpdate = () => {
    updateMutation.mutate({ type, value });
  };

  const toggleShowValue = (key: string) => {
    setShowValues((prev) => ({ ...prev, [key]: !prev[key] }));
  };

  const copyValue = (value: string) => {
    navigator.clipboard.writeText(value);
    toast.success('Copied to clipboard');
  };

  const formatDisplayValue = (env: Environment, show: boolean): string => {
    if (!show) {
      return '••••••••';
    }
    if (env.type === 'boolean') {
      return env.value === 'true' ? 'True' : 'False';
    }
    if (env.type === 'json') {
      try {
        return JSON.stringify(JSON.parse(env.value), null, 2);
      } catch {
        return env.value;
      }
    }
    return env.value.length > 50 ? env.value.slice(0, 50) + '...' : env.value;
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
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Environment</h1>
          <p className="text-muted-foreground">
            Configure environment variables for your service
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-2 size-4" />
          Add Variable
        </Button>
      </div>

      {envVars.length === 0 ? (
        <EmptyState
          icon={<Variable className="size-12" />}
          title="No environment variables"
          description="Add variables to configure your service at runtime"
          action={
            <Button onClick={() => setCreateOpen(true)}>
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
                        {formatDisplayValue(env, showValues[env.key])}
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
                          {showValues[env.key] ? (
                            <EyeOff className="size-4" />
                          ) : (
                            <Eye className="size-4" />
                          )}
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="size-8"
                          onClick={() => copyValue(env.value)}
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
                          onClick={() => {
                            setSelectedEnv(env);
                            setDeleteOpen(true);
                          }}
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

      {/* Create Dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add Variable</DialogTitle>
            <DialogDescription>
              Create a new environment variable
            </DialogDescription>
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
              />
              <FieldDescription>
                Use uppercase letters and underscores
              </FieldDescription>
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
              {getValueInput(type, value, setValue)}
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleCreate}
              disabled={!key || createMutation.isPending}
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
            <DialogTitle>Edit Variable</DialogTitle>
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel>Key</FieldLabel>
              <Input value={key} disabled />
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
              {getValueInput(type, value, setValue)}
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
        title="Delete Variable"
        description={`Are you sure you want to delete "${selectedEnv?.key}"?`}
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={() => deleteMutation.mutate()}
        isLoading={deleteMutation.isPending}
      />
    </div>
  );
}
