import { useEffect } from 'react';
import { wsClient, type UIRequestData } from '@/lib/websocket';
import { useUIDialogStore } from '@/stores/ui-dialog-store';
import { UIDialog } from '@/components/shared/ui-dialog';

export function UIDialogProvider({ children }: { children: React.ReactNode }) {
  const { addDialog, setOnResponse } = useUIDialogStore();

  useEffect(() => {
    // Set up response callback to send via WebSocket
    setOnResponse((projectId, requestId, data) => {
      wsClient.sendUIResponse(projectId, requestId, data);
    });

    // Set up WebSocket handler for UI requests
    const handleUIRequest = (projectId: string, data: UIRequestData) => {
      addDialog(projectId, data);
    };

    wsClient.setHandlers({
      onUIRequest: handleUIRequest,
    });

    return () => {
      // Note: We don't clear the handler here because other handlers might still be needed
      // The wsClient.setHandlers merges handlers, so this is safe
    };
  }, [addDialog, setOnResponse]);

  return (
    <>
      {children}
      <UIDialog />
    </>
  );
}
