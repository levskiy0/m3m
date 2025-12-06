import { useState, useCallback, useMemo } from 'react';
import type { ModelData } from '@/types';

export function useSelection(data: ModelData[]) {
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());

  const handleSelectRow = useCallback((id: string, checked: boolean) => {
    setSelectedIds(prev => {
      const next = new Set(prev);
      if (checked) {
        next.add(id);
      } else {
        next.delete(id);
      }
      return next;
    });
  }, []);

  const handleSelectAll = useCallback((checked: boolean) => {
    if (checked) {
      setSelectedIds(new Set(data.map(d => d._id)));
    } else {
      setSelectedIds(new Set());
    }
  }, [data]);

  const clearSelection = useCallback(() => {
    setSelectedIds(new Set());
  }, []);

  const allSelected = useMemo(() => {
    return data.length > 0 && data.every(d => selectedIds.has(d._id));
  }, [data, selectedIds]);

  const someSelected = selectedIds.size > 0;

  return {
    selectedIds,
    handleSelectRow,
    handleSelectAll,
    clearSelection,
    allSelected,
    someSelected,
  };
}
