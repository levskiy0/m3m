import { useState } from 'react';
import { useParams } from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';
import { Folder, FileCode, Save } from 'lucide-react';
import { toast } from 'sonner';

import { storageApi } from '@/api';
import { getLanguageFromFilename } from '@/lib/utils';
import { EditorTabs, EditorTab } from '@/components/ui/editor-tabs';
import type { StorageItem } from '@/types';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { LoadingButton } from '@/components/ui/loading-button';
import { CodeEditor } from '@/components/shared/code-editor';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { FileBrowser } from './components';

interface FileEditorTab {
  id: string;
  name: string;
  path: string;
  content: string;
  originalContent: string;
  isNew: boolean;
  currentDir: string;
}

export function StoragePage() {
  const { projectId } = useParams<{ projectId: string }>();
  const queryClient = useQueryClient();

  // Tabs state
  const [tabs, setTabs] = useState<FileEditorTab[]>([]);
  const [activeTabId, setActiveTabId] = useState<string | null>(null);
  const [savingTabId, setSavingTabId] = useState<string | null>(null);

  // Close confirmation
  const [closeTabConfirmOpen, setCloseTabConfirmOpen] = useState(false);
  const [tabToClose, setTabToClose] = useState<string | null>(null);

  // Track current directory for new files
  const [currentDir] = useState('');

  const activeTab = tabs.find((t) => t.id === activeTabId);

  const handleCreateFile = () => {
    const newTab: FileEditorTab = {
      id: `new-${Date.now()}`,
      name: 'untitled.txt',
      path: '',
      content: '',
      originalContent: '',
      isNew: true,
      currentDir: currentDir,
    };
    setTabs((prev) => [...prev, newTab]);
    setActiveTabId(newTab.id);
  };

  const handleEditFile = async (item: StorageItem) => {
    // Check if already open
    const existing = tabs.find((t) => t.path === item.path);
    if (existing) {
      setActiveTabId(existing.id);
      return;
    }

    try {
      const blob = await storageApi.download(projectId!, item.path);
      const text = await blob.text();
      const newTab: FileEditorTab = {
        id: `edit-${Date.now()}`,
        name: item.name,
        path: item.path,
        content: text,
        originalContent: text,
        isNew: false,
        currentDir: '',
      };
      setTabs((prev) => [...prev, newTab]);
      setActiveTabId(newTab.id);
    } catch {
      toast.error('Failed to load file');
    }
  };

  const handleSaveTab = async (tab: FileEditorTab) => {
    setSavingTabId(tab.id);
    try {
      if (tab.isNew) {
        await storageApi.createFile(projectId!, {
          path: tab.currentDir,
          name: tab.name,
          content: tab.content,
        });
        const newPath = tab.currentDir ? `${tab.currentDir}/${tab.name}` : tab.name;
        setTabs((prev) =>
          prev.map((t) =>
            t.id === tab.id
              ? { ...t, isNew: false, path: newPath, originalContent: tab.content }
              : t
          )
        );
      } else {
        await storageApi.updateFile(projectId!, tab.path, tab.content);
        setTabs((prev) =>
          prev.map((t) =>
            t.id === tab.id ? { ...t, originalContent: tab.content } : t
          )
        );
      }
      queryClient.invalidateQueries({ queryKey: ['storage', projectId] });
      toast.success('File saved');
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Save failed');
    } finally {
      setSavingTabId(null);
    }
  };

  const handleCloseTab = (tabId: string, force = false) => {
    const tab = tabs.find((t) => t.id === tabId);
    if (!tab) return;

    const isDirty = tab.content !== tab.originalContent;
    if (isDirty && !force) {
      setTabToClose(tabId);
      setCloseTabConfirmOpen(true);
      return;
    }

    setTabs((prev) => prev.filter((t) => t.id !== tabId));
    if (activeTabId === tabId) {
      setActiveTabId(null);
    }
  };

  const handleTabContentChange = (tabId: string, content: string) => {
    setTabs((prev) =>
      prev.map((t) => (t.id === tabId ? { ...t, content } : t))
    );
  };

  const handleTabNameChange = (tabId: string, name: string) => {
    setTabs((prev) =>
      prev.map((t) => (t.id === tabId ? { ...t, name } : t))
    );
  };

  if (!projectId) return null;

  return (
    <>
      {/* Tabs Bar */}
      <div className="w-full">
        <EditorTabs>
          <EditorTab
            active={activeTabId === null}
            onClick={() => setActiveTabId(null)}
            icon={<Folder className="size-4" />}
          >
            Files
          </EditorTab>

          {tabs.map((tab) => (
            <EditorTab
              key={tab.id}
              active={activeTabId === tab.id}
              onClick={() => setActiveTabId(tab.id)}
              icon={<FileCode className="size-4" />}
              dirty={tab.content !== tab.originalContent}
              onClose={() => handleCloseTab(tab.id)}
            >
              {tab.name}
            </EditorTab>
          ))}
        </EditorTabs>

        <Card className="flex flex-col gap-0 rounded-t-none py-0 overflow-hidden" style={{ height: 'calc(100vh - 120px)' }}>
          {activeTabId === null ? (
            <FileBrowser
              projectId={projectId}
              mode="browse"
              onEditFile={handleEditFile}
              onCreateFile={handleCreateFile}
              showUpload={true}
              className="flex-1"
            />
          ) : activeTab ? (
            <>
              {/* Editor View */}
              <div className="flex items-center justify-between px-6 py-4 border-b">
                <div className="flex items-center gap-2">
                  {activeTab.isNew && (
                    <Input
                      value={activeTab.name}
                      onChange={(e) => handleTabNameChange(activeTab.id, e.target.value)}
                      className="h-8 w-48"
                      placeholder="filename.txt"
                    />
                  )}
                  {!activeTab.isNew && (
                    <span className="text-sm text-muted-foreground">
                      {activeTab.path}
                    </span>
                  )}
                </div>
                <LoadingButton
                  size="sm"
                  onClick={() => handleSaveTab(activeTab)}
                  disabled={
                    (activeTab.content === activeTab.originalContent && !activeTab.isNew) ||
                    (activeTab.isNew && !activeTab.name)
                  }
                  loading={savingTabId === activeTab.id}
                >
                  <Save className="mr-2 size-4" />
                  Save
                </LoadingButton>
              </div>
              <div className="flex-1 min-h-0">
                <CodeEditor
                  value={activeTab.content}
                  onChange={(value) => handleTabContentChange(activeTab.id, value)}
                  language={getLanguageFromFilename(activeTab.name)}
                  height="100%"
                />
              </div>
            </>
          ) : null}
        </Card>
      </div>

      {/* Close Tab Confirm */}
      <ConfirmDialog
        open={closeTabConfirmOpen}
        onOpenChange={setCloseTabConfirmOpen}
        title="Unsaved Changes"
        description="You have unsaved changes. Are you sure you want to close this file?"
        confirmLabel="Close without saving"
        variant="destructive"
        onConfirm={() => {
          if (tabToClose) {
            handleCloseTab(tabToClose, true);
            setTabToClose(null);
          }
          setCloseTabConfirmOpen(false);
        }}
      />
    </>
  );
}
