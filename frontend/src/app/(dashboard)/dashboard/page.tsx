'use client';
import React, { useState, useEffect } from 'react';
import Link from 'next/link';
import { projectsApi, boardsApi, tasksApi, schedulersApi, resourcesApi } from '@/services/api';
import { Card, CardBody } from '@/components/ui';
import { CardSkeleton } from '@/components/ui/Skeleton';
import { FolderKanban, Columns3, CheckSquare, Clock, Users, ArrowRight } from 'lucide-react';

interface Counts { projects: number; boards: number; tasks: number; schedulers: number; resources: number }

export default function DashboardPage() {
  const [counts, setCounts] = useState<Counts>({ projects: 0, boards: 0, tasks: 0, schedulers: 0, resources: 0 });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    (async () => {
      try {
        const [p, b, t, s, r] = await Promise.all([
          projectsApi.list(), boardsApi.list(), tasksApi.list(), schedulersApi.list(), resourcesApi.list(),
        ]);
        setCounts({
          projects: (p.data || []).length,
          boards: (b.data || []).length,
          tasks: (t.data || []).length,
          schedulers: (s.data || []).length,
          resources: (r.data || []).length,
        });
      } catch { /* ignore */ }
      finally { setLoading(false); }
    })();
  }, []);

  const cards = [
    { label: 'Projects', count: counts.projects, icon: FolderKanban, href: '/projects', color: 'text-blue-600 bg-blue-50', id: 'dashboard-projects' },
    { label: 'Boards', count: counts.boards, icon: Columns3, href: '/boards', color: 'text-green-600 bg-green-50', id: 'dashboard-boards' },
    { label: 'Tasks', count: counts.tasks, icon: CheckSquare, href: '/tasks', color: 'text-orange-600 bg-orange-50', id: 'dashboard-tasks' },
    { label: 'Schedulers', count: counts.schedulers, icon: Clock, href: '/schedulers', color: 'text-purple-600 bg-purple-50', id: 'dashboard-schedulers' },
    { label: 'Resources', count: counts.resources, icon: Users, href: '/resources', color: 'text-pink-600 bg-pink-50', id: 'dashboard-resources' },
  ];

  return (
    <div e2e-test-id="dashboard-page">
      <h1 className="text-2xl font-bold text-gray-900 mb-6" e2e-test-id="dashboard-title">Dashboard</h1>

      {loading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-4">
          {[1, 2, 3, 4, 5].map((i) => <CardSkeleton key={i} />)}
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-4">
          {cards.map((c) => (
            <Link key={c.id} href={c.href} e2e-test-id={`${c.id}-link`}>
              <Card className="hover:shadow-md transition" e2e-test-id={`${c.id}-card`}>
                <CardBody>
                  <div className="flex items-center justify-between">
                    <div className={`p-2 rounded-lg ${c.color}`}><c.icon size={20} /></div>
                    <ArrowRight size={16} className="text-gray-400" />
                  </div>
                  <div className="mt-3">
                    <p className="text-2xl font-bold text-gray-900" e2e-test-id={`${c.id}-count`}>{c.count}</p>
                    <p className="text-sm text-gray-500">{c.label}</p>
                  </div>
                </CardBody>
              </Card>
            </Link>
          ))}
        </div>
      )}

      {/* Quick Actions */}
      <div className="mt-8">
        <h2 className="text-lg font-semibold text-gray-800 mb-4" e2e-test-id="dashboard-quick-actions-title">Quick Actions</h2>
        <div className="flex flex-wrap gap-3">
          <Link href="/projects" className="px-4 py-2 bg-blue-600 text-white rounded-md text-sm hover:bg-blue-700 transition" e2e-test-id="dashboard-quick-new-project">
            + New Project
          </Link>
          <Link href="/boards" className="px-4 py-2 bg-green-600 text-white rounded-md text-sm hover:bg-green-700 transition" e2e-test-id="dashboard-quick-new-board">
            + New Board
          </Link>
          <Link href="/tasks" className="px-4 py-2 bg-orange-600 text-white rounded-md text-sm hover:bg-orange-700 transition" e2e-test-id="dashboard-quick-new-task">
            + New Task
          </Link>
          <Link href="/schedulers" className="px-4 py-2 bg-purple-600 text-white rounded-md text-sm hover:bg-purple-700 transition" e2e-test-id="dashboard-quick-new-scheduler">
            + New Scheduler
          </Link>
          <Link href="/resources" className="px-4 py-2 bg-pink-600 text-white rounded-md text-sm hover:bg-pink-700 transition" e2e-test-id="dashboard-quick-new-resource">
            + New Resource
          </Link>
        </div>
      </div>
    </div>
  );
}
