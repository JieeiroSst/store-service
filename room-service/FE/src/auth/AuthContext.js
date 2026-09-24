import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import * as api from '../api/client';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  // With a stored session we don't know who it is until /me answers.
  const [loading, setLoading] = useState(api.hasSession());

  useEffect(() => {
    let cancelled = false;
    if (api.hasSession()) {
      api
        .me()
        .then((u) => !cancelled && setUser(u))
        .catch(() => {})
        .finally(() => !cancelled && setLoading(false));
    }
    // A failed refresh (or logout elsewhere) clears the session: drop the user.
    const off = api.onSessionChange((s) => {
      if (!s) setUser(null);
    });
    return () => {
      cancelled = true;
      off();
    };
  }, []);

  const login = useCallback(async (username, password) => {
    await api.login(username, password);
    setUser(await api.me());
  }, []);

  const logout = useCallback(() => api.logout(), []);

  const value = useMemo(() => ({ user, loading, login, logout }), [user, loading, login, logout]);
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export const useAuth = () => useContext(AuthContext);
