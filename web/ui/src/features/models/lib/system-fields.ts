/**
 * System fields configuration
 * Unified source of truth for system fields across the models feature
 */

import type { FieldType } from '@/types';

export interface SystemFieldConfig {
  key: string;
  label: string;
  type: FieldType;
  sortable: boolean;
  filterable: boolean;
}

/**
 * System fields that are automatically added to all models
 */
export const SYSTEM_FIELDS: readonly SystemFieldConfig[] = [
  { key: '_id', label: 'ID', type: 'string', sortable: true, filterable: false },
  { key: '_created_at', label: 'Created At', type: 'datetime', sortable: true, filterable: false },
  { key: '_updated_at', label: 'Updated At', type: 'datetime', sortable: true, filterable: false },
] as const;

export type SystemFieldKey = typeof SYSTEM_FIELDS[number]['key'];

/**
 * System field keys as array for quick checks
 */
export const SYSTEM_FIELD_KEYS = SYSTEM_FIELDS.map(f => f.key) as readonly string[];

/**
 * Labels for system fields
 */
export const SYSTEM_FIELD_LABELS: Record<string, string> = Object.fromEntries(
  SYSTEM_FIELDS.map(f => [f.key, f.label])
);

/**
 * Check if a field key is a system field
 */
export function isSystemField(key: string): boolean {
  return SYSTEM_FIELD_KEYS.includes(key);
}

/**
 * Get system field config by key
 */
export function getSystemFieldConfig(key: string): SystemFieldConfig | undefined {
  return SYSTEM_FIELDS.find(f => f.key === key);
}

/**
 * Get sortable system fields
 */
export function getSortableSystemFields(): SystemFieldConfig[] {
  return SYSTEM_FIELDS.filter(f => f.sortable);
}

/**
 * System fields that can be shown in table columns (excludes _id by default)
 */
export const TABLE_SYSTEM_FIELDS = SYSTEM_FIELDS.filter(f => f.key !== '_id');
