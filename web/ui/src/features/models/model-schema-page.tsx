/**
 * ModelSchemaPage
 * Schema editor for model fields, table config, and form config
 */

import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Save, Table, Database, FileText, Type, Settings2, Asterisk } from 'lucide-react';
import { toast } from 'sonner';
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core';
import {
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';

import { modelsApi } from '@/api';
import { useTitle } from '@/hooks';
import { Button } from '@/components/ui/button';
import { LoadingButton } from '@/components/ui/loading-button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { EditorTabs, EditorTab } from '@/components/ui/editor-tabs';

import {
  SortableSchemaField,
  TableConfigEditor,
  FormConfigEditor,
  useSchemaEditor,
} from './schema';

type ActiveTab = 'schema' | 'table' | 'form';

export function ModelSchemaPage() {
  const { projectId, modelId } = useParams<{ projectId: string; modelId: string }>();
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState<ActiveTab>('schema');

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  // Load model
  const { data: model, isLoading } = useQuery({
    queryKey: ['model', projectId, modelId],
    queryFn: () => modelsApi.get(projectId!, modelId!),
    enabled: !!projectId && !!modelId,
  });

  useTitle(model ? `${model.name} Schema` : null);

  // Load all models for ref field selection
  const { data: allModels = [] } = useQuery({
    queryKey: ['models', projectId],
    queryFn: () => modelsApi.list(projectId!),
    enabled: !!projectId,
  });

  // Schema editor state
  const {
    fields,
    tableConfig,
    formConfig,
    hasChanges,
    hasValidationErrors,
    fieldIds,
    addField,
    updateField,
    removeField,
    handleDragEnd,
    updateTableConfig,
    updateFormConfig,
    resetState,
    setHasChanges,
  } = useSchemaEditor({
    initialFields: model?.fields || [],
    initialTableConfig: model?.table_config,
    initialFormConfig: model?.form_config,
  });

  // Reset state when model changes
  useEffect(() => {
    if (model) {
      resetState(model);
    }
  }, [model, resetState]);

  // Save mutation
  const updateMutation = useMutation({
    mutationFn: () => modelsApi.update(projectId!, modelId!, {
      fields,
      table_config: tableConfig,
      form_config: formConfig,
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['model', projectId, modelId] });
      queryClient.invalidateQueries({ queryKey: ['models', projectId] });
      setHasChanges(false);
      toast.success('Model saved');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to save model');
    },
  });

  if (isLoading) {
    return (
      <div className="space-y-4 max-w-4xl">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-64" />
      </div>
    );
  }

  if (!model) {
    return <div>Model not found</div>;
  }

  return (
    <div className="space-y-4 max-w-4xl">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{model.name}</h1>
          <p className="text-muted-foreground">
            Configure schema, table view, and form settings
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button asChild variant="outline">
            <Link to={`/projects/${projectId}/models/${modelId}/data`}>
              <Table className="mr-2 size-4" />
              View Data
            </Link>
          </Button>
          <LoadingButton
            onClick={() => updateMutation.mutate()}
            disabled={!hasChanges || hasValidationErrors}
            loading={updateMutation.isPending}
          >
            <Save className="size-4" />
            Save
          </LoadingButton>
        </div>
      </div>

      {/* Tabs */}
      <EditorTabs className="px-0 mb-[-1px] relative z-10">
        <EditorTab
          active={activeTab === 'schema'}
          onClick={() => setActiveTab('schema')}
          icon={<Database className="size-4" />}
        >
          Schema
        </EditorTab>
        <EditorTab
          active={activeTab === 'table'}
          onClick={() => setActiveTab('table')}
          icon={<Table className="size-4" />}
        >
          Table View
        </EditorTab>
        <EditorTab
          active={activeTab === 'form'}
          onClick={() => setActiveTab('form')}
          icon={<FileText className="size-4" />}
        >
          Form
        </EditorTab>
      </EditorTabs>

      {/* Schema Tab */}
      {activeTab === 'schema' && (
        <Card className="rounded-t-none !mt-0">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Fields</CardTitle>
                <CardDescription>
                  Define the fields for your model schema. Drag rows to reorder.
                </CardDescription>
              </div>
              <Button variant="outline" size="sm" onClick={addField}>
                <Plus className="mr-2 size-4" />
                Add Field
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            {fields.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                No fields defined. Add a field to get started.
              </div>
            ) : (
              <DndContext
                sensors={sensors}
                collisionDetection={closestCenter}
                onDragEnd={handleDragEnd}
              >
                <div className="border rounded-lg overflow-hidden">
                  <table className="w-full text-sm">
                    <thead className="bg-muted/50">
                      <tr>
                        <th className="w-10 p-3"></th>
                        <th className="text-left font-medium p-3 w-50">
                          <div className="flex items-center gap-1.5">
                            <span>Key</span>
                          </div>
                        </th>
                        <th className="text-left font-medium p-3 w-[200px]">
                          <div className="flex items-center gap-1.5">
                            <Type className="size-4" />
                            <span>Type</span>
                          </div>
                        </th>
                        <th className="text-left font-medium p-3 w-[200px]">
                          <div className="flex items-center gap-1.5">
                            <Settings2 className="size-4" />
                            <span>Config</span>
                          </div>
                        </th>
                        <th className="text-center font-medium p-3 w-24">
                          <div className="flex items-center justify-center gap-1.5">
                            <Asterisk className="size-4" />
                          </div>
                        </th>
                        <th className="w-10 p-3"></th>
                      </tr>
                    </thead>
                    <SortableContext
                      items={fieldIds}
                      strategy={verticalListSortingStrategy}
                    >
                      <tbody>
                        {fields.map((field, index) => (
                          <SortableSchemaField
                            key={fieldIds[index] || index}
                            id={fieldIds[index] || String(index)}
                            field={field}
                            onUpdate={(updates) => updateField(index, updates)}
                            onRemove={() => removeField(index)}
                            models={allModels}
                            currentModelId={modelId}
                          />
                        ))}
                      </tbody>
                    </SortableContext>
                  </table>
                </div>
              </DndContext>
            )}
          </CardContent>
        </Card>
      )}

      {/* Table Config Tab */}
      {activeTab === 'table' && (
        <TableConfigEditor
          fields={fields}
          config={tableConfig}
          onChange={updateTableConfig}
        />
      )}

      {/* Form Config Tab */}
      {activeTab === 'form' && (
        <FormConfigEditor
          fields={fields}
          config={formConfig}
          onChange={updateFormConfig}
        />
      )}
    </div>
  );
}
