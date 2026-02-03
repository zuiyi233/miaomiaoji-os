
import React from 'react';
import { Layout } from './components/Layout';
import { ProjectProvider } from './contexts/ProjectContext';
import { AuthProvider } from './contexts/AuthContext';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { ConfirmProvider } from './contexts/ConfirmContext';

const App: React.FC = () => {
  return (
    <BrowserRouter>
      <AuthProvider>
        <ProjectProvider>
          <ConfirmProvider>
            <Routes>
              <Route path="/*" element={<Layout />} />
            </Routes>
          </ConfirmProvider>
        </ProjectProvider>
      </AuthProvider>
    </BrowserRouter>
  );
};

export default App;
