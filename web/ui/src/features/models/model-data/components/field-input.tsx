import type { ModelField, FieldView, Model } from '@/types';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';
import { Checkbox } from '@/components/ui/checkbox';
import { DatePicker, DateTimePicker } from '@/components/ui/datetime-picker';
import { getDefaultView } from '../utils';
import { RefFieldInput } from './ref-field-input';
import { FileFieldInput } from './file-field-input';

interface FieldInputProps {
  field: ModelField;
  value: unknown;
  onChange: (value: unknown) => void;
  view?: FieldView;
  projectId?: string;
  models?: Model[];
}

export function FieldInput({ field, value, onChange, view, projectId, models }: FieldInputProps) {
  const widget = view || getDefaultView(field.type);

  switch (field.type) {
    case 'file':
      if (!projectId) {
        return <Input value={(value as string) || ''} onChange={(e) => onChange(e.target.value || null)} />;
      }
      return (
        <FileFieldInput
          value={(value as string) || null}
          onChange={onChange}
          projectId={projectId}
          view={widget}
        />
      );
    case 'ref':
      if (!projectId || !models || !field.ref_model) {
        return <Input value={(value as string) || ''} onChange={(e) => onChange(e.target.value || null)} />;
      }
      return (
        <RefFieldInput
          value={(value as string) || null}
          onChange={onChange}
          projectId={projectId}
          refModelSlug={field.ref_model}
          models={models}
        />
      );
    case 'string':
      return (
        <Input
          value={(value as string) || ''}
          onChange={(e) => onChange(e.target.value)}
        />
      );
    case 'text':
      if (widget === 'tiptap' || widget === 'markdown') {
        return (
          <Textarea
            value={(value as string) || ''}
            onChange={(e) => onChange(e.target.value)}
            rows={6}
            placeholder={widget === 'markdown' ? 'Write markdown...' : 'Write content...'}
          />
        );
      }
      return (
        <Textarea
          value={(value as string) || ''}
          onChange={(e) => onChange(e.target.value)}
          rows={4}
        />
      );
    case 'number':
    case 'float':
      return (
        <Input
          type="number"
          step={field.type === 'float' ? '0.01' : '1'}
          value={(value as number)?.toString() || ''}
          onChange={(e) => onChange(parseFloat(e.target.value) || 0)}
        />
      );
    case 'bool':
      if (widget === 'checkbox') {
        return (
          <div className="flex items-center h-10">
            <Checkbox
              checked={!!value}
              onCheckedChange={onChange}
            />
          </div>
        );
      }
      return (
        <div className="flex items-center h-10">
          <Switch
            checked={!!value}
            onCheckedChange={onChange}
          />
        </div>
      );
    case 'document':
      return (
        <Textarea
          value={typeof value === 'object' ? JSON.stringify(value, null, 2) : ''}
          onChange={(e) => {
            try {
              onChange(JSON.parse(e.target.value));
            } catch {
              // Invalid JSON
            }
          }}
          rows={6}
          className="font-mono text-sm"
          placeholder="{}"
        />
      );
    case 'date':
      return (
        <DatePicker
          value={(value as string) || undefined}
          onChange={(v) => onChange(v || '')}
        />
      );
    case 'datetime':
      return (
        <DateTimePicker
          value={(value as string) || undefined}
          onChange={(v) => onChange(v || '')}
        />
      );
    default:
      return (
        <Input
          value={(value as string) || ''}
          onChange={(e) => onChange(e.target.value)}
        />
      );
  }
}
