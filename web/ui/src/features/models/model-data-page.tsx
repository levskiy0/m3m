import { useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Plus,
  Trash2,
  Edit,
  Settings,
  ChevronLeft,
  ChevronRight,
  MoreHorizontal,
  Eye,
} from 'lucide-react';
import { toast } from 'sonner';

import { modelsApi } from '@/api';
import type { ModelData, ModelField, FieldType } from '@/types';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { EmptyState } from '@/components/shared/empty-state';
import { Skeleton } from '@/components/ui/skeleton';
import { ScrollArea } from '@/components/ui/scroll-area';

interface FieldInputProps {
  field: ModelField;
  value: unknown;
  onChange: (value: unknown) => void;
}

function FieldInput({ field, value, onChange }: FieldInputProps) {
  switch (field.type) {
    case 'string':
      return (
        <Input
          value={(value as string) || ''}
          onChange={(e) => onChange(e.target.value)}
        />
      );
    case 'text':
      return (
        <Textarea
          value={(value as string) || ''}
          onChange={(e) => onChange(e.target.value)}
          rows={4}
        />
      );
    case 'number':
    case 'float':
      return (
        <Input
          type="number"
          step={field.type === 'float' ? '0.01' : '1'}
          value={(value as number)?.toString() || ''}
          onChange={(e) => onChange(parseFloat(e.target.value) || 0)}
        />
      );
    case 'bool':
      return (
        <Switch
          checked={!!value}
          onCheckedChange={onChange}
        />
      );
    case 'document':
      return (
        <Textarea
          value={typeof value === 'object' ? JSON.stringify(value, null, 2) : ''}
          onChange={(e) => {
            try {
              onChange(JSON.parse(e.target.value));
            } catch {
              // Invalid JSON, keep as string
            }
          }}
          rows={6}
          className="font-mono text-sm"
          placeholder="{}"
        />
      );
    case 'date':
      return (
        <Input
          type="date"
          value={(value as string) || ''}
          onChange={(e) => onChange(e.target.value)}
        />
      );
    case 'datetime':
      return (
        <Input
          type="datetime-local"
          value={(value as string) || ''}
          onChange={(e) => onChange(e.target.value)}
        />
      );
    default:
      return (
        <Input
          value={(value as string) || ''}
          onChange={(e) => onChange(e.target.value)}
        />
      );
  }
}

function formatCellValue(value: unknown, type: FieldType): string {
  if (value === null || value === undefined) return '-';
  if (type === 'bool') return value ? 'Yes' : 'No';
  if (type === 'document') return JSON.stringify(value).slice(0, 50) + '...';
  if (type === 'date' || type === 'datetime') {
    return new Date(value as string).toLocaleString();
  }
  const str = String(value);
  return str.length > 50 ? str.slice(0, 50) + '...' : str;
}

export function ModelDataPage() {
  const { projectId, modelId } = useParams<{ projectId: string; modelId: string }>();
  const queryClient = useQueryClient();

  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(10);

  const [createOpen, setCreateOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [viewOpen, setViewOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);

  const [selectedData, setSelectedData] = useState<ModelData | null>(null);
  const [formData, setFormData] = useState<Record<string, unknown>>({});

  const { data: model, isLoading: modelLoading } = useQuery({
    queryKey: ['model', projectId, modelId],
    queryFn: () => modelsApi.get(projectId!, modelId!),
    enabled: !!projectId && !!modelId,
  });

  const { data: dataResponse, isLoading: dataLoading } = useQuery({
    queryKey: ['model-data', projectId, modelId, page, limit],
    queryFn: () => modelsApi.listData(projectId!, modelId!, { page, limit }),
    enabled: !!projectId && !!modelId,
  });

  const createMutation = useMutation({
    mutationFn: (data: Record<string, unknown>) =>
      modelsApi.createData(projectId!, modelId!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['model-data', projectId, modelId] });
      setCreateOpen(false);
      setFormData({});
      toast.success('Record created');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to create record');
    },
  });

  const updateMutation = useMutation({
    mutationFn: () =>
      modelsApi.updateData(projectId!, modelId!, selectedData!._id, formData),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['model-data', projectId, modelId] });
      setEditOpen(false);
      setSelectedData(null);
      setFormData({});
      toast.success('Record updated');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to update record');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () =>
      modelsApi.deleteData(projectId!, modelId!, selectedData!._id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['model-data', projectId, modelId] });
      setDeleteOpen(false);
      setSelectedData(null);
      toast.success('Record deleted');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to delete record');
    },
  });

  const handleCreate = () => {
    // Initialize form with default values
    const defaults: Record<string, unknown> = {};
    model?.fields.forEach((field) => {
      if (field.defaultValue !== undefined) {
        defaults[field.key] = field.defaultValue;
      }
    });
    setFormData(defaults);
    setCreateOpen(true);
  };

  const handleEdit = (data: ModelData) => {
    setSelectedData(data);
    const editData: Record<string, unknown> = {};
    model?.fields.forEach((field) => {
      editData[field.key] = data[field.key];
    });
    setFormData(editData);
    setEditOpen(true);
  };

  const handleView = (data: ModelData) => {
    setSelectedData(data);
    setViewOpen(true);
  };

  const handleDelete = (data: ModelData) => {
    setSelectedData(data);
    setDeleteOpen(true);
  };

  const isLoading = modelLoading || dataLoading;
  const data = dataResponse?.data || [];
  const totalPages = dataResponse?.totalPages || 1;

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-64" />
      </div>
    );
  }

  if (!model) {
    return <div>Model not found</div>;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{model.name}</h1>
          <p className="text-muted-foreground">
            {dataResponse?.total || 0} records
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button asChild variant="outline">
            <Link to={`/projects/${projectId}/models/${modelId}/schema`}>
              <Settings className="mr-2 size-4" />
              Schema
            </Link>
          </Button>
          <Button onClick={handleCreate}>
            <Plus className="mr-2 size-4" />
            New Record
          </Button>
        </div>
      </div>

      <Card>
        <CardContent className="p-0">
          {data.length === 0 ? (
            <EmptyState
              title="No records"
              description="Create your first record to get started"
              action={
                <Button onClick={handleCreate}>
                  <Plus className="mr-2 size-4" />
                  Create Record
                </Button>
              }
              className="py-12"
            />
          ) : (
            <ScrollArea className="w-full">
              <Table>
                <TableHeader>
                  <TableRow>
                    {model.fields.slice(0, 5).map((field) => (
                      <TableHead key={field.key}>{field.key}</TableHead>
                    ))}
                    <TableHead className="w-12"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data.map((row) => (
                    <TableRow key={row._id}>
                      {model.fields.slice(0, 5).map((field) => (
                        <TableCell key={field.key}>
                          {formatCellValue(row[field.key], field.type)}
                        </TableCell>
                      ))}
                      <TableCell>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon" className="size-8">
                              <MoreHorizontal className="size-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => handleView(row)}>
                              <Eye className="mr-2 size-4" />
                              View
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={() => handleEdit(row)}>
                              <Edit className="mr-2 size-4" />
                              Edit
                            </DropdownMenuItem>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                              className="text-destructive"
                              onClick={() => handleDelete(row)}
                            >
                              <Trash2 className="mr-2 size-4" />
                              Delete
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </ScrollArea>
          )}
        </CardContent>
      </Card>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <Select
            value={limit.toString()}
            onValueChange={(v) => {
              setLimit(parseInt(v));
              setPage(1);
            }}
          >
            <SelectTrigger className="w-32">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="10">10 per page</SelectItem>
              <SelectItem value="25">25 per page</SelectItem>
              <SelectItem value="50">50 per page</SelectItem>
              <SelectItem value="100">100 per page</SelectItem>
            </SelectContent>
          </Select>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="icon"
              onClick={() => setPage(page - 1)}
              disabled={page === 1}
            >
              <ChevronLeft className="size-4" />
            </Button>
            <span className="text-sm">
              Page {page} of {totalPages}
            </span>
            <Button
              variant="outline"
              size="icon"
              onClick={() => setPage(page + 1)}
              disabled={page === totalPages}
            >
              <ChevronRight className="size-4" />
            </Button>
          </div>
        </div>
      )}

      {/* Create Dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Create Record</DialogTitle>
            <DialogDescription>Add a new record to {model.name}</DialogDescription>
          </DialogHeader>
          <FieldGroup>
            {model.fields.map((field) => (
              <Field key={field.key}>
                <FieldLabel>
                  {field.key}
                  {field.required && <span className="text-destructive ml-1">*</span>}
                </FieldLabel>
                <FieldInput
                  field={field}
                  value={formData[field.key]}
                  onChange={(value) =>
                    setFormData({ ...formData, [field.key]: value })
                  }
                />
              </Field>
            ))}
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={() => createMutation.mutate(formData)}
              disabled={createMutation.isPending}
            >
              {createMutation.isPending ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Dialog */}
      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Edit Record</DialogTitle>
          </DialogHeader>
          <FieldGroup>
            {model.fields.map((field) => (
              <Field key={field.key}>
                <FieldLabel>
                  {field.key}
                  {field.required && <span className="text-destructive ml-1">*</span>}
                </FieldLabel>
                <FieldInput
                  field={field}
                  value={formData[field.key]}
                  onChange={(value) =>
                    setFormData({ ...formData, [field.key]: value })
                  }
                />
              </Field>
            ))}
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={() => updateMutation.mutate()}
              disabled={updateMutation.isPending}
            >
              {updateMutation.isPending ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* View Dialog */}
      <Dialog open={viewOpen} onOpenChange={setViewOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>View Record</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            {model.fields.map((field) => (
              <div key={field.key} className="grid grid-cols-3 gap-4">
                <span className="font-medium text-muted-foreground">
                  {field.key}
                </span>
                <span className="col-span-2">
                  {field.type === 'document' ? (
                    <pre className="text-xs bg-muted p-2 rounded overflow-auto max-h-32">
                      {JSON.stringify(selectedData?.[field.key], null, 2)}
                    </pre>
                  ) : (
                    formatCellValue(selectedData?.[field.key], field.type)
                  )}
                </span>
              </div>
            ))}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setViewOpen(false)}>
              Close
            </Button>
            <Button
              onClick={() => {
                setViewOpen(false);
                handleEdit(selectedData!);
              }}
            >
              <Edit className="mr-2 size-4" />
              Edit
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirm */}
      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title="Delete Record"
        description="Are you sure you want to delete this record?"
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={() => deleteMutation.mutate()}
        isLoading={deleteMutation.isPending}
      />
    </div>
  );
}
