import { useState, useCallback } from 'react';

interface UseFormDialogOptions<T> {
  defaultValues?: Partial<T>;
}

interface UseFormDialogReturn<T> {
  isOpen: boolean;
  mode: 'create' | 'edit';
  selectedItem: T | null;
  open: () => void;
  openEdit: (item: T) => void;
  close: () => void;
}

/**
 * Hook for managing dialog state for create/edit operations.
 * Handles open/close state and tracks whether we're creating or editing.
 */
export function useFormDialog<T>(_options: UseFormDialogOptions<T> = {}): UseFormDialogReturn<T> {
  const [isOpen, setIsOpen] = useState(false);
  const [mode, setMode] = useState<'create' | 'edit'>('create');
  const [selectedItem, setSelectedItem] = useState<T | null>(null);

  const open = useCallback(() => {
    setMode('create');
    setSelectedItem(null);
    setIsOpen(true);
  }, []);

  const openEdit = useCallback((item: T) => {
    setMode('edit');
    setSelectedItem(item);
    setIsOpen(true);
  }, []);

  const close = useCallback(() => {
    setIsOpen(false);
    setSelectedItem(null);
  }, []);

  return {
    isOpen,
    mode,
    selectedItem,
    open,
    openEdit,
    close,
  };
}

interface UseDeleteDialogReturn<T> {
  isOpen: boolean;
  itemToDelete: T | null;
  open: (item: T) => void;
  close: () => void;
}

/**
 * Hook for managing delete confirmation dialog state.
 */
export function useDeleteDialog<T>(): UseDeleteDialogReturn<T> {
  const [isOpen, setIsOpen] = useState(false);
  const [itemToDelete, setItemToDelete] = useState<T | null>(null);

  const open = useCallback((item: T) => {
    setItemToDelete(item);
    setIsOpen(true);
  }, []);

  const close = useCallback(() => {
    setIsOpen(false);
    setItemToDelete(null);
  }, []);

  return {
    isOpen,
    itemToDelete,
    open,
    close,
  };
}
