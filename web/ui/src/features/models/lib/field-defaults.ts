/**
 * Field default values configuration and logic
 * Handles default value parsing, formatting, and special values like $now
 */

import type { FieldType } from '@/types';

/**
 * Special default value tokens
 */
export const DEFAULT_VALUE_TOKENS = {
  NOW: '$now',
} as const;

/**
 * Options for default value select inputs
 */
export interface DefaultValueOption {
  value: string;
  label: string;
}

/**
 * Get available default value options for a field type
 */
export function getDefaultValueOptions(type: FieldType): DefaultValueOption[] | null {
  switch (type) {
    case 'bool':
      return [
        { value: '__none__', label: 'No default' },
        { value: 'true', label: 'true' },
        { value: 'false', label: 'false' },
      ];
    case 'date':
    case 'datetime':
      return [
        { value: '__none__', label: 'No default' },
        { value: DEFAULT_VALUE_TOKENS.NOW, label: DEFAULT_VALUE_TOKENS.NOW },
      ];
    default:
      return null; // Use text input instead
  }
}

/**
 * Check if field type supports text input for default value
 */
export function supportsTextDefaultValue(type: FieldType): boolean {
  return ['string', 'text', 'number', 'float'].includes(type);
}

/**
 * Check if field type supports select input for default value
 */
export function supportsSelectDefaultValue(type: FieldType): boolean {
  return ['bool', 'date', 'datetime'].includes(type);
}

/**
 * Check if field type supports default values at all
 */
export function supportsDefaultValue(type: FieldType): boolean {
  // ref and file types don't support default values
  return !['ref', 'file', 'document'].includes(type);
}

/**
 * Parse default value from input based on field type
 */
export function parseDefaultValue(type: FieldType, inputValue: string): unknown {
  if (inputValue === '' || inputValue === '__none__') {
    return undefined;
  }

  switch (type) {
    case 'bool':
      return inputValue === 'true';
    case 'number':
      return parseInt(inputValue, 10);
    case 'float':
      return parseFloat(inputValue);
    case 'date':
    case 'datetime':
      if (inputValue === DEFAULT_VALUE_TOKENS.NOW) {
        return DEFAULT_VALUE_TOKENS.NOW;
      }
      return inputValue;
    default:
      return inputValue;
  }
}

/**
 * Format default value for display in input
 */
export function formatDefaultValueForInput(type: FieldType, value: unknown): string {
  if (value === undefined || value === null) {
    return '';
  }

  switch (type) {
    case 'bool':
      return value ? 'true' : 'false';
    case 'number':
    case 'float':
      return String(value);
    default:
      return String(value);
  }
}

/**
 * Get select value for default value dropdown
 */
export function getDefaultValueSelectValue(type: FieldType, value: unknown): string {
  if (value === undefined) {
    return '__none__';
  }

  switch (type) {
    case 'bool':
      return value ? 'true' : 'false';
    case 'date':
    case 'datetime':
      if (value === DEFAULT_VALUE_TOKENS.NOW) {
        return DEFAULT_VALUE_TOKENS.NOW;
      }
      return '__custom__';
    default:
      return String(value);
  }
}

/**
 * Get input type attribute for default value input
 */
export function getDefaultValueInputType(type: FieldType): 'text' | 'number' {
  switch (type) {
    case 'number':
    case 'float':
      return 'number';
    default:
      return 'text';
  }
}

/**
 * Get step attribute for number inputs
 */
export function getDefaultValueInputStep(type: FieldType): string {
  switch (type) {
    case 'float':
      return '0.01';
    case 'number':
      return '1';
    default:
      return '1';
  }
}
