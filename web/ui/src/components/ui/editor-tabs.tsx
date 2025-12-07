import * as React from 'react';
import { X } from 'lucide-react';
import { cn } from '@/lib/utils';

interface EditorTabsProps {
  children: React.ReactNode;
  className?: string;
}

export function EditorTabs({ children, className }: EditorTabsProps) {
  return (
    <div className={cn('flex items-end', className)}>
      {children}
    </div>
  );
}

interface EditorTabProps {
  active?: boolean;
  onClick?: () => void;
  icon?: React.ReactNode;
  children: React.ReactNode;
  dirty?: boolean;
  badge?: React.ReactNode;
  onClose?: () => void;
  className?: string;
}

export function EditorTab({
  active,
  onClick,
  icon,
  children,
  dirty,
  badge,
  onClose,
  className,
}: EditorTabProps) {
  const hasClose = !!onClose;

  return (
    <div
      className={cn(
        'group flex items-center gap-2 px-4 py-2 h-[40px] text-sm border-t border-l border-r rounded-t-xl',
        active
          ? 'border-border bg-card'
          : 'border-transparent text-muted-foreground hover:text-foreground',
        className
      )}
      style={{
        marginBottom: active ? -1 : 0,
      }}
    >
      <button
        onClick={onClick}
        className={cn('flex items-center gap-2 relative', active ? '' : 'top-[1px]', hasClose && 'cursor-pointer')}
      >
        {icon}
        <span className={cn(hasClose && 'max-w-32 truncate')}>{children}</span>
        {dirty && <span className="text-orange-500">*</span>}
        {badge}
      </button>
      {onClose && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onClose();
          }}
          className="ml-1 p-0.5 rounded-full hover:bg-muted"
        >
          <X className="size-3" />
        </button>
      )}
    </div>
  );
}

interface EditorTabPanelProps {
  children: React.ReactNode;
  className?: string;
}

export function EditorTabPanel({ children, className }: EditorTabPanelProps) {
  return (
    <div className={cn('w-full overflow-y-auto', className)}>
      {children}
    </div>
  );
}
