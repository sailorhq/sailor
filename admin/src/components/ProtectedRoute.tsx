import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

import type { Role } from '../contexts/AuthContext';

interface ProtectedRouteProps {
    children: React.ReactNode;
    requiredPermissions?: string[];
    requiredRoles?: Role[];
    fallbackPath?: string;
}

const ProtectedRoute: React.FC<ProtectedRouteProps> = ({
    children,
    requiredPermissions = [],
    requiredRoles = [],
    fallbackPath = '/login'
}) => {
    const { isAuthenticated, isLoading, hasPermission, hasRole } = useAuth();
    const location = useLocation();

    if (isLoading) {
        return (
            <div className="d-flex justify-content-center align-items-center" style={{ height: '100vh' }}>
                <div className="spinner-border text-primary" role="status">
                    <span className="visually-hidden">Loading...</span>
                </div>
            </div>
        );
    }

    if (!isAuthenticated) {
        return <Navigate to={fallbackPath} state={{ from: location }} replace />;
    }

    // Check permissions
    if (requiredPermissions.length > 0) {
        const hasAllPermissions = requiredPermissions.every(permission => hasPermission(permission));
        if (!hasAllPermissions) {
            return <Navigate to="/unauthorized" replace />;
        }
    }

    // Check roles
    if (requiredRoles.length > 0) {
        const hasRequiredRole = requiredRoles.some(role => hasRole(role));
        if (!hasRequiredRole) {
            return <Navigate to="/unauthorized" replace />;
        }
    }

    return <>{children}</>;
};

export default ProtectedRoute; 