import { useState, useEffect, useRef } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2, GripVertical, Save, Table, Database, FileText } from 'lucide-react';
import { toast } from 'sonner';
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core';
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';

import { modelsApi } from '@/api';
import { FIELD_TYPES } from '@/lib/constants';
import type { ModelField, FieldType, TableConfig, FormConfig, Model } from '@/types';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import { Field, FieldLabel } from '@/components/ui/field';
import { Skeleton } from '@/components/ui/skeleton';
import { EditorTabs, EditorTab } from '@/components/ui/editor-tabs';
import { TableConfigEditor } from './table-config-editor';
import { FormConfigEditor } from './form-config-editor';

// Component for type-aware default value input
interface DefaultValueInputProps {
  type: FieldType;
  value: unknown;
  onChange: (value: unknown) => void;
}

function DefaultValueInput({ type, value, onChange }: DefaultValueInputProps) {
  switch (type) {
    case 'bool':
      return (
        <Select
          value={value === undefined ? '__none__' : value ? 'true' : 'false'}
          onValueChange={(v) => {
            if (v === '__none__') onChange(undefined);
            else onChange(v === 'true');
          }}
        >
          <SelectTrigger>
            <SelectValue placeholder="No default" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="__none__">No default</SelectItem>
            <SelectItem value="true">true</SelectItem>
            <SelectItem value="false">false</SelectItem>
          </SelectContent>
        </Select>
      );

    case 'number':
      return (
        <Input
          type="number"
          step="1"
          value={value !== undefined ? String(value) : ''}
          onChange={(e) => {
            const v = e.target.value;
            if (v === '') onChange(undefined);
            else onChange(parseInt(v, 10));
          }}
          placeholder="No default"
        />
      );

    case 'float':
      return (
        <Input
          type="number"
          step="0.01"
          value={value !== undefined ? String(value) : ''}
          onChange={(e) => {
            const v = e.target.value;
            if (v === '') onChange(undefined);
            else onChange(parseFloat(v));
          }}
          placeholder="No default"
        />
      );

    case 'date':
      return (
        <Select
          value={value === '$now' ? '$now' : value !== undefined ? '__custom__' : '__none__'}
          onValueChange={(v) => {
            if (v === '__none__') onChange(undefined);
            else if (v === '$now') onChange('$now');
          }}
        >
          <SelectTrigger>
            <SelectValue placeholder="No default" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="__none__">No default</SelectItem>
            <SelectItem value="$now">$now</SelectItem>
          </SelectContent>
        </Select>
      );

    case 'datetime':
      return (
        <Select
          value={value === '$now' ? '$now' : value !== undefined ? '__custom__' : '__none__'}
          onValueChange={(v) => {
            if (v === '__none__') onChange(undefined);
            else if (v === '$now') onChange('$now');
          }}
        >
          <SelectTrigger>
            <SelectValue placeholder="No default" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="__none__">No default</SelectItem>
            <SelectItem value="$now">$now</SelectItem>
          </SelectContent>
        </Select>
      );

    case 'string':
    case 'text':
    default:
      return (
        <Input
          value={(value as string) || ''}
          onChange={(e) => onChange(e.target.value || undefined)}
          placeholder="No default"
        />
      );
  }
}

interface SortableSchemaFieldProps {
  id: string;
  field: ModelField;
  onUpdate: (updates: Partial<ModelField>) => void;
  onRemove: () => void;
  models: Model[];
  currentModelId?: string;
}

function SortableSchemaField({ id, field, onUpdate, onRemove, models, currentModelId }: SortableSchemaFieldProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  // Filter out current model from ref options
  const availableModels = models.filter(m => m.id !== currentModelId);

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`flex items-start gap-4 p-4 border rounded-lg bg-background ${
        isDragging ? 'opacity-50 shadow-lg z-50' : ''
      }`}
    >
      <div
        {...attributes}
        {...listeners}
        className="cursor-grab active:cursor-grabbing text-muted-foreground hover:text-foreground transition-colors"
      >
        <GripVertical className="size-5" />
      </div>
      <div className="flex-1 grid gap-4 md:grid-cols-4">
        <Field>
          <FieldLabel>Key</FieldLabel>
          <Input
            value={field.key}
            onChange={(e) =>
              onUpdate({
                key: e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, '_'),
              })
            }
            placeholder="field_name"
          />
        </Field>
        <Field>
          <FieldLabel>Type</FieldLabel>
          <Select
            value={field.type}
            onValueChange={(v) => {
              const updates: Partial<ModelField> = { type: v as FieldType };
              // Clear refModel when switching away from ref type
              if (v !== 'ref') {
                updates.refModel = undefined;
              }
              onUpdate(updates);
            }}
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {FIELD_TYPES.map((type) => (
                <SelectItem key={type.value} value={type.value}>
                  {type.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </Field>
        {field.type === 'ref' ? (
          <Field>
            <FieldLabel>Reference Model</FieldLabel>
            <Select
              value={field.refModel || ''}
              onValueChange={(v) => onUpdate({ refModel: v })}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select model" />
              </SelectTrigger>
              <SelectContent>
                {availableModels.map((m) => (
                  <SelectItem key={m.id} value={m.slug}>
                    {m.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </Field>
        ) : (
          <Field>
            <FieldLabel>Default Value</FieldLabel>
            <DefaultValueInput
              type={field.type}
              value={field.default_value}
              onChange={(value) => onUpdate({ default_value: value })}
            />
          </Field>
        )}
        <Field>
          <FieldLabel>Required</FieldLabel>
          <div className="flex items-center h-10">
            <Switch
              checked={field.required}
              onCheckedChange={(checked) => onUpdate({ required: checked })}
            />
          </div>
        </Field>
      </div>
      <Button
        variant="ghost"
        size="icon"
        className="mt-6"
        onClick={onRemove}
      >
        <Trash2 className="size-4 text-destructive" />
      </Button>
    </div>
  );
}

export function ModelSchemaPage() {
  const { projectId, modelId } = useParams<{ projectId: string; modelId: string }>();
  const queryClient = useQueryClient();

  const [fields, setFields] = useState<ModelField[]>([]);
  const fieldIdsRef = useRef<string[]>([]);
  const idCounterRef = useRef(0);

  const generateFieldId = () => {
    idCounterRef.current += 1;
    return `field-${idCounterRef.current}-${Date.now()}`;
  };

  const [tableConfig, setTableConfig] = useState<TableConfig>({
    columns: [],
    filters: [],
    sort_columns: [],
    searchable: [],
  });
  const [formConfig, setFormConfig] = useState<FormConfig>({
    field_order: [],
    hidden_fields: [],
    field_views: {},
  });
  const [hasChanges, setHasChanges] = useState(false);
  const [activeTab, setActiveTab] = useState<'schema' | 'table' | 'form'>('schema');

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  const { data: model, isLoading } = useQuery({
    queryKey: ['model', projectId, modelId],
    queryFn: () => modelsApi.get(projectId!, modelId!),
    enabled: !!projectId && !!modelId,
  });

  // Fetch all models for ref field selection
  const { data: allModels = [] } = useQuery({
    queryKey: ['models', projectId],
    queryFn: () => modelsApi.list(projectId!),
    enabled: !!projectId,
  });

  useEffect(() => {
    if (model) {
      setFields(model.fields);
      // Generate stable IDs for loaded fields
      fieldIdsRef.current = model.fields.map(() => generateFieldId());

      // Default table config
      const defaultTableConfig: TableConfig = {
        columns: model.fields.map(f => f.key),
        filters: [],
        sort_columns: model.fields.map(f => f.key),
        searchable: model.fields.filter(f => ['string', 'text'].includes(f.type)).map(f => f.key),
      };

      // Merge with saved config, ensuring all arrays exist
      setTableConfig({
        columns: model.table_config?.columns ?? defaultTableConfig.columns,
        filters: model.table_config?.filters ?? defaultTableConfig.filters,
        sort_columns: model.table_config?.sort_columns ?? defaultTableConfig.sort_columns,
        searchable: model.table_config?.searchable ?? defaultTableConfig.searchable,
      });

      // Default form config
      const defaultFormConfig: FormConfig = {
        field_order: model.fields.map(f => f.key),
        hidden_fields: [],
        field_views: {},
      };

      setFormConfig({
        field_order: model.form_config?.field_order ?? defaultFormConfig.field_order,
        hidden_fields: model.form_config?.hidden_fields ?? defaultFormConfig.hidden_fields,
        field_views: model.form_config?.field_views ?? defaultFormConfig.field_views,
      });

      setHasChanges(false);
    }
  }, [model]);

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

  const addField = () => {
    const newField: ModelField = {
      key: `field_${fields.length + 1}`,
      type: 'string',
      required: false,
    };
    const newFields = [...fields, newField];
    setFields(newFields);
    fieldIdsRef.current = [...fieldIdsRef.current, generateFieldId()];

    // Auto-add to configs
    setTableConfig(prev => ({
      ...prev,
      columns: [...prev.columns, newField.key],
      sort_columns: [...prev.sort_columns, newField.key],
      searchable: newField.type === 'string' || newField.type === 'text'
        ? [...(prev.searchable || []), newField.key]
        : prev.searchable,
    }));
    setFormConfig(prev => ({
      ...prev,
      field_order: [...prev.field_order, newField.key],
    }));

    setHasChanges(true);
  };

  const updateField = (index: number, updates: Partial<ModelField>) => {
    const oldKey = fields[index].key;
    const newFields = [...fields];
    newFields[index] = { ...newFields[index], ...updates };
    setFields(newFields);

    // Update key references in configs if key changed
    if (updates.key && updates.key !== oldKey) {
      setTableConfig(prev => ({
        columns: prev.columns.map(k => k === oldKey ? updates.key! : k),
        filters: prev.filters.map(k => k === oldKey ? updates.key! : k),
        sort_columns: prev.sort_columns.map(k => k === oldKey ? updates.key! : k),
        searchable: (prev.searchable || []).map(k => k === oldKey ? updates.key! : k),
      }));
      setFormConfig(prev => {
        const newFieldViews = { ...prev.field_views };
        if (oldKey in newFieldViews) {
          newFieldViews[updates.key!] = newFieldViews[oldKey];
          delete newFieldViews[oldKey];
        }
        return {
          field_order: prev.field_order.map(k => k === oldKey ? updates.key! : k),
          hidden_fields: prev.hidden_fields.map(k => k === oldKey ? updates.key! : k),
          field_views: newFieldViews,
        };
      });
    }

    setHasChanges(true);
  };

  const removeField = (index: number) => {
    const keyToRemove = fields[index].key;
    setFields(fields.filter((_, i) => i !== index));
    fieldIdsRef.current = fieldIdsRef.current.filter((_, i) => i !== index);

    // Remove from configs
    setTableConfig(prev => ({
      columns: prev.columns.filter(k => k !== keyToRemove),
      filters: prev.filters.filter(k => k !== keyToRemove),
      sort_columns: prev.sort_columns.filter(k => k !== keyToRemove),
      searchable: (prev.searchable || []).filter(k => k !== keyToRemove),
    }));
    setFormConfig(prev => {
      const newFieldViews = { ...prev.field_views };
      delete newFieldViews[keyToRemove];
      return {
        field_order: prev.field_order.filter(k => k !== keyToRemove),
        hidden_fields: prev.hidden_fields.filter(k => k !== keyToRemove),
        field_views: newFieldViews,
      };
    });

    setHasChanges(true);
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;

    if (over && active.id !== over.id) {
      const oldIndex = fieldIdsRef.current.indexOf(active.id as string);
      const newIndex = fieldIdsRef.current.indexOf(over.id as string);

      if (oldIndex === -1 || newIndex === -1) return;

      const newFields = arrayMove(fields, oldIndex, newIndex);
      fieldIdsRef.current = arrayMove(fieldIdsRef.current, oldIndex, newIndex);
      setFields(newFields);

      // Update configs with new field order
      const newKeys = newFields.map(f => f.key);
      setTableConfig(prev => ({
        ...prev,
        columns: newKeys.filter(k => prev.columns.includes(k)),
        sort_columns: newKeys.filter(k => prev.sort_columns.includes(k)),
      }));
      setFormConfig(prev => ({
        ...prev,
        field_order: newKeys,
      }));

      setHasChanges(true);
    }
  };

  const handleTableConfigChange = (config: TableConfig) => {
    setTableConfig(config);
    setHasChanges(true);
  };

  const handleFormConfigChange = (config: FormConfig) => {
    setFormConfig(config);
    setHasChanges(true);
  };

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
          <Button
            onClick={() => updateMutation.mutate()}
            disabled={!hasChanges || updateMutation.isPending}
          >
            <Save className="mr-2 size-4" />
            {updateMutation.isPending ? 'Saving...' : 'Save'}
          </Button>
        </div>
      </div>

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
          icon={<FileText className="size-4" />}>
          Form
        </EditorTab>
      </EditorTabs>

      {activeTab === 'schema' && (
        <Card className="rounded-t-none !mt-0">
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Fields</CardTitle>
                  <CardDescription>
                    Define the fields for your model schema
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
                  <SortableContext
                    items={fieldIdsRef.current}
                    strategy={verticalListSortingStrategy}
                  >
                    <div className="space-y-4">
                      {fields.map((field, index) => (
                        <SortableSchemaField
                          key={fieldIdsRef.current[index] || index}
                          id={fieldIdsRef.current[index] || String(index)}
                          field={field}
                          onUpdate={(updates) => updateField(index, updates)}
                          onRemove={() => removeField(index)}
                          models={allModels}
                          currentModelId={modelId}
                        />
                      ))}
                    </div>
                  </SortableContext>
                </DndContext>
              )}
            </CardContent>
          </Card>
      )}

      {activeTab === 'table' && (
        <TableConfigEditor
          fields={fields}
          config={tableConfig}
          onChange={handleTableConfigChange}
        />
      )}

      {activeTab === 'form' && (
        <FormConfigEditor
          fields={fields}
          config={formConfig}
          onChange={handleFormConfigChange}
        />
      )}
    </div>
  );
}
