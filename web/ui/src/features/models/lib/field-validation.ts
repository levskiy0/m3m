/**
 * Field validation logic
 * Handles validation of field keys, types, and values
 */

import type { ModelField, FieldType } from '@/types';

/**
 * Regex for valid field key format
 * Must start with letter, contain only letters, numbers, underscores
 */
const FIELD_KEY_REGEX = /^[a-zA-Z][a-zA-Z0-9_]*$/;

/**
 * Reserved field keys that cannot be used
 */
export const RESERVED_FIELD_KEYS = [
  '_id',
  '_created_at',
  '_updated_at',
  'id',
] as const;

/**
 * Validate field key format
 */
export function isValidFieldKey(key: string): boolean {
  if (!key) return false;
  return FIELD_KEY_REGEX.test(key);
}

/**
 * Check if field key is reserved
 */
export function isReservedFieldKey(key: string): boolean {
  return RESERVED_FIELD_KEYS.includes(key as typeof RESERVED_FIELD_KEYS[number]);
}

/**
 * Validate field key (format + not reserved)
 */
export function validateFieldKey(key: string): { valid: boolean; error?: string } {
  if (!key) {
    return { valid: false, error: 'Field key is required' };
  }

  if (!isValidFieldKey(key)) {
    return {
      valid: false,
      error: 'Key must start with a letter and contain only letters, numbers, and underscores',
    };
  }

  if (isReservedFieldKey(key)) {
    return { valid: false, error: `"${key}" is a reserved field name` };
  }

  return { valid: true };
}

/**
 * Sanitize field key input (convert to valid format)
 */
export function sanitizeFieldKey(input: string): string {
  return input.toLowerCase().replace(/[^a-z0-9_]/g, '_');
}

/**
 * Check if fields array has validation errors
 */
export function hasFieldsValidationErrors(fields: ModelField[]): boolean {
  return fields.some(f => {
    const validation = validateFieldKey(f.key);
    return !validation.valid;
  });
}

/**
 * Check for duplicate field keys
 */
export function findDuplicateFieldKeys(fields: ModelField[]): string[] {
  const seen = new Set<string>();
  const duplicates: string[] = [];

  for (const field of fields) {
    if (seen.has(field.key)) {
      duplicates.push(field.key);
    } else {
      seen.add(field.key);
    }
  }

  return duplicates;
}

/**
 * Validate all fields in the schema
 */
export interface FieldValidationResult {
  valid: boolean;
  errors: Map<number, string>;
  duplicates: string[];
}

export function validateFields(fields: ModelField[]): FieldValidationResult {
  const errors = new Map<number, string>();
  const duplicates = findDuplicateFieldKeys(fields);

  fields.forEach((field, index) => {
    const validation = validateFieldKey(field.key);
    if (!validation.valid && validation.error) {
      errors.set(index, validation.error);
    } else if (duplicates.includes(field.key)) {
      errors.set(index, 'Duplicate field key');
    }
  });

  return {
    valid: errors.size === 0 && duplicates.length === 0,
    errors,
    duplicates,
  };
}

/**
 * Check if a field type requires ref_model
 */
export function requiresRefModel(type: FieldType): boolean {
  return type === 'ref';
}

/**
 * Validate ref field configuration
 */
export function validateRefField(field: ModelField): { valid: boolean; error?: string } {
  if (field.type !== 'ref') {
    return { valid: true };
  }

  if (!field.ref_model) {
    return { valid: false, error: 'Reference model is required for ref type' };
  }

  return { valid: true };
}
