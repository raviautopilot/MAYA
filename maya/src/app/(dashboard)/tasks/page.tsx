import React, { useState, useEffect, useCallback, Suspense } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useForm, useFieldArray } from 'react-hook-form';
import { tasksApi, boardsApi, schedulersApi, resourcesApi } from '@/services/api';
import { useToastStore } from '@/store/toastStore';
import { Button, Input, Select, Modal, ConfirmDialog, Card, CardBody } from '@/components/ui';
import { CardSkeleton } from '@/components/ui/Skeleton';
import { Plus, Pencil, Trash2, CheckSquare, List, LayoutGrid, Bell, X, Calendar, Clock } from 'lucide-react';
import type { Task, TaskCreate, Board, Scheduler, Resource, Reminder } from '@/types';

type FormData = TaskCreate & { reminders?: Reminder[]; actual_time_minutes?: number };

const priorityColors: Record<string, string> = {
  Low: 'bg-gray-100 text-gray-700',
  Medium: 'bg-yellow-100 text-yellow-700',
  High: 'bg-orange-100 text-orange-700',
  Critical: 'bg-red-100 text-red-700',
};

const formatForDateTimeLocal = (dateStr?: string) => {
  if (!dateStr) return '';
  try {
    const date = new Date(dateStr);
    if (isNaN(date.getTime())) return '';
    const pad = (n: number) => n.toString().padStart(2, '0');
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
  } catch {
    return '';
  }
};

export default function TasksPage() {
  return (
    <Suspense fallback={<div className="p-6"><CardSkeleton /></div>}>
      <TasksContent />
    </Suspense>
  );
}

function TasksContent() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const boardIdParam = searchParams.get('board_id') || '';

  const [tasks, setTasks] = useState<Task[]>([]);
  const [boards, setBoards] = useState<Board[]>([]);
  const [schedulers, setSchedulers] = useState<Scheduler[]>([]);
  const [resources, setResources] = useState<Resource[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Task | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [viewMode, setViewMode] = useState<'kanban' | 'table'>('kanban');
  const [filterBoardId, setFilterBoardId] = useState(boardIdParam);
  const [filterPriority, setFilterPriority] = useState('');
  const [activeDragLane, setActiveDragLane] = useState<string | null>(null);
  const addToast = useToastStore((s) => s.addToast);

  const handleDragStart = (e: React.DragEvent, taskId: string) => {
    e.dataTransfer.setData('text/plain', taskId);
    e.dataTransfer.effectAllowed = 'move';
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
  };

  const handleDragEnter = (e: React.DragEvent, lane: string) => {
    e.preventDefault();
    setActiveDragLane(lane);
  };

  const handleDrop = async (e: React.DragEvent, targetLane: string) => {
    e.preventDefault();
    setActiveDragLane(null);
    const taskId = e.dataTransfer.getData('text/plain');
    if (taskId) {
      const task = tasks.find((t) => t.id === taskId);
      if (task && task.swimlane !== targetLane) {
        await handlePatchSwimlane(taskId, targetLane);
      }
    }
  };

  useEffect(() => {
    setFilterBoardId(boardIdParam);
  }, [boardIdParam]);

  const handleBoardFilterChange = (boardId: string) => {
    const newParams = new URLSearchParams(searchParams);
    if (boardId) {
      newParams.set('board_id', boardId);
    } else {
      newParams.delete('board_id');
    }
    setSearchParams(newParams);
  };

  const { register, handleSubmit, reset, setValue, watch, control, formState: { errors } } = useForm<FormData>({
    defaultValues: { reminders: [] },
  });
  const { fields: reminderFields, append: appendReminder, remove: removeReminder } = useFieldArray({ control, name: 'reminders' });

  const selectedBoardId = watch('board_id');
  const selectedBoard = boards.find((b) => b.id === selectedBoardId);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, string> = {};
      if (filterBoardId) params.board_id = filterBoardId;
      if (filterPriority) params.priority = filterPriority;
      const [tRes, bRes, sRes, rRes] = await Promise.all([
        tasksApi.list(Object.keys(params).length > 0 ? params : undefined),
        boardsApi.list(),
        schedulersApi.list(),
        resourcesApi.list(),
      ]);
      setTasks(tRes.data || []);
      setBoards(bRes.data || []);
      setSchedulers(sRes.data || []);
      setResources(rRes.data || []);
    } catch { addToast('Failed to load data', 'error'); }
    finally { setLoading(false); }
  }, [addToast, filterBoardId, filterPriority]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const openCreate = () => {
    setEditingTask(null);
    reset({ board_id: filterBoardId || '', swimlane: '', task_type: '', title: '', priority: 'Medium', description: '', reminders: [], due_date: '' });
    setModalOpen(true);
  };

  const openEdit = (t: Task) => {
    setEditingTask(t);
    setValue('board_id', t.board_id);
    setValue('swimlane', t.swimlane);
    setValue('task_type', t.task_type);
    setValue('title', t.title);
    setValue('description', t.description || '');
    setValue('priority', t.priority);
    setValue('assignee_id', t.assignee_id || '');
    setValue('estimation_minutes', t.estimation_minutes || 0);
    setValue('actual_time_minutes', t.actual_time_minutes || 0);
    setValue('cost', t.cost || 0);
    setValue('scheduler_id', t.scheduler_id || '');
    setValue('due_date', formatForDateTimeLocal(t.due_date));
    setValue('reminders', t.reminders || []);
    setModalOpen(true);
  };

  const onSubmit = async (data: FormData) => {
    setSubmitting(true);
    try {
      const payload = { ...data };
      if (!payload.assignee_id) delete payload.assignee_id;
      if (!payload.scheduler_id) delete payload.scheduler_id;
      if (payload.due_date) {
        const d = new Date(payload.due_date);
        if (!isNaN(d.getTime())) {
          payload.due_date = d.toISOString();
        } else {
          payload.due_date = '';
        }
      } else {
        payload.due_date = '';
      }
      if (editingTask) {
        await tasksApi.update(editingTask.id, payload);
        addToast('Task updated', 'success');
      } else {
        await tasksApi.create(payload);
        addToast('Task created', 'success');
      }
      setModalOpen(false);
      fetchData();
    } catch { addToast('Failed to save task', 'error'); }
    finally { setSubmitting(false); }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    setSubmitting(true);
    try {
      await tasksApi.delete(deleteTarget.id);
      addToast('Task deleted', 'success');
      setDeleteTarget(null);
      fetchData();
    } catch (err: unknown) {
      const errMsg = (err as { response?: { data?: { error?: string } } }).response?.data?.error || 'Failed to delete task';
      addToast(errMsg, 'error');
    }
    finally { setSubmitting(false); }
  };

  const handlePatchSwimlane = async (taskId: string, swimlane: string) => {
    try {
      await tasksApi.patch(taskId, { swimlane });
      addToast(`Moved to ${swimlane}`, 'success');
      fetchData();
    } catch { addToast('Failed to move task', 'error'); }
  };

  // Get unique swimlanes from all boards or filtered board
  const allSwimlanes = filterBoardId
    ? boards.find((b) => b.id === filterBoardId)?.swimlanes || []
    : Array.from(new Set(boards.flatMap((b) => b.swimlanes || [])));

  const kanbanSwimlanes = allSwimlanes.length > 0 ? allSwimlanes : ['To Do', 'In Progress', 'Done'];

  return (
    <div e2e-test-id="tasks-page">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2" e2e-test-id="tasks-title"><CheckSquare size={24} /> Tasks</h1>
          <p className="text-gray-500 text-sm mt-1">{tasks.length} task(s)</p>
        </div>
        <div className="flex items-center gap-3">
          <select value={filterBoardId} onChange={(e) => handleBoardFilterChange(e.target.value)} className="rounded-md border border-gray-300 px-3 py-2 text-sm" e2e-test-id="tasks-filter-board">
            <option value="">All Boards</option>
            {boards.map((b) => <option key={b.id} value={b.id}>{b.name}</option>)}
          </select>
          <select value={filterPriority} onChange={(e) => setFilterPriority(e.target.value)} className="rounded-md border border-gray-300 px-3 py-2 text-sm" e2e-test-id="tasks-filter-priority">
            <option value="">All Priorities</option>
            {['Low', 'Medium', 'High', 'Critical'].map((p) => <option key={p} value={p}>{p}</option>)}
          </select>
          <div className="flex border rounded-md overflow-hidden">
            <button onClick={() => setViewMode('kanban')} className={`p-2 ${viewMode === 'kanban' ? 'bg-blue-600 text-white' : 'bg-white text-gray-600'}`} e2e-test-id="tasks-view-kanban"><LayoutGrid size={16} /></button>
            <button onClick={() => setViewMode('table')} className={`p-2 ${viewMode === 'table' ? 'bg-blue-600 text-white' : 'bg-white text-gray-600'}`} e2e-test-id="tasks-view-table"><List size={16} /></button>
          </div>
          <Button onClick={openCreate} e2e-test-id="task-create-btn"><Plus size={16} /> New Task</Button>
        </div>
      </div>

      {loading ? (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {[1, 2, 3].map((i) => <CardSkeleton key={i} />)}
        </div>
      ) : viewMode === 'kanban' ? (
        /* ── Kanban View ── */
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4" e2e-test-id="tasks-kanban">
          {kanbanSwimlanes.map((lane) => (
            <div
              key={lane}
              className={`rounded-lg p-4 transition-all duration-200 ${
                activeDragLane === lane
                  ? 'bg-blue-50 border-2 border-dashed border-blue-300 shadow-inner'
                  : 'bg-gray-100 border-2 border-transparent'
              }`}
              onDragOver={handleDragOver}
              onDragEnter={(e) => handleDragEnter(e, lane)}
              onDragLeave={() => setActiveDragLane(null)}
              onDrop={(e) => handleDrop(e, lane)}
              e2e-test-id={`kanban-lane-${lane.replace(/\s/g, '-').toLowerCase()}`}
            >
              <h3 className="font-semibold text-gray-700 mb-3 text-sm uppercase">{lane} ({tasks.filter((t) => t.swimlane === lane).length})</h3>
              <div className="space-y-2 min-h-[150px]">
                {tasks.filter((t) => t.swimlane === lane).map((task) => {
                  const linkedScheduler = schedulers.find((s) => s.id === task.scheduler_id);
                  return (
                    <div
                      key={task.id}
                      draggable
                      onDragStart={(e) => handleDragStart(e, task.id)}
                      className="cursor-grab active:cursor-grabbing hover:scale-[1.01] transition-transform duration-150"
                    >
                      <Card className="hover:shadow-md transition" e2e-test-id={`task-card-${task.id}`}>
                        <CardBody className="p-3">
                          <div className="flex items-start justify-between">
                            <h4 className="text-sm font-medium text-gray-900">{task.title}</h4>
                            <div className="flex gap-1">
                              <button onClick={(e) => { e.stopPropagation(); openEdit(task); }} className="text-blue-500 hover:text-blue-700" e2e-test-id={`task-edit-btn-${task.id}`}><Pencil size={14} /></button>
                              <button onClick={(e) => { e.stopPropagation(); setDeleteTarget(task); }} className="text-red-500 hover:text-red-700" e2e-test-id={`task-delete-btn-${task.id}`}><Trash2 size={14} /></button>
                            </div>
                          </div>
                          {task.description && <p className="text-xs text-gray-500 mt-1 line-clamp-2">{task.description}</p>}
                          <div className="flex items-center gap-2 mt-2 flex-wrap">
                            <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${priorityColors[task.priority] || ''}`}>{task.priority}</span>
                            {task.reminders && task.reminders.length > 0 && <Bell size={12} className="text-yellow-500" />}
                          </div>
                          {task.due_date && (
                            <div className="flex items-center gap-1.5 text-xs text-gray-500 mt-2 bg-gray-50 p-1.5 rounded-md border border-gray-100">
                              <Calendar size={12} className="text-gray-400" />
                              <span>Due: {new Date(task.due_date).toLocaleString()}</span>
                            </div>
                          )}
                          {linkedScheduler && (
                            <div className="flex items-center gap-1.5 text-xs text-indigo-700 mt-1 bg-indigo-50 p-1.5 rounded-md border border-indigo-100">
                              <Clock size={12} className="text-indigo-400" />
                              <span>Recurring: {linkedScheduler.name}</span>
                            </div>
                          )}
                          {/* Move buttons */}
                          <div className="flex gap-1 mt-2">
                            {kanbanSwimlanes.filter((l) => l !== lane).map((l) => (
                              <button key={l} onClick={(e) => { e.stopPropagation(); handlePatchSwimlane(task.id, l); }} className="text-xs px-2 py-0.5 bg-gray-200 rounded hover:bg-gray-300" e2e-test-id={`task-move-${task.id}-${l.replace(/\s/g, '-').toLowerCase()}`}>
                                → {l}
                              </button>
                            ))}
                          </div>
                        </CardBody>
                      </Card>
                    </div>
                  );
                })}
              </div>
            </div>
          ))}
        </div>
      ) : (
        /* ── Table View ── */
        <Card e2e-test-id="tasks-table-card">
          <CardBody className="p-0 overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200" e2e-test-id="tasks-table">
              <thead className="bg-gray-50">
                <tr>
                  {['Title', 'Swimlane', 'Priority', 'Due Date', 'Scheduler', 'Board', 'Actions'].map((h) => (
                    <th key={h} className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {tasks.map((task) => (
                  <tr key={task.id} className="hover:bg-gray-50" e2e-test-id={`task-row-${task.id}`}>
                    <td className="px-6 py-4 text-sm font-medium">{task.title}</td>
                    <td className="px-6 py-4 text-sm">{task.swimlane}</td>
                    <td className="px-6 py-4"><span className={`px-2 py-0.5 rounded-full text-xs font-medium ${priorityColors[task.priority]}`}>{task.priority}</span></td>
                    <td className="px-6 py-4 text-sm text-gray-500">{task.due_date ? new Date(task.due_date).toLocaleString() : '—'}</td>
                    <td className="px-6 py-4 text-sm text-gray-500">{schedulers.find((s) => s.id === task.scheduler_id)?.name || '—'}</td>
                    <td className="px-6 py-4 text-sm">{boards.find((b) => b.id === task.board_id)?.name || task.board_id}</td>
                    <td className="px-6 py-4">
                      <div className="flex gap-2">
                        <button onClick={() => openEdit(task)} className="text-blue-600 hover:text-blue-800" e2e-test-id={`task-table-edit-btn-${task.id}`}><Pencil size={16} /></button>
                        <button onClick={() => setDeleteTarget(task)} className="text-red-600 hover:text-red-800" e2e-test-id={`task-table-delete-btn-${task.id}`}><Trash2 size={16} /></button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </CardBody>
        </Card>
      )}

      {/* Create/Edit Modal */}
      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title={editingTask ? 'Edit Task' : 'Create Task'} e2eTestId="task-modal">
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" e2e-test-id="task-form">
          <Input label="Title" e2e-test-id="task-title-input" {...register('title', { required: 'Title is required' })} error={errors.title?.message} />
          <Input label="Description" e2e-test-id="task-description-input" {...register('description')} />
          <Select
            label="Board"
            e2e-test-id="task-board-select"
            options={boards.map((b) => ({ value: b.id, label: b.name }))}
            {...register('board_id', { required: 'Board is required' })}
            error={errors.board_id?.message}
          />
          <Select
            label="Swimlane"
            e2e-test-id="task-swimlane-select"
            options={(selectedBoard?.swimlanes || ['To Do', 'In Progress', 'Done']).map((s) => ({ value: s, label: s }))}
            {...register('swimlane', { required: 'Swimlane is required' })}
            error={errors.swimlane?.message}
          />
          <Select
            label="Task Type"
            e2e-test-id="task-type-select"
            options={(selectedBoard?.task_types || ['Bug', 'Feature', 'Chore']).map((t) => ({ value: t, label: t }))}
            {...register('task_type', { required: 'Task type is required' })}
            error={errors.task_type?.message}
          />
          <Select
            label="Priority"
            e2e-test-id="task-priority-select"
            options={['Low', 'Medium', 'High', 'Critical'].map((p) => ({ value: p, label: p }))}
            {...register('priority', { required: 'Priority is required' })}
            error={errors.priority?.message}
          />
          <Select
            label="Assignee (Resource)"
            e2e-test-id="task-assignee-select"
            options={resources.map((r) => ({ value: r.id, label: r.name }))}
            {...register('assignee_id')}
          />
          <div className="grid grid-cols-2 gap-4">
            <Input label="Estimation (min)" type="number" e2e-test-id="task-estimation-input" {...register('estimation_minutes', { valueAsNumber: true })} />
            {editingTask && <Input label="Actual Time (min)" type="number" e2e-test-id="task-actual-time-input" {...register('actual_time_minutes', { valueAsNumber: true })} />}
          </div>
          <Input label="Cost" type="number" step="0.01" e2e-test-id="task-cost-input" {...register('cost', { valueAsNumber: true })} />
          <Select
            label="Scheduler"
            e2e-test-id="task-scheduler-select"
            options={schedulers.map((s) => ({ value: s.id, label: s.name }))}
            {...register('scheduler_id')}
          />
          <Input
            label="Due Date"
            type="datetime-local"
            e2e-test-id="task-due-date-input"
            {...register('due_date')}
          />

          {/* Reminders */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <label className="text-sm font-medium text-gray-700">Reminders (max 5)</label>
              {reminderFields.length < 5 && (
                <button type="button" onClick={() => appendReminder({ time: '', note: '' })} className="text-blue-600 text-sm flex items-center gap-1" e2e-test-id="task-add-reminder-btn">
                  <Plus size={14} /> Add
                </button>
              )}
            </div>
            {reminderFields.map((field, idx) => (
              <div key={field.id} className="flex gap-2 mb-2 items-end">
                <Input label="Time (ISO)" type="datetime-local" e2e-test-id={`task-reminder-time-${idx}`} {...register(`reminders.${idx}.time`, { required: 'Time required' })} />
                <Input label="Note" e2e-test-id={`task-reminder-note-${idx}`} {...register(`reminders.${idx}.note`)} />
                <button type="button" onClick={() => removeReminder(idx)} className="text-red-500 pb-2" e2e-test-id={`task-reminder-remove-${idx}`}><X size={16} /></button>
              </div>
            ))}
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <Button variant="secondary" type="button" onClick={() => setModalOpen(false)} e2e-test-id="task-cancel-btn">Cancel</Button>
            <Button type="submit" loading={submitting} e2e-test-id="task-submit-btn">{editingTask ? 'Update' : 'Create'}</Button>
          </div>
        </form>
      </Modal>

      <ConfirmDialog
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={handleDelete}
        title="Delete Task"
        message={`Are you sure you want to delete "${deleteTarget?.title}"?`}
        loading={submitting}
        e2eTestId="task-delete-dialog"
      />
    </div>
  );
}
