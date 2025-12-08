import { config } from '@/lib/config';
import { getToken } from '@/api/client';

export type EventType = 'monitor' | 'log' | 'running' | 'goals';

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
  onConnect?: () => void;
  onDisconnect?: () => void;
  onError?: (error: Event) => void;
}

class WebSocketClient {
  private ws: WebSocket | null = null;
  private reconnectTimeout: number | null = null;
  private handlers: WSEventHandlers = {};
  private subscribedProjects: Set<string> = new Set();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;

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
        this.subscribedProjects.forEach(projectId => {
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

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.subscribedProjects.clear();
  }

  private handleMessage(data: string): void {
    try {
      // Handle multiple messages in one (newline separated)
      const messages = data.split('\n').filter(Boolean);

      for (const msg of messages) {
        const event: WSEvent = JSON.parse(msg);

        switch (event.event.type) {
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

  subscribe(projectId: string): void {
    this.subscribedProjects.add(projectId);
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.sendSubscribe(projectId);
    }
  }

  unsubscribe(projectId: string): void {
    this.subscribedProjects.delete(projectId);
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.sendUnsubscribe(projectId);
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
}

// Singleton instance
export const wsClient = new WebSocketClient();
