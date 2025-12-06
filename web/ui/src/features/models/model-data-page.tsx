import { useState, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import {
  Plus,
  Settings,
  Search,
  Loader2,
  Eye,
  Edit,
  Table2,
  RefreshCw,
} from 'lucide-react';

import type { ModelData } from '@/types';
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
  RecordForm,
  RecordView,
  Pagination,
} from './model-data';

export function ModelDataPage() {
  const { projectId, modelId } = useParams<{ projectId: string; modelId: string }>();

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

  // Tab management
  const {
    tabs,
    activeTabId,
    setActiveTabId,
    openTab,
    closeTab,
    closeTabsForRecord,
    updateTabFormData,
    updateTabErrors,
  } = useTabs(model?.fields);

  // Column resizing
  const { handleResizeStart, getColumnWidth } = useColumnResize({ modelId });

  // Row selection
  const { selectedIds, handleSelectRow, handleSelectAll, clearSelection, allSelected, someSelected } = useSelection(data);

  // Delete confirmation dialogs
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [bulkDeleteOpen, setBulkDeleteOpen] = useState(false);
  const [deleteTargetId, setDeleteTargetId] = useState<string | null>(null);

  // Mutations
  const { createMutation, updateMutation, deleteMutation, bulkDeleteMutation } = useModelMutations({
    projectId,
    modelId,
    onCreateSuccess: closeTab,
    onUpdateSuccess: closeTab,
    onDeleteSuccess: (dataId) => {
      setDeleteOpen(false);
      setDeleteTargetId(null);
      closeTabsForRecord(dataId);
    },
    onBulkDeleteSuccess: () => {
      setBulkDeleteOpen(false);
      clearSelection();
    },
    onValidationError: updateTabErrors,
  });

  // Handlers
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
              className="pl-9"
            />
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
      <EditorTabs className="px-0 mb-[-1px] relative z-10">
        {tabs.map((tab) => (
          <EditorTab
            key={tab.id}
            active={activeTabId === tab.id}
            onClick={() => setActiveTabId(tab.id)}
            icon={
              tab.type === 'table' ? <Table2 className="size-4" /> :
              tab.type === 'view' ? <Eye className="size-4" /> :
              tab.type === 'edit' ? <Edit className="size-4" /> :
              <Plus className="size-4" />
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
      <div className="flex-1 min-h-0 min-w-0 flex flex-col overflow-hidden">
        {/* Table Tab */}
        {activeTabId === 'table' && (
          <>
            {data.length === 0 ? (
              <div className="flex-1 flex items-center justify-center border rounded-md dark:bg-zinc-950">
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
                        <Plus className="mr-2 size-4" />
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
                onEdit={handleEdit}
                onDelete={handleDelete}
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
            onEdit={handleEdit}
            onDelete={handleDelete}
          />
        )}

        {/* Edit/Create Tab Content */}
        {(activeTab?.type === 'edit' || activeTab?.type === 'create') && (
          <RecordForm
            tab={activeTab}
            orderedFormFields={orderedFormFields}
            formConfig={formConfig}
            onFormDataChange={updateTabFormData}
            onSave={() => {
              if (activeTab.type === 'create') {
                createMutation.mutate({ tabId: activeTab.id, data: activeTab.formData || {} });
              } else if (activeTab.data) {
                updateMutation.mutate({
                  tabId: activeTab.id,
                  dataId: activeTab.data._id,
                  data: activeTab.formData || {},
                });
              }
            }}
            onCancel={() => closeTab(activeTab.id)}
            isSaving={createMutation.isPending || updateMutation.isPending}
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
    </div>
  );
}
