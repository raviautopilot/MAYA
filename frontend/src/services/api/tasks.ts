import apiClient from './client';
import { APIResponse, Task, TaskCreate, TaskUpdate, TaskPatch } from '@/types';

export const tasksApi = {
  list: (params?: { board_id?: string; swimlane?: string; assignee_id?: string; priority?: string }) =>
    apiClient.get<APIResponse<Task[]>>('/tasks', { params }).then((r) => r.data),

  get: (id: string) =>
    apiClient.get<APIResponse<Task>>(`/tasks/${id}`).then((r) => r.data),

  create: (data: TaskCreate) =>
    apiClient.post<APIResponse<Task>>('/tasks', data).then((r) => r.data),

  update: (id: string, data: TaskUpdate) =>
    apiClient.put<APIResponse<Task>>(`/tasks/${id}`, data).then((r) => r.data),

  patch: (id: string, data: TaskPatch) =>
    apiClient.patch<APIResponse<Task>>(`/tasks/${id}`, data).then((r) => r.data),

  delete: (id: string) =>
    apiClient.delete<APIResponse<{ message: string }>>(`/tasks/${id}`).then((r) => r.data),
};
