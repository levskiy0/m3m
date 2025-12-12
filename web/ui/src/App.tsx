import { AuthProvider } from '@/providers/auth-provider';
import { QueryProvider } from '@/providers/query-provider';
import { ThemeProvider } from '@/providers/theme-provider';
import { WebSocketProvider } from '@/providers/websocket-provider';
import { UIDialogProvider } from '@/providers/ui-dialog-provider';
import { AppRouter } from '@/routes';

function App() {
  return (
    <ThemeProvider>
      <QueryProvider>
        <AuthProvider>
          <WebSocketProvider>
            <UIDialogProvider>
              <AppRouter />
            </UIDialogProvider>
          </WebSocketProvider>
        </AuthProvider>
      </QueryProvider>
    </ThemeProvider>
  );
}

export default App;
