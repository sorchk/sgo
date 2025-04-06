import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import apiService from '../services/api';

interface AuthContextType {
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (clientId: string, secret: string) => Promise<void>;
  logout: () => void;
  error: string | null;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      validateToken();
    } else {
      setIsLoading(false);
    }
  }, []);

  const validateToken = async () => {
    try {
      await apiService.auth.validateToken();
      setIsAuthenticated(true);
    } catch (error) {
      localStorage.removeItem('token');
      setIsAuthenticated(false);
    } finally {
      setIsLoading(false);
    }
  };

  const login = async (clientId: string, secret: string) => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await apiService.auth.login(clientId, secret);
      const { token } = response.data.data;
      localStorage.setItem('token', token);
      setIsAuthenticated(true);
    } catch (error: any) {
      setError(error.response?.data?.error || 'Login failed');
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  const logout = () => {
    localStorage.removeItem('token');
    setIsAuthenticated(false);
  };

  return (
    <AuthContext.Provider value={{ isAuthenticated, isLoading, login, logout, error }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
