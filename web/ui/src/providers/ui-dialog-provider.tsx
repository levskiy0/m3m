import { useEffect } from 'react';
import { toast } from 'sonner';
import { wsClient, type UIRequestData, type UIRequestOptions, type UIFormUpdateOptions } from '@/lib/websocket';
import { useUIDialogStore } from '@/stores/ui-dialog-store';
import { UIDialog } from '@/components/shared/ui-dialog';

export function UIDialogProvider({ children }: { children: React.ReactNode }) {
  const { addDialog, updateFormState, setOnResponse } = useUIDialogStore();

  useEffect(() => {
    // Set up response callback to send via WebSocket
    setOnResponse((projectId, requestId, data) => {
      wsClient.sendUIResponse(projectId, requestId, data);
    });

    // Set up WebSocket handler for UI requests
    const handleUIRequest = (projectId: string, data: UIRequestData) => {
      // Handle toast separately - it uses sonner directly
      if (data.dialogType === 'toast') {
        const options = data.options as UIRequestOptions;
        const { text, severity = 'info' } = options;
        const toastFn = {
          info: toast.info,
          success: toast.success,
          warning: toast.warning,
          error: toast.error,
        }[severity] || toast;
        toastFn(text || '');
        return;
      }

      // Handle form updates (loading, errors, close)
      if (data.dialogType === 'form_update') {
        const update = data.options as UIFormUpdateOptions;
        updateFormState(data.requestId, update);
        return;
      }

      // All other dialogs go through the store
      addDialog(projectId, data);
    };

    wsClient.setHandlers({
      onUIRequest: handleUIRequest,
    });

    return () => {
      // Note: We don't clear the handler here because other handlers might still be needed
      // The wsClient.setHandlers merges handlers, so this is safe
    };
  }, [addDialog, updateFormState, setOnResponse]);

  return (
    <>
      {children}
      <UIDialog />
    </>
  );
}
