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
      <div className="relative flex items-center">
        <div className="relative flex-1">
          {isImage ? (
            <Image className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
          ) : (
            <File className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
          )}
          <Input
            type="text"
            value={filename || ''}
            placeholder={isImage ? 'Select image...' : 'Select file...'}
            readOnly
            onClick={handleOpenDialog}
            className="pl-9 pr-20 cursor-pointer truncate"
          />
          <div className="absolute right-1 top-1/2 -translate-y-1/2 flex items-center gap-0.5">
            {value && (
              <>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7"
                  onClick={(e) => { e.stopPropagation(); window.open(value, '_blank'); }}
                  title="Open file"
                >
                  <ExternalLink className="h-3.5 w-3.5" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7"
                  onClick={(e) => { e.stopPropagation(); handleClear(); }}
                  title="Clear"
                >
                  <X className="h-3.5 w-3.5" />
                </Button>
              </>
            )}
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              onClick={handleOpenDialog}
              title="Browse files"
            >
              <FolderOpen className="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
      </div>

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
