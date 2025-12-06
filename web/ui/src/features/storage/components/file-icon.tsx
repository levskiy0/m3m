import {
  Folder,
  File,
  FileText,
  FileCode,
  FileImage,
  FileArchive,
  FileAudio,
  FileVideo,
} from 'lucide-react';
import { getFileExtension, isImageFile } from '@/lib/utils';
import type { StorageItem } from '@/types';

interface FileIconProps {
  item: StorageItem;
  size?: 'sm' | 'md';
}

export function FileIcon({ item, size = 'md' }: FileIconProps) {
  const iconSize = size === 'sm' ? 'size-5' : 'size-6';
  const thumbSize = size === 'sm' ? 'size-5' : 'size-6';

  if (item.is_dir) {
    return <Folder className={`${iconSize} text-blue-500 shrink-0`} />;
  }

  const ext = getFileExtension(item.name);

  // Images - show thumbnail
  if (isImageFile(item.name) && item.thumb_url) {
    return (
      <div className={`${thumbSize} shrink-0 rounded overflow-hidden bg-muted flex items-center justify-center`}>
        <img
          src={item.thumb_url}
          alt={item.name}
          className="h-full w-full object-cover"
          onError={(e) => {
            e.currentTarget.style.display = 'none';
            e.currentTarget.parentElement?.classList.add('fallback');
          }}
        />
        <FileImage className={`${iconSize} text-muted-foreground hidden`} />
      </div>
    );
  }

  // Text/documents
  if (['txt', 'md', 'doc', 'docx', 'pdf', 'rtf'].includes(ext)) {
    return <FileText className={`${iconSize} text-orange-500 shrink-0`} />;
  }

  // Code
  if (['js', 'ts', 'jsx', 'tsx', 'json', 'html', 'css', 'py', 'go', 'rs', 'java', 'c', 'cpp', 'h', 'yml', 'yaml', 'xml', 'sh', 'bash'].includes(ext)) {
    return <FileCode className={`${iconSize} text-green-500 shrink-0`} />;
  }

  // Archives
  if (['zip', 'rar', '7z', 'tar', 'gz', 'bz2'].includes(ext)) {
    return <FileArchive className={`${iconSize} text-yellow-500 shrink-0`} />;
  }

  // Audio
  if (['mp3', 'wav', 'ogg', 'flac', 'm4a', 'aac'].includes(ext)) {
    return <FileAudio className={`${iconSize} text-purple-500 shrink-0`} />;
  }

  // Video
  if (['mp4', 'webm', 'mkv', 'avi', 'mov', 'wmv'].includes(ext)) {
    return <FileVideo className={`${iconSize} text-pink-500 shrink-0`} />;
  }

  return <File className={`${iconSize} text-muted-foreground shrink-0`} />;
}
