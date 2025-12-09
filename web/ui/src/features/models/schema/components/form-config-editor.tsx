/**
 * FormConfigEditor component
 * Configure form: field order, visibility, widgets (table-based layout)
 */

import { GripVertical, Eye, LayoutGrid } from 'lucide-react';
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
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  getDefaultFieldView,
  getAvailableFieldViews,
  toggleHiddenField,
  setFieldView,
  isFieldHidden,
} from '../../lib';

interface SortableRowProps {
  field: ModelField;
  isHidden: boolean;
  currentView: string;
  availableViews: { value: string; label: string }[];
  onToggleHidden: () => void;
  onViewChange: (view: FieldView) => void;
}

function SortableRow({
  field,
  isHidden,
  currentView,
  availableViews,
  onToggleHidden,
  onViewChange,
}: SortableRowProps) {
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
    opacity: isDragging ? 0.5 : 1,
  };

  return (
    <tr
      ref={setNodeRef}
      style={style}
      className={`border-t bg-background ${isHidden ? 'bg-muted/30' : ''} ${isDragging ? 'z-50' : ''}`}
    >
      <td className="p-3 w-10">
        <div
          {...attributes}
          {...listeners}
          className="cursor-grab active:cursor-grabbing text-muted-foreground hover:text-foreground transition-colors"
        >
          <GripVertical className="size-4" />
        </div>
      </td>
      <td className="p-3">
        <div className="flex flex-col">
          <span className={`font-mono text-sm ${isHidden ? 'text-muted-foreground' : ''}`}>
            {field.key}
          </span>
          <span className="text-xs text-muted-foreground">
            {field.type}{field.required ? ' (required)' : ''}
          </span>
        </div>
      </td>
      <td className="p-3">
        <Select
          value={currentView}
          onValueChange={(v) => onViewChange(v as FieldView)}
          disabled={availableViews.length <= 1}
        >
          <SelectTrigger className="h-8 w-36">
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
      </td>
      <td className="p-3 text-center">
        <div className="flex items-center justify-center">
          <Checkbox
            checked={!isHidden}
            onCheckedChange={() => onToggleHidden()}
          />
        </div>
      </td>
    </tr>
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

  // Get all fields in order (including hidden)
  const allOrderedFields = (() => {
    const fieldMap = new Map(fields.map(f => [f.key, f]));
    const ordered: ModelField[] = [];

    for (const key of config.field_order) {
      const field = fieldMap.get(key);
      if (field) {
        ordered.push(field);
        fieldMap.delete(key);
      }
    }

    for (const field of fieldMap.values()) {
      ordered.push(field);
    }

    return ordered;
  })();

  const handleToggleHidden = (key: string) => {
    onChange(toggleHiddenField(config, key));
  };

  const handleViewChange = (key: string, view: FieldView) => {
    onChange(setFieldView(config, key, view));
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;

    if (over && active.id !== over.id) {
      const currentOrder = allOrderedFields.map(f => f.key);
      const oldIndex = currentOrder.indexOf(active.id as string);
      const newIndex = currentOrder.indexOf(over.id as string);
      const newOrder = arrayMove(currentOrder, oldIndex, newIndex);
      onChange({ ...config, field_order: newOrder });
    }
  };

  return (
    <Card className="rounded-t-none !mt-0">
      <CardHeader>
        <CardTitle>Form Configuration</CardTitle>
        <CardDescription>
          Configure the order, visibility, and widget types for form fields. Drag rows to reorder.
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
            <div className="border rounded-lg overflow-hidden">
              <table className="w-full text-sm">
                <thead className="bg-muted/50">
                  <tr>
                    <th className="w-10 p-3"></th>
                    <th className="text-left font-medium p-3">Field</th>
                    <th className="text-left font-medium p-3 w-44">
                      <div className="flex items-center gap-1.5">
                        <LayoutGrid className="size-4" />
                        <span>Widget</span>
                      </div>
                    </th>
                    <th className="text-center font-medium p-3 w-28">
                      <div className="flex items-center justify-center gap-1.5">
                        <Eye className="size-4" />
                        <span>Visible</span>
                      </div>
                    </th>
                  </tr>
                </thead>
                <SortableContext
                  items={allOrderedFields.map(f => f.key)}
                  strategy={verticalListSortingStrategy}
                >
                  <tbody>
                    {allOrderedFields.map((field) => {
                      const isHidden = isFieldHidden(config, field.key);
                      const currentView = config.field_views[field.key] || getDefaultFieldView(field.type);
                      const availableViews = getAvailableFieldViews(field.type);

                      return (
                        <SortableRow
                          key={field.key}
                          field={field}
                          isHidden={isHidden}
                          currentView={currentView}
                          availableViews={availableViews}
                          onToggleHidden={() => handleToggleHidden(field.key)}
                          onViewChange={(view) => handleViewChange(field.key, view)}
                        />
                      );
                    })}
                  </tbody>
                </SortableContext>
              </table>
            </div>
          </DndContext>
        )}
      </CardContent>
    </Card>
  );
}
