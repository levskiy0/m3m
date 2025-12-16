import { useState } from 'react';
import { File, Plus, MoreHorizontal, Pencil, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field';
import { LoadingButton } from '@/components/ui/loading-button';
import { cn } from '@/lib/utils';
import type { CodeFile } from '@/types';

interface FileSidebarProps {
  files: CodeFile[];
  activeFileName: string;
  dirtyFiles: Set<string>;
  onFileSelect: (fileName: string) => void;
  onFileCreate: (name: string) => Promise<void>;
  onFileRename: (oldName: string, newName: string) => Promise<void>;
  onFileDelete: (name: string) => Promise<void>;
  disabled?: boolean;
}

export function FileSidebar({
  files,
  activeFileName,
  dirtyFiles,
  onFileSelect,
  onFileCreate,
  onFileRename,
  onFileDelete,
  disabled,
}: FileSidebarProps) {
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [renameDialogOpen, setRenameDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedFile, setSelectedFile] = useState<string | null>(null);
  const [newFileName, setNewFileName] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleCreate = async () => {
    if (!newFileName.trim()) return;
    setIsLoading(true);
    try {
      await onFileCreate(newFileName.trim());
      setCreateDialogOpen(false);
      setNewFileName('');
    } finally {
      setIsLoading(false);
    }
  };

  const handleRename = async () => {
    if (!selectedFile || !newFileName.trim()) return;
    setIsLoading(true);
    try {
      await onFileRename(selectedFile, newFileName.trim());
      setRenameDialogOpen(false);
      setSelectedFile(null);
      setNewFileName('');
    } finally {
      setIsLoading(false);
    }
  };

  const handleDelete = async () => {
    if (!selectedFile) return;
    setIsLoading(true);
    try {
      await onFileDelete(selectedFile);
      setDeleteDialogOpen(false);
      setSelectedFile(null);
    } finally {
      setIsLoading(false);
    }
  };

  const openRenameDialog = (fileName: string) => {
    setSelectedFile(fileName);
    setNewFileName(fileName);
    setRenameDialogOpen(true);
  };

  const openDeleteDialog = (fileName: string) => {
    setSelectedFile(fileName);
    setDeleteDialogOpen(true);
  };

  return (
    <div className="flex flex-col h-full border-r bg-muted/30">
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2 border-b">
        <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
          Files
        </span>
        <Button
          variant="ghost"
          size="icon"
          className="size-6"
          onClick={() => {
            setNewFileName('');
            setCreateDialogOpen(true);
          }}
          disabled={disabled}
          title="New file"
        >
          <Plus className="size-3.5" />
        </Button>
      </div>

      {/* File List */}
      <div className="flex-1 overflow-y-auto py-1">
        {files.map((file) => {
          const isMain = file.name === 'main';
          const isActive = file.name === activeFileName;
          const isDirty = dirtyFiles.has(file.name);

          return (
            <div
              key={file.name}
              className={cn(
                'group flex items-center gap-1 px-2 py-1 mx-1 rounded-sm cursor-pointer hover:bg-accent',
                isActive && 'bg-accent'
              )}
              onClick={() => onFileSelect(file.name)}
            >
              <File className="size-3.5 flex-shrink-0 text-muted-foreground" />
              <span className={cn('flex-1 text-sm truncate', isActive && 'font-medium')}>
                {file.name}
                {isDirty && <span className="text-amber-500 ml-0.5">*</span>}
              </span>
              {!isMain && (
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="size-5 opacity-0 group-hover:opacity-100"
                      onClick={(e) => e.stopPropagation()}
                      disabled={disabled}
                    >
                      <MoreHorizontal className="size-3" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={() => openRenameDialog(file.name)}>
                      <Pencil className="mr-2 size-4" />
                      Rename
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      className="text-destructive"
                      onClick={() => openDeleteDialog(file.name)}
                    >
                      <Trash2 className="mr-2 size-4" />
                      Delete
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              )}
            </div>
          );
        })}
      </div>

      {/* Create File Dialog */}
      <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create File</DialogTitle>
            <DialogDescription>Create a new file in this branch</DialogDescription>
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel>File Name</FieldLabel>
              <Input
                value={newFileName}
                onChange={(e) => setNewFileName(e.target.value)}
                placeholder="utils"
                onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
              />
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateDialogOpen(false)}>
              Cancel
            </Button>
            <LoadingButton
              onClick={handleCreate}
              disabled={!newFileName.trim()}
              loading={isLoading}
            >
              Create
            </LoadingButton>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Rename File Dialog */}
      <Dialog open={renameDialogOpen} onOpenChange={setRenameDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Rename File</DialogTitle>
            <DialogDescription>Enter a new name for "{selectedFile}"</DialogDescription>
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel>New Name</FieldLabel>
              <Input
                value={newFileName}
                onChange={(e) => setNewFileName(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleRename()}
              />
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRenameDialogOpen(false)}>
              Cancel
            </Button>
            <LoadingButton
              onClick={handleRename}
              disabled={!newFileName.trim() || newFileName === selectedFile}
              loading={isLoading}
            >
              Rename
            </LoadingButton>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete File Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete File</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete "{selectedFile}"? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
              Cancel
            </Button>
            <LoadingButton variant="destructive" onClick={handleDelete} loading={isLoading}>
              Delete
            </LoadingButton>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
