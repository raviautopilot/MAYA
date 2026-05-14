'use client';
import React, { useState, useEffect, useCallback } from 'react';
import { useForm } from 'react-hook-form';
import { schedulersApi, tasksApi } from '@/services/api';
import { useToastStore } from '@/store/toastStore';
import { Button, Input, Select, Modal, Table, ConfirmDialog, Card, CardBody } from '@/components/ui';
import { TableSkeleton } from '@/components/ui/Skeleton';
import { Plus, Pencil, Trash2, Clock } from 'lucide-react';
import type { Scheduler, SchedulerCreate, Task } from '@/types';

export default function SchedulersPage() {
  const [schedulers, setSchedulers] = useState<Scheduler[]>([]);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingScheduler, setEditingScheduler] = useState<Scheduler | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Scheduler | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const addToast = useToastStore((s) => s.addToast);

  const { register, handleSubmit, reset, setValue, watch, formState: { errors } } = useForm<SchedulerCreate>();
  const schedType = watch('type');

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [sRes, tRes] = await Promise.all([schedulersApi.list(), tasksApi.list()]);
      setSchedulers(sRes.data || []);
      setTasks(tRes.data || []);
    } catch { addToast('Failed to load schedulers', 'error'); }
    finally { setLoading(false); }
  }, [addToast]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const openCreate = () => {
    setEditingScheduler(null);
    reset({ name: '', type: 'daily', cron_expression: '', linked_task_template_id: '' });
    setModalOpen(true);
  };

  const openEdit = (s: Scheduler) => {
    setEditingScheduler(s);
    setValue('name', s.name);
    setValue('type', s.type);
    setValue('cron_expression', s.cron_expression || '');
    setValue('linked_task_template_id', s.linked_task_template_id || '');
    setModalOpen(true);
  };

  const onSubmit = async (data: SchedulerCreate) => {
    setSubmitting(true);
    try {
      const payload = { ...data };
      if (data.type !== 'cron') delete payload.cron_expression;
      if (!payload.linked_task_template_id) delete payload.linked_task_template_id;
      if (editingScheduler) {
        await schedulersApi.update(editingScheduler.id, payload);
        addToast('Scheduler updated', 'success');
      } else {
        await schedulersApi.create(payload);
        addToast('Scheduler created', 'success');
      }
      setModalOpen(false);
      fetchData();
    } catch { addToast('Failed to save scheduler', 'error'); }
    finally { setSubmitting(false); }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    setSubmitting(true);
    try {
      await schedulersApi.delete(deleteTarget.id);
      addToast('Scheduler deleted', 'success');
      setDeleteTarget(null);
      fetchData();
    } catch { addToast('Failed to delete scheduler', 'error'); }
    finally { setSubmitting(false); }
  };

  const columns = [
    { key: 'name', header: 'Name', render: (s: Scheduler) => <span className="font-medium">{s.name}</span> },
    { key: 'type', header: 'Type', render: (s: Scheduler) => (
      <span className="inline-flex px-2 py-0.5 rounded-full text-xs font-medium bg-indigo-100 text-indigo-700">{s.type}</span>
    )},
    { key: 'cron_expression', header: 'Cron', render: (s: Scheduler) => s.cron_expression || '—' },
    { key: 'next_run', header: 'Next Run', render: (s: Scheduler) => s.next_run ? new Date(s.next_run).toLocaleString() : '—' },
    { key: 'linked_task_template_id', header: 'Linked Task', render: (s: Scheduler) => {
      const t = tasks.find((tk) => tk.id === s.linked_task_template_id);
      return t ? t.title : s.linked_task_template_id || '—';
    }},
    { key: 'actions', header: 'Actions', render: (s: Scheduler) => (
      <div className="flex gap-2">
        <button onClick={(e) => { e.stopPropagation(); openEdit(s); }} className="text-blue-600 hover:text-blue-800" e2e-test-id={`scheduler-edit-btn-${s.id}`}><Pencil size={16} /></button>
        <button onClick={(e) => { e.stopPropagation(); setDeleteTarget(s); }} className="text-red-600 hover:text-red-800" e2e-test-id={`scheduler-delete-btn-${s.id}`}><Trash2 size={16} /></button>
      </div>
    )},
  ];

  return (
    <div e2e-test-id="schedulers-page">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2" e2e-test-id="schedulers-title"><Clock size={24} /> Schedulers</h1>
          <p className="text-gray-500 text-sm mt-1">{schedulers.length} scheduler(s)</p>
        </div>
        <Button onClick={openCreate} e2e-test-id="scheduler-create-btn"><Plus size={16} /> New Scheduler</Button>
      </div>

      <Card e2e-test-id="schedulers-card">
        <CardBody className="p-0">
          {loading ? <div className="p-6"><TableSkeleton /></div> : (
            <Table columns={columns} data={schedulers} keyField="id" e2e-test-id="schedulers-table" />
          )}
        </CardBody>
      </Card>

      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title={editingScheduler ? 'Edit Scheduler' : 'Create Scheduler'} e2eTestId="scheduler-modal">
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" e2e-test-id="scheduler-form">
          <Input label="Name" e2e-test-id="scheduler-name-input" {...register('name', { required: 'Name is required' })} error={errors.name?.message} />
          <Select
            label="Type"
            e2e-test-id="scheduler-type-select"
            options={['cron', 'daily', 'weekly', 'monthly', 'yearly'].map((t) => ({ value: t, label: t.charAt(0).toUpperCase() + t.slice(1) }))}
            {...register('type', { required: 'Type is required' })}
            error={errors.type?.message}
          />
          {schedType === 'cron' && (
            <Input label="Cron Expression" placeholder="0 9 * * 1" e2e-test-id="scheduler-cron-input" {...register('cron_expression')} />
          )}
          <Select
            label="Linked Task Template"
            e2e-test-id="scheduler-linked-task-select"
            options={tasks.map((t) => ({ value: t.id, label: t.title }))}
            {...register('linked_task_template_id')}
          />
          <div className="flex justify-end gap-3 pt-2">
            <Button variant="secondary" type="button" onClick={() => setModalOpen(false)} e2e-test-id="scheduler-cancel-btn">Cancel</Button>
            <Button type="submit" loading={submitting} e2e-test-id="scheduler-submit-btn">{editingScheduler ? 'Update' : 'Create'}</Button>
          </div>
        </form>
      </Modal>

      <ConfirmDialog
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={handleDelete}
        title="Delete Scheduler"
        message={`Are you sure you want to delete "${deleteTarget?.name}"?`}
        loading={submitting}
        e2eTestId="scheduler-delete-dialog"
      />
    </div>
  );
}
