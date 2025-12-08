/**
 * Models feature business logic
 * Exports all utilities, configurations, and operations
 */

// System fields
export * from './system-fields';

// Field configuration
export * from './field-config';

// Field defaults
export * from './field-defaults';

// Field validation
export * from './field-validation';

// Table config operations
export * from './table-config';

// Form config operations
export {
  createDefaultFormConfig,
  mergeFormConfig,
  toggleHiddenField,
  setFieldView,
  reorderFormFields,
  getOrderedFormFields,
  getDefaultFieldView,
  getAvailableFieldViews,
  getFieldView,
  isFieldHidden,
  updateConfigOnFieldRename as updateFormConfigOnFieldRename,
  updateConfigOnFieldRemove as updateFormConfigOnFieldRemove,
  updateConfigOnFieldAdd as updateFormConfigOnFieldAdd,
} from './form-config';
