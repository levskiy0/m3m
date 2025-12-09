import { useState } from 'react';
import { Check, ChevronsUpDown } from 'lucide-react';
import type { ModelField, FieldView, Model } from '@/types';
import { cn } from '@/lib/utils';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';
import { Checkbox } from '@/components/ui/checkbox';
import { Button } from '@/components/ui/button';
import { DatePicker, DateTimePicker } from '@/components/ui/datetime-picker';
import { Label } from '@/components/ui/label';
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
} from '@/components/ui/command';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { MultiSelect } from '@/components/ui/multi-select';
import { getDefaultView } from '../utils';
import { RefFieldInput } from './ref-field-input';
import { FileFieldInput } from './file-field-input';

interface SelectComboboxProps {
  value: string;
  onChange: (value: unknown) => void;
  options: string[];
}

function SelectCombobox({ value, onChange, options }: SelectComboboxProps) {
  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="w-full justify-between font-normal"
        >
          {value || 'Select...'}
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[var(--radix-popover-trigger-width)] p-0" align="start">
        <Command>
          <CommandInput placeholder="Search..." />
          <CommandList>
            <CommandEmpty>No option found</CommandEmpty>
            <CommandGroup>
              {options.map((option) => (
                <CommandItem
                  key={option}
                  value={option}
                  onSelect={() => {
                    onChange(option === value ? '' : option);
                    setOpen(false);
                  }}
                >
                  <Check
                    className={cn(
                      'mr-2 h-4 w-4',
                      value === option ? 'opacity-100' : 'opacity-0'
                    )}
                  />
                  {option}
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}

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
          <div className="flex items-center">
            <Checkbox
              checked={!!value}
              onCheckedChange={onChange}
            />
          </div>
        );
      }
      return (
        <div className="flex items-center">
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
    case 'select':
      if (widget === 'combobox') {
        return (
          <SelectCombobox
            value={(value as string) || ''}
            onChange={onChange}
            options={field.options || []}
          />
        );
      }
      if (widget === 'radiogroup') {
        return (
          <RadioGroup
            value={(value as string) || ''}
            onValueChange={onChange}
            className="flex flex-col gap-2"
          >
            {(field.options || []).map((option) => (
              <div key={option} className="flex items-center gap-2">
                <RadioGroupItem value={option} id={`${field.key}-${option}`} />
                <Label htmlFor={`${field.key}-${option}`} className="font-normal cursor-pointer">
                  {option}
                </Label>
              </div>
            ))}
          </RadioGroup>
        );
      }
      return (
        <Select
          value={(value as string) || ''}
          onValueChange={onChange}
        >
          <SelectTrigger>
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            {(field.options || []).map((option) => (
              <SelectItem key={option} value={option}>
                {option}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      );
    case 'multiselect':
      if (widget === 'checkboxgroup') {
        const selectedValues = Array.isArray(value) ? (value as string[]) : [];
        const handleToggle = (option: string, checked: boolean) => {
          if (checked) {
            onChange([...selectedValues, option]);
          } else {
            onChange(selectedValues.filter((v) => v !== option));
          }
        };
        return (
          <div className="flex flex-col gap-2">
            {(field.options || []).map((option) => (
              <div key={option} className="flex items-center gap-2">
                <Checkbox
                  id={`${field.key}-${option}`}
                  checked={selectedValues.includes(option)}
                  onCheckedChange={(checked) => handleToggle(option, !!checked)}
                />
                <Label htmlFor={`${field.key}-${option}`} className="font-normal cursor-pointer">
                  {option}
                </Label>
              </div>
            ))}
          </div>
        );
      }
      return (
        <MultiSelect
          options={field.options || []}
          value={Array.isArray(value) ? (value as string[]) : []}
          onChange={onChange}
          placeholder="Select..."
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
