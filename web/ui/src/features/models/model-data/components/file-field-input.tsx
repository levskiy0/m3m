import { useState } from 'react';
import { FolderOpen, X, File, Image, ExternalLink } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { FileBrowser } from '@/features/storage/components';
import type { FieldView, StorageItem } from '@/types';

interface FileFieldInputProps {
  value: string | null;
  onChange: (value: string | null) => void;
  projectId: string;
  view?: FieldView;
}

export function FileFieldInput({
  value,
  onChange,
  projectId,
  view = 'file',
}: FileFieldInputProps) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedItem, setSelectedItem] = useState<StorageItem | null>(null);

  const isImage = view === 'image';
  const accept = isImage ? 'images' : 'all';

  const handleSelect = (item: StorageItem) => {
    setSelectedItem(item);
  };

  const handleConfirm = () => {
    if (selectedItem?.url) {
      onChange(selectedItem.url);
      setDialogOpen(false);
      setSelectedItem(null);
    }
  };

  const handleClear = () => {
    onChange(null);
  };

  const handleOpenDialog = () => {
    setSelectedItem(null);
    setDialogOpen(true);
  };

  // Extract filename from path
  const filename = value?.split('/').pop() || '';

  return (
    <div className="space-y-2">
      {value ? (
        <div className="flex items-center gap-2 border rounded-md p-2">
          {isImage ? (
            <Image className="h-4 w-4 text-muted-foreground shrink-0" />
          ) : (
            <File className="h-4 w-4 text-muted-foreground shrink-0" />
          )}
          <span className="flex-1 truncate text-sm">{filename}</span>
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={() => window.open(value, '_blank')}
            title="Open file"
          >
            <ExternalLink className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={handleClear}
            title="Clear"
          >
            <X className="h-4 w-4" />
          </Button>
        </div>
      ) : (
        <div className="flex gap-2">
          <Input
            type="text"
            placeholder={isImage ? 'Select image...' : 'Select file...'}
            value=""
            readOnly
            onClick={handleOpenDialog}
            className="cursor-pointer"
          />
          <Button
            variant="outline"
            size="icon"
            onClick={handleOpenDialog}
            title="Browse files"
          >
            <FolderOpen className="h-4 w-4" />
          </Button>
        </div>
      )}

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-2xl max-h-[80vh] flex flex-col">
          <DialogHeader>
            <DialogTitle>
              {isImage ? 'Select Image' : 'Select File'}
            </DialogTitle>
          </DialogHeader>
          <div className="flex-1 min-h-0 border rounded-md overflow-hidden">
            <FileBrowser
              projectId={projectId}
              mode="select"
              selectedPath={selectedItem?.path || null}
              onSelect={handleSelect}
              accept={accept as 'all' | 'images'}
              showUpload={true}
              className="h-[400px]"
            />
          </div>
          {selectedItem && (
            <div className="text-sm text-muted-foreground">
              Selected: <span className="font-medium text-foreground">{selectedItem.name}</span>
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleConfirm} disabled={!selectedItem?.url}>
              Select
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
