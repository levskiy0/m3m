/**
 * SortableSchemaField component
 * Draggable field row for schema editor (table-based layout)
 */

import { GripVertical, Trash2, Plus, X, Settings2 } from 'lucide-react';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';

import type { ModelField, Model, FieldType } from '@/types';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';
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
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from '@/components/ui/command';
import { useState } from 'react';
import { DefaultValueInput } from './default-value-input';
import {
  getFieldTypeOptions,
  hasDefaultValueSupport,
  hasRefModelSupport,
  hasOptionsSupport,
  isValidFieldKey,
  sanitizeFieldKey,
  cleanFieldOnTypeChange,
} from '../../lib';

interface SelectOptionsEditorProps {
  options: string[];
  onChange: (options: string[]) => void;
}

function SelectOptionsEditor({ options, onChange }: SelectOptionsEditorProps) {
  const [inputValue, setInputValue] = useState('');

  const addOption = (value: string) => {
    const trimmed = value.trim();
    if (trimmed && !options.includes(trimmed)) {
      onChange([...options, trimmed]);
    }
    setInputValue('');
  };

  const removeOption = (option: string) => {
    onChange(options.filter(o => o !== option));
  };

  const trimmedInput = inputValue.trim();
  const hasExactMatch = trimmedInput && options.includes(trimmedInput);
  const canAddNew = trimmedInput && !hasExactMatch;

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline" size="sm" className="h-8 justify-start gap-1.5">
          <Settings2 className="size-3.5" />
          {options.length > 0 ? (
            <span className="text-xs">{options.length} options</span>
          ) : (
            <span className="text-xs text-muted-foreground">Add options</span>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-64 p-0" align="start">
        <Command>
          <CommandInput
            placeholder="Add option..."
            value={inputValue}
            onValueChange={setInputValue}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && inputValue) {
                e.preventDefault();
                addOption(inputValue);
              }
            }}
          />
          <CommandList>
            <CommandEmpty>
              {inputValue ? (
                <Button
                  variant="ghost"
                  className="w-full justify-start"
                  onClick={() => addOption(inputValue)}
                >
                  <Plus className="mr-2 size-4" />
                  Add "{inputValue}"
                </Button>
              ) : (
                'Type to add option'
              )}
            </CommandEmpty>
            {canAddNew && options.length > 0 && (
              <CommandGroup>
                <CommandItem
                  onSelect={() => addOption(inputValue)}
                  className="text-primary"
                >
                  <Plus className="mr-2 size-4" />
                  Add "{trimmedInput}"
                </CommandItem>
              </CommandGroup>
            )}
            {options.length > 0 && (
              <CommandGroup>
                {options.map((option) => (
                  <CommandItem
                    key={option}
                    className="justify-between"
                    onSelect={() => removeOption(option)}
                  >
                    <span>{option}</span>
                    <X className="size-4 text-muted-foreground" />
                  </CommandItem>
                ))}
              </CommandGroup>
            )}
            {options.length > 0 && (
              <>
                <CommandSeparator />
                <CommandGroup>
                  <CommandItem
                    onSelect={() => onChange([])}
                    className="justify-center text-center text-muted-foreground"
                  >
                    Clear all
                  </CommandItem>
                </CommandGroup>
              </>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}

interface SortableSchemaFieldProps {
  id: string;
  field: ModelField;
  onUpdate: (updates: Partial<ModelField>) => void;
  onRemove: () => void;
  models: Model[];
  currentModelId?: string;
}

export function SortableSchemaField({
  id,
  field,
  onUpdate,
  onRemove,
  models,
  currentModelId,
}: SortableSchemaFieldProps) {
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
    opacity: isDragging ? 0.5 : 1,
  };

  // Filter out current model from ref options
  const availableModels = models.filter(m => m.id !== currentModelId);
  const fieldTypes = getFieldTypeOptions();
  const keyError = field.key && !isValidFieldKey(field.key);
  const showRefModel = hasRefModelSupport(field.type);
  const showDefaultValue = hasDefaultValueSupport(field.type);
  const showOptions = hasOptionsSupport(field.type);

  const handleKeyChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onUpdate({ key: sanitizeFieldKey(e.target.value) });
  };

  const handleTypeChange = (newType: FieldType) => {
    const updates = cleanFieldOnTypeChange(field, newType);
    onUpdate(updates);
  };

  return (
    <tr
      ref={setNodeRef}
      style={style}
      className={`border-t bg-background ${isDragging ? 'z-50' : ''}`}
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
        <Input
          value={field.key}
          onChange={handleKeyChange}
          placeholder="field_name"
          className={`h-8 font-mono text-sm ${keyError ? 'border-destructive focus-visible:ring-destructive' : ''}`}
        />
      </td>
      <td className="p-3">
        <Select value={field.type} onValueChange={handleTypeChange}>
          <SelectTrigger className="h-8">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {fieldTypes.map((type) => (
              <SelectItem key={type.value} value={type.value}>
                {type.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </td>
      <td className="p-3">
        {showRefModel ? (
          <Select
            value={field.ref_model || ''}
            onValueChange={(v) => onUpdate({ ref_model: v })}
          >
            <SelectTrigger className="h-8">
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
        ) : showOptions ? (
          <SelectOptionsEditor
            options={field.options || []}
            onChange={(options) => onUpdate({ options })}
          />
        ) : showDefaultValue ? (
          <DefaultValueInput
            type={field.type}
            value={field.default_value}
            onChange={(value) => onUpdate({ default_value: value })}
            className="h-8"
          />
        ) : (
          <span className="text-muted-foreground text-sm">—</span>
        )}
      </td>
      <td className="p-3 text-center">
        <div className="flex items-center justify-center">
          <Checkbox
            checked={field.required}
            onCheckedChange={(checked) => onUpdate({ required: !!checked })}
          />
        </div>
      </td>
      <td className="p-3 w-10">
        <Button
          variant="ghost"
          size="icon"
          className="h-8 w-8 text-muted-foreground hover:text-destructive"
          onClick={onRemove}
        >
          <Trash2 className="size-4" />
        </Button>
      </td>
    </tr>
  );
}
