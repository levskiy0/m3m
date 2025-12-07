/**
 * Constants for model-data feature
 * Re-exports system fields from lib and adds data-specific constants
 */

import type { FilterCondition } from '@/types';

// Re-export system fields from the centralized source
export {
  SYSTEM_FIELDS,
  SYSTEM_FIELD_KEYS,
  SYSTEM_FIELD_LABELS,
  isSystemField,
  getSystemFieldConfig,
  type SystemFieldKey,
  type SystemFieldConfig,
} from '../lib/system-fields';

// Backwards compatibility: array of system field keys
export const SYSTEM_FIELD_KEYS_ARRAY = ['_id', '_created_at', '_updated_at'] as const;
export type SystemField = typeof SYSTEM_FIELD_KEYS_ARRAY[number];

/**
 * Filter operators by field type
 */
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
  ref: [
    { value: 'eq', label: '=' },
    { value: 'ne', label: '!=' },
  ],
  select: [
    { value: 'eq', label: '=' },
    { value: 'ne', label: '!=' },
  ],
  multiselect: [
    { value: 'in', label: 'Contains' },
    { value: 'eq', label: '=' },
    { value: 'ne', label: '!=' },
  ],
};

/**
 * Get filter operators for a field type
 */
export function getFilterOperators(type: string) {
  return FILTER_OPERATORS[type] || FILTER_OPERATORS.string;
}
