/**
 * Field configuration and metadata
 * Centralized configuration for field types and their properties
 */

import type { FieldType, FieldView, ModelField } from '@/types';

/**
 * Field type metadata
 */
export interface FieldTypeConfig {
  value: FieldType;
  label: string;
  hasDefaultValue: boolean;
  hasRefModel: boolean;
  isSearchable: boolean;
  defaultView: FieldView;
}

/**
 * Complete field type configurations
 */
export const FIELD_TYPE_CONFIGS: Record<FieldType, FieldTypeConfig> = {
  string: {
    value: 'string',
    label: 'String',
    hasDefaultValue: true,
    hasRefModel: false,
    isSearchable: true,
    defaultView: 'input',
  },
  text: {
    value: 'text',
    label: 'Text',
    hasDefaultValue: true,
    hasRefModel: false,
    isSearchable: true,
    defaultView: 'textarea',
  },
  number: {
    value: 'number',
    label: 'Number',
    hasDefaultValue: true,
    hasRefModel: false,
    isSearchable: false,
    defaultView: 'input',
  },
  float: {
    value: 'float',
    label: 'Float',
    hasDefaultValue: true,
    hasRefModel: false,
    isSearchable: false,
    defaultView: 'input',
  },
  bool: {
    value: 'bool',
    label: 'Boolean',
    hasDefaultValue: true,
    hasRefModel: false,
    isSearchable: false,
    defaultView: 'switch',
  },
  document: {
    value: 'document',
    label: 'Document (JSON)',
    hasDefaultValue: false,
    hasRefModel: false,
    isSearchable: false,
    defaultView: 'json',
  },
  file: {
    value: 'file',
    label: 'File',
    hasDefaultValue: false,
    hasRefModel: false,
    isSearchable: false,
    defaultView: 'file',
  },
  ref: {
    value: 'ref',
    label: 'Reference',
    hasDefaultValue: false,
    hasRefModel: true,
    isSearchable: false,
    defaultView: 'select',
  },
  date: {
    value: 'date',
    label: 'Date',
    hasDefaultValue: true,
    hasRefModel: false,
    isSearchable: false,
    defaultView: 'datepicker',
  },
  datetime: {
    value: 'datetime',
    label: 'DateTime',
    hasDefaultValue: true,
    hasRefModel: false,
    isSearchable: false,
    defaultView: 'datetimepicker',
  },
};

/**
 * Get all field types as array for select options
 */
export function getFieldTypeOptions(): { value: FieldType; label: string }[] {
  return Object.values(FIELD_TYPE_CONFIGS).map(config => ({
    value: config.value,
    label: config.label,
  }));
}

/**
 * Get field type config
 */
export function getFieldTypeConfig(type: FieldType): FieldTypeConfig {
  return FIELD_TYPE_CONFIGS[type];
}

/**
 * Check if field type has default value support
 */
export function hasDefaultValueSupport(type: FieldType): boolean {
  return FIELD_TYPE_CONFIGS[type]?.hasDefaultValue ?? false;
}

/**
 * Check if field type requires ref_model
 */
export function hasRefModelSupport(type: FieldType): boolean {
  return FIELD_TYPE_CONFIGS[type]?.hasRefModel ?? false;
}

/**
 * Check if field type is searchable
 */
export function isSearchableType(type: FieldType): boolean {
  return FIELD_TYPE_CONFIGS[type]?.isSearchable ?? false;
}

/**
 * Get default view for field type
 */
export function getDefaultView(type: FieldType): FieldView {
  return FIELD_TYPE_CONFIGS[type]?.defaultView ?? 'input';
}

/**
 * Create a new empty field with defaults
 */
export function createEmptyField(index: number): ModelField {
  return {
    key: `field_${index + 1}`,
    type: 'string',
    required: false,
  };
}

/**
 * Clean field when type changes
 * Removes properties that don't apply to new type
 */
export function cleanFieldOnTypeChange(
  field: ModelField,
  newType: FieldType
): Partial<ModelField> {
  const updates: Partial<ModelField> = { type: newType };

  // Clear ref_model when switching away from ref type
  if (newType !== 'ref') {
    updates.ref_model = undefined;
  }

  // Clear default_value for types that don't support it
  if (!hasDefaultValueSupport(newType)) {
    updates.default_value = undefined;
  }

  return updates;
}

/**
 * Format field for display
 */
export function formatFieldDisplay(field: ModelField): string {
  const config = getFieldTypeConfig(field.type);
  let display = `${field.key} (${config.label})`;
  if (field.required) {
    display += ' *';
  }
  return display;
}
