import { useState, useMemo, useEffect, useRef } from 'react';
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
  Search,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  Loader2,
} from 'lucide-react';
import { toast } from 'sonner';

import { modelsApi } from '@/api';
import { ApiValidationError } from '@/api/client';
import type { ModelData, ModelField, FieldType, FormConfig, TableConfig, FieldView } from '@/types';

// System fields that can be displayed in tables and views
const SYSTEM_FIELDS = ['_id', '_created_at', '_updated_at'] as const;
type SystemField = typeof SYSTEM_FIELDS[number];

const SYSTEM_FIELD_LABELS: Record<SystemField, string> = {
  '_id': 'ID',
  '_created_at': 'Created At',
  '_updated_at': 'Updated At',
};
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
import { Checkbox } from '@/components/ui/checkbox';
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { EmptyState } from '@/components/shared/empty-state';
import { Skeleton } from '@/components/ui/skeleton';
import { ScrollArea } from '@/components/ui/scroll-area';
import { DatePicker, DateTimePicker } from '@/components/ui/datetime-picker';

interface FieldInputProps {
  field: ModelField;
  value: unknown;
  onChange: (value: unknown) => void;
  view?: FieldView;
}

function FieldInput({ field, value, onChange, view }: FieldInputProps) {
  // Determine widget to use based on view or default
  const widget = view || getDefaultView(field.type);

  switch (field.type) {
    case 'string':
      return (
        <Input
          value={(value as string) || ''}
          onChange={(e) => onChange(e.target.value)}
        />
      );
    case 'text':
      if (widget === 'tiptap' || widget === 'markdown') {
        // For now, use textarea for rich text (can be enhanced later)
        return (
          <Textarea
            value={(value as string) || ''}
            onChange={(e) => onChange(e.target.value)}
            rows={6}
            placeholder={widget === 'markdown' ? 'Write markdown...' : 'Write content...'}
          />
        );
      }
      return (
        <Textarea
          value={(value as string) || ''}
          onChange={(e) => onChange(e.target.value)}
          rows={4}
        />
      );
    case 'number':
    case 'float':
      if (widget === 'slider') {
        // For now, use input for slider (can be enhanced later)
        return (
          <Input
            type="number"
            step={field.type === 'float' ? '0.01' : '1'}
            value={(value as number)?.toString() || ''}
            onChange={(e) => onChange(parseFloat(e.target.value) || 0)}
          />
        );
      }
      return (
        <Input
          type="number"
          step={field.type === 'float' ? '0.01' : '1'}
          value={(value as number)?.toString() || ''}
          onChange={(e) => onChange(parseFloat(e.target.value) || 0)}
        />
      );
    case 'bool':
      if (widget === 'checkbox') {
        return (
          <div className="flex items-center h-10">
            <Checkbox
              checked={!!value}
              onCheckedChange={onChange}
            />
          </div>
        );
      }
      return (
        <div className="flex items-center h-10">
          <Switch
            checked={!!value}
            onCheckedChange={onChange}
          />
        </div>
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
        <DatePicker
          value={(value as string) || undefined}
          onChange={(v) => onChange(v || '')}
        />
      );
    case 'datetime':
      return (
        <DateTimePicker
          value={(value as string) || undefined}
          onChange={(v) => onChange(v || '')}
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

function getDefaultView(type: FieldType): FieldView {
  switch (type) {
    case 'string': return 'input';
    case 'text': return 'textarea';
    case 'number': return 'input';
    case 'float': return 'input';
    case 'bool': return 'switch';
    case 'document': return 'json';
    case 'date': return 'datepicker';
    case 'datetime': return 'datetimepicker';
    case 'file': return 'file';
    case 'ref': return 'select';
    default: return 'input';
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

function formatSystemFieldValue(key: SystemField, value: unknown): string {
  if (value === null || value === undefined) return '-';
  if (key === '_id') {
    const str = String(value);
    return str.length > 24 ? str.slice(0, 24) + '...' : str;
  }
  if (key === '_created_at' || key === '_updated_at') {
    return new Date(value as string).toLocaleString();
  }
  return String(value);
}

function isSystemField(key: string): key is SystemField {
  return SYSTEM_FIELDS.includes(key as SystemField);
}

export function ModelDataPage() {
  const { projectId, modelId } = useParams<{ projectId: string; modelId: string }>();
  const queryClient = useQueryClient();

  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(10);
  const [searchInput, setSearchInput] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [sortField, setSortField] = useState<string | null>(null);
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc');
  const searchInputRef = useRef<HTMLInputElement>(null);

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => {
      if (searchInput !== searchQuery) {
        setSearchQuery(searchInput);
        setPage(1);
      }
    }, 300);
    return () => clearTimeout(timer);
  }, [searchInput, searchQuery]);

  const [createOpen, setCreateOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [viewOpen, setViewOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);

  const [selectedData, setSelectedData] = useState<ModelData | null>(null);
  const [formData, setFormData] = useState<Record<string, unknown>>({});
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const { data: model, isLoading: modelLoading } = useQuery({
    queryKey: ['model', projectId, modelId],
    queryFn: () => modelsApi.get(projectId!, modelId!),
    enabled: !!projectId && !!modelId,
  });

  // Get table config or defaults - must be before data query
  const tableConfig = useMemo((): TableConfig => {
    const defaultColumns = model?.fields.map(f => f.key) || [];
    const defaultSearchable = model?.fields.filter(f => ['string', 'text'].includes(f.type)).map(f => f.key) || [];

    return {
      columns: model?.table_config?.columns ?? defaultColumns,
      filters: model?.table_config?.filters ?? [],
      sort_columns: model?.table_config?.sort_columns ?? defaultColumns,
      searchable: model?.table_config?.searchable ?? defaultSearchable,
    };
  }, [model]);

  // Get form config or defaults
  const formConfig = useMemo((): FormConfig => {
    const defaultFieldOrder = model?.fields.map(f => f.key) || [];

    return {
      field_order: model?.form_config?.field_order ?? defaultFieldOrder,
      hidden_fields: model?.form_config?.hidden_fields ?? [],
      field_views: model?.form_config?.field_views ?? {},
    };
  }, [model]);

  const { data: dataResponse, isLoading: dataLoading, isFetching } = useQuery({
    queryKey: ['model-data', projectId, modelId, page, limit, sortField, sortOrder, searchQuery],
    queryFn: () => {
      // Use queryData for search, listData otherwise
      if (searchQuery && tableConfig.searchable && tableConfig.searchable.length > 0) {
        return modelsApi.queryData(projectId!, modelId!, {
          page,
          limit,
          sort: sortField || undefined,
          order: sortOrder,
          search: searchQuery,
          searchIn: tableConfig.searchable,
        });
      }
      return modelsApi.listData(projectId!, modelId!, {
        page,
        limit,
        sort: sortField || undefined,
        order: sortOrder,
      });
    },
    enabled: !!projectId && !!modelId,
    staleTime: 0,
    placeholderData: (prev) => prev, // Keep previous data while fetching
  });

  // Get visible columns based on tableConfig (regular fields)
  const visibleColumns = useMemo(() => {
    if (!model) return [];
    const fieldMap = new Map(model.fields.map(f => [f.key, f]));
    return tableConfig.columns
      .filter(key => !isSystemField(key))
      .map(key => fieldMap.get(key))
      .filter((f): f is ModelField => f !== undefined);
  }, [model, tableConfig]);

  // Get visible system columns based on tableConfig
  const visibleSystemColumns = useMemo(() => {
    return tableConfig.columns.filter(key => isSystemField(key)) as SystemField[];
  }, [tableConfig]);

  // Get ordered fields for forms based on formConfig
  const orderedFormFields = useMemo(() => {
    if (!model) return [];
    const fieldMap = new Map(model.fields.map(f => [f.key, f]));
    const ordered: ModelField[] = [];

    // First add fields in specified order
    for (const key of formConfig.field_order) {
      const field = fieldMap.get(key);
      if (field && !formConfig.hidden_fields.includes(key)) {
        ordered.push(field);
        fieldMap.delete(key);
      }
    }

    // Then add remaining fields not in the order and not hidden
    for (const [key, field] of fieldMap) {
      if (!formConfig.hidden_fields.includes(key)) {
        ordered.push(field);
      }
    }

    return ordered;
  }, [model, formConfig]);

  const createMutation = useMutation({
    mutationFn: (data: Record<string, unknown>) =>
      modelsApi.createData(projectId!, modelId!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['model-data', projectId, modelId] });
      setCreateOpen(false);
      setFormData({});
      setFieldErrors({});
      toast.success('Record created');
    },
    onError: (err) => {
      if (err instanceof ApiValidationError) {
        const errors: Record<string, string> = {};
        err.details.forEach(d => {
          errors[d.field] = d.message;
        });
        setFieldErrors(errors);
        toast.error('Validation failed. Please check the form.');
      } else {
        toast.error(err instanceof Error ? err.message : 'Failed to create record');
      }
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
      setFieldErrors({});
      toast.success('Record updated');
    },
    onError: (err) => {
      if (err instanceof ApiValidationError) {
        const errors: Record<string, string> = {};
        err.details.forEach(d => {
          errors[d.field] = d.message;
        });
        setFieldErrors(errors);
        toast.error('Validation failed. Please check the form.');
      } else {
        toast.error(err instanceof Error ? err.message : 'Failed to update record');
      }
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
    // Don't pre-fill defaults - backend applies them
    setFormData({});
    setFieldErrors({});
    setCreateOpen(true);
  };

  const handleEdit = (data: ModelData) => {
    setSelectedData(data);
    const editData: Record<string, unknown> = {};
    model?.fields.forEach((field) => {
      editData[field.key] = data[field.key];
    });
    setFormData(editData);
    setFieldErrors({});
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

  const handleSort = (fieldKey: string) => {
    if (!tableConfig.sort_columns.includes(fieldKey)) return;

    if (sortField === fieldKey) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(fieldKey);
      setSortOrder('asc');
    }
    setPage(1);
  };

  // Only show loading skeleton on initial load, not during search/filter
  const isLoading = modelLoading || (dataLoading && !dataResponse);
  const data = dataResponse?.data || [];
  const totalPages = dataResponse?.totalPages || 1;
  const hasSearchable = (tableConfig.searchable?.length || 0) > 0;

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

      {/* Search bar */}
      {hasSearchable && (
        <div className="flex items-center gap-2">
          <div className="relative flex-1 max-w-sm">
            {isFetching && searchQuery ? (
              <Loader2 className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground animate-spin" />
            ) : (
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
            )}
            <Input
              ref={searchInputRef}
              placeholder={`Search in ${tableConfig.searchable?.join(', ')}...`}
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              className="pl-9"
            />
          </div>
        </div>
      )}

      <Card>
        <CardContent className="p-0">
          {data.length === 0 ? (
            searchQuery ? (
              <EmptyState
                title="No results found"
                description={`No records match "${searchQuery}"`}
                className="py-12"
              />
            ) : (
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
            )
          ) : (
            <ScrollArea className="w-full">
              <Table>
                <TableHeader>
                  <TableRow>
                    {visibleColumns.map((field) => {
                      const isSortable = tableConfig.sort_columns.includes(field.key);
                      const isCurrentSort = sortField === field.key;

                      return (
                        <TableHead
                          key={field.key}
                          className={isSortable ? 'cursor-pointer select-none hover:bg-muted/50' : ''}
                          onClick={() => isSortable && handleSort(field.key)}
                        >
                          <div className="flex items-center gap-1">
                            {field.key}
                            {isSortable && (
                              <span className="text-muted-foreground">
                                {isCurrentSort ? (
                                  sortOrder === 'asc' ? (
                                    <ArrowUp className="size-4" />
                                  ) : (
                                    <ArrowDown className="size-4" />
                                  )
                                ) : (
                                  <ArrowUpDown className="size-3 opacity-50" />
                                )}
                              </span>
                            )}
                          </div>
                        </TableHead>
                      );
                    })}
                    {visibleSystemColumns.map((key) => {
                      const isSortable = tableConfig.sort_columns.includes(key);
                      const isCurrentSort = sortField === key;

                      return (
                        <TableHead
                          key={key}
                          className={`text-muted-foreground ${isSortable ? 'cursor-pointer select-none hover:bg-muted/50' : ''}`}
                          onClick={() => isSortable && handleSort(key)}
                        >
                          <div className="flex items-center gap-1">
                            {SYSTEM_FIELD_LABELS[key]}
                            {isSortable && (
                              <span className="text-muted-foreground">
                                {isCurrentSort ? (
                                  sortOrder === 'asc' ? (
                                    <ArrowUp className="size-4" />
                                  ) : (
                                    <ArrowDown className="size-4" />
                                  )
                                ) : (
                                  <ArrowUpDown className="size-3 opacity-50" />
                                )}
                              </span>
                            )}
                          </div>
                        </TableHead>
                      );
                    })}
                    <TableHead className="w-12"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data.map((row) => (
                    <TableRow key={row._id}>
                      {visibleColumns.map((field) => (
                        <TableCell key={field.key}>
                          {formatCellValue(row[field.key], field.type)}
                        </TableCell>
                      ))}
                      {visibleSystemColumns.map((key) => (
                        <TableCell key={key} className="text-muted-foreground text-xs font-mono">
                          {formatSystemFieldValue(key, row[key])}
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
            {orderedFormFields.map((field) => (
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
                  view={formConfig.field_views[field.key]}
                />
                {fieldErrors[field.key] && (
                  <p className="text-sm text-destructive mt-1">{fieldErrors[field.key]}</p>
                )}
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
            {orderedFormFields.map((field) => (
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
                  view={formConfig.field_views[field.key]}
                />
                {fieldErrors[field.key] && (
                  <p className="text-sm text-destructive mt-1">{fieldErrors[field.key]}</p>
                )}
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
            {/* System fields at the top */}
            <div className="pb-3 mb-3 border-b space-y-2">
              {SYSTEM_FIELDS.map((key) => (
                <div key={key} className="grid grid-cols-3 gap-4">
                  <span className="font-medium text-muted-foreground text-sm">
                    {SYSTEM_FIELD_LABELS[key]}
                  </span>
                  <span className="col-span-2 font-mono text-xs text-muted-foreground">
                    {formatSystemFieldValue(key, selectedData?.[key])}
                  </span>
                </div>
              ))}
            </div>
            {/* Regular fields */}
            {orderedFormFields.map((field) => (
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
