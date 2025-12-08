import { useState, useCallback, useEffect, useRef } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import {
  Plus,
  Settings,
  Search,
  Loader2,
  Eye,
  Table2,
  RefreshCw,
  X,
} from 'lucide-react';

import type { ModelData } from '@/types';
import { modelsApi } from '@/api';
import { useTitle } from '@/hooks';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { EmptyState } from '@/components/shared/empty-state';
import { Skeleton } from '@/components/ui/skeleton';
import { EditorTabs, EditorTab } from '@/components/ui/editor-tabs';

import {
  useModelData,
  useModelMutations,
  useTabs,
  useColumnResize,
  useSelection,
  DataTable,
  FilterPopover,
  ActiveFilterBadges,
  RecordView,
  RecordCreate,
  Pagination,
} from './model-data';
import type { ModelField } from '@/types';

// Normalize form data before sending to API
function normalizeFormData(
  formData: Record<string, unknown>,
  fields: ModelField[]
): Record<string, unknown> {
  const result: Record<string, unknown> = {};

  for (const field of fields) {
    const value = formData[field.key];

    switch (field.type) {
      case 'ref':
      case 'file':
        // Convert empty strings and undefined to null
        result[field.key] = (value === '' || value === undefined) ? null : value;
        break;
      case 'bool':
        // Ensure boolean value (undefined/null -> false)
        result[field.key] = !!value;
        break;
      default:
        result[field.key] = value;
    }
  }

  return result;
}

export function ModelDataPage() {
  const { projectId, modelId } = useParams<{ projectId: string; modelId: string }>();

  // Load all models for ref field lookups
  const { data: allModels = [] } = useQuery({
    queryKey: ['models', projectId],
    queryFn: () => modelsApi.list(projectId!),
    enabled: !!projectId,
  });

  // Data and model state
  const {
    model,
    data,
    isFetching,
    isLoading,
    page,
    setPage,
    limit,
    setLimit,
    total,
    totalPages,
    searchInput,
    setSearchInput,
    searchQuery,
    searchInputRef,
    hasSearchable,
    sortField,
    sortOrder,
    handleSort,
    activeFilters,
    setActiveFilters,
    filterableFields,
    tableConfig,
    formConfig,
    visibleColumns,
    visibleSystemColumns,
    orderedFormFields,
    refetch,
  } = useModelData({ projectId, modelId });

  useTitle(model?.name);

  // Tab management
  const {
    tabs,
    activeTabId,
    setActiveTabId,
    openTab,
    closeTab,
    closeTabsForRecord,
    updateTabErrors,
    markTabAsChanged,
    clearTabChanges,
    resetTabs,
    hasUnsavedChanges,
  } = useTabs(model?.fields);

  // Column resizing
  const { handleResizeStart, getColumnWidth } = useColumnResize({ modelId });

  // Row selection
  const { selectedIds, handleSelectRow, handleSelectAll, clearSelection, allSelected, someSelected } = useSelection(data);

  // Delete confirmation dialogs
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [bulkDeleteOpen, setBulkDeleteOpen] = useState(false);
  const [deleteTargetId, setDeleteTargetId] = useState<string | null>(null);

  // Unsaved changes confirmation dialog
  const [unsavedChangesDialogOpen, setUnsavedChangesDialogOpen] = useState(false);
  const prevModelIdRef = useRef(modelId);

  // Reset tabs when model changes
  useEffect(() => {
    if (prevModelIdRef.current !== modelId) {
      // Model changed - check for unsaved changes
      if (hasUnsavedChanges) {
        // Show confirmation dialog
        setUnsavedChangesDialogOpen(true);
      } else {
        // No unsaved changes - reset tabs immediately
        resetTabs();
      }
      prevModelIdRef.current = modelId;
    }
  }, [modelId, hasUnsavedChanges, resetTabs]);

  // Handle unsaved changes dialog confirmation
  const handleDiscardChanges = useCallback(() => {
    resetTabs();
    setUnsavedChangesDialogOpen(false);
  }, [resetTabs]);

  // Clear selection when switching tabs or models
  useEffect(() => {
    clearSelection();
  }, [activeTabId, modelId, clearSelection]);

  // Field errors for inline editing in view tabs
  const [viewFieldErrors, setViewFieldErrors] = useState<Record<string, Record<string, string>>>({});

  // Field errors for create tabs
  const [createFieldErrors, setCreateFieldErrors] = useState<Record<string, Record<string, string>>>({});

  // Mutations
  const { createMutation, updateMutation, deleteMutation, bulkDeleteMutation } = useModelMutations({
    projectId,
    modelId,
    onCreateSuccess: (tabId, closeAfterSave) => {
      // Clear create field errors on success
      setCreateFieldErrors(prev => {
        const next = { ...prev };
        delete next[tabId];
        return next;
      });
      // Clear hasChanges flag
      clearTabChanges(tabId);
      if (closeAfterSave) {
        closeTab(tabId);
        setActiveTabId('table');
      }
    },
    onUpdateSuccess: (tabId, closeAfterSave, dataId) => {
      // Clear field errors on success
      if (dataId) {
        setViewFieldErrors(prev => {
          const next = { ...prev };
          delete next[dataId];
          return next;
        });
      }
      // Clear hasChanges flag
      clearTabChanges(tabId);
      if (closeAfterSave && dataId) {
        closeTabsForRecord(dataId);
      }
    },
    onDeleteSuccess: (dataId) => {
      setDeleteOpen(false);
      setDeleteTargetId(null);
      closeTabsForRecord(dataId);
    },
    onBulkDeleteSuccess: () => {
      setBulkDeleteOpen(false);
      clearSelection();
    },
    onValidationError: (tabId, errors) => {
      const tab = tabs.find(t => t.id === tabId);
      if (tab?.type === 'create') {
        // For create tabs, store errors by tab ID
        setCreateFieldErrors(prev => ({ ...prev, [tabId]: errors }));
      } else if (tab?.data?._id) {
        // For view tabs, store errors by data ID
        setViewFieldErrors(prev => ({ ...prev, [tab.data!._id]: errors }));
      }
      updateTabErrors(tabId, errors);
    },
  });

  // Handlers
  const handleCreate = useCallback(() => {
    openTab('create');
  }, [openTab]);

  const handleView = useCallback((data: ModelData) => {
    openTab('view', data);
  }, [openTab]);

  const handleDelete = useCallback((data: ModelData) => {
    setDeleteTargetId(data._id);
    setDeleteOpen(true);
  }, []);

  const handleRemoveFilter = useCallback((index: number) => {
    setActiveFilters(prev => prev.filter((_, i) => i !== index));
    setPage(1);
  }, [setActiveFilters, setPage]);

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

  const activeTab = tabs.find(t => t.id === activeTabId);

  return (
    <div className="h-full flex flex-col overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{model.name}</h1>
          <p className="text-muted-foreground">{total} records</p>
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

      {/* Toolbar: Search, Filters */}
      <div className="flex flex-wrap items-center justify-between gap-2 mb-4">
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="icon"
            onClick={() => refetch()}
            disabled={isFetching}
          >
            <RefreshCw className={`size-4 ${isFetching ? 'animate-spin' : ''}`} />
          </Button>
          <FilterPopover
            activeFilters={activeFilters}
            filterableFields={filterableFields}
            modelFields={model.fields}
            onFiltersChange={setActiveFilters}
            onPageReset={() => setPage(1)}
            projectId={projectId}
            models={allModels}
          />
        </div>

        {/* Search */}
        {hasSearchable && (
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
              className={`pl-9 ${searchInput ? 'pr-8' : ''}`}
            />
            {searchInput && (
              <Button
                variant="ghost"
                size="icon"
                className="absolute right-1 top-1/2 -translate-y-1/2 h-7 w-7"
                onClick={() => setSearchInput('')}
              >
                <X className="size-4" />
              </Button>
            )}
          </div>
        )}
      </div>

      {/* Active Filter Badges */}
      <ActiveFilterBadges
        activeFilters={activeFilters}
        modelFields={model.fields}
        onRemoveFilter={handleRemoveFilter}
      />

      {/* Tabs Bar */}
      <EditorTabs className="px-0 mb-0 relative z-10">
        {tabs.map((tab) => (
          <EditorTab
            key={tab.id}
            active={activeTabId === tab.id}
            onClick={() => setActiveTabId(tab.id)}
            icon={
              tab.type === 'table' ? <Table2 className="size-4" /> :
              tab.type === 'view' ? <Eye className="size-4" /> :
              <Plus className="size-4" />
            }
            onClose={tab.id !== 'table' ? () => closeTab(tab.id) : undefined}
            className={tab.type === 'table' ? "bg-background" : ""}
            dirty={tab.hasChanges}
          >
            {tab.title}
          </EditorTab>
        ))}
      </EditorTabs>

      {/* Tab Content */}
      <div className="min-h-0 min-w-0 flex flex-col overflow-hidden">
        {/* Table Tab */}
        {activeTabId === 'table' && (
            <>
              {data.length === 0 ? (
                  <div className="flex-1 flex items-center justify-center border rounded-xl rounded-t-none bg-background">
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
                  <DataTable
                      data={data}
                      visibleColumns={visibleColumns}
                      visibleSystemColumns={visibleSystemColumns}
                      tableConfig={tableConfig}
                      sortField={sortField}
                      sortOrder={sortOrder}
                      selectedIds={selectedIds}
                      allSelected={allSelected}
                      getColumnWidth={getColumnWidth}
                      onSort={handleSort}
                      onSelectRow={handleSelectRow}
                      onSelectAll={handleSelectAll}
                      onView={handleView}
                      onEdit={handleView}
                      onDelete={handleDelete}
                      onDeleteSelected={() => setBulkDeleteOpen(true)}
                      onResizeStart={handleResizeStart}
                  />
              )}

              {/* Pagination */}
              <Pagination
                  page={page}
                  totalPages={totalPages}
                  total={total}
                  limit={limit}
                  selectedCount={selectedIds.size}
                  onPageChange={setPage}
                  onLimitChange={setLimit}
                  onBulkDelete={someSelected ? () => setBulkDeleteOpen(true) : undefined}
              />
            </>
        )}

        {/* View Tab Content */}
        {activeTab?.type === 'view' && activeTab.data && (
          <RecordView
            data={activeTab.data}
            orderedFormFields={orderedFormFields}
            formConfig={formConfig}
            onSave={(formData, closeAfterSave) => {
              const normalizedData = model?.fields
                ? normalizeFormData(formData, model.fields)
                : formData;

              updateMutation.mutate({
                tabId: activeTab.id,
                dataId: activeTab.data!._id,
                data: normalizedData,
                closeAfterSave,
              });
            }}
            onDelete={handleDelete}
            isSaving={updateMutation.isPending}
            fieldErrors={viewFieldErrors[activeTab.data._id]}
            projectId={projectId}
            models={allModels}
            onFormChange={(hasChanges) => markTabAsChanged(activeTab.id, hasChanges)}
          />
        )}

        {/* Create Tab Content */}
        {activeTab?.type === 'create' && (
          <RecordCreate
            orderedFormFields={orderedFormFields}
            formConfig={formConfig}
            onSave={(formData, closeAfterSave) => {
              const normalizedData = model?.fields
                ? normalizeFormData(formData, model.fields)
                : formData;

              createMutation.mutate({
                tabId: activeTab.id,
                data: normalizedData,
                closeAfterSave,
              });
            }}
            onCancel={() => closeTab(activeTab.id)}
            isSaving={createMutation.isPending}
            fieldErrors={createFieldErrors[activeTab.id]}
            projectId={projectId}
            models={allModels}
            onFormChange={(hasChanges) => markTabAsChanged(activeTab.id, hasChanges)}
          />
        )}
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
        onConfirm={() => bulkDeleteMutation.mutate(Array.from(selectedIds))}
        isLoading={bulkDeleteMutation.isPending}
      />

      {/* Unsaved Changes Confirm */}
      <ConfirmDialog
        open={unsavedChangesDialogOpen}
        onOpenChange={setUnsavedChangesDialogOpen}
        title="Unsaved Changes"
        description="You have unsaved changes. Do you want to discard them?"
        confirmLabel="Discard"
        variant="destructive"
        onConfirm={handleDiscardChanges}
      />
    </div>
  );
}
