import { useState, useRef, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Upload,
  FolderPlus,
  FilePlus,
  Download,
  Trash2,
  Edit,
  Link,
  Folder,
  File,
  FileText,
  FileCode,
  FileImage,
  MoreHorizontal,
  ChevronRight,
  Grid,
  List,
  Home,
} from 'lucide-react';
import { toast } from 'sonner';

import { storageApi } from '@/api';
import { EDITABLE_MIME_TYPES, IMAGE_MIME_TYPES } from '@/lib/constants';
import type { StorageItem } from '@/types';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field';
import { CodeEditor } from '@/components/shared/code-editor';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { EmptyState } from '@/components/shared/empty-state';
import { Skeleton } from '@/components/ui/skeleton';
import { cn } from '@/lib/utils';

function getFileIcon(item: StorageItem) {
  if (item.isDir) return Folder;
  if (IMAGE_MIME_TYPES.includes(item.mimeType || '')) return FileImage;
  if (EDITABLE_MIME_TYPES.includes(item.mimeType || '')) return FileCode;
  if (item.mimeType?.startsWith('text/')) return FileText;
  return File;
}

function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

export function StoragePage() {
  const { projectId } = useParams<{ projectId: string }>();
  const queryClient = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [currentPath, setCurrentPath] = useState('');
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [selectedItem, setSelectedItem] = useState<StorageItem | null>(null);

  // Dialogs
  const [createFolderOpen, setCreateFolderOpen] = useState(false);
  const [createFileOpen, setCreateFileOpen] = useState(false);
  const [renameOpen, setRenameOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [previewOpen, setPreviewOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [linkOpen, setLinkOpen] = useState(false);

  // Form state
  const [newName, setNewName] = useState('');
  const [newFileName, setNewFileName] = useState('');
  const [newFileContent, setNewFileContent] = useState('');
  const [editContent, setEditContent] = useState('');
  const [publicLink, setPublicLink] = useState('');

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

  const generateLinkMutation = useMutation({
    mutationFn: () =>
      storageApi.generateLink(projectId!, { path: selectedItem!.path }),
    onSuccess: (data) => {
      setPublicLink(data.url);
      setLinkOpen(true);
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to generate link');
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
    if (item.isDir) {
      setCurrentPath(item.path);
    } else if (IMAGE_MIME_TYPES.includes(item.mimeType || '')) {
      setSelectedItem(item);
      setPreviewOpen(true);
    } else if (EDITABLE_MIME_TYPES.includes(item.mimeType || '')) {
      handleEditFile(item);
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
    if (a.isDir && !b.isDir) return -1;
    if (!a.isDir && b.isDir) return 1;
    return a.name.localeCompare(b.name);
  });

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Storage</h1>
          <p className="text-muted-foreground">Manage project files and assets</p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" onClick={() => setCreateFolderOpen(true)}>
            <FolderPlus className="mr-2 size-4" />
            New Folder
          </Button>
          <Button variant="outline" onClick={() => setCreateFileOpen(true)}>
            <FilePlus className="mr-2 size-4" />
            New File
          </Button>
          <Button onClick={() => fileInputRef.current?.click()}>
            <Upload className="mr-2 size-4" />
            Upload
          </Button>
          <input
            ref={fileInputRef}
            type="file"
            className="hidden"
            onChange={handleFileUpload}
          />
        </div>
      </div>

      {/* Breadcrumb & View Toggle */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-1 text-sm">
          <Button
            variant="ghost"
            size="sm"
            className="h-7"
            onClick={() => setCurrentPath('')}
          >
            <Home className="size-4" />
          </Button>
          {pathSegments.map((segment, index) => (
            <div key={index} className="flex items-center">
              <ChevronRight className="size-4 text-muted-foreground" />
              <Button
                variant="ghost"
                size="sm"
                className="h-7"
                onClick={() => navigateToPath(index)}
              >
                {segment}
              </Button>
            </div>
          ))}
        </div>
        <div className="flex items-center gap-1 border rounded-md p-1">
          <Button
            variant={viewMode === 'grid' ? 'secondary' : 'ghost'}
            size="icon"
            className="size-7"
            onClick={() => setViewMode('grid')}
          >
            <Grid className="size-4" />
          </Button>
          <Button
            variant={viewMode === 'list' ? 'secondary' : 'ghost'}
            size="icon"
            className="size-7"
            onClick={() => setViewMode('list')}
          >
            <List className="size-4" />
          </Button>
        </div>
      </div>

      {/* Content */}
      {isLoading ? (
        <div className={cn('grid gap-4', viewMode === 'grid' && 'md:grid-cols-4 lg:grid-cols-6')}>
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <Skeleton key={i} className="h-24" />
          ))}
        </div>
      ) : sortedItems.length === 0 ? (
        <EmptyState
          icon={<Folder className="size-12" />}
          title="This folder is empty"
          description="Upload files or create a new folder to get started"
        />
      ) : viewMode === 'grid' ? (
        <div className="grid gap-4 md:grid-cols-4 lg:grid-cols-6">
          {sortedItems.map((item) => {
            const Icon = getFileIcon(item);
            return (
              <Card
                key={item.path}
                className="cursor-pointer transition-colors hover:bg-muted/50"
                onClick={() => handleItemClick(item)}
              >
                <CardContent className="p-4">
                  <div className="flex flex-col items-center gap-2">
                    <Icon
                      className={cn(
                        'size-10',
                        item.isDir ? 'text-amber-500' : 'text-muted-foreground'
                      )}
                    />
                    <span className="text-sm font-medium truncate w-full text-center">
                      {item.name}
                    </span>
                    {!item.isDir && (
                      <span className="text-xs text-muted-foreground">
                        {formatFileSize(item.size)}
                      </span>
                    )}
                  </div>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="absolute top-2 right-2 size-7 opacity-0 group-hover:opacity-100"
                      >
                        <MoreHorizontal className="size-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent>
                      {!item.isDir && (
                        <>
                          <DropdownMenuItem
                            onClick={(e) => {
                              e.stopPropagation();
                              handleDownload(item);
                            }}
                          >
                            <Download className="mr-2 size-4" />
                            Download
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={(e) => {
                              e.stopPropagation();
                              setSelectedItem(item);
                              generateLinkMutation.mutate();
                            }}
                          >
                            <Link className="mr-2 size-4" />
                            Get Link
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                        </>
                      )}
                      <DropdownMenuItem
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedItem(item);
                          setNewName(item.name);
                          setRenameOpen(true);
                        }}
                      >
                        <Edit className="mr-2 size-4" />
                        Rename
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        className="text-destructive"
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedItem(item);
                          setDeleteOpen(true);
                        }}
                      >
                        <Trash2 className="mr-2 size-4" />
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </CardContent>
              </Card>
            );
          })}
        </div>
      ) : (
        <Card>
          <CardContent className="p-0">
            <div className="divide-y">
              {sortedItems.map((item) => {
                const Icon = getFileIcon(item);
                return (
                  <div
                    key={item.path}
                    className="flex items-center gap-4 p-3 cursor-pointer hover:bg-muted/50"
                    onClick={() => handleItemClick(item)}
                  >
                    <Icon
                      className={cn(
                        'size-5',
                        item.isDir ? 'text-amber-500' : 'text-muted-foreground'
                      )}
                    />
                    <div className="flex-1 min-w-0">
                      <p className="font-medium truncate">{item.name}</p>
                    </div>
                    {!item.isDir && (
                      <span className="text-sm text-muted-foreground">
                        {formatFileSize(item.size)}
                      </span>
                    )}
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
                        <Button variant="ghost" size="icon" className="size-8">
                          <MoreHorizontal className="size-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        {!item.isDir && (
                          <>
                            <DropdownMenuItem onClick={() => handleDownload(item)}>
                              <Download className="mr-2 size-4" />
                              Download
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              onClick={() => {
                                setSelectedItem(item);
                                generateLinkMutation.mutate();
                              }}
                            >
                              <Link className="mr-2 size-4" />
                              Get Link
                            </DropdownMenuItem>
                            <DropdownMenuSeparator />
                          </>
                        )}
                        <DropdownMenuItem
                          onClick={() => {
                            setSelectedItem(item);
                            setNewName(item.name);
                            setRenameOpen(true);
                          }}
                        >
                          <Edit className="mr-2 size-4" />
                          Rename
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          className="text-destructive"
                          onClick={() => {
                            setSelectedItem(item);
                            setDeleteOpen(true);
                          }}
                        >
                          <Trash2 className="mr-2 size-4" />
                          Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>
      )}

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

      {/* Preview Dialog */}
      <Dialog open={previewOpen} onOpenChange={setPreviewOpen}>
        <DialogContent className="max-w-4xl">
          <DialogHeader>
            <DialogTitle>{selectedItem?.name}</DialogTitle>
          </DialogHeader>
          {selectedItem && IMAGE_MIME_TYPES.includes(selectedItem.mimeType || '') && (
            <div className="flex items-center justify-center">
              <img
                src={storageApi.getDownloadUrl(projectId!, selectedItem.path)}
                alt={selectedItem.name}
                className="max-h-[60vh] object-contain"
              />
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setPreviewOpen(false)}>
              Close
            </Button>
            <Button onClick={() => selectedItem && handleDownload(selectedItem)}>
              <Download className="mr-2 size-4" />
              Download
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Link Dialog */}
      <Dialog open={linkOpen} onOpenChange={setLinkOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Public Link</DialogTitle>
            <DialogDescription>
              Share this link to allow others to download the file
            </DialogDescription>
          </DialogHeader>
          <div className="flex items-center gap-2">
            <Input value={publicLink} readOnly className="font-mono text-sm" />
            <Button
              variant="outline"
              onClick={() => {
                navigator.clipboard.writeText(publicLink);
                toast.success('Link copied');
              }}
            >
              Copy
            </Button>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setLinkOpen(false)}>
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirm */}
      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title={`Delete ${selectedItem?.isDir ? 'Folder' : 'File'}`}
        description={`Are you sure you want to delete "${selectedItem?.name}"?${
          selectedItem?.isDir ? ' All contents will be deleted.' : ''
        }`}
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={() => deleteMutation.mutate()}
        isLoading={deleteMutation.isPending}
      />
    </div>
  );
}
