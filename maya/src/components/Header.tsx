import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/store/authStore';
import { Menu, LogOut, User } from 'lucide-react';

interface HeaderProps {
  onMenuClick: () => void;
}

export default function Header({ onMenuClick }: HeaderProps) {
  const navigate = useNavigate();
  const { user, logout } = useAuthStore();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <header className="h-16 bg-white border-b border-gray-200 flex items-center justify-between px-6" e2e-test-id="header">
      <button className="lg:hidden text-gray-600 hover:text-gray-900" onClick={onMenuClick} e2e-test-id="header-menu-btn">
        <Menu size={24} />
      </button>
      <div className="hidden lg:block" />

      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2 text-sm text-gray-600" e2e-test-id="header-user-info">
          <User size={16} />
          <span>{user?.email || 'User'}</span>
        </div>
        <button
          onClick={handleLogout}
          className="flex items-center gap-1 text-sm text-gray-500 hover:text-red-600 transition"
          e2e-test-id="header-logout-btn"
        >
          <LogOut size={16} />
          Logout
        </button>
      </div>
    </header>
  );
}
