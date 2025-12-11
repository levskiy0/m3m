import { useEffect } from 'react';
import { wsClient } from '@/lib/websocket';
import { useAuth } from '@/providers/auth-provider';

export function WebSocketProvider({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();

  useEffect(() => {
    if (isAuthenticated) {
      wsClient.connect();
    } else {
      wsClient.disconnect();
    }

    return () => {
      wsClient.disconnect();
    };
  }, [isAuthenticated]);

  return <>{children}</>;
}
