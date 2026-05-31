'use client';
import React, { useState, useEffect, useCallback } from 'react';
import { useForm } from 'react-hook-form';
import { projectsApi } from '@/services/api';
import { useToastStore } from '@/store/toastStore';
import { Button, Input, Select, Modal, Table, ConfirmDialog, Card, CardBody } from '@/components/ui';
import { TableSkeleton } from '@/components/ui/Skeleton';
import { Plus, Pencil, Trash2, FolderKanban } from 'lucide-react';
import type { Project, ProjectCreate } from '@/types';

export default function ProjectsPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingProject, setEditingProject] = useState<Project | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Project | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const addToast = useToastStore((s) => s.addToast);

  const { register, handleSubmit, reset, setValue, formState: { errors } } = useForm<ProjectCreate>();

  const fetchProjects = useCallback(async () => {
    setLoading(true);
    try {
      const res = await projectsApi.list();
      setProjects(res.data || []);
    } catch { addToast('Failed to load projects', 'error'); }
    finally { setLoading(false); }
  }, [addToast]);

  useEffect(() => { fetchProjects(); }, [fetchProjects]);

  const openCreate = () => {
    setEditingProject(null);
    reset({ name: '', description: '', type: 'personal', start_date: '', end_date: '', estimated_hours: 0 });
    setModalOpen(true);
  };

  const openEdit = (p: Project) => {
    setEditingProject(p);
    setValue('name', p.name);
    setValue('description', p.description);
    setValue('type', p.type);
    setValue('start_date', p.start_date ? p.start_date.split('T')[0] : '');
    setValue('end_date', p.end_date ? p.end_date.split('T')[0] : '');
    setValue('estimated_hours', p.estimated_hours ?? 0);
    setModalOpen(true);
  };

  const onSubmit = async (data: ProjectCreate) => {
    setSubmitting(true);
    const payload = {
      ...data,
      start_date: data.start_date ? new Date(data.start_date).toISOString() : undefined,
      end_date: data.end_date ? new Date(data.end_date).toISOString() : undefined,
      estimated_hours: data.estimated_hours ? Number(data.estimated_hours) : 0,
    };
    try {
      if (editingProject) {
        await projectsApi.update(editingProject.id, payload);
        addToast('Project updated', 'success');
      } else {
        await projectsApi.create(payload);
        addToast('Project created', 'success');
      }
      setModalOpen(false);
      fetchProjects();
    } catch { addToast('Failed to save project', 'error'); }
    finally { setSubmitting(false); }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    setSubmitting(true);
    try {
      await projectsApi.delete(deleteTarget.id);
      addToast('Project deleted', 'success');
      setDeleteTarget(null);
      fetchProjects();
    } catch (err: unknown) {
      const errMsg = (err as { response?: { data?: { error?: string } } }).response?.data?.error || 'Failed to delete project';
      addToast(errMsg, 'error');
    }
    finally { setSubmitting(false); }
  };

  const columns = [
    { key: 'name', header: 'Name', render: (p: Project) => <span className="font-medium">{p.name}</span> },
    { key: 'type', header: 'Type', render: (p: Project) => (
      <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${p.type === 'personal' ? 'bg-green-100 text-green-700' : 'bg-blue-100 text-blue-700'}`}>
        {p.type}
      </span>
    )},
    { key: 'description', header: 'Description' },
    { key: 'created_at', header: 'Created', render: (p: Project) => new Date(p.created_at).toLocaleDateString() },
    { key: 'actions', header: 'Actions', render: (p: Project) => (
      <div className="flex gap-2">
        <button onClick={(e) => { e.stopPropagation(); openEdit(p); }} className="text-blue-600 hover:text-blue-800" e2e-test-id={`project-edit-btn-${p.id}`}><Pencil size={16} /></button>
        <button onClick={(e) => { e.stopPropagation(); setDeleteTarget(p); }} className="text-red-600 hover:text-red-800" e2e-test-id={`project-delete-btn-${p.id}`}><Trash2 size={16} /></button>
      </div>
    )},
  ];

  return (
    <div e2e-test-id="projects-page">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2" e2e-test-id="projects-title">
            <FolderKanban size={24} /> Projects
          </h1>
          <p className="text-gray-500 text-sm mt-1">{projects.length} project(s)</p>
        </div>
        <Button onClick={openCreate} e2e-test-id="project-create-btn"><Plus size={16} /> New Project</Button>
      </div>

      <Card e2e-test-id="projects-card">
        <CardBody className="p-0">
          {loading ? <div className="p-6"><TableSkeleton /></div> : (
            <Table columns={columns} data={projects} keyField="id" e2e-test-id="projects-table" />
          )}
        </CardBody>
      </Card>

      {/* Create/Edit Modal */}
      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title={editingProject ? 'Edit Project' : 'Create Project'} e2eTestId="project-modal">
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" e2e-test-id="project-form">
          <Input label="Name" e2e-test-id="project-name-input" {...register('name', { required: 'Name is required' })} error={errors.name?.message} />
          <Input label="Description" e2e-test-id="project-description-input" {...register('description')} />
          <Select
            label="Type"
            e2e-test-id="project-type-select"
            options={[{ value: 'personal', label: 'Personal' }, { value: 'professional', label: 'Professional' }]}
            {...register('type', { required: 'Type is required' })}
            error={errors.type?.message}
          />
          {editingProject && (
            <>
              <div className="grid grid-cols-2 gap-4">
                <Input type="date" label="Start Date" e2e-test-id="project-start-date-input" {...register('start_date')} />
                <Input type="date" label="End Date" e2e-test-id="project-end-date-input" {...register('end_date')} />
              </div>
              <Input type="number" step="0.5" label="Estimated Hours" e2e-test-id="project-estimated-hours-input" {...register('estimated_hours')} />
            </>
          )}
          <div className="flex justify-end gap-3 pt-2">
            <Button variant="secondary" type="button" onClick={() => setModalOpen(false)} e2e-test-id="project-cancel-btn">Cancel</Button>
            <Button type="submit" loading={submitting} e2e-test-id="project-submit-btn">{editingProject ? 'Update' : 'Create'}</Button>
          </div>
        </form>
      </Modal>

      <ConfirmDialog
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={handleDelete}
        title="Delete Project"
        message={`Are you sure you want to delete "${deleteTarget?.name}"? This action cannot be undone.`}
        loading={submitting}
        e2eTestId="project-delete-dialog"
      />
    </div>
  );
}
