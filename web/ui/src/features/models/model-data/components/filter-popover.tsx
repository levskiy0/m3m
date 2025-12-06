import { Filter, Plus, X } from 'lucide-react';
import { formatFieldLabel } from '@/lib/format';
import type { ModelField, FilterCondition } from '@/types';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { DatePicker, DateTimePicker } from '@/components/ui/datetime-picker';
import { FILTER_OPERATORS } from '../constants';
import type { ActiveFilter } from '../types';

interface FilterPopoverProps {
  activeFilters: ActiveFilter[];
  filterableFields: ModelField[];
  modelFields: ModelField[];
  onFiltersChange: (filters: ActiveFilter[]) => void;
  onPageReset: () => void;
}

export function FilterPopover({
  activeFilters,
  filterableFields,
  modelFields,
  onFiltersChange,
  onPageReset,
}: FilterPopoverProps) {
  const handleAddFilter = () => {
    const firstField = filterableFields[0];
    if (!firstField) return;
    onFiltersChange([...activeFilters, { field: firstField.key, operator: 'eq', value: '' }]);
  };

  const handleRemoveFilter = (index: number) => {
    onFiltersChange(activeFilters.filter((_, i) => i !== index));
    onPageReset();
  };

  const handleUpdateFilter = (index: number, updates: Partial<ActiveFilter>) => {
    const newFilters = [...activeFilters];
    newFilters[index] = { ...newFilters[index], ...updates };
    onFiltersChange(newFilters);
  };

  const handleClearAll = () => {
    onFiltersChange([]);
    onPageReset();
  };

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline" size="sm">
          <Filter className="mr-2 size-4" />
          Filter
          {activeFilters.length > 0 && (
            <Badge variant="secondary" className="ml-2 px-1.5 py-0 text-xs">
              {activeFilters.length}
            </Badge>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <div className="p-3 space-y-2 min-w-[400px]">
          {activeFilters.length === 0 ? (
            <p className="text-sm text-muted-foreground py-2">No filters applied</p>
          ) : (
            activeFilters.map((f, index) => {
              const field = modelFields.find(fld => fld.key === f.field);
              const operators = FILTER_OPERATORS[field?.type || 'string'] || FILTER_OPERATORS.string;
              return (
                <div key={index} className="flex items-center gap-1.5">
                  {index > 0 && (
                    <span className="text-xs text-muted-foreground w-7 text-right">and</span>
                  )}
                  {index === 0 && <span className="w-7" />}
                  <Select
                    value={f.field}
                    onValueChange={(v) => {
                      handleUpdateFilter(index, { field: v, operator: 'eq', value: '' });
                    }}
                  >
                    <SelectTrigger className="h-8 w-[120px] text-xs">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {filterableFields.map(fld => (
                        <SelectItem key={fld.key} value={fld.key}>
                          {formatFieldLabel(fld.key)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <Select
                    value={f.operator}
                    onValueChange={(v) => {
                      handleUpdateFilter(index, { operator: v as FilterCondition['operator'] });
                      onPageReset();
                    }}
                  >
                    <SelectTrigger className="h-8 w-[70px] text-xs">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {operators.map(op => (
                        <SelectItem key={op.value} value={op.value}>
                          {op.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  {field?.type === 'bool' ? (
                    <Select
                      value={String(f.value)}
                      onValueChange={(v) => {
                        handleUpdateFilter(index, { value: v === 'true' });
                        onPageReset();
                      }}
                    >
                      <SelectTrigger className="h-8 w-[80px] text-xs">
                        <SelectValue placeholder="..." />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="true">Yes</SelectItem>
                        <SelectItem value="false">No</SelectItem>
                      </SelectContent>
                    </Select>
                  ) : field?.type === 'date' ? (
                    <div className="w-[130px]">
                      <DatePicker
                        value={String(f.value) || undefined}
                        onChange={(v) => {
                          handleUpdateFilter(index, { value: v || '' });
                          onPageReset();
                        }}
                      />
                    </div>
                  ) : field?.type === 'datetime' ? (
                    <div className="w-[180px]">
                      <DateTimePicker
                        value={String(f.value) || undefined}
                        onChange={(v) => {
                          handleUpdateFilter(index, { value: v || '' });
                          onPageReset();
                        }}
                      />
                    </div>
                  ) : (
                    <Input
                      type={field?.type === 'number' || field?.type === 'float' ? 'number' : 'text'}
                      value={String(f.value)}
                      onChange={(e) => {
                        const val = field?.type === 'number' ? parseInt(e.target.value, 10) || 0
                          : field?.type === 'float' ? parseFloat(e.target.value) || 0
                          : e.target.value;
                        handleUpdateFilter(index, { value: val });
                      }}
                      onBlur={onPageReset}
                      onKeyDown={(e) => e.key === 'Enter' && onPageReset()}
                      placeholder="value"
                      className="h-8 w-[120px] text-xs"
                    />
                  )}
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 shrink-0"
                    onClick={() => handleRemoveFilter(index)}
                  >
                    <X className="size-3.5" />
                  </Button>
                </div>
              );
            })
          )}
          <div className="flex items-center gap-2 pt-2 border-t">
            <Button
              variant="ghost"
              size="sm"
              className="h-7 text-xs"
              onClick={handleAddFilter}
            >
              <Plus className="mr-1 size-3" />
              Add
            </Button>
            {activeFilters.length > 0 && (
              <Button
                variant="ghost"
                size="sm"
                className="h-7 text-xs text-muted-foreground"
                onClick={handleClearAll}
              >
                Clear all
              </Button>
            )}
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
}

interface ActiveFilterBadgesProps {
  activeFilters: ActiveFilter[];
  modelFields: ModelField[];
  onRemoveFilter: (index: number) => void;
}

export function ActiveFilterBadges({
  activeFilters,
  modelFields,
  onRemoveFilter,
}: ActiveFilterBadgesProps) {
  if (activeFilters.length === 0) return null;

  return (
    <div className="flex items-center gap-2 mb-4 flex-wrap min-h-[22px]">
      {activeFilters.map((f, i) => {
        const field = modelFields.find(fld => fld.key === f.field);
        const opLabel = FILTER_OPERATORS[field?.type || 'string']
          ?.find(op => op.value === f.operator)?.label || f.operator;
        return (
          <Badge key={i} variant="secondary" className="gap-1 font-normal">
            <span className="text-muted-foreground">{formatFieldLabel(f.field)}</span>
            <span>{opLabel}</span>
            <span className="font-medium">{String(f.value)}</span>
            <button
              onClick={() => onRemoveFilter(i)}
              className="ml-0.5 hover:text-destructive"
            >
              <X className="size-3" />
            </button>
          </Badge>
        );
      })}
    </div>
  );
}
