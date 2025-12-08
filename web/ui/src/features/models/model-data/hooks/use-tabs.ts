import { useState, useCallback, useMemo } from 'react';
import type { ModelData, ModelField } from '@/types';
import type { Tab, TabType } from '../types';

const INITIAL_TABS: Tab[] = [{ id: 'table', type: 'table', title: 'Table' }];

export function useTabs(fields: ModelField[] = []) {
  const [tabs, setTabs] = useState<Tab[]>(INITIAL_TABS);
  const [activeTabId, setActiveTabId] = useState('table');

  // Check if any tab has unsaved changes
  const hasUnsavedChanges = useMemo(() => {
    return tabs.some(t => t.hasChanges);
  }, [tabs]);

  // Get list of tab IDs with unsaved changes
  const tabsWithChanges = useMemo(() => {
    return tabs.filter(t => t.hasChanges).map(t => t.id);
  }, [tabs]);

  const openTab = useCallback((type: TabType, data?: ModelData) => {
    const tabId = type === 'create' ? `create-${Date.now()}` :
                  type === 'table' ? 'table' :
                  `${type}-${data?._id}`;

    // Check if tab already exists
    const existingTab = tabs.find(t => t.id === tabId);
    if (existingTab) {
      setActiveTabId(tabId);
      return;
    }

    const title = type === 'create' ? 'New Record' :
                  type === 'edit' ? `Edit: ${data?._id.slice(-6)}` :
                  type === 'view' ? `View: ${data?._id.slice(-6)}` : 'Table';

    const formData: Record<string, unknown> = {};
    if (type === 'edit' && data) {
      fields.forEach((field) => {
        formData[field.key] = data[field.key];
      });
    }

    const newTab: Tab = { id: tabId, type, title, data, formData, fieldErrors: {} };

    setTabs(prev => {
      let newTabs = [...prev, newTab];
      // Close view tab when opening edit for the same record
      if (type === 'edit' && data) {
        const viewTabId = `view-${data._id}`;
        newTabs = newTabs.filter(t => t.id !== viewTabId);
      }
      return newTabs;
    });
    setActiveTabId(tabId);
  }, [tabs, fields]);

  const closeTab = useCallback((tabId: string) => {
    if (tabId === 'table') return; // Can't close table tab

    setTabs(prev => {
      const newTabs = prev.filter(t => t.id !== tabId);
      // If closing active tab, switch to previous or table
      if (activeTabId === tabId) {
        const idx = prev.findIndex(t => t.id === tabId);
        const newActiveIdx = Math.max(0, idx - 1);
        setActiveTabId(newTabs[newActiveIdx]?.id || 'table');
      }
      return newTabs;
    });
  }, [activeTabId]);

  const closeTabsForRecord = useCallback((recordId: string) => {
    setTabs(prev => prev.filter(t => !t.id.includes(recordId)));
    setActiveTabId('table');
  }, []);

  const updateTabFormData = useCallback((tabId: string, formData: Record<string, unknown>) => {
    setTabs(prev => prev.map(t => t.id === tabId ? { ...t, formData } : t));
  }, []);

  const updateTabErrors = useCallback((tabId: string, fieldErrors: Record<string, string>) => {
    setTabs(prev => prev.map(t => t.id === tabId ? { ...t, fieldErrors } : t));
  }, []);

  const getActiveTab = useCallback(() => {
    return tabs.find(t => t.id === activeTabId);
  }, [tabs, activeTabId]);

  // Mark a tab as having unsaved changes
  const markTabAsChanged = useCallback((tabId: string, changed: boolean = true) => {
    setTabs(prev => prev.map(t => t.id === tabId ? { ...t, hasChanges: changed } : t));
  }, []);

  // Clear hasChanges flag for a tab
  const clearTabChanges = useCallback((tabId: string) => {
    setTabs(prev => prev.map(t => t.id === tabId ? { ...t, hasChanges: false } : t));
  }, []);

  // Reset tabs to initial state (only table tab)
  const resetTabs = useCallback(() => {
    setTabs(INITIAL_TABS);
    setActiveTabId('table');
  }, []);

  return {
    tabs,
    activeTabId,
    setActiveTabId,
    openTab,
    closeTab,
    closeTabsForRecord,
    updateTabFormData,
    updateTabErrors,
    getActiveTab,
    markTabAsChanged,
    clearTabChanges,
    resetTabs,
    hasUnsavedChanges,
    tabsWithChanges,
  };
}
