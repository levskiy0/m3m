import { config } from '@/lib/config';
import { getToken } from '@/api/client';
import type { ActionRuntimeState } from '@/types';

export type EventType = 'monitor' | 'log' | 'running' | 'goals' | 'actions' | 'ui_request';

export type UIDialogType = 'alert' | 'confirm' | 'prompt' | 'form' | 'toast' | 'form_update';

export interface UIRequestData {
  requestId: string;
  dialogType: UIDialogType;
  options: UIRequestOptions | UIFormUpdateOptions;
}

export interface UIRequestOptions {
  title?: string;
  text?: string;
  severity?: 'info' | 'success' | 'warning' | 'error';
  icon?: string;
  yes?: string;
  no?: string;
  placeholder?: string;
  defaultValue?: string;
  schema?: UIFormField[];
  actions?: UIFormAction[];
}

export interface UIFormUpdateOptions {
  loading?: boolean;
  errors?: Record<string, string>;
  close?: boolean;
}

export interface UIFormField {
  name: string;
  type: 'input' | 'textarea' | 'checkbox' | 'select' | 'combobox' | 'radiogroup' | 'date' | 'datetime';
  label?: string;
  hint?: string;
  colspan?: number | 'full';
  required?: boolean;
  placeholder?: string;
  defaultValue?: unknown;
  options?: string[] | { label: string; value: string }[];
}

export interface UIFormAction {
  label: string;
  variant?: 'default' | 'outline' | 'destructive';
  action: string;
}

export interface WSEvent {
  projectId: string;
  event: {
    type: EventType;
    data: unknown;
  };
}

export interface WSEventHandlers {
  onMonitor?: (projectId: string, data: unknown) => void;
  onLog?: (projectId: string, data: { hasNewLogs: boolean }) => void;
  onRunning?: (projectId: string, data: { running: boolean }) => void;
  onGoals?: (projectId: string, data: unknown) => void;
  onActions?: (projectId: string, data: ActionRuntimeState[]) => void;
  onUIRequest?: (projectId: string, data: UIRequestData) => void;
  onConnect?: () => void;
  onDisconnect?: () => void;
  onError?: (error: Event) => void;
}

class WebSocketClient {
  private ws: WebSocket | null = null;
  private reconnectTimeout: number | null = null;
  private handlers: WSEventHandlers = {};
  private subscriptionCounts: Map<string, number> = new Map(); // reference counting
  private pendingUnsubscribes: Map<string, number> = new Map(); // debounce unsubscribes
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private unsubscribeDelay = 500; // ms to wait before actually unsubscribing
  private _sessionId: string | null = null; // unique session ID for this connection

  private getWSUrl(): string {
    // config.apiURL is like http://localhost:8080 or https://example.com
    const apiUrl = config.apiURL;

    // Convert http(s):// to ws(s)://
    let wsUrl = apiUrl.replace(/^http/, 'ws');

    // Append /api/ws path
    wsUrl = wsUrl.replace(/\/$/, '') + '/api/ws';

    return wsUrl;
  }

  connect(): void {
    // Prevent multiple connections
    if (this.ws?.readyState === WebSocket.OPEN || this.ws?.readyState === WebSocket.CONNECTING) {
      return;
    }

    const token = getToken();
    if (!token) {
      console.warn('WebSocket: No auth token, skipping connection');
      return;
    }

    try {
      // Include token in URL for WebSocket auth
      const wsUrl = `${this.getWSUrl()}?token=${token}`;
      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.reconnectAttempts = 0;
        this.handlers.onConnect?.();

        // Re-subscribe to projects
        this.subscriptionCounts.forEach((_, projectId) => {
          this.sendSubscribe(projectId);
        });
      };

      this.ws.onclose = () => {
        console.log('WebSocket disconnected');
        this.handlers.onDisconnect?.();
        this.scheduleReconnect();
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        this.handlers.onError?.(error);
      };

      this.ws.onmessage = (event) => {
        this.handleMessage(event.data);
      };
    } catch (error) {
      console.error('WebSocket connection error:', error);
      this.scheduleReconnect();
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.warn('WebSocket: Max reconnect attempts reached');
      return;
    }

    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
    }

    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts);
    this.reconnectTimeout = window.setTimeout(() => {
      this.reconnectAttempts++;
      console.log(`WebSocket: Reconnecting (attempt ${this.reconnectAttempts})`);
      this.connect();
    }, delay);
  }

  disconnect(): void {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }

    // Clear pending unsubscribes
    this.pendingUnsubscribes.forEach((timeout) => clearTimeout(timeout));
    this.pendingUnsubscribes.clear();

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.subscriptionCounts.clear();
    this._sessionId = null;
  }

  private handleMessage(data: string): void {
    try {
      // Handle multiple messages in one (newline separated)
      const messages = data.split('\n').filter(Boolean);

      for (const msg of messages) {
        const parsed = JSON.parse(msg);

        // Handle session message (sent on connect)
        if (parsed.type === 'session' && parsed.sessionId) {
          this._sessionId = parsed.sessionId;
          console.log('WebSocket: Session ID received:', this._sessionId);
          continue;
        }

        // Handle regular events
        const event = parsed as WSEvent;

        switch (event.event?.type) {
          case 'monitor':
            this.handlers.onMonitor?.(event.projectId, event.event.data);
            break;
          case 'log':
            this.handlers.onLog?.(event.projectId, event.event.data as { hasNewLogs: boolean });
            break;
          case 'running':
            this.handlers.onRunning?.(event.projectId, event.event.data as { running: boolean });
            break;
          case 'goals':
            this.handlers.onGoals?.(event.projectId, event.event.data);
            break;
          case 'actions':
            this.handlers.onActions?.(event.projectId, event.event.data as ActionRuntimeState[]);
            break;
          case 'ui_request':
            this.handlers.onUIRequest?.(event.projectId, event.event.data as UIRequestData);
            break;
        }
      }
    } catch (error) {
      console.error('WebSocket: Failed to parse message:', error, data);
    }
  }

  private send(message: object): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    }
  }

  private sendSubscribe(projectId: string): void {
    this.send({ action: 'subscribe', projectId });
  }

  private sendUnsubscribe(projectId: string): void {
    this.send({ action: 'unsubscribe', projectId });
  }

  sendUIResponse(projectId: string, requestId: string, data: unknown): void {
    this.send({ action: 'ui_response', projectId, requestId, data });
  }

  subscribe(projectId: string): void {
    // Cancel any pending unsubscribe
    const pendingTimeout = this.pendingUnsubscribes.get(projectId);
    if (pendingTimeout) {
      clearTimeout(pendingTimeout);
      this.pendingUnsubscribes.delete(projectId);
    }

    const count = this.subscriptionCounts.get(projectId) || 0;
    this.subscriptionCounts.set(projectId, count + 1);

    // Only send subscribe on first subscription
    if (count === 0 && this.ws?.readyState === WebSocket.OPEN) {
      this.sendSubscribe(projectId);
    }
  }

  unsubscribe(projectId: string): void {
    const count = this.subscriptionCounts.get(projectId) || 0;
    if (count <= 1) {
      // Last subscriber - schedule unsubscribe with delay
      this.subscriptionCounts.delete(projectId);

      // Debounce: wait before actually unsubscribing
      const timeout = window.setTimeout(() => {
        this.pendingUnsubscribes.delete(projectId);
        // Only unsubscribe if still not re-subscribed
        if (!this.subscriptionCounts.has(projectId) && this.ws?.readyState === WebSocket.OPEN) {
          this.sendUnsubscribe(projectId);
        }
      }, this.unsubscribeDelay);

      this.pendingUnsubscribes.set(projectId, timeout);
    } else {
      // Still have other subscribers
      this.subscriptionCounts.set(projectId, count - 1);
    }
  }

  setHandlers(handlers: WSEventHandlers): void {
    this.handlers = { ...this.handlers, ...handlers };
  }

  clearHandlers(): void {
    this.handlers = {};
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  get sessionId(): string | null {
    return this._sessionId;
  }
}

// Singleton instance
export const wsClient = new WebSocketClient();
