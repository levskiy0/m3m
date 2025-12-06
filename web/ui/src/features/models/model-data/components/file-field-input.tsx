import { useState, useRef } from 'react';
import { Upload, X, File, Image, Loader2, ExternalLink } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { storageApi } from '@/api/storage';
import type { FieldView } from '@/types';

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
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const isImage = view === 'image';
  const accept = isImage ? 'image/*' : undefined;

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setIsUploading(true);
    setError(null);

    try {
      // Upload to 'uploads' folder in project storage
      const result = await storageApi.upload(projectId, 'uploads', file);
      onChange(result.path);
    } catch (err) {
      setError('Failed to upload file');
      console.error('Upload error:', err);
    } finally {
      setIsUploading(false);
      // Reset input so same file can be selected again
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    }
  };

  const handleClear = () => {
    onChange(null);
    setError(null);
  };

  const handleClick = () => {
    fileInputRef.current?.click();
  };

  // Extract filename from path
  const filename = value?.split('/').pop() || '';

  return (
    <div className="space-y-2">
      <input
        ref={fileInputRef}
        type="file"
        accept={accept}
        onChange={handleFileSelect}
        className="hidden"
      />

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
            onClick={() => {
              // Open file in new tab using the storage URL
              window.open(`/api/projects/${projectId}/storage/download/${value}`, '_blank');
            }}
          >
            <ExternalLink className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={handleClear}
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
            onClick={handleClick}
            className="cursor-pointer"
          />
          <Button
            variant="outline"
            size="icon"
            onClick={handleClick}
            disabled={isUploading}
          >
            {isUploading ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Upload className="h-4 w-4" />
            )}
          </Button>
        </div>
      )}

      {error && (
        <p className="text-sm text-destructive">{error}</p>
      )}
    </div>
  );
}
