import { GripVertical, Eye, EyeOff } from 'lucide-react';
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

import type { ModelField, FormConfig, FieldView } from '@/types';
import { FIELD_VIEWS } from '@/lib/constants';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

interface SortableFieldItemProps {
  field: ModelField;
  isHidden: boolean;
  currentView: string;
  availableViews: { value: string; label: string }[];
  onToggleHidden: () => void;
  onViewChange: (view: FieldView) => void;
}

function SortableFieldItem({
  field,
  isHidden,
  currentView,
  availableViews,
  onToggleHidden,
  onViewChange,
}: SortableFieldItemProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: field.key });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`flex items-center gap-3 p-3 border rounded-lg transition-colors ${
        isHidden ? 'bg-muted/30 opacity-60' : 'bg-background'
      } ${isDragging ? 'opacity-50 shadow-lg z-50' : ''}`}
    >
      <div
        {...attributes}
        {...listeners}
        className="cursor-grab active:cursor-grabbing text-muted-foreground hover:text-foreground transition-colors"
      >
        <GripVertical className="size-5" />
      </div>

      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="font-mono text-sm font-medium">{field.key}</span>
          <Badge variant="outline" className="text-xs">
            {field.type}
          </Badge>
          {field.required && (
            <Badge variant="secondary" className="text-xs">
              required
            </Badge>
          )}
        </div>
      </div>

      <div className="flex items-center gap-3">
        <Select
          value={currentView}
          onValueChange={(v) => onViewChange(v as FieldView)}
          disabled={availableViews.length <= 1}
        >
          <SelectTrigger className="w-40 h-8">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {availableViews.map((view) => (
              <SelectItem key={view.value} value={view.value}>
                {view.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Button
          variant={isHidden ? 'secondary' : 'ghost'}
          size="icon"
          className="h-8 w-8"
          onClick={onToggleHidden}
          title={isHidden ? 'Show in form' : 'Hide from form'}
        >
          {isHidden ? (
            <EyeOff className="size-4" />
          ) : (
            <Eye className="size-4" />
          )}
        </Button>
      </div>
    </div>
  );
}

interface FormConfigEditorProps {
  fields: ModelField[];
  config: FormConfig;
  onChange: (config: FormConfig) => void;
}

export function FormConfigEditor({ fields, config, onChange }: FormConfigEditorProps) {
  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  // Get ordered fields list based on config.field_order
  const getOrderedFields = () => {
    const fieldMap = new Map(fields.map((f) => [f.key, f]));
    const ordered: ModelField[] = [];

    // First add fields in specified order
    for (const key of config.field_order) {
      const field = fieldMap.get(key);
      if (field) {
        ordered.push(field);
        fieldMap.delete(key);
      }
    }

    // Then add remaining fields not in the order
    for (const field of fieldMap.values()) {
      ordered.push(field);
    }

    return ordered;
  };

  const orderedFields = getOrderedFields();

  const toggleHidden = (key: string) => {
    const hidden_fields = config.hidden_fields.includes(key)
      ? config.hidden_fields.filter((k) => k !== key)
      : [...config.hidden_fields, key];
    onChange({ ...config, hidden_fields });
  };

  const setFieldView = (key: string, view: FieldView | '') => {
    const field_views = { ...config.field_views };
    if (view === '') {
      delete field_views[key];
    } else {
      field_views[key] = view;
    }
    onChange({ ...config, field_views });
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;

    if (over && active.id !== over.id) {
      const currentOrder = orderedFields.map((f) => f.key);
      const oldIndex = currentOrder.indexOf(active.id as string);
      const newIndex = currentOrder.indexOf(over.id as string);
      const newOrder = arrayMove(currentOrder, oldIndex, newIndex);
      onChange({ ...config, field_order: newOrder });
    }
  };

  const getDefaultView = (field: ModelField): string => {
    const views = FIELD_VIEWS[field.type];
    return views?.[0]?.value || 'input';
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Form Configuration</CardTitle>
        <CardDescription>
          Configure the order, visibility, and widget types for form fields
        </CardDescription>
      </CardHeader>
      <CardContent>
        {fields.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            No fields defined. Add fields in the Schema tab first.
          </div>
        ) : (
          <DndContext
            sensors={sensors}
            collisionDetection={closestCenter}
            onDragEnd={handleDragEnd}
          >
            <SortableContext
              items={orderedFields.map((f) => f.key)}
              strategy={verticalListSortingStrategy}
            >
              <div className="space-y-2">
                {orderedFields.map((field) => {
                  const isHidden = config.hidden_fields.includes(field.key);
                  const currentView = config.field_views[field.key] || getDefaultView(field);
                  const availableViews = FIELD_VIEWS[field.type] || [];

                  return (
                    <SortableFieldItem
                      key={field.key}
                      field={field}
                      isHidden={isHidden}
                      currentView={currentView}
                      availableViews={availableViews}
                      onToggleHidden={() => toggleHidden(field.key)}
                      onViewChange={(view) => setFieldView(field.key, view)}
                    />
                  );
                })}
              </div>
            </SortableContext>
          </DndContext>
        )}

        <div className="mt-4 text-xs text-muted-foreground space-y-1">
          <p><strong>Order:</strong> Drag fields to reorder them in the form</p>
          <p><strong>Widget:</strong> Choose how the field is displayed in forms</p>
          <p><strong>Visibility:</strong> Click the eye icon to show/hide fields from forms</p>
        </div>
      </CardContent>
    </Card>
  );
}
