import { formatDateTime } from '@/lib/format';
import type { FieldType, FieldView } from '@/types';
import type { SystemField } from './constants';

export function getDefaultView(type: FieldType): FieldView {
  switch (type) {
    case 'string': return 'input';
    case 'text': return 'textarea';
    case 'number': return 'input';
    case 'float': return 'input';
    case 'bool': return 'switch';
    case 'document': return 'json';
    case 'date': return 'datepicker';
    case 'datetime': return 'datetimepicker';
    case 'file': return 'file';
    case 'ref': return 'select';
    default: return 'input';
  }
}

export function formatCellValue(value: unknown, type: FieldType): string {
  if (value === null || value === undefined) return '-';
  if (type === 'bool') return value ? 'Yes' : 'No';
  if (type === 'document') return JSON.stringify(value).slice(0, 50) + '...';
  if (type === 'date' || type === 'datetime') {
    return formatDateTime(value as string);
  }
  const str = String(value);
  return str.length > 50 ? str.slice(0, 50) + '...' : str;
}

export function formatSystemFieldValue(key: SystemField, value: unknown): string {
  if (value === null || value === undefined) return '-';
  if (key === '_id') {
    const str = String(value);
    return str.length > 24 ? str.slice(0, 24) + '...' : str;
  }
  if (key === '_created_at' || key === '_updated_at') {
    return formatDateTime(value as string);
  }
  return String(value);
}
