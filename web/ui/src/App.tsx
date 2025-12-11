import { AuthProvider } from '@/providers/auth-provider';
import { QueryProvider } from '@/providers/query-provider';
import { ThemeProvider } from '@/providers/theme-provider';
import { WebSocketProvider } from '@/providers/websocket-provider';
import { AppRouter } from '@/routes';

function App() {
  return (
    <ThemeProvider>
      <QueryProvider>
        <AuthProvider>
          <WebSocketProvider>
            <AppRouter />
          </WebSocketProvider>
        </AuthProvider>
      </QueryProvider>
    </ThemeProvider>
  );
}

export default App;
