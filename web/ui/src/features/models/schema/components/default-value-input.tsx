/**
 * DefaultValueInput component
 * Type-aware input for field default values
 */

import type { FieldType } from '@/types';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  getDefaultValueOptions,
  supportsTextDefaultValue,
  getDefaultValueSelectValue,
  parseDefaultValue,
  getDefaultValueInputType,
  getDefaultValueInputStep,
} from '../../lib';

interface DefaultValueInputProps {
  type: FieldType;
  value: unknown;
  onChange: (value: unknown) => void;
}

export function DefaultValueInput({ type, value, onChange }: DefaultValueInputProps) {
  const options = getDefaultValueOptions(type);

  // Use select for types with predefined options
  if (options) {
    const selectValue = getDefaultValueSelectValue(type, value);

    return (
      <Select
        value={selectValue}
        onValueChange={(v) => onChange(parseDefaultValue(type, v))}
      >
        <SelectTrigger>
          <SelectValue placeholder="No default" />
        </SelectTrigger>
        <SelectContent>
          {options.map((opt) => (
            <SelectItem key={opt.value} value={opt.value}>
              {opt.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    );
  }

  // Use text/number input for other types
  if (supportsTextDefaultValue(type)) {
    const inputType = getDefaultValueInputType(type);
    const step = getDefaultValueInputStep(type);

    return (
      <Input
        type={inputType}
        step={step}
        value={value !== undefined ? String(value) : ''}
        onChange={(e) => {
          const v = e.target.value;
          onChange(parseDefaultValue(type, v));
        }}
        placeholder="No default"
      />
    );
  }

  // Fallback for unsupported types
  return (
    <Input
      value={(value as string) || ''}
      onChange={(e) => onChange(e.target.value || undefined)}
      placeholder="No default"
    />
  );
}
