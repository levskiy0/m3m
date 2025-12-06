import { useState, useEffect, useRef, useCallback, useMemo } from 'react';

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
  const columnWidthsKey = `model-column-widths-${modelId}`;

  // Use useMemo to derive state from modelId changes
  const initialWidths = useMemo(() => loadColumnWidths(modelId), [modelId]);
  const [columnWidths, setColumnWidths] = useState<Record<string, number>>(initialWidths);

  // Reset when modelId changes
  const prevModelIdRef = useRef(modelId);
  if (prevModelIdRef.current !== modelId) {
    prevModelIdRef.current = modelId;
    // This is a render-time update pattern, not in useEffect
    setColumnWidths(loadColumnWidths(modelId));
  }

  const resizingRef = useRef<{ column: string; startX: number; startWidth: number } | null>(null);

  // Save column widths to localStorage
  useEffect(() => {
    if (modelId && Object.keys(columnWidths).length > 0) {
      localStorage.setItem(columnWidthsKey, JSON.stringify(columnWidths));
    }
  }, [columnWidths, columnWidthsKey, modelId]);

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
