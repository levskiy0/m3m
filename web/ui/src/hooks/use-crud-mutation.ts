import { useMutation, useQueryClient, type QueryKey } from '@tanstack/react-query';
import { toast } from 'sonner';

interface UseCrudMutationOptions<TData, TVariables> {
  mutationFn: (variables: TVariables) => Promise<TData>;
  queryKey: QueryKey;
  successMessage?: string;
  errorMessage?: string;
  onSuccess?: (data: TData, variables: TVariables) => void;
  onError?: (error: Error) => void;
}

/**
 * Wrapper around useMutation with common CRUD patterns:
 * - Invalidates query on success
 * - Shows toast notifications
 * - Handles errors with proper messages
 */
export function useCrudMutation<TData = unknown, TVariables = void>({
  mutationFn,
  queryKey,
  successMessage,
  errorMessage = 'Operation failed',
  onSuccess,
  onError,
}: UseCrudMutationOptions<TData, TVariables>) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn,
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries({ queryKey });
      if (successMessage) {
        toast.success(successMessage);
      }
      onSuccess?.(data, variables);
    },
    onError: (err) => {
      const message = err instanceof Error ? err.message : errorMessage;
      toast.error(message);
      onError?.(err instanceof Error ? err : new Error(message));
    },
  });
}

interface UseCrudMutationsOptions<TCreateData, TCreateVars, TUpdateData, TUpdateVars, TDeleteData, TDeleteVars> {
  queryKey: QueryKey;
  create?: {
    mutationFn: (variables: TCreateVars) => Promise<TCreateData>;
    successMessage?: string;
    onSuccess?: (data: TCreateData) => void;
  };
  update?: {
    mutationFn: (variables: TUpdateVars) => Promise<TUpdateData>;
    successMessage?: string;
    onSuccess?: (data: TUpdateData) => void;
  };
  delete?: {
    mutationFn: (variables: TDeleteVars) => Promise<TDeleteData>;
    successMessage?: string;
    onSuccess?: () => void;
  };
}

/**
 * Creates a set of CRUD mutations for a resource.
 * Useful when you need create, update, and delete operations on the same entity.
 */
export function useCrudMutations<
  TCreateData = unknown,
  TCreateVars = void,
  TUpdateData = unknown,
  TUpdateVars = void,
  TDeleteData = unknown,
  TDeleteVars = void,
>({
  queryKey,
  create,
  update,
  delete: del,
}: UseCrudMutationsOptions<TCreateData, TCreateVars, TUpdateData, TUpdateVars, TDeleteData, TDeleteVars>) {
  const createMutation = create
    ? useCrudMutation<TCreateData, TCreateVars>({
        mutationFn: create.mutationFn,
        queryKey,
        successMessage: create.successMessage,
        onSuccess: create.onSuccess,
      })
    : undefined;

  const updateMutation = update
    ? useCrudMutation<TUpdateData, TUpdateVars>({
        mutationFn: update.mutationFn,
        queryKey,
        successMessage: update.successMessage,
        onSuccess: update.onSuccess,
      })
    : undefined;

  const deleteMutation = del
    ? useCrudMutation<TDeleteData, TDeleteVars>({
        mutationFn: del.mutationFn,
        queryKey,
        successMessage: del.successMessage,
        onSuccess: del.onSuccess,
      })
    : undefined;

  return {
    createMutation,
    updateMutation,
    deleteMutation,
  };
}
