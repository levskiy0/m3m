import { X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

interface FileTabsProps {
  openFiles: string[];
  activeFileName: string;
  dirtyFiles: Set<string>;
  onFileSelect: (fileName: string) => void;
  onFileClose: (fileName: string) => void;
}

export function FileTabs({
  openFiles,
  activeFileName,
  dirtyFiles,
  onFileSelect,
  onFileClose,
}: FileTabsProps) {
  if (openFiles.length === 0) return null;

  return (
    <div className="flex items-center gap-0.5 h-[41px] px-3 border-b bg-muted/30 overflow-x-auto">
      {openFiles.map((fileName) => {
        const isActive = fileName === activeFileName;
        const isDirty = dirtyFiles.has(fileName);
        const isMain = fileName === 'main';

        return (
          <div
            key={fileName}
            className={cn(
              'group flex items-center gap-1 px-2 py-1 rounded-sm cursor-pointer text-sm transition-colors',
              isActive
                ? 'bg-background text-foreground border shadow-sm'
                : 'text-muted-foreground hover:text-foreground hover:bg-accent'
            )}
            onClick={() => onFileSelect(fileName)}
          >
            <span className={cn('max-w-32 truncate', isActive && 'font-medium')}>
              {fileName}
            </span>
            {isDirty && <span className="text-amber-500">*</span>}
            {!isMain && (
              <Button
                variant="ghost"
                size="icon"
                className={cn(
                  'size-4 p-0 ml-1 rounded-sm',
                  isActive ? 'opacity-60 hover:opacity-100' : 'opacity-0 group-hover:opacity-60 hover:!opacity-100'
                )}
                onClick={(e) => {
                  e.stopPropagation();
                  onFileClose(fileName);
                }}
              >
                <X className="size-3" />
              </Button>
            )}
          </div>
        );
      })}
    </div>
  );
}
