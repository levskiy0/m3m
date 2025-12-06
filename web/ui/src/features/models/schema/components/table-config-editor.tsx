/**
 * TableConfigEditor component
 * Configure table view: columns, filters, sorting, search
 */

import { useMemo } from 'react';
import { Columns3, Filter, ArrowUpDown, Search, GripVertical } from 'lucide-react';
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

import type { ModelField, TableConfig } from '@/types';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Checkbox } from '@/components/ui/checkbox';
import {
  buildOrderedColumnItems,
  toggleColumn,
  toggleFilter,
  toggleSortable,
  toggleSearchable,
  isSearchableFieldType,
  type ColumnItem,
} from '../../lib';

interface SortableRowProps {
  item: ColumnItem;
  config: TableConfig;
  fields: ModelField[];
  onToggleColumn: (key: string) => void;
  onToggleFilter: (key: string) => void;
  onToggleSortable: (key: string) => void;
  onToggleSearchable: (key: string) => void;
}

function SortableRow({
  item,
  config,
  fields,
  onToggleColumn,
  onToggleFilter,
  onToggleSortable,
  onToggleSearchable,
}: SortableRowProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: item.key });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  const isSearchableType = isSearchableFieldType(item.type);
  const isInColumns = config.columns.includes(item.key);

  return (
    <tr
      ref={setNodeRef}
      style={style}
      className={`border-t ${item.isSystem ? 'bg-muted/30' : ''} ${isDragging ? 'z-50' : ''}`}
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
          <span className={`font-mono text-sm ${item.isSystem ? 'text-muted-foreground' : ''}`}>
            {item.isSystem ? item.label : item.key}
          </span>
          <span className="text-xs text-muted-foreground">
            {item.type}{item.isSystem ? ' (system)' : ''}
          </span>
        </div>
      </td>
      <td className="p-3 text-center">
        <div className="flex items-center justify-center">
          <Checkbox
            id={`col-${item.key}`}
            checked={isInColumns}
            onCheckedChange={() => onToggleColumn(item.key)}
          />
        </div>
      </td>
      <td className="p-3 text-center">
        <div className="flex items-center justify-center">
          <Checkbox
            id={`filter-${item.key}`}
            checked={config.filters.includes(item.key)}
            onCheckedChange={() => onToggleFilter(item.key)}
            disabled={item.isSystem}
          />
        </div>
      </td>
      <td className="p-3 text-center">
        <div className="flex items-center justify-center">
          <Checkbox
            id={`sort-${item.key}`}
            checked={config.sort_columns.includes(item.key)}
            onCheckedChange={() => onToggleSortable(item.key)}
          />
        </div>
      </td>
      <td className="p-3 text-center">
        <div className="flex items-center justify-center">
          <Checkbox
            id={`search-${item.key}`}
            checked={(config.searchable || []).includes(item.key)}
            onCheckedChange={() => onToggleSearchable(item.key)}
            disabled={!isSearchableType || item.isSystem}
          />
        </div>
      </td>
    </tr>
  );
}

interface TableConfigEditorProps {
  fields: ModelField[];
  config: TableConfig;
  onChange: (config: TableConfig) => void;
}

export function TableConfigEditor({ fields, config, onChange }: TableConfigEditorProps) {
  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  const orderedItems = useMemo(
    () => buildOrderedColumnItems(fields, config),
    [fields, config]
  );

  const handleToggleColumn = (key: string) => onChange(toggleColumn(config, key));
  const handleToggleFilter = (key: string) => onChange(toggleFilter(config, key));
  const handleToggleSortable = (key: string) => onChange(toggleSortable(config, key));
  const handleToggleSearchable = (key: string) => onChange(toggleSearchable(config, key, fields));

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;

    if (over && active.id !== over.id) {
      const oldIndex = orderedItems.findIndex(i => i.key === active.id);
      const newIndex = orderedItems.findIndex(i => i.key === over.id);

      if (oldIndex === -1 || newIndex === -1) return;

      const newOrder = arrayMove(orderedItems, oldIndex, newIndex);
      const newColumns = newOrder
        .filter(i => config.columns.includes(i.key))
        .map(i => i.key);

      onChange({ ...config, columns: newColumns });
    }
  };

  return (
    <Card className="rounded-t-none !mt-0">
      <CardHeader>
        <CardTitle>Table Configuration</CardTitle>
        <CardDescription>
          Configure how data is displayed in the table view. Drag rows to reorder columns.
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
                    <th className="text-center font-medium p-3 w-28">
                      <div className="flex items-center justify-center gap-1.5">
                        <Columns3 className="size-4" />
                        <span>Column</span>
                      </div>
                    </th>
                    <th className="text-center font-medium p-3 w-28">
                      <div className="flex items-center justify-center gap-1.5">
                        <Filter className="size-4" />
                        <span>Filter</span>
                      </div>
                    </th>
                    <th className="text-center font-medium p-3 w-28">
                      <div className="flex items-center justify-center gap-1.5">
                        <ArrowUpDown className="size-4" />
                        <span>Sort</span>
                      </div>
                    </th>
                    <th className="text-center font-medium p-3 w-28">
                      <div className="flex items-center justify-center gap-1.5">
                        <Search className="size-4" />
                        <span>Search</span>
                      </div>
                    </th>
                  </tr>
                </thead>
                <SortableContext
                  items={orderedItems.map(i => i.key)}
                  strategy={verticalListSortingStrategy}
                >
                  <tbody>
                    {orderedItems.map((item) => (
                      <SortableRow
                        key={item.key}
                        item={item}
                        config={config}
                        fields={fields}
                        onToggleColumn={handleToggleColumn}
                        onToggleFilter={handleToggleFilter}
                        onToggleSortable={handleToggleSortable}
                        onToggleSearchable={handleToggleSearchable}
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
  );
}
