import { useMutation } from '@tanstack/react-query';
import { Zap, ChevronDown, Loader2 } from 'lucide-react';
import { toast } from 'sonner';

import { actionsApi } from '@/api/actions';
import { cn } from '@/lib/utils';
import type { Action, ActionState } from '@/types';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

interface ActionsDropdownProps {
  projectSlug: string;
  actions: Action[];
  actionStates: Map<string, ActionState>;
}

export function ActionsDropdown({
  projectSlug,
  actions,
  actionStates,
}: ActionsDropdownProps) {
  const triggerMutation = useMutation({
    mutationFn: (actionSlug: string) => actionsApi.trigger(projectSlug, actionSlug),
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to trigger action');
    },
  });

  if (actions.length === 0) {
    return null;
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm">
          <Zap className="mr-2 size-4" />
          Actions
          <ChevronDown className="ml-2 size-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-48">
        {actions.map((action) => {
          const state = actionStates.get(action.slug) || 'enabled';
          const isDisabled = state === 'disabled';
          const isLoading = state === 'loading';

          return (
            <DropdownMenuItem
              key={action.id}
              disabled={isDisabled || isLoading || triggerMutation.isPending}
              onClick={() => triggerMutation.mutate(action.slug)}
              className={cn(
                'cursor-pointer',
                isDisabled && 'opacity-50 cursor-not-allowed'
              )}
            >
              {isLoading ? (
                <Loader2 className="mr-2 size-4 animate-spin" />
              ) : (
                <Zap className="mr-2 size-4" />
              )}
              {action.name}
            </DropdownMenuItem>
          );
        })}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
