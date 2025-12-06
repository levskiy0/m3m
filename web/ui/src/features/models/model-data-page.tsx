import { useState, useMemo, useEffect, useRef, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Plus,
  Trash2,
  Edit,
  Settings,
  Search,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  Loader2,
  Filter,
  X,
  Eye,
  MoreHorizontal,
  Table2,
} from 'lucide-react';
import { toast } from 'sonner';

import { modelsApi } from '@/api';
import { formatDateTime, formatFieldLabel } from '@/lib/format';
import { ApiValidationError } from '@/api/client';
import type { ModelData, ModelField, FieldType, FormConfig, TableConfig, FieldView, FilterCondition } from '@/types';

// System fields that can be displayed in tables and views
const SYSTEM_FIELDS = ['_id', '_created_at', '_updated_at'] as const;
type SystemField = typeof SYSTEM_FIELDS[number];

const SYSTEM_FIELD_LABELS: Record<SystemField, string> = {
  '_id': 'ID',
  '_created_at': 'Created At',
  '_updated_at': 'Updated At',
};

import { Button } from '@/components/ui/button';
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
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { EmptyState } from '@/components/shared/empty-state';
import { Skeleton } from '@/components/ui/skeleton';
import { DatePicker, DateTimePicker } from '@/components/ui/datetime-picker';
import { EditorTabs, EditorTab } from '@/components/ui/editor-tabs';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { Badge } from '@/components/ui/badge';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

// Filter operators by field type
const FILTER_OPERATORS: Record<string, { value: FilterCondition['operator']; label: string }[]> = {
  string: [
    { value: 'eq', label: '=' },
    { value: 'ne', label: '!=' },
    { value: 'contains', label: 'Contains' },
    { value: 'startsWith', label: 'Starts with' },
    { value: 'endsWith', label: 'Ends with' },
  ],
  text: [
    { value: 'eq', label: '=' },
    { value: 'ne', label: '!=' },
    { value: 'contains', label: 'Contains' },
  ],
  number: [
    { value: 'eq', label: '=' },
    { value: 'ne', label: '!=' },
    { value: 'gt', label: '>' },
    { value: 'gte', label: '>=' },
    { value: 'lt', label: '<' },
    { value: 'lte', label: '<=' },
  ],
  float: [
    { value: 'eq', label: '=' },
    { value: 'ne', label: '!=' },
    { value: 'gt', label: '>' },
    { value: 'gte', label: '>=' },
    { value: 'lt', label: '<' },
    { value: 'lte', label: '<=' },
  ],
  bool: [
    { value: 'eq', label: '=' },
  ],
  date: [
    { value: 'eq', label: '=' },
    { value: 'ne', label: '!=' },
    { value: 'gt', label: '>' },
    { value: 'gte', label: '>=' },
    { value: 'lt', label: '<' },
    { value: 'lte', label: '<=' },
  ],
  datetime: [
    { value: 'eq', label: '=' },
    { value: 'ne', label: '!=' },
    { value: 'gt', label: '>' },
    { value: 'gte', label: '>=' },
    { value: 'lt', label: '<' },
    { value: 'lte', label: '<=' },
  ],
};

interface FieldInputProps {
  field: ModelField;
  value: unknown;
  onChange: (value: unknown) => void;
  view?: FieldView;
}

function FieldInput({ field, value, onChange, view }: FieldInputProps) {
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
              // Invalid JSON
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
    return formatDateTime(value as string);
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
    return formatDateTime(value as string);
  }
  return String(value);
}

function isSystemField(key: string): key is SystemField {
  return SYSTEM_FIELDS.includes(key as SystemField);
}

// Tab types for file-manager style interface
type TabType = 'table' | 'view' | 'edit' | 'create';

interface Tab {
  id: string;
  type: TabType;
  title: string;
  data?: ModelData;
  formData?: Record<string, unknown>;
  fieldErrors?: Record<string, string>;
}

interface ActiveFilter {
  field: string;
  operator: FilterCondition['operator'];
  value: unknown;
}

export function ModelDataPage() {
  const { projectId, modelId } = useParams<{ projectId: string; modelId: string }>();
  const queryClient = useQueryClient();

  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(25);
  const [searchInput, setSearchInput] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [sortField, setSortField] = useState<string | null>(null);
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc');
  const searchInputRef = useRef<HTMLInputElement>(null);

  // Filters
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);

  // Selection for bulk actions
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());

  // Tabs state (file-manager style)
  const [tabs, setTabs] = useState<Tab[]>([{ id: 'table', type: 'table', title: 'Table' }]);
  const [activeTabId, setActiveTabId] = useState('table');

  // Delete confirmation
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [bulkDeleteOpen, setBulkDeleteOpen] = useState(false);
  const [deleteTargetId, setDeleteTargetId] = useState<string | null>(null);

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

  const { data: model, isLoading: modelLoading } = useQuery({
    queryKey: ['model', projectId, modelId],
    queryFn: () => modelsApi.get(projectId!, modelId!),
    enabled: !!projectId && !!modelId,
  });

  // Get table config or defaults
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

  // Build filter conditions for API
  const filterConditions = useMemo((): FilterCondition[] => {
    return activeFilters.map(f => ({
      field: f.field,
      operator: f.operator,
      value: f.value,
    }));
  }, [activeFilters]);

  const { data: dataResponse, isLoading: dataLoading, isFetching } = useQuery({
    queryKey: ['model-data', projectId, modelId, page, limit, sortField, sortOrder, searchQuery, filterConditions],
    queryFn: () => {
      // Always use queryData for advanced filtering
      return modelsApi.queryData(projectId!, modelId!, {
        page,
        limit,
        sort: sortField || undefined,
        order: sortOrder,
        search: searchQuery || undefined,
        searchIn: searchQuery && tableConfig.searchable?.length ? tableConfig.searchable : undefined,
        filters: filterConditions.length > 0 ? filterConditions : undefined,
      });
    },
    enabled: !!projectId && !!modelId,
    staleTime: 0,
    placeholderData: (prev) => prev,
  });

  // Get visible columns based on tableConfig
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

    for (const key of formConfig.field_order) {
      const field = fieldMap.get(key);
      if (field && !formConfig.hidden_fields.includes(key)) {
        ordered.push(field);
        fieldMap.delete(key);
      }
    }

    for (const [key, field] of fieldMap) {
      if (!formConfig.hidden_fields.includes(key)) {
        ordered.push(field);
      }
    }

    return ordered;
  }, [model, formConfig]);

  // Get filterable fields
  const filterableFields = useMemo(() => {
    if (!model) return [];
    return model.fields.filter(f => tableConfig.filters.includes(f.key));
  }, [model, tableConfig]);

  // Tab management functions
  const openTab = useCallback((type: TabType, data?: ModelData) => {
    const tabId = type === 'create' ? `create-${Date.now()}` :
                  type === 'table' ? 'table' :
                  `${type}-${data?._id}`;

    // Check if tab already exists
    const existingTab = tabs.find(t => t.id === tabId);
    if (existingTab) {
      setActiveTabId(tabId);
      return;
    }

    const title = type === 'create' ? 'New Record' :
                  type === 'edit' ? `Edit: ${data?._id.slice(-6)}` :
                  type === 'view' ? `View: ${data?._id.slice(-6)}` : 'Table';

    const formData: Record<string, unknown> = {};
    if (type === 'edit' && data) {
      model?.fields.forEach((field) => {
        formData[field.key] = data[field.key];
      });
    }

    const newTab: Tab = { id: tabId, type, title, data, formData, fieldErrors: {} };
    setTabs(prev => [...prev, newTab]);
    setActiveTabId(tabId);
  }, [tabs, model]);

  const closeTab = useCallback((tabId: string) => {
    if (tabId === 'table') return; // Can't close table tab

    setTabs(prev => {
      const newTabs = prev.filter(t => t.id !== tabId);
      // If closing active tab, switch to previous or table
      if (activeTabId === tabId) {
        const idx = prev.findIndex(t => t.id === tabId);
        const newActiveIdx = Math.max(0, idx - 1);
        setActiveTabId(newTabs[newActiveIdx]?.id || 'table');
      }
      return newTabs;
    });
  }, [activeTabId]);

  const updateTabFormData = useCallback((tabId: string, formData: Record<string, unknown>) => {
    setTabs(prev => prev.map(t => t.id === tabId ? { ...t, formData } : t));
  }, []);

  const updateTabErrors = useCallback((tabId: string, fieldErrors: Record<string, string>) => {
    setTabs(prev => prev.map(t => t.id === tabId ? { ...t, fieldErrors } : t));
  }, []);

  const createMutation = useMutation({
    mutationFn: (params: { tabId: string; data: Record<string, unknown> }) =>
      modelsApi.createData(projectId!, modelId!, params.data),
    onSuccess: (_, { tabId }) => {
      queryClient.invalidateQueries({ queryKey: ['model-data', projectId, modelId] });
      closeTab(tabId);
      toast.success('Record created');
    },
    onError: (err, { tabId }) => {
      if (err instanceof ApiValidationError) {
        const errors: Record<string, string> = {};
        err.details.forEach(d => {
          errors[d.field] = d.message;
        });
        updateTabErrors(tabId, errors);
        toast.error('Validation failed. Please check the form.');
      } else {
        toast.error(err instanceof Error ? err.message : 'Failed to create record');
      }
    },
  });

  const updateMutation = useMutation({
    mutationFn: (params: { tabId: string; dataId: string; data: Record<string, unknown> }) =>
      modelsApi.updateData(projectId!, modelId!, params.dataId, params.data),
    onSuccess: (_, { tabId }) => {
      queryClient.invalidateQueries({ queryKey: ['model-data', projectId, modelId] });
      closeTab(tabId);
      toast.success('Record updated');
    },
    onError: (err, { tabId }) => {
      if (err instanceof ApiValidationError) {
        const errors: Record<string, string> = {};
        err.details.forEach(d => {
          errors[d.field] = d.message;
        });
        updateTabErrors(tabId, errors);
        toast.error('Validation failed. Please check the form.');
      } else {
        toast.error(err instanceof Error ? err.message : 'Failed to update record');
      }
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (dataId: string) =>
      modelsApi.deleteData(projectId!, modelId!, dataId),
    onSuccess: (_, dataId) => {
      queryClient.invalidateQueries({ queryKey: ['model-data', projectId, modelId] });
      setDeleteOpen(false);
      setDeleteTargetId(null);
      // Close any tabs related to this record
      setTabs(prev => prev.filter(t => !t.id.includes(dataId)));
      toast.success('Record deleted');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to delete record');
    },
  });

  const bulkDeleteMutation = useMutation({
    mutationFn: () =>
      modelsApi.bulkDeleteData(projectId!, modelId!, Array.from(selectedIds)),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['model-data', projectId, modelId] });
      setBulkDeleteOpen(false);
      setSelectedIds(new Set());
      toast.success(`${result.deleted_count} records deleted`);
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to delete records');
    },
  });

  const handleCreate = useCallback(() => {
    openTab('create');
  }, [openTab]);

  const handleEdit = useCallback((data: ModelData) => {
    openTab('edit', data);
  }, [openTab]);

  const handleView = useCallback((data: ModelData) => {
    openTab('view', data);
  }, [openTab]);

  const handleDelete = useCallback((data: ModelData) => {
    setDeleteTargetId(data._id);
    setDeleteOpen(true);
  }, []);

  const handleSort = useCallback((fieldKey: string) => {
    if (!tableConfig.sort_columns.includes(fieldKey)) return;

    if (sortField === fieldKey) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(fieldKey);
      setSortOrder('asc');
    }
    setPage(1);
  }, [sortField, sortOrder, tableConfig.sort_columns]);

  const handleAddFilter = useCallback(() => {
    // Add a new empty filter row with the first available field
    const firstField = filterableFields[0];
    if (!firstField) return;

    setActiveFilters(prev => [...prev, { field: firstField.key, operator: 'eq', value: '' }]);
  }, [filterableFields]);

  const handleRemoveFilter = useCallback((index: number) => {
    setActiveFilters(prev => prev.filter((_, i) => i !== index));
    setPage(1);
  }, []);

  const handleSelectRow = useCallback((id: string, checked: boolean) => {
    setSelectedIds(prev => {
      const next = new Set(prev);
      if (checked) {
        next.add(id);
      } else {
        next.delete(id);
      }
      return next;
    });
  }, []);

  const isLoading = modelLoading || (dataLoading && !dataResponse);
  const data = useMemo(() => dataResponse?.data || [], [dataResponse?.data]);
  const totalPages = dataResponse?.totalPages || 1;
  const total = dataResponse?.total || 0;
  const hasSearchable = (tableConfig.searchable?.length || 0) > 0;
  const hasFilters = filterableFields.length > 0;
  const allSelected = data.length > 0 && data.every(d => selectedIds.has(d._id));
  const someSelected = selectedIds.size > 0;

  const handleSelectAll = useCallback((checked: boolean) => {
    if (checked) {
      setSelectedIds(new Set(data.map(d => d._id)));
    } else {
      setSelectedIds(new Set());
    }
  }, [data]);

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
      <div className="h-full flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between mb-4">
          <div>
            <h1 className="text-2xl font-bold tracking-tight">{model.name}</h1>
            <p className="text-muted-foreground">
              {total} records
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Button asChild variant="outline">
              <Link to={`/projects/${projectId}/models/${modelId}/schema`}>
                <Settings className="mr-2 size-4"/>
                Schema
              </Link>
            </Button>
            <Button onClick={handleCreate}>
              <Plus className="mr-2 size-4"/>
              New Record
            </Button>
          </div>
        </div>

        {/* Toolbar: Search, Filters, Bulk Actions */}
        <div className="flex flex-wrap items-center gap-2 mb-4">
          {/* Search */}
          {hasSearchable && (
              <div className="relative flex-1 max-w-sm">
                {isFetching && searchQuery ? (
                    <Loader2
                        className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground animate-spin"/>
                ) : (
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground"/>
                )}
                <Input
                    ref={searchInputRef}
                    placeholder={`Search in ${tableConfig.searchable?.join(', ')}...`}
                    value={searchInput}
                    onChange={(e) => setSearchInput(e.target.value)}
                    className="pl-9"
                />
              </div>
          )}

        </div>

        {/* Filter Popover & Active Filters Display */}
        {hasFilters && (
            <div className="flex items-center gap-2 mb-4 flex-wrap">
              <Popover>
                <PopoverTrigger asChild>
                  <Button variant="outline" size="sm">
                    <Filter className="mr-2 size-4"/>
                    Filter
                    {activeFilters.length > 0 && (
                        <Badge variant="secondary" className="ml-2 px-1.5 py-0 text-xs">
                          {activeFilters.length}
                        </Badge>
                    )}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                  <div className="p-3 space-y-2 min-w-[400px]">
                    {activeFilters.length === 0 ? (
                        <p className="text-sm text-muted-foreground py-2">No filters applied</p>
                    ) : (
                        activeFilters.map((f, index) => {
                          const field = model.fields.find(fld => fld.key === f.field);
                          const operators = FILTER_OPERATORS[field?.type || 'string'] || FILTER_OPERATORS.string;
                          return (
                              <div key={index} className="flex items-center gap-1.5">
                                {index > 0 && (
                                    <span className="text-xs text-muted-foreground w-7 text-right">and</span>
                                )}
                                {index === 0 && <span className="w-7"/>}
                                <Select
                                    value={f.field}
                                    onValueChange={(v) => {
                                      const newFilters = [...activeFilters];
                                      newFilters[index] = {...newFilters[index], field: v, operator: 'eq', value: ''};
                                      setActiveFilters(newFilters);
                                    }}
                                >
                                  <SelectTrigger className="h-8 w-[120px] text-xs">
                                    <SelectValue/>
                                  </SelectTrigger>
                                  <SelectContent>
                                    {filterableFields.map(fld => (
                                        <SelectItem key={fld.key} value={fld.key}>
                                          {formatFieldLabel(fld.key)}
                                        </SelectItem>
                                    ))}
                                  </SelectContent>
                                </Select>
                                <Select
                                    value={f.operator}
                                    onValueChange={(v) => {
                                      const newFilters = [...activeFilters];
                                      newFilters[index] = {
                                        ...newFilters[index],
                                        operator: v as FilterCondition['operator']
                                      };
                                      setActiveFilters(newFilters);
                                      setPage(1);
                                    }}
                                >
                                  <SelectTrigger className="h-8 w-[70px] text-xs">
                                    <SelectValue/>
                                  </SelectTrigger>
                                  <SelectContent>
                                    {operators.map(op => (
                                        <SelectItem key={op.value} value={op.value}>
                                          {op.label}
                                        </SelectItem>
                                    ))}
                                  </SelectContent>
                                </Select>
                                {field?.type === 'bool' ? (
                                    <Select
                                        value={String(f.value)}
                                        onValueChange={(v) => {
                                          const newFilters = [...activeFilters];
                                          newFilters[index] = {...newFilters[index], value: v === 'true'};
                                          setActiveFilters(newFilters);
                                          setPage(1);
                                        }}
                                    >
                                      <SelectTrigger className="h-8 w-[80px] text-xs">
                                        <SelectValue placeholder="..."/>
                                      </SelectTrigger>
                                      <SelectContent>
                                        <SelectItem value="true">Yes</SelectItem>
                                        <SelectItem value="false">No</SelectItem>
                                      </SelectContent>
                                    </Select>
                                ) : field?.type === 'date' ? (
                                    <div className="w-[130px]">
                                      <DatePicker
                                          value={String(f.value) || undefined}
                                          onChange={(v) => {
                                            const newFilters = [...activeFilters];
                                            newFilters[index] = {...newFilters[index], value: v || ''};
                                            setActiveFilters(newFilters);
                                            setPage(1);
                                          }}
                                      />
                                    </div>
                                ) : field?.type === 'datetime' ? (
                                    <div className="w-[180px]">
                                      <DateTimePicker
                                          value={String(f.value) || undefined}
                                          onChange={(v) => {
                                            const newFilters = [...activeFilters];
                                            newFilters[index] = {...newFilters[index], value: v || ''};
                                            setActiveFilters(newFilters);
                                            setPage(1);
                                          }}
                                      />
                                    </div>
                                ) : (
                                    <Input
                                        type={field?.type === 'number' || field?.type === 'float' ? 'number' : 'text'}
                                        value={String(f.value)}
                                        onChange={(e) => {
                                          const newFilters = [...activeFilters];
                                          const val = field?.type === 'number' ? parseInt(e.target.value, 10) || 0
                                              : field?.type === 'float' ? parseFloat(e.target.value) || 0
                                                  : e.target.value;
                                          newFilters[index] = {...newFilters[index], value: val};
                                          setActiveFilters(newFilters);
                                        }}
                                        onBlur={() => setPage(1)}
                                        onKeyDown={(e) => e.key === 'Enter' && setPage(1)}
                                        placeholder="value"
                                        className="h-8 w-[120px] text-xs"
                                    />
                                )}
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="h-8 w-8 shrink-0"
                                    onClick={() => handleRemoveFilter(index)}
                                >
                                  <X className="size-3.5"/>
                                </Button>
                              </div>
                          );
                        })
                    )}
                    <div className="flex items-center gap-2 pt-2 border-t">
                      <Button
                          variant="ghost"
                          size="sm"
                          className="h-7 text-xs"
                          onClick={handleAddFilter}
                      >
                        <Plus className="mr-1 size-3"/>
                        Add
                      </Button>
                      {activeFilters.length > 0 && (
                          <Button
                              variant="ghost"
                              size="sm"
                              className="h-7 text-xs text-muted-foreground"
                              onClick={() => {
                                setActiveFilters([]);
                                setPage(1);
                              }}
                          >
                            Clear all
                          </Button>
                      )}
                    </div>
                  </div>
                </PopoverContent>
              </Popover>

              {/* Show active filters as compact badges */}
              {activeFilters.map((f, i) => {
                const opLabel = FILTER_OPERATORS[model.fields.find(fld => fld.key === f.field)?.type || 'string']
                    ?.find(op => op.value === f.operator)?.label || f.operator;
                return (
                    <Badge key={i} variant="secondary" className="gap-1 font-normal">
                      <span className="text-muted-foreground">{formatFieldLabel(f.field)}</span>
                      <span>{opLabel}</span>
                      <span className="font-medium">{String(f.value)}</span>
                      <button
                          onClick={() => handleRemoveFilter(i)}
                          className="ml-0.5 hover:text-destructive"
                      >
                        <X className="size-3"/>
                      </button>
                    </Badge>
                );
              })}
            </div>
        )}

        {/* Tabs Bar */}
        <EditorTabs className="px-0 mb-[-1px] relative z-10">
          {tabs.map((tab) => (
              <EditorTab
                  key={tab.id}
                  active={activeTabId === tab.id}
                  onClick={() => setActiveTabId(tab.id)}
                  icon={
                    tab.type === 'table' ? <Table2 className="size-4"/> :
                        tab.type === 'view' ? <Eye className="size-4"/> :
                            tab.type === 'edit' ? <Edit className="size-4"/> :
                                <Plus className="size-4"/>
                  }
                  onClose={tab.id !== 'table' ? () => closeTab(tab.id) : undefined}
                  className="bg-background"
              >
                {tab.title}
              </EditorTab>
          ))}
        </EditorTabs>

        <div className="border-b mb-4 relative top-[1px]"></div>

        {/* Tab Content */}
        <div className="flex-1 min-h-0 flex flex-col">
          {/* Table Tab */}
          {activeTabId === 'table' && (
              <>
                {data.length === 0 ? (
                    <div className="flex-1 flex items-center justify-center border rounded-md">
                      {searchQuery || activeFilters.length > 0 ? (
                          <EmptyState
                              title="No results found"
                              description={searchQuery ? `No records match "${searchQuery}"` : "No records match the current filters"}
                          />
                      ) : (
                          <EmptyState
                              title="No records"
                              description="Create your first record to get started"
                              action={
                                <Button onClick={handleCreate}>
                                  <Plus className="mr-2 size-4"/>
                                  Create Record
                                </Button>
                              }
                          />
                      )}
                    </div>
                ) : (
                    <div className="overflow-hidden rounded-md border">
                      <Table
                          wrapperClassName="h-[calc(100vh-380px)] overflow-auto [&_thead]:sticky [&_thead]:top-0 [&_thead]:z-10 [&_thead]:bg-background">
                        <TableHeader>
                          <TableRow>
                            <TableHead className="w-12 min-w-12 bg-background">
                              <Checkbox
                                  checked={allSelected}
                                  onCheckedChange={handleSelectAll}
                              />
                            </TableHead>
                            {visibleColumns.map((field) => {
                              const isSortable = tableConfig.sort_columns.includes(field.key);
                              const isCurrentSort = sortField === field.key;
                              return (
                                  <TableHead
                                      key={field.key}
                                      className={`whitespace-nowrap bg-background ${isSortable ? 'cursor-pointer select-none hover:bg-muted/50' : ''}`}
                                      onClick={() => isSortable && handleSort(field.key)}
                                  >
                                    <div className="flex items-center gap-2">
                                      {formatFieldLabel(field.key)}
                                      {isSortable && (
                                          <span className="text-muted-foreground">
                                  {isCurrentSort ? (sortOrder === 'asc' ? <ArrowUp className="size-4"/> :
                                      <ArrowDown className="size-4"/>) : <ArrowUpDown className="size-3 opacity-50"/>}
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
                                      className={`whitespace-nowrap text-muted-foreground bg-background ${isSortable ? 'cursor-pointer select-none hover:bg-muted/50' : ''}`}
                                      onClick={() => isSortable && handleSort(key)}
                                  >
                                    <div className="flex items-center gap-2">
                                      {SYSTEM_FIELD_LABELS[key]}
                                      {isSortable && (
                                          <span className="text-muted-foreground">
                                  {isCurrentSort ? (sortOrder === 'asc' ? <ArrowUp className="size-4"/> :
                                      <ArrowDown className="size-4"/>) : <ArrowUpDown className="size-3 opacity-50"/>}
                                </span>
                                      )}
                                    </div>
                                  </TableHead>
                              );
                            })}
                            <TableHead className="w-12 bg-background"/>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {data.map((row) => (
                              <TableRow
                                  key={row._id}
                                  className="cursor-pointer text-md hover:bg-muted/50"
                                  onDoubleClick={() => handleView(row)}
                              >
                                <TableCell className="w-12 min-w-12" onClick={(e) => e.stopPropagation()}>
                                  <Checkbox
                                      checked={selectedIds.has(row._id)}
                                      onCheckedChange={(checked) => handleSelectRow(row._id, !!checked)}
                                  />
                                </TableCell>
                                {visibleColumns.map((field) => (
                                    <TableCell key={field.key} className="max-w-[300px]">
                            <span className="block truncate font-mono whitespace-nowrap">
                              {formatCellValue(row[field.key], field.type)}
                            </span>
                                    </TableCell>
                                ))}
                                {visibleSystemColumns.map((key) => (
                                    <TableCell key={key} className="text-muted-foreground font-mono whitespace-nowrap">
                                      {formatSystemFieldValue(key, row[key])}
                                    </TableCell>
                                ))}
                                <TableCell className="w-12" onClick={(e) => e.stopPropagation()}>
                                  <DropdownMenu>
                                    <DropdownMenuTrigger asChild>
                                      <Button variant="ghost" size="icon" className="size-8">
                                        <MoreHorizontal className="size-4"/>
                                      </Button>
                                    </DropdownMenuTrigger>
                                    <DropdownMenuContent align="end">
                                      <DropdownMenuItem onClick={() => handleView(row)}>
                                        <Eye className="mr-2 size-4"/>
                                        View
                                      </DropdownMenuItem>
                                      <DropdownMenuItem onClick={() => handleEdit(row)}>
                                        <Edit className="mr-2 size-4"/>
                                        Edit
                                      </DropdownMenuItem>
                                      <DropdownMenuItem onClick={() => handleDelete(row)}
                                                        className="text-destructive focus:text-destructive">
                                        <Trash2 className="mr-2 size-4"/>
                                        Delete
                                      </DropdownMenuItem>
                                    </DropdownMenuContent>
                                  </DropdownMenu>
                                </TableCell>
                              </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </div>
                )}

                {/* Pagination */}
                <div className="flex items-center justify-between py-4">
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-muted-foreground">
                      {selectedIds.size > 0 ? `${selectedIds.size} of ${total} selected` : `${total} rows`}
                    </span>
                    {someSelected && (
                        <Button
                            variant="destructive"
                            size="sm"
                            onClick={() => setBulkDeleteOpen(true)}
                        >
                          <Trash2 className="mr-2 size-4"/>
                          Delete Selected
                        </Button>
                    )}
                  </div>
                  <div className="flex items-center gap-4">
                    <Select value={limit.toString()} onValueChange={(v) => {
                      setLimit(parseInt(v));
                      setPage(1);
                    }}>
                      <SelectTrigger className="w-[100px]"><SelectValue/></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="25">25</SelectItem>
                        <SelectItem value="50">50</SelectItem>
                        <SelectItem value="100">100</SelectItem>
                      </SelectContent>
                    </Select>
                    <div className="flex items-center gap-2">
                      <Button variant="outline" size="sm" onClick={() => setPage(page - 1)}
                              disabled={page === 1}>Previous</Button>
                      <span className="text-sm">{page} / {totalPages || 1}</span>
                      <Button variant="outline" size="sm" onClick={() => setPage(page + 1)}
                              disabled={page >= totalPages}>Next</Button>
                    </div>
                  </div>
                </div>
              </>
          )}

          {/* View Tab Content */}
          {tabs.filter(t => t.type === 'view' && t.id === activeTabId).map((tab) => (
              <div key={tab.id} className="flex-1 overflow-auto">
                <div className="max-w-2xl space-y-4">
                  {/* Actions */}
                  <div className="flex items-center gap-2">
                    <Button variant="outline" size="sm" onClick={() => tab.data && handleEdit(tab.data)}>
                      <Edit className="mr-2 size-4"/>
                      Edit
                    </Button>
                    <Button variant="outline" size="sm" onClick={() => tab.data && handleDelete(tab.data)}
                            className="text-destructive hover:text-destructive">
                      <Trash2 className="mr-2 size-4"/>
                      Delete
                    </Button>
                  </div>
                  {/* Fields table */}
                  <div className="rounded-md border overflow-hidden">
                    <Table>
                      <TableBody>
                        {/* Regular fields */}
                        {orderedFormFields.map((field) => (
                            <TableRow key={field.key}>
                              <TableCell className="w-1/3 font-medium text-muted-foreground bg-muted/30">
                                {formatFieldLabel(field.key)}
                              </TableCell>
                              <TableCell>
                                {field.type === 'document' ? (
                                    <pre className="text-xs bg-muted p-2 rounded overflow-auto max-h-32">
                                      {JSON.stringify(tab.data?.[field.key], null, 2)}
                                    </pre>
                                ) : (
                                    <span className="font-mono text-sm">
                                      {formatCellValue(tab.data?.[field.key], field.type) || '—'}
                                    </span>
                                )}
                              </TableCell>
                            </TableRow>
                        ))}
                        {/* System fields */}
                        {SYSTEM_FIELDS.map((key) => (
                            <TableRow key={key}>
                              <TableCell className="w-1/3 font-medium text-muted-foreground bg-muted/30">
                                {SYSTEM_FIELD_LABELS[key]}
                              </TableCell>
                              <TableCell className="font-mono text-xs text-muted-foreground">
                                {formatSystemFieldValue(key, tab.data?.[key])}
                              </TableCell>
                            </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </div>
                </div>
              </div>
          ))}

          {/* Edit/Create Tab Content */}
          {tabs.filter(t => (t.type === 'edit' || t.type === 'create') && t.id === activeTabId).map((tab) => (
              <div key={tab.id} className="flex-1 overflow-auto">
                <div className="max-w-2xl space-y-4">
                  {orderedFormFields.map((field) => (
                      <div key={field.key} className="space-y-2">
                        <label className="text-sm font-medium">
                          {formatFieldLabel(field.key)}
                          {field.required && <span className="text-destructive ml-1">*</span>}
                        </label>
                        <FieldInput
                            field={field}
                            value={tab.formData?.[field.key]}
                            onChange={(value) => updateTabFormData(tab.id, {...tab.formData, [field.key]: value})}
                            view={formConfig.field_views[field.key]}
                        />
                        {tab.fieldErrors?.[field.key] && (
                            <p className="text-sm text-destructive">{tab.fieldErrors[field.key]}</p>
                        )}
                      </div>
                  ))}
                  {/* Actions */}
                  <div className="flex items-center gap-2">
                    <Button variant="outline" onClick={() => closeTab(tab.id)}>Cancel</Button>
                    <Button
                        onClick={() => {
                          if (tab.type === 'create') {
                            createMutation.mutate({tabId: tab.id, data: tab.formData || {}});
                          } else if (tab.data) {
                            updateMutation.mutate({tabId: tab.id, dataId: tab.data._id, data: tab.formData || {}});
                          }
                        }}
                        disabled={createMutation.isPending || updateMutation.isPending}
                    >
                      {(createMutation.isPending || updateMutation.isPending) ? 'Saving...' : 'Save'}
                    </Button>
                  </div>
                </div>
              </div>
          ))}
        </div>

        {/* Delete Confirm */}
        <ConfirmDialog
            open={deleteOpen}
            onOpenChange={setDeleteOpen}
            title="Delete Record"
            description="Are you sure you want to delete this record?"
            confirmLabel="Delete"
            variant="destructive"
            onConfirm={() => deleteTargetId && deleteMutation.mutate(deleteTargetId)}
            isLoading={deleteMutation.isPending}
        />

        {/* Bulk Delete Confirm */}
        <ConfirmDialog
            open={bulkDeleteOpen}
            onOpenChange={setBulkDeleteOpen}
            title="Delete Records"
            description={`Are you sure you want to delete ${selectedIds.size} selected records?`}
            confirmLabel="Delete All"
            variant="destructive"
            onConfirm={() => bulkDeleteMutation.mutate()}
            isLoading={bulkDeleteMutation.isPending}
        />
      </div>
  );
}
