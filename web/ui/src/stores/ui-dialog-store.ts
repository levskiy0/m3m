import { create } from 'zustand';
import type { UIRequestData, UIFormUpdateOptions } from '@/lib/websocket';

export interface UIDialogRequest extends UIRequestData {
  projectId: string;
}

// Form state for managing loading/errors
export interface FormState {
  loading: boolean;
  errors: Record<string, string>;
}

interface UIDialogState {
  // Queue of pending dialogs
  dialogs: UIDialogRequest[];
  // Current dialog being displayed
  currentDialog: UIDialogRequest | null;
  // Form state (loading, errors) keyed by requestId
  formStates: Map<string, FormState>;
  // Add a new dialog request to the queue
  addDialog: (projectId: string, request: UIRequestData) => void;
  // Update form state (loading, errors, close)
  updateFormState: (requestId: string, update: UIFormUpdateOptions) => void;
  // Respond to current dialog (for forms, doesn't close - waits for backend)
  respondToDialog: (data: unknown) => void;
  // Close current dialog without response (for alert)
  closeDialog: () => void;
  // Get form state for a request
  getFormState: (requestId: string) => FormState;
  // Response callback (set by provider)
  onResponse: ((projectId: string, requestId: string, data: unknown) => void) | null;
  setOnResponse: (
    callback: (projectId: string, requestId: string, data: unknown) => void
  ) => void;
}

const defaultFormState: FormState = { loading: false, errors: {} };

export const useUIDialogStore = create<UIDialogState>((set, get) => ({
  dialogs: [],
  currentDialog: null,
  formStates: new Map(),
  onResponse: null,

  addDialog: (projectId, request) => {
    const dialog: UIDialogRequest = { ...request, projectId };
    set((state) => {
      const newDialogs = [...state.dialogs, dialog];
      // Initialize form state for forms
      const newFormStates = new Map(state.formStates);
      if (request.dialogType === 'form') {
        newFormStates.set(request.requestId, { loading: false, errors: {} });
      }
      return {
        dialogs: newDialogs,
        formStates: newFormStates,
        // If no current dialog, show this one immediately
        currentDialog: state.currentDialog || dialog,
      };
    });
  },

  updateFormState: (requestId, update) => {
    // Handle close - remove dialog
    if (update.close) {
      const { currentDialog, dialogs, formStates } = get();
      const remaining = dialogs.filter((d) => d.requestId !== requestId);
      const newFormStates = new Map(formStates);
      newFormStates.delete(requestId);

      set({
        dialogs: remaining,
        formStates: newFormStates,
        currentDialog: currentDialog?.requestId === requestId
          ? (remaining[0] || null)
          : currentDialog,
      });
      return;
    }

    // Update loading/errors
    set((state) => {
      const newFormStates = new Map(state.formStates);
      const current = newFormStates.get(requestId) || { loading: false, errors: {} };
      newFormStates.set(requestId, {
        loading: update.loading !== undefined ? update.loading : current.loading,
        errors: update.errors !== undefined ? update.errors : current.errors,
      });
      return { formStates: newFormStates };
    });
  },

  respondToDialog: (data) => {
    const { currentDialog, onResponse, dialogs } = get();
    if (!currentDialog || !onResponse) return;

    // Send response to backend
    onResponse(currentDialog.projectId, currentDialog.requestId, data);

    // For forms, don't close - wait for backend to send close or errors
    if (currentDialog.dialogType === 'form') {
      return;
    }

    // For non-forms, remove dialog and show next
    const remaining = dialogs.filter(
      (d) => d.requestId !== currentDialog.requestId
    );
    set({
      dialogs: remaining,
      currentDialog: remaining[0] || null,
    });
  },

  closeDialog: () => {
    const { currentDialog, dialogs, formStates } = get();
    if (!currentDialog) return;

    // Remove current dialog and show next one
    const remaining = dialogs.filter(
      (d) => d.requestId !== currentDialog.requestId
    );
    const newFormStates = new Map(formStates);
    newFormStates.delete(currentDialog.requestId);

    set({
      dialogs: remaining,
      formStates: newFormStates,
      currentDialog: remaining[0] || null,
    });
  },

  getFormState: (requestId) => {
    return get().formStates.get(requestId) || defaultFormState;
  },

  setOnResponse: (callback) => {
    set({ onResponse: callback });
  },
}));
