import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import ApplicationsPage from './ApplicationsPage';
import SettingsPage from './SettingsPage';
import ApplicationInfoPage from './ApplicationInfoPage';

const DashboardRoutes: React.FC = () => (
    <Routes>
        <Route path="apps" element={<ApplicationsPage />} />
        <Route path="apps/:app" element={<ApplicationInfoPage />} />
        <Route path="settings" element={<SettingsPage />} />
        <Route path="*" element={<Navigate to="applications" replace />} />
    </Routes>
);

export default DashboardRoutes; 