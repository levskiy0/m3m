import { useMemo } from 'react';
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
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

const colorClasses: Record<string, string> = {
  blue: 'text-blue-500',
  green: 'text-green-500',
  yellow: 'text-yellow-500',
  red: 'text-red-500',
  purple: 'text-purple-500',
  orange: 'text-orange-500',
};

interface ActionsDropdownProps {
  projectId: string;
  actions: Action[];
  actionStates: Map<string, ActionState>;
}

export function ActionsDropdown({
  projectId,
  actions,
  actionStates,
}: ActionsDropdownProps) {
  const triggerMutation = useMutation({
    mutationFn: (actionSlug: string) => actionsApi.trigger(projectId, actionSlug),
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to trigger action');
    },
  });

  // Filter actions that should be shown in menu and group them
  const groupedActions = useMemo(() => {
    const visible = actions.filter((a) => a.show_in_menu);
    const groups: Record<string, Action[]> = {};
    const ungrouped: Action[] = [];

    for (const action of visible) {
      if (action.group) {
        if (!groups[action.group]) {
          groups[action.group] = [];
        }
        groups[action.group].push(action);
      } else {
        ungrouped.push(action);
      }
    }

    return { groups, ungrouped };
  }, [actions]);

  const visibleCount = groupedActions.ungrouped.length +
    Object.values(groupedActions.groups).reduce((acc, arr) => acc + arr.length, 0);

  if (visibleCount === 0) {
    return null;
  }

  const renderActionItem = (action: Action) => {
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
          <Zap className={cn('mr-2 size-4', action.color ? colorClasses[action.color] : '')} />
        )}
        {action.name}
      </DropdownMenuItem>
    );
  };

  const groupNames = Object.keys(groupedActions.groups);

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
        {/* Render ungrouped actions first */}
        {groupedActions.ungrouped.map(renderActionItem)}

        {/* Render grouped actions with separators */}
        {groupNames.map((groupName, idx) => (
          <div key={groupName}>
            {(idx > 0 || groupedActions.ungrouped.length > 0) && (
              <DropdownMenuSeparator />
            )}
            <DropdownMenuLabel className="text-xs text-muted-foreground">
              {groupName}
            </DropdownMenuLabel>
            {groupedActions.groups[groupName].map(renderActionItem)}
          </div>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
