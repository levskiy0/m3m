/**
 * useEnvEditor hook
 * Manages state and operations for the environment editor
 */

import { useState, useRef, useCallback } from 'react';
import { arrayMove } from '@dnd-kit/sortable';
import type { DragEndEvent } from '@dnd-kit/core';

import type { Environment, EnvType } from '@/types';
import type { EnvRowData } from '../components';

interface UseEnvEditorOptions {
  initialEnvVars: Environment[];
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
  const [hasChanges, setHasChanges] = useState(false);
  const [deletedKeys, setDeletedKeys] = useState<string[]>([]);

  // Stable IDs for sortable
  const envIdsRef = useRef<string[]>([]);
  const idCounterRef = useRef(0);

  // Generate unique env ID
  const generateEnvId = useCallback(() => {
    idCounterRef.current += 1;
    return `env-${idCounterRef.current}-${Date.now()}`;
  }, []);

  // Initialize refs on first render
  if (envIdsRef.current.length === 0 && initialEnvVars.length > 0) {
    envIdsRef.current = initialEnvVars.map(() => generateEnvId());
  }

  // Reset state when env vars change from server
  const resetState = useCallback(
    (newEnvVars: Environment[]) => {
      setEnvVars(
        newEnvVars.map((e) => ({
          key: e.key,
          type: e.type,
          value: e.value,
          isNew: false,
        }))
      );
      envIdsRef.current = newEnvVars.map(() => generateEnvId());
      setHasChanges(false);
      setDeletedKeys([]);
    },
    [generateEnvId]
  );

  // Add new env var
  const addEnvVar = useCallback(() => {
    const newEnv: EnvRowData = {
      key: '',
      type: 'string' as EnvType,
      value: '',
      isNew: true,
    };

    setEnvVars((prev) => [...prev, newEnv]);
    envIdsRef.current = [...envIdsRef.current, generateEnvId()];
    setHasChanges(true);
  }, [generateEnvId]);

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
    envIdsRef.current = envIdsRef.current.filter((_, i) => i !== index);
    setHasChanges(true);
  }, []);

  // Handle drag end for reordering
  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      const { active, over } = event;

      if (over && active.id !== over.id) {
        const oldIndex = envIdsRef.current.indexOf(active.id as string);
        const newIndex = envIdsRef.current.indexOf(over.id as string);

        if (oldIndex === -1 || newIndex === -1) return;

        const newEnvVars = arrayMove(envVars, oldIndex, newIndex);
        envIdsRef.current = arrayMove(envIdsRef.current, oldIndex, newIndex);
        setEnvVars(newEnvVars);
        setHasChanges(true);
      }
    },
    [envVars]
  );

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

    // Refs (for sortable)
    envIds: envIdsRef.current,

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
