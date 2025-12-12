import { useEffect } from 'react';
import { wsClient, type ServerTime } from '@/lib/websocket';
import { useAuth } from '@/providers/auth-provider';
import { useServerTimeStore } from '@/stores/server-time-store';

export function WebSocketProvider({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();
  const setServerTime = useServerTimeStore((state) => state.setServerTime);

  useEffect(() => {
    if (isAuthenticated) {
      wsClient.setHandlers({
        onTime: (data: ServerTime) => {
          setServerTime(data);
        },
      });
      wsClient.connect();
    } else {
      wsClient.disconnect();
    }

    return () => {
      wsClient.disconnect();
    };
  }, [isAuthenticated, setServerTime]);

  return <>{children}</>;
}
