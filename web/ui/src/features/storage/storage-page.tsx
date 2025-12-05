import { useState, useRef, useCallback, useMemo } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Upload,
  FolderPlus,
  FilePlus,
  Download,
  Trash2,
  Pencil,
  Link,
  Folder,
  FolderOpen,
  ArrowUp,
  File,
  FileText,
  FileCode,
  FileImage,
  FileArchive,
  FileAudio,
  FileVideo,
  RefreshCw,
  ChevronRight,
  Home,
  ExternalLink,
  X,
  Save,
  Move,
  Check,
} from 'lucide-react';
import { toast } from 'sonner';

import { storageApi } from '@/api';
import { EDITABLE_MIME_TYPES } from '@/lib/constants';
import type { StorageItem } from '@/types';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from '@/components/ui/context-menu';
import { Input } from '@/components/ui/input';
import { Field, FieldLabel } from '@/components/ui/field';
import { CodeEditor } from '@/components/shared/code-editor';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { cn } from '@/lib/utils';

interface EditorTab {
  id: string;
  name: string;
  path: string; // empty for new files
  content: string;
  originalContent: string;
  isNew: boolean;
}

function getFileExtension(filename: string): string {
  return filename.split('.').pop()?.toLowerCase() || '';
}

function isImageFile(filename: string): boolean {
  const ext = getFileExtension(filename);
  return ['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg', 'ico', 'bmp'].includes(ext);
}

function formatFileSize(bytes: number): string {
  if (bytes === 0) return '-';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function getLanguageFromFilename(filename: string): string {
  if (filename.endsWith('.json')) return 'json';
  if (filename.endsWith('.yaml') || filename.endsWith('.yml')) return 'yaml';
  if (filename.endsWith('.js')) return 'javascript';
  if (filename.endsWith('.ts')) return 'typescript';
  if (filename.endsWith('.html')) return 'html';
  if (filename.endsWith('.css')) return 'css';
  if (filename.endsWith('.md')) return 'markdown';
  return 'plaintext';
}

export function StoragePage() {
  const { projectId } = useParams<{ projectId: string }>();
  const queryClient = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [currentPath, setCurrentPath] = useState('');
  const [selectedItem, setSelectedItem] = useState<StorageItem | null>(null);

  // Multi-select
  const [selectedPaths, setSelectedPaths] = useState<Set<string>>(new Set());
  const [lastSelectedIndex, setLastSelectedIndex] = useState<number | null>(null);

  // Drag & drop
  const [isDragging, setIsDragging] = useState(false);
  const dragCounter = useRef(0);

  // Move dialog
  const [moveDialogOpen, setMoveDialogOpen] = useState(false);
  const [moveTargetPath, setMoveTargetPath] = useState('');

  // Tabs
  const [tabs, setTabs] = useState<EditorTab[]>([]);
  const [activeTabId, setActiveTabId] = useState<string | null>(null);

  // Dialogs
  const [createFolderOpen, setCreateFolderOpen] = useState(false);
  const [renameOpen, setRenameOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [closeTabConfirmOpen, setCloseTabConfirmOpen] = useState(false);
  const [tabToClose, setTabToClose] = useState<string | null>(null);

  // Form state
  const [newName, setNewName] = useState('');

  const activeTab = tabs.find((t) => t.id === activeTabId);

  const { data: items = [], isLoading } = useQuery({
    queryKey: ['storage', projectId, currentPath],
    queryFn: () => storageApi.list(projectId!, currentPath),
    enabled: !!projectId,
  });

  // Sorted items (folders first, then alphabetically)
  const sortedItems = useMemo(() => {
    return [...items].sort((a, b) => {
      if (a.is_dir && !b.is_dir) return -1;
      if (!a.is_dir && b.is_dir) return 1;
      return a.name.localeCompare(b.name);
    });
  }, [items]);

  // Get all folders for move dialog
  const { data: allFolders = [] } = useQuery({
    queryKey: ['storage-folders', projectId],
    queryFn: async () => {
      const getAllFolders = async (path: string): Promise<StorageItem[]> => {
        const items = await storageApi.list(projectId!, path);
        const folders = items.filter((i) => i.is_dir);
        const subFolders: StorageItem[] = [];
        for (const folder of folders) {
          const sub = await getAllFolders(folder.path);
          subFolders.push(...sub);
        }
        return [...folders, ...subFolders];
      };
      return getAllFolders('');
    },
    enabled: !!projectId && moveDialogOpen,
  });

  // Mutations
  const createFolderMutation = useMutation({
    mutationFn: (name: string) =>
      storageApi.mkdir(projectId!, { path: currentPath, name }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['storage', projectId] });
      setCreateFolderOpen(false);
      setNewName('');
      toast.success('Folder created');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to create folder');
    },
  });

  const uploadMutation = useMutation({
    mutationFn: (file: File) => storageApi.upload(projectId!, currentPath, file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['storage', projectId] });
      toast.success('File uploaded');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Upload failed');
    },
  });

  const renameMutation = useMutation({
    mutationFn: () =>
      storageApi.rename(projectId!, { path: selectedItem!.path, newName }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['storage', projectId] });
      setRenameOpen(false);
      setSelectedItem(null);
      setNewName('');
      toast.success('Renamed successfully');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Rename failed');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (path: string) => storageApi.delete(projectId!, path),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['storage', projectId] });
      setDeleteOpen(false);
      setSelectedItem(null);
      toast.success('Deleted successfully');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Delete failed');
    },
  });

  const deleteSelectedMutation = useMutation({
    mutationFn: async () => {
      const paths = Array.from(selectedPaths);
      for (const path of paths) {
        await storageApi.delete(projectId!, path);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['storage', projectId] });
      setDeleteOpen(false);
      setSelectedPaths(new Set());
      toast.success('Deleted successfully');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Delete failed');
    },
  });

  const moveMutation = useMutation({
    mutationFn: async (targetDir: string) => {
      const paths = Array.from(selectedPaths);
      for (const path of paths) {
        await storageApi.move(projectId!, path, targetDir);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['storage', projectId] });
      setMoveDialogOpen(false);
      setSelectedPaths(new Set());
      setMoveTargetPath('');
      toast.success('Moved successfully');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Move failed');
    },
  });

  const [savingTabId, setSavingTabId] = useState<string | null>(null);

  const handleSaveTab = async (tab: EditorTab) => {
    setSavingTabId(tab.id);
    try {
      if (tab.isNew) {
        await storageApi.createFile(projectId!, {
          path: currentPath,
          name: tab.name,
          content: tab.content,
        });
        // Update tab to not be new anymore
        const newPath = currentPath ? `${currentPath}/${tab.name}` : tab.name;
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

  const handleFileUpload = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (file) {
        uploadMutation.mutate(file);
      }
      e.target.value = '';
    },
    [uploadMutation]
  );

  const handleItemClick = (item: StorageItem) => {
    if (item.is_dir) {
      setCurrentPath(item.path);
      setSelectedPaths(new Set());
      setLastSelectedIndex(null);
    }
  };

  // Multi-select handler
  const handleItemSelect = useCallback(
    (item: StorageItem, index: number, e: React.MouseEvent) => {
      const isMetaKey = e.metaKey || e.ctrlKey;
      const isShiftKey = e.shiftKey;

      if (isShiftKey && lastSelectedIndex !== null) {
        // Shift-click: select range
        const start = Math.min(lastSelectedIndex, index);
        const end = Math.max(lastSelectedIndex, index);
        const newSelection = new Set(selectedPaths);
        for (let i = start; i <= end; i++) {
          newSelection.add(sortedItems[i].path);
        }
        setSelectedPaths(newSelection);
      } else if (isMetaKey) {
        // Ctrl/Cmd-click: toggle selection
        const newSelection = new Set(selectedPaths);
        if (newSelection.has(item.path)) {
          newSelection.delete(item.path);
        } else {
          newSelection.add(item.path);
        }
        setSelectedPaths(newSelection);
        setLastSelectedIndex(index);
      } else {
        // Regular click: single selection
        setSelectedPaths(new Set([item.path]));
        setLastSelectedIndex(index);
      }
    },
    [selectedPaths, lastSelectedIndex, sortedItems]
  );

  // Clear selection when clicking empty area
  const handleClearSelection = useCallback((e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      setSelectedPaths(new Set());
      setLastSelectedIndex(null);
    }
  }, []);

  // Drag & drop handlers for file upload
  const handleDragEnter = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    dragCounter.current++;
    if (e.dataTransfer.types.includes('Files')) {
      setIsDragging(true);
    }
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    dragCounter.current--;
    if (dragCounter.current === 0) {
      setIsDragging(false);
    }
  }, []);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  }, []);

  const handleDrop = useCallback(
    async (e: React.DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      dragCounter.current = 0;
      setIsDragging(false);

      const files = Array.from(e.dataTransfer.files);
      if (files.length === 0) return;

      for (const file of files) {
        try {
          await storageApi.upload(projectId!, currentPath, file);
        } catch {
          toast.error(`Failed to upload ${file.name}`);
        }
      }
      queryClient.invalidateQueries({ queryKey: ['storage', projectId] });
      toast.success(`Uploaded ${files.length} file(s)`);
    },
    [projectId, currentPath, queryClient]
  );

  const handleCreateFile = () => {
    const newTab: EditorTab = {
      id: `new-${Date.now()}`,
      name: 'untitled.txt',
      path: '',
      content: '',
      originalContent: '',
      isNew: true,
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
      const newTab: EditorTab = {
        id: `edit-${Date.now()}`,
        name: item.name,
        path: item.path,
        content: text,
        originalContent: text,
        isNew: false,
      };
      setTabs((prev) => [...prev, newTab]);
      setActiveTabId(newTab.id);
    } catch {
      toast.error('Failed to load file');
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

  const handleDownload = (item: StorageItem) => {
    if (item.download_url) {
      window.open(item.download_url, '_blank');
    }
  };

  const pathSegments = currentPath ? currentPath.split('/').filter(Boolean) : [];

  const navigateToPath = (index: number) => {
    if (index < 0) {
      setCurrentPath('');
    } else {
      setCurrentPath(pathSegments.slice(0, index + 1).join('/'));
    }
  };

  const handleOpenInNewTab = (item: StorageItem) => {
    if (item.url) {
      window.open(item.url, '_blank');
    }
  };

  const handleCopyLink = async (item: StorageItem) => {
    if (!item.url) return;
    try {
      await navigator.clipboard.writeText(item.url);
      toast.success('Link copied');
    } catch {
      toast.error('Failed to copy link');
    }
  };

  const getFileIcon = (item: StorageItem) => {
    if (item.is_dir) {
      return <Folder className="size-6 text-blue-500 shrink-0" />;
    }

    const ext = getFileExtension(item.name);

    // Images - show thumbnail
    if (isImageFile(item.name) && item.thumb_url) {
      return (
        <div className="size-6 shrink-0 rounded overflow-hidden bg-muted flex items-center justify-center">
          <img
            src={item.thumb_url}
            alt={item.name}
            className="h-full w-full object-cover"
            onError={(e) => {
              e.currentTarget.style.display = 'none';
              e.currentTarget.parentElement?.classList.add('fallback');
            }}
          />
          <FileImage className="size-6 text-muted-foreground hidden" />
        </div>
      );
    }

    // Text/documents
    if (['txt', 'md', 'doc', 'docx', 'pdf', 'rtf'].includes(ext)) {
      return <FileText className="size-6 text-orange-500 shrink-0" />;
    }

    // Code
    if (['js', 'ts', 'jsx', 'tsx', 'json', 'html', 'css', 'py', 'go', 'rs', 'java', 'c', 'cpp', 'h', 'yml', 'yaml', 'xml', 'sh', 'bash'].includes(ext)) {
      return <FileCode className="size-6 text-green-500 shrink-0" />;
    }

    // Archives
    if (['zip', 'rar', '7z', 'tar', 'gz', 'bz2'].includes(ext)) {
      return <FileArchive className="size-6 text-yellow-500 shrink-0" />;
    }

    // Audio
    if (['mp3', 'wav', 'ogg', 'flac', 'm4a', 'aac'].includes(ext)) {
      return <FileAudio className="size-6 text-purple-500 shrink-0" />;
    }

    // Video
    if (['mp4', 'webm', 'mkv', 'avi', 'mov', 'wmv'].includes(ext)) {
      return <FileVideo className="size-8 text-pink-500 shrink-0" />;
    }

    return <File className="size-6 text-muted-foreground shrink-0" />;
  };

  return (
      <>
        <Card className="flex flex-col gap-0" style={{height: 'calc(100vh - 85px)'}}>
          {/* Tabs Bar */}
          <div className="flex items-end border-b px-4">
            {/* Files tab */}
            <button
                onClick={() => setActiveTabId(null)}
                className={cn(
                    'flex items-center gap-2 px-4 py-2 text-sm border-t border-l border-r',
                    activeTabId === null
                        ? 'border-border bg-background'
                        : 'border-transparent text-muted-foreground hover:text-foreground'
                )}
                style={{
                  borderTopLeftRadius: 6,
                  borderTopRightRadius: 6,
                  marginBottom: activeTabId === null ? -1 : 0,
                }}
            >
              <Folder className="size-4"/>
              Files
            </button>

            {/* Open file tabs */}
            {tabs.map((tab) => {
              const isDirty = tab.content !== tab.originalContent;
              const isActive = activeTabId === tab.id;
              return (
                  <div
                      key={tab.id}
                      className={cn(
                          'group flex items-center gap-3 px-4 py-2 text-sm border-t border-l border-r',
                          isActive
                              ? 'border-border bg-background'
                              : 'border-transparent text-muted-foreground hover:text-foreground'
                      )}
                      style={{
                        borderTopLeftRadius: 6,
                        borderTopRightRadius: 6,
                        marginBottom: isActive ? -1 : 0,
                      }}
                  >
                    <button
                        onClick={() => setActiveTabId(tab.id)}
                        className="flex items-center gap-2"
                    >
                      <FileCode className="size-4"/>
                      <span className="max-w-32 truncate">{tab.name}</span>
                      {isDirty && <span className="text-orange-500">*</span>}
                    </button>
                    <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleCloseTab(tab.id);
                        }}
                        className={cn(
                            'ml-1 p-0.5 rounded hover:bg-muted',
                            isActive ? '' : 'opacity-0 group-hover:opacity-100'
                        )}
                    >
                      <X className="size-3"/>
                    </button>
                  </div>
              );
            })}
          </div>

          {/* Content */}
          {activeTabId === null ? (
              <>
                {/* File Browser */}
                <CardHeader className="py-4 grid-rows-[none]">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-1 text-sm text-muted-foreground">
                      <Button
                          variant="ghost"
                          size="sm"
                          className="h-6 p-2"
                          onClick={() => setCurrentPath('')}
                      >
                        <Home className="size-3"/>
                      </Button>
                      {pathSegments.map((segment, index) => (
                          <div key={index} className="flex items-center">
                            <ChevronRight className="size-3"/>
                            <Button
                                variant="ghost"
                                size="sm"
                                className="h-6 px-2"
                                onClick={() => navigateToPath(index)}
                                disabled={index === pathSegments.length - 1}
                            >
                              {segment}
                            </Button>
                          </div>
                      ))}
                    </div>
                    <div className="flex items-center gap-2">
                      <Button
                          variant="outline"
                          size="icon"
                          onClick={() => queryClient.invalidateQueries({queryKey: ['storage', projectId]})}
                          title="Refresh"
                      >
                        <RefreshCw className="size-4"/>
                      </Button>
                      <Button
                          variant="outline"
                          size="icon"
                          onClick={() => setCreateFolderOpen(true)}
                          title="New folder"
                      >
                        <FolderPlus className="size-4"/>
                      </Button>
                      <Button
                          variant="outline"
                          size="icon"
                          onClick={handleCreateFile}
                          title="New file"
                      >
                        <FilePlus className="size-4"/>
                      </Button>
                      <Button
                          variant="outline"
                          onClick={() => fileInputRef.current?.click()}
                          disabled={uploadMutation.isPending}
                      >
                        <Upload className="mr-2 size-4"/>
                        {uploadMutation.isPending ? 'Uploading...' : 'Upload'}
                      </Button>
                      <input
                          ref={fileInputRef}
                          type="file"
                          className="hidden"
                          onChange={handleFileUpload}
                          multiple
                      />
                    </div>
                  </div>
                </CardHeader>

                <CardContent
                    className={cn(
                        'flex-1 overflow-hidden relative',
                        isDragging && 'bg-primary/5'
                    )}
                    onDragEnter={handleDragEnter}
                    onDragLeave={handleDragLeave}
                    onDragOver={handleDragOver}
                    onDrop={handleDrop}
                    onClick={handleClearSelection}
                >
                  {/* Drag overlay */}
                  {isDragging && (
                      <div
                          className="absolute inset-0 flex items-center justify-center bg-primary/10 border-1 border-dashed border-primary z-10 pointer-events-none">
                        <div className="text-center">
                          <p className="text-lg font-medium text-primary">Drop files to upload</p>
                        </div>
                      </div>
                  )}

                  {isLoading ? (
                      <div className="flex items-center justify-center h-48">
                        <p className="text-muted-foreground">Loading...</p>
                      </div>
                  ) : sortedItems.length === 0 && !currentPath ? (
                      <div className="flex flex-col items-center justify-center h-48 text-center h-full">
                        <Folder className="size-12 text-muted-foreground/50 mb-4"/>
                        <p className="text-muted-foreground mb-2">This folder is empty</p>
                        <p className="text-sm text-muted-foreground">
                          Upload files or create a new folder to get started
                        </p>
                      </div>
                  ) : (
                      <div className="divide-y">
                        {/* Go up row */}
                        {currentPath && (
                            <div
                                className="flex items-center gap-3 p-3 hover:bg-muted/50 cursor-pointer select-none"
                                onClick={() => {
                                  const parentPath = currentPath.split('/').slice(0, -1).join('/');
                                  setCurrentPath(parentPath);
                                  setSelectedPaths(new Set());
                                  setLastSelectedIndex(null);
                                }}
                            >
                              <ArrowUp className="size-6 text-muted-foreground shrink-0"/>
                              <span className="text-muted-foreground">..</span>
                            </div>
                        )}
                        {sortedItems.map((item, index) => {
                          const isSelected = selectedPaths.has(item.path);
                          return (
                              <ContextMenu key={item.path}>
                                <ContextMenuTrigger>
                                  <div
                                      className={cn(
                                          'flex items-center justify-between gap-3 p-3 hover:bg-muted/50 cursor-pointer select-none',
                                          index % 2 === 1 && 'bg-muted/30',
                                          isSelected && 'bg-primary/20 hover:bg-primary/25'
                                      )}
                                      onClick={(e) => handleItemSelect(item, index, e)}
                                      onDoubleClick={() => {
                                        if (item.is_dir) {
                                          handleItemClick(item);
                                        } else if (EDITABLE_MIME_TYPES.includes(item.mime_type || '')) {
                                          handleEditFile(item);
                                        }
                                      }}
                                  >
                                    <div className="flex items-center gap-3 flex-1 min-w-0">
                                      {isSelected ? (
                                          <div
                                              className="size-6 shrink-0 rounded bg-primary flex items-center justify-center">
                                            <Check className="size-4 text-primary-foreground"/>
                                          </div>
                                      ) : (
                                          getFileIcon(item)
                                      )}
                                      <span className="truncate">{item.name}</span>
                                    </div>
                                    {!item.is_dir && (
                                        <div className="flex items-center gap-6 text-sm text-muted-foreground">
                                          <span className="w-20 text-right">{formatFileSize(item.size)}</span>
                                          <span className="w-24">
                                {new Date(item.updated_at).toLocaleDateString()}
                              </span>
                                        </div>
                                    )}
                                  </div>
                                </ContextMenuTrigger>
                                <ContextMenuContent>
                                  {item.is_dir ? (
                                      <ContextMenuItem onClick={() => handleItemClick(item)}>
                                        <Folder className="mr-2 size-4"/>
                                        Open
                                      </ContextMenuItem>
                                  ) : (
                                      <>
                                        <ContextMenuItem onClick={() => handleOpenInNewTab(item)}>
                                          <ExternalLink className="mr-2 size-4"/>
                                          Open in new tab
                                        </ContextMenuItem>
                                        <ContextMenuItem onClick={() => handleCopyLink(item)}>
                                          <Link className="mr-2 size-4"/>
                                          Copy link
                                        </ContextMenuItem>
                                        <ContextMenuItem onClick={() => handleDownload(item)}>
                                          <Download className="mr-2 size-4"/>
                                          Download
                                        </ContextMenuItem>
                                        {EDITABLE_MIME_TYPES.includes(item.mime_type || '') && (
                                            <ContextMenuItem onClick={() => handleEditFile(item)}>
                                              <FileCode className="mr-2 size-4"/>
                                              Edit
                                            </ContextMenuItem>
                                        )}
                                      </>
                                  )}
                                  <ContextMenuSeparator/>
                                  <ContextMenuItem
                                      onClick={() => {
                                        setSelectedItem(item);
                                        setNewName(item.name);
                                        setRenameOpen(true);
                                      }}
                                  >
                                    <Pencil className="mr-2 size-4"/>
                                    Rename
                                  </ContextMenuItem>
                                  <ContextMenuItem
                                      onClick={() => {
                                        setSelectedPaths(new Set([item.path]));
                                        setMoveTargetPath('');
                                        setMoveDialogOpen(true);
                                      }}
                                  >
                                    <Move className="mr-2 size-4"/>
                                    Move
                                  </ContextMenuItem>
                                  <ContextMenuSeparator/>
                                  <ContextMenuItem
                                      variant="destructive"
                                      onClick={() => {
                                        setSelectedItem(item);
                                        setDeleteOpen(true);
                                      }}
                                  >
                                    <Trash2 className="mr-2 size-4"/>
                                    Delete
                                  </ContextMenuItem>
                                </ContextMenuContent>
                              </ContextMenu>
                          );
                        })}
                      </div>
                  )}
                </CardContent>

                {/* Selection toolbar - bottom */}
                {selectedPaths.size > 0 && (
                    <div className="flex items-center justify-between px-4 py-2 bg-primary/10 border-t">
                <span className="text-sm font-medium">
                  {selectedPaths.size} item{selectedPaths.size > 1 ? 's' : ''} selected
                </span>
                      <div className="flex items-center gap-2">
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => {
                              setMoveTargetPath('');
                              setMoveDialogOpen(true);
                            }}
                        >
                          <Move className="mr-2 size-4"/>
                          Move
                        </Button>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setDeleteOpen(true)}
                        >
                          <Trash2 className="mr-2 size-4"/>
                          Delete
                        </Button>
                        <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => {
                              setSelectedPaths(new Set());
                              setLastSelectedIndex(null);
                            }}
                        >
                          <X className="mr-2 size-4"/>
                          Clear
                        </Button>
                      </div>
                    </div>
                )}
              </>
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
                  <Button
                      size="sm"
                      onClick={() => handleSaveTab(activeTab)}
                      disabled={
                          savingTabId === activeTab.id ||
                          (activeTab.content === activeTab.originalContent && !activeTab.isNew) ||
                          (activeTab.isNew && !activeTab.name)
                      }
                  >
                    <Save className="mr-2 size-4"/>
                    {savingTabId === activeTab.id ? 'Saving...' : 'Save'}
                  </Button>
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

        {/* Create Folder Dialog */}
        <Dialog open={createFolderOpen} onOpenChange={setCreateFolderOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create Folder</DialogTitle>
            </DialogHeader>
            <Field>
              <FieldLabel>Folder Name</FieldLabel>
              <Input
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  placeholder="my-folder"
              />
            </Field>
            <DialogFooter>
              <Button variant="outline" onClick={() => setCreateFolderOpen(false)}>
                Cancel
              </Button>
              <Button
                  onClick={() => createFolderMutation.mutate(newName)}
                  disabled={!newName || createFolderMutation.isPending}
              >
                Create
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Rename Dialog */}
        <Dialog open={renameOpen} onOpenChange={setRenameOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Rename</DialogTitle>
            </DialogHeader>
            <Field>
              <FieldLabel>New Name</FieldLabel>
              <Input
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
              />
            </Field>
            <DialogFooter>
              <Button variant="outline" onClick={() => setRenameOpen(false)}>
                Cancel
              </Button>
              <Button
                  onClick={() => renameMutation.mutate()}
                  disabled={!newName || renameMutation.isPending}
              >
                Rename
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Delete Confirm */}
        <ConfirmDialog
            open={deleteOpen}
            onOpenChange={(open) => {
              setDeleteOpen(open);
              if (!open) setSelectedItem(null);
            }}
            title={
              selectedPaths.size > 1
                  ? `Delete ${selectedPaths.size} items`
                  : `Delete ${selectedItem?.is_dir ? 'Folder' : 'File'}`
            }
            description={
              selectedPaths.size > 1
                  ? `Are you sure you want to delete ${selectedPaths.size} selected items?`
                  : `Are you sure you want to delete "${selectedItem?.name}"?${
                      selectedItem?.is_dir ? ' All contents will be deleted.' : ''
                  }`
            }
            confirmLabel="Delete"
            variant="destructive"
            onConfirm={() => {
              if (selectedPaths.size > 1 || (selectedPaths.size === 1 && !selectedItem)) {
                deleteSelectedMutation.mutate();
              } else if (selectedItem) {
                deleteMutation.mutate(selectedItem.path);
              }
            }}
            isLoading={deleteMutation.isPending || deleteSelectedMutation.isPending}
        />

        {/* Move Dialog */}
        <Dialog open={moveDialogOpen} onOpenChange={setMoveDialogOpen}>
          <DialogContent className="max-w-md">
            <DialogHeader>
              <DialogTitle>Move {selectedPaths.size} item{selectedPaths.size > 1 ? 's' : ''}</DialogTitle>
            </DialogHeader>
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground mb-4">Select destination folder:</p>
              <div className="border rounded-lg max-h-64 overflow-auto">
                {/* Root folder option */}
                <button
                    type="button"
                    onClick={() => setMoveTargetPath('')}
                    className={cn(
                        'w-full flex items-center gap-2 px-3 py-2 text-left hover:bg-muted/50 transition-colors',
                        moveTargetPath === '' && 'bg-primary/20'
                    )}
                >
                  <Home className="size-4 text-muted-foreground"/>
                  <span className="font-medium">Root</span>
                  {moveTargetPath === '' && <Check className="size-4 ml-auto text-primary"/>}
                </button>
                {/* Folder list */}
                {allFolders
                    .filter((folder) => !selectedPaths.has(folder.path))
                    .sort((a, b) => a.path.localeCompare(b.path))
                    .map((folder) => (
                        <button
                            key={folder.path}
                            type="button"
                            onClick={() => setMoveTargetPath(folder.path)}
                            className={cn(
                                'w-full flex items-center gap-2 px-3 py-2 text-left hover:bg-muted/50 transition-colors',
                                moveTargetPath === folder.path && 'bg-primary/20'
                            )}
                        >
                          <FolderOpen className="size-4 text-blue-500"/>
                          <span className="truncate">{folder.path}</span>
                          {moveTargetPath === folder.path && (
                              <Check className="size-4 ml-auto text-primary shrink-0"/>
                          )}
                        </button>
                    ))}
                {allFolders.length === 0 && (
                    <p className="text-sm text-muted-foreground p-3">No folders available</p>
                )}
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setMoveDialogOpen(false)}>
                Cancel
              </Button>
              <Button
                  onClick={() => moveMutation.mutate(moveTargetPath)}
                  disabled={moveMutation.isPending}
              >
                <Move className="mr-2 size-4"/>
                {moveMutation.isPending ? 'Moving...' : 'Move'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

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
