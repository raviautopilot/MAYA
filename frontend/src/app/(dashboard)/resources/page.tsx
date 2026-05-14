'use client';
import React, { useState, useEffect, useCallback } from 'react';
import { useForm } from 'react-hook-form';
import { resourcesApi } from '@/services/api';
import { useToastStore } from '@/store/toastStore';
import { Button, Input, Select, Modal, Table, ConfirmDialog, Card, CardBody } from '@/components/ui';
import { TableSkeleton } from '@/components/ui/Skeleton';
import { Plus, Pencil, Trash2, Users } from 'lucide-react';
import type { Resource, ResourceCreate } from '@/types';

export default function ResourcesPage() {
  const [resources, setResources] = useState<Resource[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingResource, setEditingResource] = useState<Resource | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Resource | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [linkedItemsInput, setLinkedItemsInput] = useState('');
  const addToast = useToastStore((s) => s.addToast);

  const { register, handleSubmit, reset, setValue, formState: { errors } } = useForm<ResourceCreate>();

  const fetchResources = useCallback(async () => {
    setLoading(true);
    try {
      const res = await resourcesApi.list();
      setResources(res.data || []);
    } catch { addToast('Failed to load resources', 'error'); }
    finally { setLoading(false); }
  }, [addToast]);

  useEffect(() => { fetchResources(); }, [fetchResources]);

  const openCreate = () => {
    setEditingResource(null);
    reset({ name: '', type: 'Global' });
    setLinkedItemsInput('');
    setModalOpen(true);
  };

  const openEdit = (r: Resource) => {
    setEditingResource(r);
    setValue('name', r.name);
    setValue('type', r.type);
    setLinkedItemsInput((r.linked_items || []).join(', '));
    setModalOpen(true);
  };

  const onSubmit = async (data: ResourceCreate) => {
    setSubmitting(true);
    const payload: ResourceCreate = {
      ...data,
      linked_items: linkedItemsInput ? linkedItemsInput.split(',').map((s) => s.trim()).filter(Boolean) : undefined,
    };
    try {
      if (editingResource) {
        await resourcesApi.update(editingResource.id, payload);
        addToast('Resource updated', 'success');
      } else {
        await resourcesApi.create(payload);
        addToast('Resource created', 'success');
      }
      setModalOpen(false);
      fetchResources();
    } catch { addToast('Failed to save resource', 'error'); }
    finally { setSubmitting(false); }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    setSubmitting(true);
    try {
      await resourcesApi.delete(deleteTarget.id);
      addToast('Resource deleted', 'success');
      setDeleteTarget(null);
      fetchResources();
    } catch { addToast('Failed to delete resource', 'error'); }
    finally { setSubmitting(false); }
  };

  const columns = [
    { key: 'name', header: 'Name', render: (r: Resource) => <span className="font-medium">{r.name}</span> },
    { key: 'type', header: 'Type', render: (r: Resource) => (
      <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${r.type === 'Global' ? 'bg-green-100 text-green-700' : r.type === 'Project' ? 'bg-blue-100 text-blue-700' : 'bg-purple-100 text-purple-700'}`}>
        {r.type}
      </span>
    )},
    { key: 'linked_items', header: 'Linked Items', render: (r: Resource) => (r.linked_items || []).length > 0 ? `${(r.linked_items || []).length} item(s)` : '—' },
    { key: 'created_at', header: 'Created', render: (r: Resource) => new Date(r.created_at).toLocaleDateString() },
    { key: 'actions', header: 'Actions', render: (r: Resource) => (
      <div className="flex gap-2">
        <button onClick={(e) => { e.stopPropagation(); openEdit(r); }} className="text-blue-600 hover:text-blue-800" e2e-test-id={`resource-edit-btn-${r.id}`}><Pencil size={16} /></button>
        <button onClick={(e) => { e.stopPropagation(); setDeleteTarget(r); }} className="text-red-600 hover:text-red-800" e2e-test-id={`resource-delete-btn-${r.id}`}><Trash2 size={16} /></button>
      </div>
    )},
  ];

  return (
    <div e2e-test-id="resources-page">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2" e2e-test-id="resources-title"><Users size={24} /> Resources</h1>
          <p className="text-gray-500 text-sm mt-1">{resources.length} resource(s)</p>
        </div>
        <Button onClick={openCreate} e2e-test-id="resource-create-btn"><Plus size={16} /> New Resource</Button>
      </div>

      <Card e2e-test-id="resources-card">
        <CardBody className="p-0">
          {loading ? <div className="p-6"><TableSkeleton /></div> : (
            <Table columns={columns} data={resources} keyField="id" e2e-test-id="resources-table" />
          )}
        </CardBody>
      </Card>

      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title={editingResource ? 'Edit Resource' : 'Create Resource'} e2eTestId="resource-modal">
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" e2e-test-id="resource-form">
          <Input label="Name" e2e-test-id="resource-name-input" {...register('name', { required: 'Name is required' })} error={errors.name?.message} />
          <Select
            label="Type"
            e2e-test-id="resource-type-select"
            options={[{ value: 'Global', label: 'Global' }, { value: 'Project', label: 'Project' }, { value: 'Task', label: 'Task' }]}
            {...register('type', { required: 'Type is required' })}
            error={errors.type?.message}
          />
          <Input
            label="Linked Items (comma-separated UUIDs)"
            placeholder="uuid1, uuid2"
            value={linkedItemsInput}
            onChange={(e) => setLinkedItemsInput(e.target.value)}
            e2e-test-id="resource-linked-items-input"
          />
          <div className="flex justify-end gap-3 pt-2">
            <Button variant="secondary" type="button" onClick={() => setModalOpen(false)} e2e-test-id="resource-cancel-btn">Cancel</Button>
            <Button type="submit" loading={submitting} e2e-test-id="resource-submit-btn">{editingResource ? 'Update' : 'Create'}</Button>
          </div>
        </form>
      </Modal>

      <ConfirmDialog
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={handleDelete}
        title="Delete Resource"
        message={`Are you sure you want to delete "${deleteTarget?.name}"?`}
        loading={submitting}
        e2eTestId="resource-delete-dialog"
      />
    </div>
  );
}
