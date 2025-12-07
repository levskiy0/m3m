import { useState, useEffect, useRef, useCallback } from 'react';

interface UseColumnResizeOptions {
  modelId: string | undefined;
}

function loadColumnWidths(modelId: string | undefined): Record<string, number> {
  if (!modelId) return {};
  try {
    const saved = localStorage.getItem(`model-column-widths-${modelId}`);
    return saved ? JSON.parse(saved) : {};
  } catch {
    return {};
  }
}

export function useColumnResize({ modelId }: UseColumnResizeOptions) {
  const [columnWidths, setColumnWidths] = useState<Record<string, number>>(() => loadColumnWidths(modelId));
  const resizingRef = useRef<{ column: string; startX: number; startWidth: number } | null>(null);

  // Track which modelId the current columnWidths belong to
  // This prevents saving old model's widths to new model's localStorage
  const loadedModelIdRef = useRef<string | undefined>(undefined);

  // Load column widths when modelId changes
  useEffect(() => {
    loadedModelIdRef.current = undefined; // Reset before loading
    const loadedWidths = loadColumnWidths(modelId);
    // eslint-disable-next-line react-hooks/set-state-in-effect -- Required to sync state with localStorage on model change
    setColumnWidths(loadedWidths);
    loadedModelIdRef.current = modelId; // Mark as loaded for this model
  }, [modelId]);

  // Save column widths to localStorage - only when data belongs to current model
  useEffect(() => {
    // Only save if:
    // 1. We have a valid modelId
    // 2. The columnWidths belong to the current model (not stale data from previous model)
    // 3. There's actually data to save
    if (modelId && loadedModelIdRef.current === modelId && Object.keys(columnWidths).length > 0) {
      localStorage.setItem(`model-column-widths-${modelId}`, JSON.stringify(columnWidths));
    }
  }, [columnWidths, modelId]);

  const handleResizeStart = useCallback((e: React.MouseEvent, column: string, currentWidth: number) => {
    e.preventDefault();
    e.stopPropagation();
    resizingRef.current = { column, startX: e.clientX, startWidth: currentWidth };

    const handleMouseMove = (e: MouseEvent) => {
      if (!resizingRef.current) return;
      const delta = e.clientX - resizingRef.current.startX;
      const newWidth = Math.max(50, resizingRef.current.startWidth + delta);
      setColumnWidths(prev => ({ ...prev, [resizingRef.current!.column]: newWidth }));
    };

    const handleMouseUp = () => {
      resizingRef.current = null;
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
  }, []);

  const getColumnWidth = useCallback((key: string, defaultWidth: number = 150) => {
    return columnWidths[key] || defaultWidth;
  }, [columnWidths]);

  return {
    columnWidths,
    handleResizeStart,
    getColumnWidth,
  };
}
