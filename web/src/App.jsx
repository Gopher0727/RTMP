import React from 'react';
import { AuthProvider } from './context/AuthContext';
import { MessageProvider } from './context/MessageContext';
import AppRoutes from './components/Routes';
import './App.css';

function App() {
  return (
    <AuthProvider>
      <MessageProvider>
        <AppRoutes />
      </MessageProvider>
    </AuthProvider>
  );
}

export default App;
