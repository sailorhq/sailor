import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import ProtectedRoute from '../components/ProtectedRoute';
import ApplicationsPage from './ApplicationsPage';
import SettingsPage from './SettingsPage';
import ApplicationInfoPage from './ApplicationInfoPage';

const DashboardRoutes: React.FC = () => (
    <Routes>

        <Route path="apps" element={
            <ProtectedRoute requiredRoles={['admin', 'user']}>
                <ApplicationsPage />
            </ProtectedRoute>
        } />
        <Route path="apps/:app" element={
            <ProtectedRoute requiredRoles={['admin', 'user']}>
                <ApplicationInfoPage />
            </ProtectedRoute>
        } />


        {/* Settings page - only for admin and user roles */}
        <Route
            path="settings"
            element={
                <ProtectedRoute requiredRoles={['admin']}>
                    <SettingsPage />
                </ProtectedRoute>
            }
        />

        <Route path="*" element={<Navigate to="apps" replace />} />
    </Routes>
);

export default DashboardRoutes; 