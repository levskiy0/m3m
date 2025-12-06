/**
 * FormConfig operations and utilities
 * Business logic for managing form configuration
 */

import type { ModelField, FormConfig, FieldView, FieldType } from '@/types';
import { FIELD_VIEWS } from '@/lib/constants';

/**
 * Create default form config from model fields
 */
export function createDefaultFormConfig(fields: ModelField[]): FormConfig {
  return {
    field_order: fields.map(f => f.key),
    hidden_fields: [],
    field_views: {},
  };
}

/**
 * Merge saved config with defaults
 */
export function mergeFormConfig(
  saved: Partial<FormConfig> | undefined,
  defaults: FormConfig
): FormConfig {
  return {
    field_order: saved?.field_order ?? defaults.field_order,
    hidden_fields: saved?.hidden_fields ?? defaults.hidden_fields,
    field_views: saved?.field_views ?? defaults.field_views,
  };
}

/**
 * Toggle field visibility in form
 */
export function toggleHiddenField(config: FormConfig, key: string): FormConfig {
  const hidden_fields = config.hidden_fields.includes(key)
    ? config.hidden_fields.filter(k => k !== key)
    : [...config.hidden_fields, key];
  return { ...config, hidden_fields };
}

/**
 * Set field view widget
 */
export function setFieldView(
  config: FormConfig,
  key: string,
  view: FieldView | ''
): FormConfig {
  const field_views = { ...config.field_views };
  if (view === '') {
    delete field_views[key];
  } else {
    field_views[key] = view;
  }
  return { ...config, field_views };
}

/**
 * Reorder form fields
 */
export function reorderFormFields(
  config: FormConfig,
  oldIndex: number,
  newIndex: number
): FormConfig {
  const field_order = [...config.field_order];
  const [removed] = field_order.splice(oldIndex, 1);
  field_order.splice(newIndex, 0, removed);
  return { ...config, field_order };
}

/**
 * Get ordered fields for form display
 */
export function getOrderedFormFields(
  fields: ModelField[],
  config: FormConfig
): ModelField[] {
  const fieldMap = new Map(fields.map(f => [f.key, f]));
  const ordered: ModelField[] = [];

  // First add fields in specified order
  for (const key of config.field_order) {
    const field = fieldMap.get(key);
    if (field && !config.hidden_fields.includes(key)) {
      ordered.push(field);
      fieldMap.delete(key);
    }
  }

  // Then add remaining fields not in the order
  for (const [key, field] of fieldMap) {
    if (!config.hidden_fields.includes(key)) {
      ordered.push(field);
    }
  }

  return ordered;
}

/**
 * Get default view for a field type
 */
export function getDefaultFieldView(type: FieldType): FieldView {
  const views = FIELD_VIEWS[type];
  return (views?.[0]?.value as FieldView) || 'input';
}

/**
 * Get available views for a field type
 */
export function getAvailableFieldViews(type: FieldType): { value: string; label: string }[] {
  return FIELD_VIEWS[type] || [];
}

/**
 * Get the view to use for a field (configured or default)
 */
export function getFieldView(config: FormConfig, field: ModelField): FieldView {
  return config.field_views[field.key] || getDefaultFieldView(field.type);
}

/**
 * Update config when a field key is renamed
 */
export function updateConfigOnFieldRename(
  config: FormConfig,
  oldKey: string,
  newKey: string
): FormConfig {
  const newFieldViews = { ...config.field_views };
  if (oldKey in newFieldViews) {
    newFieldViews[newKey] = newFieldViews[oldKey];
    delete newFieldViews[oldKey];
  }

  return {
    field_order: config.field_order.map(k => k === oldKey ? newKey : k),
    hidden_fields: config.hidden_fields.map(k => k === oldKey ? newKey : k),
    field_views: newFieldViews,
  };
}

/**
 * Update config when a field is removed
 */
export function updateConfigOnFieldRemove(config: FormConfig, key: string): FormConfig {
  const newFieldViews = { ...config.field_views };
  delete newFieldViews[key];

  return {
    field_order: config.field_order.filter(k => k !== key),
    hidden_fields: config.hidden_fields.filter(k => k !== key),
    field_views: newFieldViews,
  };
}

/**
 * Update config when a new field is added
 */
export function updateConfigOnFieldAdd(config: FormConfig, field: ModelField): FormConfig {
  return {
    ...config,
    field_order: [...config.field_order, field.key],
  };
}

/**
 * Check if a field is hidden in form
 */
export function isFieldHidden(config: FormConfig, key: string): boolean {
  return config.hidden_fields.includes(key);
}
