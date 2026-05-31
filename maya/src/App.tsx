import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import Home from '@/app/page';
import LoginPage from '@/app/login/page';
import DashboardLayout from '@/app/(dashboard)/layout';
import DashboardPage from '@/app/(dashboard)/dashboard/page';
import ProjectsPage from '@/app/(dashboard)/projects/page';
import BoardsPage from '@/app/(dashboard)/boards/page';
import TasksPage from '@/app/(dashboard)/tasks/page';
import SchedulersPage from '@/app/(dashboard)/schedulers/page';
import ResourcesPage from '@/app/(dashboard)/resources/page';

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/login" element={<LoginPage />} />
        <Route element={<DashboardLayout />}>
          <Route path="/dashboard" element={<DashboardPage />} />
          <Route path="/projects" element={<ProjectsPage />} />
          <Route path="/boards" element={<BoardsPage />} />
          <Route path="/tasks" element={<TasksPage />} />
          <Route path="/schedulers" element={<SchedulersPage />} />
          <Route path="/resources" element={<ResourcesPage />} />
        </Route>
        {/* Redirect unknown routes to Home */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
