'use client';
import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Sidebar from '@/components/Sidebar';
import Header from '@/components/Header';
import { ToastContainer } from '@/components/ui';
import { useAuthStore } from '@/store/authStore';

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const { hydrate } = useAuthStore();
  const router = useRouter();
  const [ready, setReady] = useState(false);

  useEffect(() => {
    hydrate();
    const token = localStorage.getItem('token');
    if (!token) {
      router.replace('/login');
    } else {
      setReady(true);
    }
  }, [hydrate, router]);

  if (!ready) return null;

  return (
    <div className="flex h-screen bg-gray-50">
      <Sidebar open={sidebarOpen} onClose={() => setSidebarOpen(false)} />
      <div className="flex-1 flex flex-col overflow-hidden">
        <Header onMenuClick={() => setSidebarOpen(true)} />
        <main className="flex-1 overflow-y-auto p-6" e2e-test-id="main-content">
          {children}
          <ToastContainer />
        </main>
      </div>
    </div>
  );
}
