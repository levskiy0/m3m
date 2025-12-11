import { create } from 'zustand';
import type { UIRequestData } from '@/lib/websocket';

export interface UIDialogRequest extends UIRequestData {
  projectId: string;
}

interface UIDialogState {
  // Queue of pending dialogs
  dialogs: UIDialogRequest[];
  // Current dialog being displayed
  currentDialog: UIDialogRequest | null;
  // Add a new dialog request to the queue
  addDialog: (projectId: string, request: UIRequestData) => void;
  // Respond to current dialog and remove it
  respondToDialog: (data: unknown) => void;
  // Close current dialog without response (for alert)
  closeDialog: () => void;
  // Response callback (set by provider)
  onResponse: ((projectId: string, requestId: string, data: unknown) => void) | null;
  setOnResponse: (
    callback: (projectId: string, requestId: string, data: unknown) => void
  ) => void;
}

export const useUIDialogStore = create<UIDialogState>((set, get) => ({
  dialogs: [],
  currentDialog: null,
  onResponse: null,

  addDialog: (projectId, request) => {
    const dialog: UIDialogRequest = { ...request, projectId };
    set((state) => {
      const newDialogs = [...state.dialogs, dialog];
      return {
        dialogs: newDialogs,
        // If no current dialog, show this one immediately
        currentDialog: state.currentDialog || dialog,
      };
    });
  },

  respondToDialog: (data) => {
    const { currentDialog, onResponse, dialogs } = get();
    if (currentDialog && onResponse) {
      onResponse(currentDialog.projectId, currentDialog.requestId, data);
    }

    // Remove current dialog and show next one
    const remaining = dialogs.filter(
      (d) => d.requestId !== currentDialog?.requestId
    );
    set({
      dialogs: remaining,
      currentDialog: remaining[0] || null,
    });
  },

  closeDialog: () => {
    const { currentDialog, dialogs } = get();
    // Remove current dialog and show next one
    const remaining = dialogs.filter(
      (d) => d.requestId !== currentDialog?.requestId
    );
    set({
      dialogs: remaining,
      currentDialog: remaining[0] || null,
    });
  },

  setOnResponse: (callback) => {
    set({ onResponse: callback });
  },
}));
