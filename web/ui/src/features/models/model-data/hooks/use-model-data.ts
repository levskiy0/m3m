import { useState, useMemo, useEffect, useRef } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

import { modelsApi } from '@/api';
import { ApiValidationError } from '@/api/client';
import type { FilterCondition, TableConfig, FormConfig, ModelField } from '@/types';
import { isSystemField, type SystemField } from '../constants';
import type { ActiveFilter } from '../types';

interface UseModelDataOptions {
  projectId: string | undefined;
  modelId: string | undefined;
}

export function useModelData({ projectId, modelId }: UseModelDataOptions) {
  const queryClient = useQueryClient();

  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(25);
  const [searchInput, setSearchInput] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [sortField, setSortField] = useState<string | null>(null);
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc');
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
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

  const isLoading = modelLoading || (dataLoading && !dataResponse);
  const data = dataResponse?.data || [];
  const totalPages = dataResponse?.totalPages || 1;
  const total = dataResponse?.total || 0;
  const hasSearchable = (tableConfig.searchable?.length || 0) > 0;

  return {
    // Model
    model,
    modelLoading,

    // Data
    data,
    dataLoading,
    isFetching,
    isLoading,

    // Pagination
    page,
    setPage,
    limit,
    setLimit,
    total,
    totalPages,

    // Search
    searchInput,
    setSearchInput,
    searchQuery,
    searchInputRef,
    hasSearchable,

    // Sort
    sortField,
    sortOrder,
    handleSort,

    // Filters
    activeFilters,
    setActiveFilters,
    filterableFields,

    // Configs
    tableConfig,
    formConfig,

    // Columns
    visibleColumns,
    visibleSystemColumns,
    orderedFormFields,

    // Query client for mutations
    queryClient,
  };
}

interface UseModelMutationsOptions {
  projectId: string | undefined;
  modelId: string | undefined;
  onCreateSuccess?: (tabId: string) => void;
  onUpdateSuccess?: (tabId: string) => void;
  onDeleteSuccess?: (dataId: string) => void;
  onBulkDeleteSuccess?: (count: number) => void;
  onValidationError?: (tabId: string, errors: Record<string, string>) => void;
}

export function useModelMutations({
  projectId,
  modelId,
  onCreateSuccess,
  onUpdateSuccess,
  onDeleteSuccess,
  onBulkDeleteSuccess,
  onValidationError,
}: UseModelMutationsOptions) {
  const queryClient = useQueryClient();

  const createMutation = useMutation({
    mutationFn: (params: { tabId: string; data: Record<string, unknown> }) =>
      modelsApi.createData(projectId!, modelId!, params.data),
    onSuccess: (_, { tabId }) => {
      queryClient.invalidateQueries({ queryKey: ['model-data', projectId, modelId] });
      onCreateSuccess?.(tabId);
      toast.success('Record created');
    },
    onError: (err, { tabId }) => {
      if (err instanceof ApiValidationError) {
        const errors: Record<string, string> = {};
        err.details.forEach(d => {
          errors[d.field] = d.message;
        });
        onValidationError?.(tabId, errors);
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
      onUpdateSuccess?.(tabId);
      toast.success('Record updated');
    },
    onError: (err, { tabId }) => {
      if (err instanceof ApiValidationError) {
        const errors: Record<string, string> = {};
        err.details.forEach(d => {
          errors[d.field] = d.message;
        });
        onValidationError?.(tabId, errors);
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
      onDeleteSuccess?.(dataId);
      toast.success('Record deleted');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to delete record');
    },
  });

  const bulkDeleteMutation = useMutation({
    mutationFn: (ids: string[]) =>
      modelsApi.bulkDeleteData(projectId!, modelId!, ids),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['model-data', projectId, modelId] });
      onBulkDeleteSuccess?.(result.deleted_count);
      toast.success(`${result.deleted_count} records deleted`);
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to delete records');
    },
  });

  return {
    createMutation,
    updateMutation,
    deleteMutation,
    bulkDeleteMutation,
  };
}
