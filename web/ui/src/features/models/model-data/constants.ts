import type { FilterCondition } from '@/types';

// System fields that can be displayed in tables and views
export const SYSTEM_FIELDS = ['_id', '_created_at', '_updated_at'] as const;
export type SystemField = typeof SYSTEM_FIELDS[number];

export const SYSTEM_FIELD_LABELS: Record<SystemField, string> = {
  '_id': 'ID',
  '_created_at': 'Created At',
  '_updated_at': 'Updated At',
};

// Filter operators by field type
export const FILTER_OPERATORS: Record<string, { value: FilterCondition['operator']; label: string }[]> = {
  string: [
    { value: 'eq', label: '=' },
    { value: 'ne', label: '!=' },
    { value: 'contains', label: 'Contains' },
    { value: 'startsWith', label: 'Starts with' },
    { value: 'endsWith', label: 'Ends with' },
  ],
  text: [
    { value: 'eq', label: '=' },
    { value: 'ne', label: '!=' },
    { value: 'contains', label: 'Contains' },
  ],
  number: [
    { value: 'eq', label: '=' },
    { value: 'ne', label: '!=' },
    { value: 'gt', label: '>' },
    { value: 'gte', label: '>=' },
    { value: 'lt', label: '<' },
    { value: 'lte', label: '<=' },
  ],
  float: [
    { value: 'eq', label: '=' },
    { value: 'ne', label: '!=' },
    { value: 'gt', label: '>' },
    { value: 'gte', label: '>=' },
    { value: 'lt', label: '<' },
    { value: 'lte', label: '<=' },
  ],
  bool: [
    { value: 'eq', label: '=' },
  ],
  date: [
    { value: 'eq', label: '=' },
    { value: 'ne', label: '!=' },
    { value: 'gt', label: '>' },
    { value: 'gte', label: '>=' },
    { value: 'lt', label: '<' },
    { value: 'lte', label: '<=' },
  ],
  datetime: [
    { value: 'eq', label: '=' },
    { value: 'ne', label: '!=' },
    { value: 'gt', label: '>' },
    { value: 'gte', label: '>=' },
    { value: 'lt', label: '<' },
    { value: 'lte', label: '<=' },
  ],
};

export function isSystemField(key: string): key is SystemField {
  return SYSTEM_FIELDS.includes(key as SystemField);
}
