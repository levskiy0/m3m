/**
 * useEnvEditor hook
 * Manages state and operations for the environment editor
 */

import { useState, useCallback } from 'react';
import { arrayMove } from '@dnd-kit/sortable';
import type { DragEndEvent } from '@dnd-kit/core';

import type { Environment, EnvType } from '@/types';
import type { EnvRowData } from '../components';

interface UseEnvEditorOptions {
  initialEnvVars: Environment[];
}

// Counter for generating unique IDs
let idCounter = 0;
function generateEnvId() {
  idCounter += 1;
  return `env-${idCounter}-${Date.now()}`;
}

export function useEnvEditor({ initialEnvVars }: UseEnvEditorOptions) {
  const [envVars, setEnvVars] = useState<EnvRowData[]>(
    initialEnvVars.map((e) => ({
      key: e.key,
      type: e.type,
      value: e.value,
      isNew: false,
    }))
  );
  const [envIds, setEnvIds] = useState<string[]>(() =>
    initialEnvVars.map(() => generateEnvId())
  );
  const [hasChanges, setHasChanges] = useState(false);
  const [deletedKeys, setDeletedKeys] = useState<string[]>([]);

  // Reset state when env vars change from server
  const resetState = useCallback((newEnvVars: Environment[]) => {
    setEnvVars(
      newEnvVars.map((e) => ({
        key: e.key,
        type: e.type,
        value: e.value,
        isNew: false,
      }))
    );
    setEnvIds(newEnvVars.map(() => generateEnvId()));
    setHasChanges(false);
    setDeletedKeys([]);
  }, []);

  // Add new env var
  const addEnvVar = useCallback(() => {
    const newEnv: EnvRowData = {
      key: '',
      type: 'string' as EnvType,
      value: '',
      isNew: true,
    };

    setEnvVars((prev) => [...prev, newEnv]);
    setEnvIds((prev) => [...prev, generateEnvId()]);
    setHasChanges(true);
  }, []);

  // Update env var
  const updateEnvVar = useCallback(
    (index: number, updates: Partial<EnvRowData>) => {
      setEnvVars((prev) => {
        const newEnvVars = [...prev];
        newEnvVars[index] = { ...newEnvVars[index], ...updates };
        return newEnvVars;
      });
      setHasChanges(true);
    },
    []
  );

  // Remove env var
  const removeEnvVar = useCallback((index: number) => {
    setEnvVars((prev) => {
      const envToRemove = prev[index];
      // Track deleted keys for existing vars (not new ones)
      if (!envToRemove.isNew && envToRemove.key) {
        setDeletedKeys((keys) => [...keys, envToRemove.key]);
      }
      return prev.filter((_, i) => i !== index);
    });
    setEnvIds((prev) => prev.filter((_, i) => i !== index));
    setHasChanges(true);
  }, []);

  // Handle drag end for reordering
  const handleDragEnd = useCallback((event: DragEndEvent) => {
    const { active, over } = event;

    if (over && active.id !== over.id) {
      setEnvIds((currentIds) => {
        const oldIndex = currentIds.indexOf(active.id as string);
        const newIndex = currentIds.indexOf(over.id as string);

        if (oldIndex === -1 || newIndex === -1) return currentIds;

        setEnvVars((currentEnvVars) => arrayMove(currentEnvVars, oldIndex, newIndex));
        setHasChanges(true);

        return arrayMove(currentIds, oldIndex, newIndex);
      });
    }
  }, []);

  // Get changes for saving
  const getChanges = useCallback(() => {
    const toCreate = envVars.filter((e) => e.isNew && e.key);
    const toUpdate = envVars.filter((e) => !e.isNew && e.key);
    const toDelete = deletedKeys;

    return { toCreate, toUpdate, toDelete };
  }, [envVars, deletedKeys]);

  // Validation
  const hasValidationErrors = envVars.some((e) => !e.key);

  // Check for duplicate keys
  const hasDuplicateKeys = (() => {
    const keys = envVars.map((e) => e.key).filter(Boolean);
    return new Set(keys).size !== keys.length;
  })();

  return {
    // State
    envVars,
    hasChanges,
    hasValidationErrors: hasValidationErrors || hasDuplicateKeys,
    deletedKeys,

    // IDs for sortable
    envIds,

    // Actions
    addEnvVar,
    updateEnvVar,
    removeEnvVar,
    handleDragEnd,
    resetState,
    setHasChanges,
    getChanges,
  };
}
