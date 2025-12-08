/**
 * EnvironmentPage
 * Inline table editor for environment variables (similar to model schema editor)
 */

import { useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Save, Variable, Key, Type, Settings2 } from 'lucide-react';
import { toast } from 'sonner';
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core';
import {
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';

import { environmentApi } from '@/api';
import { queryKeys } from '@/lib/query-keys';
import { Button } from '@/components/ui/button';
import { LoadingButton } from '@/components/ui/loading-button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { EmptyState } from '@/components/shared/empty-state';
import { PageHeader } from '@/components/shared/page-header';

import { SortableEnvRow } from './components';
import { useEnvEditor } from './hooks';

export function EnvironmentPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const queryClient = useQueryClient();

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  // Load env vars
  const { data: envVars = [], isLoading } = useQuery({
    queryKey: queryKeys.environment.all(projectId!),
    queryFn: () => environmentApi.list(projectId!),
    enabled: !!projectId,
  });

  // Editor state
  const {
    envVars: editableEnvVars,
    hasChanges,
    hasValidationErrors,
    envIds,
    addEnvVar,
    updateEnvVar,
    removeEnvVar,
    handleDragEnd,
    resetState,
    setHasChanges,
  } = useEnvEditor({ initialEnvVars: envVars });

  // Reset state when env vars change from server
  useEffect(() => {
    if (envVars.length > 0 || !isLoading) {
      resetState(envVars);
    }
  }, [envVars, isLoading, resetState]);

  // Save mutation - bulk update all env vars
  const saveMutation = useMutation({
    mutationFn: async () => {
      const items = editableEnvVars
        .filter((env) => env.key) // Filter out empty keys
        .map((env, index) => ({
          key: env.key,
          type: env.type,
          value: env.value,
          order: index,
        }));

      return environmentApi.bulkUpdate(projectId!, { items });
    },
    onSuccess: (data) => {
      queryClient.setQueryData(queryKeys.environment.all(projectId!), data);
      setHasChanges(false);
      toast.success('Environment saved');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to save');
    },
  });

  if (isLoading) {
    return (
      <div className="space-y-4 max-w-4xl">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-64" />
      </div>
    );
  }

  return (
    <div className="space-y-4 max-w-4xl">
      {/* Header */}
      <PageHeader
        title="Environment"
        description="Configure environment variables for your service"
        action={
          <LoadingButton
            onClick={() => saveMutation.mutate()}
            disabled={!hasChanges || hasValidationErrors}
            loading={saveMutation.isPending}
          >
            <Save className="size-4" />
            Save
          </LoadingButton>
        }
      />

      {editableEnvVars.length === 0 ? (
        <EmptyState
          icon={<Variable className="size-12" />}
          title="No environment variables"
          description="Add variables to configure your service at runtime"
          action={
            <Button onClick={addEnvVar}>
              <Plus className="mr-2 size-4" />
              Add Variable
            </Button>
          }
        />
      ) : (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Variables</CardTitle>
                <CardDescription>
                  Define environment variables for your service. Drag rows to reorder.
                </CardDescription>
              </div>
              <Button variant="outline" size="sm" onClick={addEnvVar}>
                <Plus className="mr-2 size-4" />
                Add Variable
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <DndContext
              sensors={sensors}
              collisionDetection={closestCenter}
              onDragEnd={handleDragEnd}
            >
              <div className="border rounded-lg overflow-hidden">
                <table className="w-full text-sm">
                  <thead className="bg-muted/50">
                    <tr>
                      <th className="w-10 p-3"></th>
                      <th className="text-left font-medium p-3 w-[200px]">
                        <div className="flex items-center gap-1.5">
                          <Key className="size-4" />
                          <span>Key</span>
                        </div>
                      </th>
                      <th className="text-left font-medium p-3 w-[140px]">
                        <div className="flex items-center gap-1.5">
                          <Type className="size-4" />
                          <span>Type</span>
                        </div>
                      </th>
                      <th className="text-left font-medium p-3">
                        <div className="flex items-center gap-1.5">
                          <Settings2 className="size-4" />
                          <span>Value</span>
                        </div>
                      </th>
                      <th className="w-32 p-3"></th>
                    </tr>
                  </thead>
                  <SortableContext
                    items={envIds}
                    strategy={verticalListSortingStrategy}
                  >
                    <tbody>
                      {editableEnvVars.map((env, index) => (
                        <SortableEnvRow
                          key={envIds[index] || index}
                          id={envIds[index] || String(index)}
                          env={env}
                          onUpdate={(updates) => updateEnvVar(index, updates)}
                          onRemove={() => removeEnvVar(index)}
                        />
                      ))}
                    </tbody>
                  </SortableContext>
                </table>
              </div>
            </DndContext>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
