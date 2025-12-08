/**
 * useSchemaEditor hook
 * Manages state and operations for the schema editor
 */
/* eslint-disable react-hooks/refs -- this hook intentionally exposes ref values for sortable IDs */

import { useState, useRef, useCallback } from 'react';
import { arrayMove } from '@dnd-kit/sortable';
import type { DragEndEvent } from '@dnd-kit/core';

import type { ModelField, TableConfig, FormConfig, Model } from '@/types';
import {
  createDefaultTableConfig,
  mergeTableConfig,
  updateConfigOnFieldRename as updateTableConfigOnRename,
  updateConfigOnFieldRemove as updateTableConfigOnRemove,
  updateConfigOnFieldAdd as updateTableConfigOnAdd,
  isSearchableFieldType,
} from '../../lib/table-config';
import {
  createDefaultFormConfig,
  mergeFormConfig,
  updateConfigOnFieldRename as updateFormConfigOnRename,
  updateConfigOnFieldRemove as updateFormConfigOnRemove,
  updateConfigOnFieldAdd as updateFormConfigOnAdd,
} from '../../lib/form-config';
import {
  createEmptyField,
  hasFieldsValidationErrors,
} from '../../lib';

interface UseSchemaEditorOptions {
  initialFields: ModelField[];
  initialTableConfig?: TableConfig;
  initialFormConfig?: FormConfig;
}

export function useSchemaEditor({
  initialFields,
  initialTableConfig,
  initialFormConfig,
}: UseSchemaEditorOptions) {
  const [fields, setFields] = useState<ModelField[]>(initialFields);
  const [hasChanges, setHasChanges] = useState(false);

  // Stable IDs for sortable
  const fieldIdsRef = useRef<string[]>([]);
  const fieldKeysRef = useRef<string[]>([]);
  const idCounterRef = useRef(0);

  // Configs
  const defaultTableConfig = createDefaultTableConfig(initialFields);
  const defaultFormConfig = createDefaultFormConfig(initialFields);

  const [tableConfig, setTableConfig] = useState<TableConfig>(
    mergeTableConfig(initialTableConfig, defaultTableConfig)
  );
  const [formConfig, setFormConfig] = useState<FormConfig>(
    mergeFormConfig(initialFormConfig, defaultFormConfig)
  );

  // Generate unique field ID
  const generateFieldId = useCallback(() => {
    idCounterRef.current += 1;
    return `field-${idCounterRef.current}-${Date.now()}`;
  }, []);

  // Initialize refs on first render
   
  if (fieldIdsRef.current.length === 0 && initialFields.length > 0) {
     
    fieldIdsRef.current = initialFields.map(() => generateFieldId());
     
    fieldKeysRef.current = initialFields.map(f => f.key);
  }

  // Reset state when model changes
  const resetState = useCallback((model: Model) => {
    setFields(model.fields);
    fieldIdsRef.current = model.fields.map(() => generateFieldId());
    fieldKeysRef.current = model.fields.map(f => f.key);

    const newDefaultTableConfig = createDefaultTableConfig(model.fields);
    const newDefaultFormConfig = createDefaultFormConfig(model.fields);

    setTableConfig(mergeTableConfig(model.table_config, newDefaultTableConfig));
    setFormConfig(mergeFormConfig(model.form_config, newDefaultFormConfig));
    setHasChanges(false);
  }, [generateFieldId]);

  // Add new field
  const addField = useCallback(() => {
    const newField = createEmptyField(fields.length);
    const newFields = [...fields, newField];

    setFields(newFields);
    fieldIdsRef.current = [...fieldIdsRef.current, generateFieldId()];
    fieldKeysRef.current = [...fieldKeysRef.current, newField.key];

    setTableConfig(prev => updateTableConfigOnAdd(prev, newField));
    setFormConfig(prev => updateFormConfigOnAdd(prev, newField));
    setHasChanges(true);
  }, [fields, generateFieldId]);

  // Update field
  const updateField = useCallback((index: number, updates: Partial<ModelField>) => {
    const oldKey = fieldKeysRef.current[index];
    const oldField = fields[index];
    const newFields = [...fields];
    newFields[index] = { ...newFields[index], ...updates };
    setFields(newFields);

    // Update key references if key changed
    if (updates.key && updates.key !== oldKey) {
      fieldKeysRef.current[index] = updates.key;

      setTableConfig(prev => updateTableConfigOnRename(prev, oldKey, updates.key!));
      setFormConfig(prev => updateFormConfigOnRename(prev, oldKey, updates.key!));
    }

    // If type changed to non-searchable, remove from searchable
    if (updates.type && updates.type !== oldField.type) {
      const fieldKey = updates.key || oldKey;
      if (!isSearchableFieldType(updates.type)) {
        setTableConfig(prev => ({
          ...prev,
          searchable: prev.searchable.filter(k => k !== fieldKey),
        }));
      }
    }

    setHasChanges(true);
  }, [fields]);

  // Remove field
  const removeField = useCallback((index: number) => {
    const keyToRemove = fieldKeysRef.current[index];

    setFields(prev => prev.filter((_, i) => i !== index));
    fieldIdsRef.current = fieldIdsRef.current.filter((_, i) => i !== index);
    fieldKeysRef.current = fieldKeysRef.current.filter((_, i) => i !== index);

    setTableConfig(prev => updateTableConfigOnRemove(prev, keyToRemove));
    setFormConfig(prev => updateFormConfigOnRemove(prev, keyToRemove));
    setHasChanges(true);
  }, []);

  // Handle drag end for field reordering
  const handleDragEnd = useCallback((event: DragEndEvent) => {
    const { active, over } = event;

    if (over && active.id !== over.id) {
      const oldIndex = fieldIdsRef.current.indexOf(active.id as string);
      const newIndex = fieldIdsRef.current.indexOf(over.id as string);

      if (oldIndex === -1 || newIndex === -1) return;

      const newFields = arrayMove(fields, oldIndex, newIndex);
      fieldIdsRef.current = arrayMove(fieldIdsRef.current, oldIndex, newIndex);
      fieldKeysRef.current = arrayMove(fieldKeysRef.current, oldIndex, newIndex);
      setFields(newFields);

      // Update configs with new field order
      const newKeys = newFields.map(f => f.key);
      setTableConfig(prev => ({
        ...prev,
        columns: newKeys.filter(k => prev.columns.includes(k)),
        sort_columns: newKeys.filter(k => prev.sort_columns.includes(k)),
      }));
      setFormConfig(prev => ({
        ...prev,
        field_order: newKeys,
      }));

      setHasChanges(true);
    }
  }, [fields]);

  // Update table config
  const updateTableConfig = useCallback((config: TableConfig) => {
    setTableConfig(config);
    setHasChanges(true);
  }, []);

  // Update form config
  const updateFormConfig = useCallback((config: FormConfig) => {
    setFormConfig(config);
    setHasChanges(true);
  }, []);

  // Validation
  const hasValidationErrors = hasFieldsValidationErrors(fields);

  return {
    // State
    fields,
    tableConfig,
    formConfig,
    hasChanges,
    hasValidationErrors,

    // Refs (for sortable) - access the ref inside the returned object
     
    fieldIds: fieldIdsRef.current,

    // Actions
    addField,
    updateField,
    removeField,
    handleDragEnd,
    updateTableConfig,
    updateFormConfig,
    resetState,
    setHasChanges,
  };
}
