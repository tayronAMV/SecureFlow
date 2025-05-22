import { useState } from 'react';
import { ThemeProvider } from './contexts/ThemeContext';
import { AuthProvider } from './contexts/AuthContext';
import { WebSocketProvider } from './contexts/WebSocketContext';
import AppRoutes from './routes/AppRoutes';
import { Toaster } from '@/components/ui/sonner';

function App() {
  return (
    <ThemeProvider>
      <AuthProvider>
        <WebSocketProvider>
          <AppRoutes />
          <Toaster />
        </WebSocketProvider>
      </AuthProvider>
    </ThemeProvider>
  );
}

export default App;