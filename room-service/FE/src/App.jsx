import { Navigate, Route, Routes } from 'react-router-dom';
import { AuthProvider, useAuth } from './auth/AuthContext';
import LoginPage from './pages/LoginPage';
import RoomsPage from './pages/RoomsPage';
import ChatPage from './pages/ChatPage';
import EmptyChat from './pages/EmptyChat';

function RequireAuth({ children }) {
  const { user, loading } = useAuth();
  if (loading) return <div className="center muted">Loading…</div>;
  return user ? children : <Navigate to="/login" replace />;
}

function LoginRoute() {
  const { user, loading } = useAuth();
  if (loading) return <div className="center muted">Loading…</div>;
  return user ? <Navigate to="/" replace /> : <LoginPage />;
}

export default function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route path="/login" element={<LoginRoute />} />
        <Route
          path="/"
          element={
            <RequireAuth>
              <RoomsPage />
            </RequireAuth>
          }
        >
          <Route index element={<EmptyChat />} />
          <Route path="rooms/:roomId" element={<ChatPage />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </AuthProvider>
  );
}
