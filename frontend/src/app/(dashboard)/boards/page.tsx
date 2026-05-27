'use client';
import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { useForm } from 'react-hook-form';
import { boardsApi, projectsApi } from '@/services/api';
import { useToastStore } from '@/store/toastStore';
import { Button, Input, Select, Modal, Table, ConfirmDialog, Card, CardBody } from '@/components/ui';
import { TableSkeleton } from '@/components/ui/Skeleton';
import { Plus, Pencil, Trash2, Columns3 } from 'lucide-react';
import type { Board, BoardCreate, Project } from '@/types';

export default function BoardsPage() {
  const [boards, setBoards] = useState<Board[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingBoard, setEditingBoard] = useState<Board | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Board | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [filterProjectId, setFilterProjectId] = useState('');
  const [swimlanesInput, setSwimlanesInput] = useState('');
  const [taskTypesInput, setTaskTypesInput] = useState('');
  const addToast = useToastStore((s) => s.addToast);

  const { register, handleSubmit, reset, setValue, formState: { errors } } = useForm<BoardCreate>();

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [bRes, pRes] = await Promise.all([
        boardsApi.list(filterProjectId || undefined),
        projectsApi.list(),
      ]);
      setBoards(bRes.data || []);
      setProjects(pRes.data || []);
    } catch { addToast('Failed to load data', 'error'); }
    finally { setLoading(false); }
  }, [addToast, filterProjectId]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const projectName = (pid: string) => projects.find((p) => p.id === pid)?.name || pid;

  const openCreate = () => {
    setEditingBoard(null);
    reset({ name: '', project_id: '' });
    setSwimlanesInput('');
    setTaskTypesInput('');
    setModalOpen(true);
  };

  const openEdit = (b: Board) => {
    setEditingBoard(b);
    setValue('name', b.name);
    setValue('project_id', b.project_id);
    setSwimlanesInput((b.swimlanes || []).join(', '));
    setTaskTypesInput((b.task_types || []).join(', '));
    setModalOpen(true);
  };

  const onSubmit = async (data: BoardCreate) => {
    setSubmitting(true);
    const payload: BoardCreate = {
      ...data,
      swimlanes: swimlanesInput ? swimlanesInput.split(',').map((s) => s.trim()).filter(Boolean) : undefined,
      task_types: taskTypesInput ? taskTypesInput.split(',').map((s) => s.trim()).filter(Boolean) : undefined,
    };
    try {
      if (editingBoard) {
        await boardsApi.update(editingBoard.id, payload);
        addToast('Board updated', 'success');
      } else {
        await boardsApi.create(payload);
        addToast('Board created', 'success');
      }
      setModalOpen(false);
      fetchData();
    } catch { addToast('Failed to save board', 'error'); }
    finally { setSubmitting(false); }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    setSubmitting(true);
    try {
      await boardsApi.delete(deleteTarget.id);
      addToast('Board deleted', 'success');
      setDeleteTarget(null);
      fetchData();
    } catch { addToast('Failed to delete board', 'error'); }
    finally { setSubmitting(false); }
  };

  const columns = [
    { key: 'name', header: 'Name', render: (b: Board) => (
      <Link href={`/tasks?board_id=${b.id}`} className="font-medium text-blue-600 hover:text-blue-800 hover:underline" e2e-test-id={`board-name-link-${b.id}`}>
        {b.name}
      </Link>
    ) },
    { key: 'project_id', header: 'Project', render: (b: Board) => projectName(b.project_id) },
    { key: 'swimlanes', header: 'Swimlanes', render: (b: Board) => (
      <div className="flex flex-wrap gap-1">{(b.swimlanes || []).map((s) => (
        <span key={s} className="inline-flex px-2 py-0.5 rounded-full text-xs bg-gray-100 text-gray-600">{s}</span>
      ))}</div>
    )},
    { key: 'task_types', header: 'Task Types', render: (b: Board) => (
      <div className="flex flex-wrap gap-1">{(b.task_types || []).map((t) => (
        <span key={t} className="inline-flex px-2 py-0.5 rounded-full text-xs bg-purple-100 text-purple-600">{t}</span>
      ))}</div>
    )},
    { key: 'actions', header: 'Actions', render: (b: Board) => (
      <div className="flex gap-2">
        <button onClick={(e) => { e.stopPropagation(); openEdit(b); }} className="text-blue-600 hover:text-blue-800" e2e-test-id={`board-edit-btn-${b.id}`}><Pencil size={16} /></button>
        <button onClick={(e) => { e.stopPropagation(); setDeleteTarget(b); }} className="text-red-600 hover:text-red-800" e2e-test-id={`board-delete-btn-${b.id}`}><Trash2 size={16} /></button>
      </div>
    )},
  ];

  return (
    <div e2e-test-id="boards-page">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2" e2e-test-id="boards-title"><Columns3 size={24} /> Boards</h1>
          <p className="text-gray-500 text-sm mt-1">{boards.length} board(s)</p>
        </div>
        <div className="flex items-center gap-3">
          <select
            value={filterProjectId}
            onChange={(e) => setFilterProjectId(e.target.value)}
            className="rounded-md border border-gray-300 px-3 py-2 text-sm"
            e2e-test-id="boards-filter-project"
          >
            <option value="">All Projects</option>
            {projects.map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}
          </select>
          <Button onClick={openCreate} e2e-test-id="board-create-btn"><Plus size={16} /> New Board</Button>
        </div>
      </div>

      <Card e2e-test-id="boards-card">
        <CardBody className="p-0">
          {loading ? <div className="p-6"><TableSkeleton /></div> : (
            <Table columns={columns} data={boards} keyField="id" e2e-test-id="boards-table" />
          )}
        </CardBody>
      </Card>

      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title={editingBoard ? 'Edit Board' : 'Create Board'} e2eTestId="board-modal">
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" e2e-test-id="board-form">
          <Input label="Name" e2e-test-id="board-name-input" {...register('name', { required: 'Name is required' })} error={errors.name?.message} />
          <Select
            label="Project"
            e2e-test-id="board-project-select"
            options={projects.map((p) => ({ value: p.id, label: p.name }))}
            {...register('project_id', { required: 'Project is required' })}
            error={errors.project_id?.message}
          />
          <Input
            label="Swimlanes (comma-separated)"
            placeholder="To Do, In Progress, Done"
            value={swimlanesInput}
            onChange={(e) => setSwimlanesInput(e.target.value)}
            e2e-test-id="board-swimlanes-input"
          />
          <Input
            label="Task Types (comma-separated)"
            placeholder="Bug, Feature, Chore"
            value={taskTypesInput}
            onChange={(e) => setTaskTypesInput(e.target.value)}
            e2e-test-id="board-task-types-input"
          />
          <div className="flex justify-end gap-3 pt-2">
            <Button variant="secondary" type="button" onClick={() => setModalOpen(false)} e2e-test-id="board-cancel-btn">Cancel</Button>
            <Button type="submit" loading={submitting} e2e-test-id="board-submit-btn">{editingBoard ? 'Update' : 'Create'}</Button>
          </div>
        </form>
      </Modal>

      <ConfirmDialog
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={handleDelete}
        title="Delete Board"
        message={`Are you sure you want to delete "${deleteTarget?.name}"?`}
        loading={submitting}
        e2eTestId="board-delete-dialog"
      />
    </div>
  );
}
