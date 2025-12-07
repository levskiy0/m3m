import * as React from 'react';
import { Button, type ButtonProps } from '@/components/ui/button';
import { Spinner } from '@/components/ui/spinner';
import { cn } from '@/lib/utils';

interface LoadingButtonProps extends ButtonProps {
  loading?: boolean;
  children: React.ReactNode;
}

export function LoadingButton({
  loading,
  children,
  disabled,
  className,
  ...props
}: LoadingButtonProps) {
  return (
    <Button
      disabled={disabled || loading}
      className={cn('relative', className)}
      {...props}
    >
      <span className={cn('inline-flex items-center gap-2', loading && 'opacity-0')}>
        {children}
      </span>
      {loading && (
        <Spinner className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2" />
      )}
    </Button>
  );
}
