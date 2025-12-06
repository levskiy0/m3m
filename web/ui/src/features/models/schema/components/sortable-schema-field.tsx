/**
 * SortableSchemaField component
 * Draggable field row for schema editor
 */

import { GripVertical, Trash2 } from 'lucide-react';
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
import { DefaultValueInput } from './default-value-input';
import {
  getFieldTypeOptions,
  hasDefaultValueSupport,
  hasRefModelSupport,
  isValidFieldKey,
  sanitizeFieldKey,
  cleanFieldOnTypeChange,
} from '../../lib';

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
