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

// System fields that can be displayed in tables
const SYSTEM_FIELDS = [
  { key: '_created_at', label: 'Created At', type: 'datetime' },
  { key: '_updated_at', label: 'Updated At', type: 'datetime' },
] as const;

type ColumnItem = {
  key: string;
  label: string;
  type: string;
  isSystem: boolean;
};

interface SortableRowProps {
  item: ColumnItem;
  config: TableConfig;
  onToggleColumn: (key: string) => void;
  onToggleFilter: (key: string) => void;
  onToggleSortable: (key: string) => void;
  onToggleSearchable: (key: string) => void;
}

function SortableRow({
  item,
  config,
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

  const isSearchableType = ['string', 'text'].includes(item.type);
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
            disabled={(!isSearchableType && !(config.searchable || []).includes(item.key)) || item.isSystem}
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

  // Create ordered list of all columns (fields + system fields)
  // Order is based on config.columns array, with remaining items at the end
  const orderedItems = useMemo((): ColumnItem[] => {
    const fieldItems: ColumnItem[] = fields.map(f => ({
      key: f.key,
      label: f.key,
      type: f.type,
      isSystem: false,
    }));

    const systemItems: ColumnItem[] = SYSTEM_FIELDS.map(sf => ({
      key: sf.key,
      label: sf.label,
      type: sf.type,
      isSystem: true,
    }));

    const allItems = [...fieldItems, ...systemItems];
    const allKeys = allItems.map(i => i.key);

    // Order based on config.columns, then add remaining
    const ordered: ColumnItem[] = [];
    const seen = new Set<string>();

    // First add items in config.columns order
    for (const key of config.columns) {
      const item = allItems.find(i => i.key === key);
      if (item && !seen.has(key)) {
        ordered.push(item);
        seen.add(key);
      }
    }

    // Then add remaining items that aren't in columns yet
    for (const item of allItems) {
      if (!seen.has(item.key)) {
        ordered.push(item);
        seen.add(item.key);
      }
    }

    return ordered;
  }, [fields, config.columns]);

  const toggleColumn = (key: string) => {
    const columns = config.columns.includes(key)
      ? config.columns.filter((c) => c !== key)
      : [...config.columns, key];
    onChange({ ...config, columns });
  };

  const toggleFilter = (key: string) => {
    const filters = config.filters.includes(key)
      ? config.filters.filter((f) => f !== key)
      : [...config.filters, key];
    onChange({ ...config, filters });
  };

  const toggleSortable = (key: string) => {
    const sort_columns = config.sort_columns.includes(key)
      ? config.sort_columns.filter((s) => s !== key)
      : [...config.sort_columns, key];
    onChange({ ...config, sort_columns });
  };

  const toggleSearchable = (key: string) => {
    const searchable = (config.searchable || []).includes(key)
      ? (config.searchable || []).filter((s) => s !== key)
      : [...(config.searchable || []), key];
    onChange({ ...config, searchable });
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;

    if (over && active.id !== over.id) {
      const oldIndex = orderedItems.findIndex(i => i.key === active.id);
      const newIndex = orderedItems.findIndex(i => i.key === over.id);

      if (oldIndex === -1 || newIndex === -1) return;

      const newOrder = arrayMove(orderedItems, oldIndex, newIndex);

      // Update columns array to reflect new order (only for visible columns)
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
                        onToggleColumn={toggleColumn}
                        onToggleFilter={toggleFilter}
                        onToggleSortable={toggleSortable}
                        onToggleSearchable={toggleSearchable}
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
