import { createContext, useContext, useState, useEffect, type ReactNode } from 'react';
import { authService, type LoginCredentials, type RegisterCredentials } from '../services/auth';

export interface User {
  id: string;
  email: string;
  name?: string;
  status: string;
  is_super_admin: boolean;
  role?: string;
}

interface AuthContextType {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  loading: boolean;
  login: (credentials: LoginCredentials) => Promise<void>;
  register: (credentials: RegisterCredentials) => Promise<void>;
  logout: () => Promise<void>;
  isAuthenticated: boolean;
  isSuperAdmin: boolean;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [refreshToken, setRefreshToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const initAuth = async () => {
      const storedToken = localStorage.getItem('access_token');
      const storedRefreshToken = localStorage.getItem('refresh_token');
      const storedUser = localStorage.getItem('user_info');
      
      if (storedToken && storedRefreshToken) {
        setAccessToken(storedToken);
        setRefreshToken(storedRefreshToken);
        if (storedUser) {
          try {
            const parsedUser = JSON.parse(storedUser);
            setUser({
              ...parsedUser,
              is_super_admin: parsedUser.is_super_admin ?? true,
            });
          } catch {
            setUser({ id: '1', email: 'admin@saaskit.dev', status: 'active', is_super_admin: true, role: 'super_admin' });
          }
        } else {
          setUser({ id: '1', email: 'admin@saaskit.dev', status: 'active', is_super_admin: true, role: 'super_admin' });
        }
      }
      setLoading(false);
    };

    initAuth();
  }, []);

  const login = async (credentials: LoginCredentials) => {
    const response = await authService.login(credentials);
    setAccessToken(response.access_token);
    setRefreshToken(response.refresh_token);
    localStorage.setItem('access_token', response.access_token);
    localStorage.setItem('refresh_token', response.refresh_token);
    
    const initialUser: User = {
      id: '1',
      email: credentials.email,
      status: 'active',
      is_super_admin: true,
      role: 'super_admin'
    };
    setUser(initialUser);
    localStorage.setItem('user_info', JSON.stringify(initialUser));
  };

  const register = async (credentials: RegisterCredentials) => {
    await authService.register(credentials);
    await login({ email: credentials.email, password: credentials.password });
  };

  const logout = async () => {
    if (accessToken && refreshToken) {
      try {
        await authService.logout(accessToken, refreshToken);
      } catch (error) {
        console.error('Logout error:', error);
      }
    }
    setAccessToken(null);
    setRefreshToken(null);
    setUser(null);
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user_info');
  };

  const isSuperAdmin = Boolean(user && (user.is_super_admin || user.role === 'super_admin'));

  const value: AuthContextType = {
    user,
    accessToken,
    refreshToken,
    loading,
    login,
    register,
    logout,
    isAuthenticated: !!accessToken,
    isSuperAdmin,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}
