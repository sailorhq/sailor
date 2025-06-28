import React, { createContext, useContext, useState, useEffect } from 'react';
import type { ReactNode } from 'react';

export type Role = 'admin' | 'user' | 'viewer';

interface User {
    id: string;
    username: string;
    email: string;
    roles: Role[];
    permissions: string[];
    token: string;
}

interface AuthContextType {
    user: User | null;
    isAuthenticated: boolean;
    isLoading: boolean;
    login: (userData: User) => void;
    logout: () => void;
    hasPermission: (permission: string) => boolean;
    hasRole: (role: Role) => boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);
const SAILOR_USER_KEY = 'sailor-user';

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};

interface AuthProviderProps {
    children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
    const [user, setUser] = useState<User | null>(null);
    const [isLoading, setIsLoading] = useState(true);


    useEffect(() => {
        // Check for existing token on app load
        const storedUser = localStorage.getItem(SAILOR_USER_KEY);
        if (storedUser) {
            const user = JSON.parse(storedUser) as User;
            if (user) {
                validateToken(user);
            } else {
                setIsLoading(false);
            }
        }
    }, []);

    const validateToken = async (user: User) => {
        try {
            const response = await fetch('http://localhost:7766/api/v1/auth.validate', {
                method: 'POST',
                headers: {
                    'x-token': user.token,
                    'x-username': user.username,
                }
            });

            if (response.status == 200) {
                const userData = await response.json();
                setUser(userData);
            } else {
                localStorage.removeItem(SAILOR_USER_KEY);
            }
        } catch (error) {
            console.error('Token validation error:', error);
            localStorage.removeItem(SAILOR_USER_KEY);
        } finally {
            setIsLoading(false);
        }
    };

    const login = (userData: User) => {
        localStorage.setItem(SAILOR_USER_KEY, JSON.stringify(userData));
        setUser(userData);
    };

    const logout = () => {
        localStorage.removeItem(SAILOR_USER_KEY);
        setUser(null);
    };

    const hasPermission = (permission: string): boolean => {
        return user?.permissions.includes(permission) || false;
    };

    const hasRole = (role: Role): boolean => {
        return user?.roles.includes(role) || false;
    };

    const value: AuthContextType = {
        user,
        isAuthenticated: user !== null && user !== undefined,
        isLoading,
        login,
        logout,
        hasPermission,
        hasRole
    };

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    );
}; 