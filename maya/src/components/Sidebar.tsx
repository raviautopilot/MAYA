import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { LayoutDashboard, FolderKanban, Columns3, CheckSquare, Clock, Users, X } from 'lucide-react';

const navItems = [
  { href: '/dashboard', label: 'Dashboard', icon: LayoutDashboard, id: 'nav-dashboard' },
  { href: '/projects', label: 'Projects', icon: FolderKanban, id: 'nav-projects' },
  { href: '/boards', label: 'Boards', icon: Columns3, id: 'nav-boards' },
  { href: '/tasks', label: 'Tasks', icon: CheckSquare, id: 'nav-tasks' },
  { href: '/schedulers', label: 'Schedulers', icon: Clock, id: 'nav-schedulers' },
  { href: '/resources', label: 'Resources', icon: Users, id: 'nav-resources' },
];

interface SidebarProps {
  open: boolean;
  onClose: () => void;
}

export default function Sidebar({ open, onClose }: SidebarProps) {
  const location = useLocation();
  const pathname = location.pathname;

  return (
    <>
      {/* Mobile overlay */}
      {open && <div className="fixed inset-0 bg-black/30 z-40 lg:hidden" onClick={onClose} />}

      <aside
        className={`fixed lg:static inset-y-0 left-0 z-50 w-64 bg-gray-900 text-white transform transition-transform lg:transform-none ${open ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'}`}
        e2e-test-id="sidebar"
      >
        <div className="flex items-center justify-between h-16 px-6 border-b border-gray-800">
          <Link to="/projects" className="text-xl font-bold" e2e-test-id="sidebar-logo">
            <LayoutDashboard className="inline mr-2" size={20} />
            MyKanban
          </Link>
          <button className="lg:hidden text-gray-400 hover:text-white" onClick={onClose} e2e-test-id="sidebar-close-btn">
            <X size={20} />
          </button>
        </div>

        <nav className="mt-6 px-3 space-y-1" e2e-test-id="sidebar-nav">
          {navItems.map((item) => {
            const active = pathname.startsWith(item.href);
            return (
              <Link
                key={item.href}
                to={item.href}
                onClick={onClose}
                className={`flex items-center gap-3 px-3 py-2.5 rounded-md text-sm font-medium transition ${active ? 'bg-blue-600 text-white' : 'text-gray-300 hover:bg-gray-800 hover:text-white'}`}
                e2e-test-id={item.id}
              >
                <item.icon size={18} />
                {item.label}
              </Link>
            );
          })}
        </nav>
      </aside>
    </>
  );
}
