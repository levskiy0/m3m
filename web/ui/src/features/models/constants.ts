import type { FieldType } from '@/types';

export const FIELD_TYPES: { value: FieldType; label: string }[] = [
  { value: 'string', label: 'String' },
  { value: 'text', label: 'Text' },
  { value: 'number', label: 'Number' },
  { value: 'float', label: 'Float' },
  { value: 'bool', label: 'Boolean' },
  { value: 'document', label: 'Document (JSON)' },
  { value: 'file', label: 'File' },
  { value: 'ref', label: 'Reference' },
  { value: 'date', label: 'Date' },
  { value: 'datetime', label: 'DateTime' },
  { value: 'select', label: 'Select' },
  { value: 'multiselect', label: 'Multi Select' },
];

// Field views per field type (what widgets can be used for each type)
// IMPORTANT: Keep in sync with backend validViews in model_schema_validation.go
export const FIELD_VIEWS: Record<FieldType, { value: string; label: string }[]> = {
  string: [
    { value: 'input', label: 'Input' },
    { value: 'select', label: 'Select' },
  ],
  text: [
    { value: 'textarea', label: 'Textarea' },
    // { value: 'tiptap', label: 'Rich Text Editor' },
    // { value: 'markdown', label: 'Markdown' },
  ],
  number: [
    { value: 'input', label: 'Input' },
    { value: 'slider', label: 'Slider' },
  ],
  float: [
    { value: 'input', label: 'Input' },
    { value: 'slider', label: 'Slider' },
  ],
  bool: [
    { value: 'checkbox', label: 'Checkbox' },
    { value: 'switch', label: 'Switch' },
  ],
  date: [
    { value: 'datepicker', label: 'Date Picker' },
  ],
  datetime: [
    { value: 'datetimepicker', label: 'DateTime Picker' },
  ],
  file: [
    { value: 'file', label: 'File Upload' },
    { value: 'image', label: 'Image Picker' },
  ],
  ref: [
    { value: 'select', label: 'Select' },
  ],
  document: [
    { value: 'json', label: 'JSON Editor' },
  ],
  select: [
    { value: 'select', label: 'Select' },
    { value: 'combobox', label: 'Combobox' },
    { value: 'radiogroup', label: 'Radio Group' },
  ],
  multiselect: [
    { value: 'multiselect', label: 'Multi Select' },
    { value: 'checkboxgroup', label: 'Checkbox Group' },
  ],
};
