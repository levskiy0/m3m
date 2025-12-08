import { useMemo, useState, useRef } from 'react';
import {
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  Eye,
  Edit,
  Trash2,
  MoreHorizontal,
} from 'lucide-react';
import { formatFieldLabel } from '@/lib/format';
import { cn } from '@/lib/utils';
import type { ModelData, ModelField, TableConfig } from '@/types';
import { Checkbox } from '@/components/ui/checkbox';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from '@/components/ui/context-menu';
import { SYSTEM_FIELD_LABELS, type SystemField } from '../constants';
import { formatCellValue, formatSystemFieldValue } from '../utils';

interface DataTableProps {
  data: ModelData[];
  visibleColumns: ModelField[];
  visibleSystemColumns: SystemField[];
  tableConfig: TableConfig;
  sortField: string | null;
  sortOrder: 'asc' | 'desc';
  selectedIds: Set<string>;
  allSelected: boolean;
  getColumnWidth: (key: string, defaultWidth?: number) => number;
  onSort: (fieldKey: string) => void;
  onSelectRow: (id: string, checked: boolean) => void;
  onSelectAll: (checked: boolean) => void;
  onView: (data: ModelData) => void;
  onEdit: (data: ModelData) => void;
  onDelete: (data: ModelData) => void;
  onDeleteSelected: () => void;
  onResizeStart: (e: React.MouseEvent, column: string, currentWidth: number) => void;
}

export function DataTable({
  data,
  visibleColumns,
  visibleSystemColumns,
  tableConfig,
  sortField,
  sortOrder,
  selectedIds,
  allSelected,
  getColumnWidth,
  onSort,
  onSelectRow,
  onSelectAll,
  onView,
  onEdit,
  onDelete,
  onDeleteSelected,
  onResizeStart,
}: DataTableProps) {
  // Reset focus when data changes using a key derived from data
  const dataKey = data.map(d => d._id).join(',');
  const [focusedId, setFocusedId] = useState<string | null>(null);
  const prevDataKeyRef = useRef(dataKey);

  // Reset focus when data key changes
  // eslint-disable-next-line react-hooks/refs -- intentional: tracking previous value pattern
  if (prevDataKeyRef.current !== dataKey) {
    // eslint-disable-next-line react-hooks/refs -- intentional: updating ref to track changes
    prevDataKeyRef.current = dataKey;
    if (focusedId !== null) {
      setFocusedId(null);
    }
  }

  // Calculate total table width
  const tableWidth = useMemo(() => {
    const checkboxWidth = 48; // w-12
    const actionsWidth = 48; // w-12
    const regularColumnsWidth = visibleColumns.reduce((sum, field) => sum + getColumnWidth(field.key), 0);
    const systemColumnsWidth = visibleSystemColumns.reduce((sum, key) => sum + getColumnWidth(key, 180), 0);
    return checkboxWidth + regularColumnsWidth + systemColumnsWidth + actionsWidth;
  }, [visibleColumns, visibleSystemColumns, getColumnWidth]);

  return (
    <div className="rounded-xl rounded-t-none border overflow-x-auto bg-background">
      <Table
        className="table-fixed"
        style={{ width: tableWidth, minWidth: tableWidth }}
        wrapperClassName="h-[calc(100vh-315px)] overflow-y-auto [&_thead]:sticky [&_thead]:top-0 [&_thead]:z-10 [&_thead]:bg-background"
      >
        <TableHeader>
          <TableRow>
            <TableHead className="w-12 min-w-12 bg-background">
              <Checkbox
                checked={allSelected}
                onCheckedChange={onSelectAll}
              />
            </TableHead>
            {visibleColumns.map((field) => {
              const isSortable = tableConfig.sort_columns.includes(field.key);
              const isCurrentSort = sortField === field.key;
              const width = getColumnWidth(field.key);
              return (
                <TableHead
                  key={field.key}
                  className={`relative whitespace-nowrap bg-background ${isSortable ? 'cursor-pointer select-none hover:bg-muted/50' : ''}`}
                  style={{ width, minWidth: width, maxWidth: width }}
                  onClick={() => isSortable && onSort(field.key)}
                >
                  <div className="flex items-center gap-2 overflow-hidden">
                    <span className="truncate">{formatFieldLabel(field.key)}</span>
                    {isSortable && (
                      <span className="text-muted-foreground shrink-0">
                        {isCurrentSort ? (
                          sortOrder === 'asc' ? <ArrowUp className="size-4" /> : <ArrowDown className="size-4" />
                        ) : (
                          <ArrowUpDown className="size-3 opacity-50" />
                        )}
                      </span>
                    )}
                  </div>
                  {/* Resize handle */}
                  <div
                    className="absolute right-0 top-0 h-full w-1 cursor-col-resize hover:bg-primary/50 active:bg-primary"
                    onMouseDown={(e) => onResizeStart(e, field.key, width)}
                    onClick={(e) => e.stopPropagation()}
                  />
                </TableHead>
              );
            })}
            {visibleSystemColumns.map((key) => {
              const isSortable = tableConfig.sort_columns.includes(key);
              const isCurrentSort = sortField === key;
              const width = getColumnWidth(key, 180);
              return (
                <TableHead
                  key={key}
                  className={`relative whitespace-nowrap text-muted-foreground bg-background ${isSortable ? 'cursor-pointer select-none hover:bg-muted/50' : ''}`}
                  style={{ width, minWidth: width, maxWidth: width }}
                  onClick={() => isSortable && onSort(key)}
                >
                  <div className="flex items-center gap-2 overflow-hidden">
                    <span className="truncate">{SYSTEM_FIELD_LABELS[key]}</span>
                    {isSortable && (
                      <span className="text-muted-foreground shrink-0">
                        {isCurrentSort ? (
                          sortOrder === 'asc' ? <ArrowUp className="size-4" /> : <ArrowDown className="size-4" />
                        ) : (
                          <ArrowUpDown className="size-3 opacity-50" />
                        )}
                      </span>
                    )}
                  </div>
                  <div
                    className="absolute right-0 top-0 h-full w-1 cursor-col-resize hover:bg-primary/50 active:bg-primary"
                    onMouseDown={(e) => onResizeStart(e, key, width)}
                    onClick={(e) => e.stopPropagation()}
                  />
                </TableHead>
              );
            })}
            <TableHead className="w-12 bg-background" />
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((row) => (
            <ContextMenu key={row._id}>
              <ContextMenuTrigger asChild>
                <TableRow
                  className={cn(
                    "cursor-pointer text-md hover:bg-muted/50",
                    selectedIds.has(row._id) && "bg-blue-500/10",
                    focusedId === row._id && "bg-blue-500/20 hover:bg-blue-500/25"
                  )}
                  onClick={() => setFocusedId(row._id)}
                  onDoubleClick={() => onView(row)}
                >
                  <TableCell className="w-12 min-w-12" onClick={(e) => e.stopPropagation()}>
                    <Checkbox
                      checked={selectedIds.has(row._id)}
                      onCheckedChange={(checked) => onSelectRow(row._id, !!checked)}
                    />
                  </TableCell>
                  {visibleColumns.map((field) => {
                    const width = getColumnWidth(field.key);
                    return (
                      <TableCell
                        key={field.key}
                        style={{ width, minWidth: width, maxWidth: width }}
                      >
                        <span className="block truncate font-mono whitespace-nowrap">
                          {formatCellValue(row[field.key], field.type)}
                        </span>
                      </TableCell>
                    );
                  })}
                  {visibleSystemColumns.map((key) => {
                    const width = getColumnWidth(key, 180);
                    return (
                      <TableCell
                        key={key}
                        className="text-muted-foreground font-mono whitespace-nowrap"
                        style={{ width, minWidth: width, maxWidth: width }}
                      >
                        <span className="block truncate font-mono whitespace-nowrap">
                          {formatSystemFieldValue(key, row[key])}
                        </span>
                      </TableCell>
                    );
                  })}
                  <TableCell className="w-12" onClick={(e) => e.stopPropagation()}>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="size-8">
                          <MoreHorizontal className="size-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => onView(row)}>
                          <Eye className="mr-2 size-4" />
                          View
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => onEdit(row)}>
                          <Edit className="mr-2 size-4" />
                          Edit
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          onClick={() => onDelete(row)}
                          className="text-destructive focus:text-destructive"
                        >
                          <Trash2 className="mr-2 size-4" />
                          Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              </ContextMenuTrigger>
              <ContextMenuContent>
                <ContextMenuItem onClick={() => onView(row)}>
                  <Eye className="mr-2 size-4" />
                  View
                </ContextMenuItem>
                <ContextMenuItem onClick={() => onEdit(row)}>
                  <Edit className="mr-2 size-4" />
                  Edit
                </ContextMenuItem>
                <ContextMenuSeparator />
                <ContextMenuItem
                  onClick={() => onDelete(row)}
                  className="text-destructive focus:text-destructive"
                >
                  <Trash2 className="mr-2 size-4" />
                  Delete
                </ContextMenuItem>
                {selectedIds.size > 0 && (
                  <ContextMenuItem
                    onClick={onDeleteSelected}
                    className="text-destructive focus:text-destructive"
                  >
                    <Trash2 className="mr-2 size-4" />
                    Delete Selected ({selectedIds.size})
                  </ContextMenuItem>
                )}
              </ContextMenuContent>
            </ContextMenu>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
