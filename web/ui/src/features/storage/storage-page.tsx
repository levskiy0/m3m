import { useState, useRef, useCallback } from 'react';
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
import { Textarea } from '@/components/ui/textarea';
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field';
import { CodeEditor } from '@/components/shared/code-editor';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';

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

export function StoragePage() {
  const { projectId } = useParams<{ projectId: string }>();
  const queryClient = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [currentPath, setCurrentPath] = useState('');
  const [selectedItem, setSelectedItem] = useState<StorageItem | null>(null);

  // Dialogs
  const [createFolderOpen, setCreateFolderOpen] = useState(false);
  const [createFileOpen, setCreateFileOpen] = useState(false);
  const [renameOpen, setRenameOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);

  // Form state
  const [newName, setNewName] = useState('');
  const [newFileName, setNewFileName] = useState('');
  const [newFileContent, setNewFileContent] = useState('');
  const [editContent, setEditContent] = useState('');

  const { data: items = [], isLoading } = useQuery({
    queryKey: ['storage', projectId, currentPath],
    queryFn: () => storageApi.list(projectId!, currentPath),
    enabled: !!projectId,
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

  const createFileMutation = useMutation({
    mutationFn: () =>
      storageApi.createFile(projectId!, {
        path: currentPath,
        name: newFileName,
        content: newFileContent,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['storage', projectId] });
      setCreateFileOpen(false);
      setNewFileName('');
      setNewFileContent('');
      toast.success('File created');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to create file');
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
    mutationFn: () => storageApi.delete(projectId!, selectedItem!.path),
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

  const updateFileMutation = useMutation({
    mutationFn: () =>
      storageApi.updateFile(projectId!, selectedItem!.path, editContent),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['storage', projectId] });
      setEditOpen(false);
      setSelectedItem(null);
      toast.success('File saved');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Save failed');
    },
  });

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
    }
  };

  const handleEditFile = async (item: StorageItem) => {
    try {
      const blob = await storageApi.download(projectId!, item.path);
      const text = await blob.text();
      setSelectedItem(item);
      setEditContent(text);
      setEditOpen(true);
    } catch (err) {
      toast.error('Failed to load file');
    }
  };

  const handleDownload = async (item: StorageItem) => {
    try {
      const blob = await storageApi.download(projectId!, item.path);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = item.name;
      a.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      toast.error('Download failed');
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

  const sortedItems = [...items].sort((a, b) => {
    if (a.is_dir && !b.is_dir) return -1;
    if (!a.is_dir && b.is_dir) return 1;
    return a.name.localeCompare(b.name);
  });

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
      return <Folder className="size-5 text-blue-500 shrink-0" />;
    }

    const ext = getFileExtension(item.name);

    // Images - show thumbnail
    if (isImageFile(item.name) && item.url) {
      return (
        <div className="size-8 shrink-0 rounded overflow-hidden bg-muted flex items-center justify-center">
          <img
            src={item.url}
            alt={item.name}
            className="h-full w-full object-cover"
            onError={(e) => {
              e.currentTarget.style.display = 'none';
              e.currentTarget.parentElement?.classList.add('fallback');
            }}
          />
          <FileImage className="size-4 text-muted-foreground hidden" />
        </div>
      );
    }

    // Text/documents
    if (['txt', 'md', 'doc', 'docx', 'pdf', 'rtf'].includes(ext)) {
      return <FileText className="size-5 text-orange-500 shrink-0" />;
    }

    // Code
    if (['js', 'ts', 'jsx', 'tsx', 'json', 'html', 'css', 'py', 'go', 'rs', 'java', 'c', 'cpp', 'h', 'yml', 'yaml', 'xml', 'sh', 'bash'].includes(ext)) {
      return <FileCode className="size-5 text-green-500 shrink-0" />;
    }

    // Archives
    if (['zip', 'rar', '7z', 'tar', 'gz', 'bz2'].includes(ext)) {
      return <FileArchive className="size-5 text-yellow-500 shrink-0" />;
    }

    // Audio
    if (['mp3', 'wav', 'ogg', 'flac', 'm4a', 'aac'].includes(ext)) {
      return <FileAudio className="size-5 text-purple-500 shrink-0" />;
    }

    // Video
    if (['mp4', 'webm', 'mkv', 'avi', 'mov', 'wmv'].includes(ext)) {
      return <FileVideo className="size-5 text-pink-500 shrink-0" />;
    }

    return <File className="size-5 text-muted-foreground shrink-0" />;
  };

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-1 text-sm text-muted-foreground">
              <Button
                variant="ghost"
                size="sm"
                className="h-6 px-2"
                onClick={() => setCurrentPath('')}
              >
                <Home className="size-3" />
              </Button>
              {pathSegments.map((segment, index) => (
                <div key={index} className="flex items-center">
                  <ChevronRight className="size-3" />
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
                onClick={() => queryClient.invalidateQueries({ queryKey: ['storage', projectId] })}
                title="Refresh"
              >
                <RefreshCw className="size-4" />
              </Button>
              <Button
                variant="outline"
                size="icon"
                onClick={() => setCreateFolderOpen(true)}
                title="New folder"
              >
                <FolderPlus className="size-4" />
              </Button>
              <Button
                variant="outline"
                size="icon"
                onClick={() => setCreateFileOpen(true)}
                title="New file"
              >
                <FilePlus className="size-4" />
              </Button>
              <Button
                variant="outline"
                onClick={() => fileInputRef.current?.click()}
                disabled={uploadMutation.isPending}
              >
                <Upload className="mr-2 size-4" />
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
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center h-48">
              <p className="text-muted-foreground">Loading...</p>
            </div>
          ) : sortedItems.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-48 text-center">
              <Folder className="size-12 text-muted-foreground/50 mb-4" />
              <p className="text-muted-foreground mb-2">This folder is empty</p>
              <p className="text-sm text-muted-foreground">
                Upload files or create a new folder to get started
              </p>
            </div>
          ) : (
            <div className="border rounded-lg divide-y">
              {sortedItems.map((item) => (
                <ContextMenu key={item.path}>
                  <ContextMenuTrigger>
                    <div
                      className={`flex items-center justify-between gap-3 p-3 hover:bg-muted/50 ${
                        item.is_dir ? 'cursor-pointer' : ''
                      }`}
                      onDoubleClick={() => handleItemClick(item)}
                    >
                      <div className="flex items-center gap-3 flex-1 min-w-0">
                        {getFileIcon(item)}
                        <span className="truncate">{item.name}</span>
                      </div>
                      <div className="flex items-center gap-6 text-sm text-muted-foreground">
                        <span className="w-20 text-right">{formatFileSize(item.size)}</span>
                        <span className="w-24">
                          {new Date(item.updated_at).toLocaleDateString()}
                        </span>
                      </div>
                    </div>
                  </ContextMenuTrigger>
                  <ContextMenuContent>
                    {item.is_dir ? (
                      <ContextMenuItem onClick={() => handleItemClick(item)}>
                        <Folder className="mr-2 size-4" />
                        Open
                      </ContextMenuItem>
                    ) : (
                      <>
                        <ContextMenuItem onClick={() => handleOpenInNewTab(item)}>
                          <ExternalLink className="mr-2 size-4" />
                          Open in new tab
                        </ContextMenuItem>
                        <ContextMenuItem onClick={() => handleCopyLink(item)}>
                          <Link className="mr-2 size-4" />
                          Copy link
                        </ContextMenuItem>
                        <ContextMenuItem onClick={() => handleDownload(item)}>
                          <Download className="mr-2 size-4" />
                          Download
                        </ContextMenuItem>
                        {EDITABLE_MIME_TYPES.includes(item.mime_type || '') && (
                          <ContextMenuItem onClick={() => handleEditFile(item)}>
                            <FileCode className="mr-2 size-4" />
                            Edit
                          </ContextMenuItem>
                        )}
                      </>
                    )}
                    <ContextMenuSeparator />
                    <ContextMenuItem
                      onClick={() => {
                        setSelectedItem(item);
                        setNewName(item.name);
                        setRenameOpen(true);
                      }}
                    >
                      <Pencil className="mr-2 size-4" />
                      Rename
                    </ContextMenuItem>
                    <ContextMenuSeparator />
                    <ContextMenuItem
                      variant="destructive"
                      onClick={() => {
                        setSelectedItem(item);
                        setDeleteOpen(true);
                      }}
                    >
                      <Trash2 className="mr-2 size-4" />
                      Delete
                    </ContextMenuItem>
                  </ContextMenuContent>
                </ContextMenu>
              ))}
            </div>
          )}
        </CardContent>
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

      {/* Create File Dialog */}
      <Dialog open={createFileOpen} onOpenChange={setCreateFileOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Create File</DialogTitle>
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel>File Name</FieldLabel>
              <Input
                value={newFileName}
                onChange={(e) => setNewFileName(e.target.value)}
                placeholder="config.json"
              />
            </Field>
            <Field>
              <FieldLabel>Content</FieldLabel>
              <Textarea
                value={newFileContent}
                onChange={(e) => setNewFileContent(e.target.value)}
                placeholder="File content..."
                rows={10}
                className="font-mono text-sm"
              />
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateFileOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={() => createFileMutation.mutate()}
              disabled={!newFileName || createFileMutation.isPending}
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

      {/* Edit File Dialog */}
      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent className="max-w-4xl h-[80vh]">
          <DialogHeader>
            <DialogTitle>Edit: {selectedItem?.name}</DialogTitle>
          </DialogHeader>
          <div className="flex-1 min-h-0 border rounded-md overflow-hidden">
            <CodeEditor
              value={editContent}
              onChange={setEditContent}
              language={
                selectedItem?.name?.endsWith('.json')
                  ? 'json'
                  : selectedItem?.name?.endsWith('.yaml') ||
                    selectedItem?.name?.endsWith('.yml')
                  ? 'yaml'
                  : 'plaintext'
              }
              height="100%"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={() => updateFileMutation.mutate()}
              disabled={updateFileMutation.isPending}
            >
              Save
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirm */}
      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title={`Delete ${selectedItem?.is_dir ? 'Folder' : 'File'}`}
        description={`Are you sure you want to delete "${selectedItem?.name}"?${
          selectedItem?.is_dir ? ' All contents will be deleted.' : ''
        }`}
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={() => deleteMutation.mutate()}
        isLoading={deleteMutation.isPending}
      />
    </>
  );
}
