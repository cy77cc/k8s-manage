import React, { createContext, useContext, useEffect, useMemo, useState } from 'react';
import type { ReactNode } from 'react';
import { authApi } from '../../api/modules/auth';
import type { AuthUser, LoginParams, RegisterParams } from '../../api/modules/auth';

interface AuthContextType {
  user: AuthUser | null;
  token: string | null;
  loading: boolean;
  isAuthenticated: boolean;
  login: (payload: LoginParams) => Promise<void>;
  register: (payload: RegisterParams) => Promise<void>;
  logout: () => void;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [token, setToken] = useState<string | null>(localStorage.getItem('token'));
  const [loading, setLoading] = useState(true);

  const persistSession = (nextToken: string, nextUser: AuthUser, permissions?: string[]) => {
    localStorage.setItem('token', nextToken);
    localStorage.setItem('user', JSON.stringify(nextUser));
    if (permissions) {
      localStorage.setItem('permissions', JSON.stringify(permissions));
    }
    setToken(nextToken);
    setUser(nextUser);
  };

  const clearSession = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    localStorage.removeItem('permissions');
    setToken(null);
    setUser(null);
  };

  const refreshUser = async () => {
    if (!localStorage.getItem('token')) {
      clearSession();
      return;
    }
    const res = await authApi.getMe();
    setUser(res.data);
    localStorage.setItem('user', JSON.stringify(res.data));
    if (res.data.permissions) {
      localStorage.setItem('permissions', JSON.stringify(res.data.permissions));
    }
  };

  const login = async (payload: LoginParams) => {
    const res = await authApi.login(payload);
    persistSession(res.data.token, res.data.user, res.data.permissions);
  };

  const register = async (payload: RegisterParams) => {
    const res = await authApi.register(payload);
    persistSession(res.data.token, res.data.user, res.data.permissions);
  };

  const logout = () => {
    clearSession();
  };

  useEffect(() => {
    let mounted = true;
    const bootstrap = async () => {
      try {
        const storedToken = localStorage.getItem('token');
        if (!storedToken) {
          clearSession();
          return;
        }
        const userText = localStorage.getItem('user');
        if (userText && mounted) {
          setUser(JSON.parse(userText) as AuthUser);
        }
        await refreshUser();
      } catch {
        clearSession();
      } finally {
        if (mounted) {
          setLoading(false);
        }
      }
    };
    bootstrap();
    return () => {
      mounted = false;
    };
  }, []);

  const value = useMemo(
    () => ({
      user,
      token,
      loading,
      isAuthenticated: Boolean(token && user),
      login,
      register,
      logout,
      refreshUser,
    }),
    [user, token, loading],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
};
