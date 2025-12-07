/**
 * SortableSchemaField component
 * Draggable field row for schema editor
 */

import { GripVertical, Trash2, Plus, X } from 'lucide-react';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';

import type { ModelField, Model, FieldType } from '@/types';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import { Field, FieldLabel } from '@/components/ui/field';
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
import { Badge } from '@/components/ui/badge';
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

  // Check if exact match exists (not just substring match)
  const trimmedInput = inputValue.trim();
  const hasExactMatch = trimmedInput && options.includes(trimmedInput);
  const canAddNew = trimmedInput && !hasExactMatch;

  return (
    <Field>
      <FieldLabel>Options</FieldLabel>
      <Popover>
        <PopoverTrigger asChild>
          <Button variant="outline" className="w-full justify-start h-auto min-h-10">
            {options.length > 0 ? (
              <div className="flex flex-wrap gap-1">
                {options.slice(0, 2).map((option) => (
                  <Badge key={option} variant="secondary" className="rounded-sm px-1 font-normal">
                    {option}
                  </Badge>
                ))}
                {options.length > 2 && (
                  <Badge variant="secondary" className="rounded-sm px-1 font-normal">
                    +{options.length - 2}
                  </Badge>
                )}
              </div>
            ) : (
              <span className="text-muted-foreground">Add options...</span>
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
              {/* Show "Add" button when there are filtered results but no exact match */}
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
    </Field>
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
            onChange={handleKeyChange}
            placeholder="field_name"
            className={keyError ? 'border-destructive focus-visible:ring-destructive' : ''}
          />
        </Field>

        <Field>
          <FieldLabel>Type</FieldLabel>
          <Select value={field.type} onValueChange={handleTypeChange}>
            <SelectTrigger>
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
        </Field>

        {showRefModel ? (
          <Field>
            <FieldLabel>Reference Model</FieldLabel>
            <Select
              value={field.ref_model || ''}
              onValueChange={(v) => onUpdate({ ref_model: v })}
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
        ) : showOptions ? (
          <SelectOptionsEditor
            options={field.options || []}
            onChange={(options) => onUpdate({ options })}
          />
        ) : showDefaultValue ? (
          <Field>
            <FieldLabel>Default Value</FieldLabel>
            <DefaultValueInput
              type={field.type}
              value={field.default_value}
              onChange={(value) => onUpdate({ default_value: value })}
            />
          </Field>
        ) : (
          <Field>
            <FieldLabel className="text-muted-foreground">-</FieldLabel>
            <div className="h-10" />
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
