/**
 * TableConfig operations and utilities
 * Business logic for managing table configuration
 */

import type { ModelField, TableConfig } from '@/types';
import { isSystemField, TABLE_SYSTEM_FIELDS, SYSTEM_FIELD_KEYS } from './system-fields';

/**
 * Column item for table config editor
 */
export interface ColumnItem {
  key: string;
  label: string;
  type: string;
  isSystem: boolean;
}

/**
 * Create default table config from model fields
 */
export function createDefaultTableConfig(fields: ModelField[]): TableConfig {
  const fieldKeys = fields.map(f => f.key);
  const searchableFields = fields
    .filter(f => ['string', 'text'].includes(f.type))
    .map(f => f.key);

  return {
    columns: fieldKeys,
    filters: [],
    sort_columns: [...fieldKeys, ...SYSTEM_FIELD_KEYS],
    searchable: searchableFields,
  };
}

/**
 * Merge saved config with defaults, ensuring all arrays exist
 */
export function mergeTableConfig(
  saved: Partial<TableConfig> | undefined,
  defaults: TableConfig
): TableConfig {
  return {
    columns: saved?.columns ?? defaults.columns,
    filters: saved?.filters ?? defaults.filters,
    sort_columns: saved?.sort_columns ?? defaults.sort_columns,
    searchable: saved?.searchable ?? defaults.searchable,
  };
}

/**
 * Build ordered column items list for table config editor
 * Order is based on config.columns array, with remaining items at the end
 */
export function buildOrderedColumnItems(
  fields: ModelField[],
  config: TableConfig
): ColumnItem[] {
  // Build field items
  const fieldItems: ColumnItem[] = fields.map(f => ({
    key: f.key,
    label: f.key,
    type: f.type,
    isSystem: false,
  }));

  // Build system field items
  const systemItems: ColumnItem[] = TABLE_SYSTEM_FIELDS.map(sf => ({
    key: sf.key,
    label: sf.label,
    type: sf.type,
    isSystem: true,
  }));

  const allItems = [...fieldItems, ...systemItems];

  // Order based on config.columns, then add remaining
  const ordered: ColumnItem[] = [];
  const seen = new Set<string>();

  // First add items in config.columns order
  for (const key of config.columns) {
    const item = allItems.find(i => i.key === key);
    if (item && !seen.has(key)) {
      ordered.push(item);
      seen.add(key);
    }
  }

  // Then add remaining items that aren't in columns yet
  for (const item of allItems) {
    if (!seen.has(item.key)) {
      ordered.push(item);
      seen.add(item.key);
    }
  }

  return ordered;
}

/**
 * Toggle column visibility in config
 */
export function toggleColumn(config: TableConfig, key: string): TableConfig {
  const columns = config.columns.includes(key)
    ? config.columns.filter(c => c !== key)
    : [...config.columns, key];
  return { ...config, columns };
}

/**
 * Toggle filter availability for a field
 */
export function toggleFilter(config: TableConfig, key: string): TableConfig {
  const filters = config.filters.includes(key)
    ? config.filters.filter(f => f !== key)
    : [...config.filters, key];
  return { ...config, filters };
}

/**
 * Toggle sortable for a field
 */
export function toggleSortable(config: TableConfig, key: string): TableConfig {
  const sort_columns = config.sort_columns.includes(key)
    ? config.sort_columns.filter(s => s !== key)
    : [...config.sort_columns, key];
  return { ...config, sort_columns };
}

/**
 * Toggle searchable for a field
 */
export function toggleSearchable(
  config: TableConfig,
  key: string,
  fields: ModelField[]
): TableConfig {
  const field = fields.find(f => f.key === key);
  const isSearchableType = field && ['string', 'text'].includes(field.type);

  // Only allow adding if it's a searchable type
  if (!isSearchableType && !(config.searchable || []).includes(key)) {
    return config;
  }

  const searchable = (config.searchable || []).includes(key)
    ? (config.searchable || []).filter(s => s !== key)
    : [...(config.searchable || []), key];
  return { ...config, searchable };
}

/**
 * Reorder columns after drag and drop
 */
export function reorderColumns(
  config: TableConfig,
  orderedItems: ColumnItem[],
  oldIndex: number,
  newIndex: number
): TableConfig {
  // Calculate new order
  const newOrder = [...orderedItems];
  const [removed] = newOrder.splice(oldIndex, 1);
  newOrder.splice(newIndex, 0, removed);

  // Update columns array to reflect new order (only for visible columns)
  const newColumns = newOrder
    .filter(i => config.columns.includes(i.key))
    .map(i => i.key);

  return { ...config, columns: newColumns };
}

/**
 * Update config when a field key is renamed
 */
export function updateConfigOnFieldRename(
  config: TableConfig,
  oldKey: string,
  newKey: string
): TableConfig {
  return {
    columns: config.columns.map(k => k === oldKey ? newKey : k),
    filters: config.filters.map(k => k === oldKey ? newKey : k),
    sort_columns: config.sort_columns.map(k => k === oldKey ? newKey : k),
    searchable: (config.searchable || []).map(k => k === oldKey ? newKey : k),
  };
}

/**
 * Update config when a field is removed
 */
export function updateConfigOnFieldRemove(config: TableConfig, key: string): TableConfig {
  return {
    columns: config.columns.filter(k => k !== key),
    filters: config.filters.filter(k => k !== key),
    sort_columns: config.sort_columns.filter(k => k !== key),
    searchable: (config.searchable || []).filter(k => k !== key),
  };
}

/**
 * Update config when a new field is added
 */
export function updateConfigOnFieldAdd(
  config: TableConfig,
  field: ModelField
): TableConfig {
  const isSearchableType = ['string', 'text'].includes(field.type);

  return {
    ...config,
    columns: [...config.columns, field.key],
    sort_columns: [...config.sort_columns, field.key],
    searchable: isSearchableType
      ? [...(config.searchable || []), field.key]
      : config.searchable,
  };
}

/**
 * Check if a field type is searchable
 */
export function isSearchableFieldType(type: string): boolean {
  return ['string', 'text'].includes(type);
}

/**
 * Get visible model columns (excluding system fields)
 */
export function getVisibleModelColumns(
  fields: ModelField[],
  config: TableConfig
): ModelField[] {
  const fieldMap = new Map(fields.map(f => [f.key, f]));
  return config.columns
    .filter(key => !isSystemField(key))
    .map(key => fieldMap.get(key))
    .filter((f): f is ModelField => f !== undefined);
}

/**
 * Get visible system columns
 */
export function getVisibleSystemColumns(config: TableConfig): string[] {
  return config.columns.filter(key => isSystemField(key));
}
