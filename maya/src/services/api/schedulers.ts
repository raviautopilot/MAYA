import apiClient from './client';
import { APIResponse, Scheduler, SchedulerCreate, SchedulerUpdate } from '@/types';

export const schedulersApi = {
  list: () =>
    apiClient.get<APIResponse<Scheduler[]>>('/schedulers').then((r) => r.data),

  get: (id: string) =>
    apiClient.get<APIResponse<Scheduler>>(`/schedulers/${id}`).then((r) => r.data),

  create: (data: SchedulerCreate) =>
    apiClient.post<APIResponse<Scheduler>>('/schedulers', data).then((r) => r.data),

  update: (id: string, data: SchedulerUpdate) =>
    apiClient.put<APIResponse<Scheduler>>(`/schedulers/${id}`, data).then((r) => r.data),

  delete: (id: string) =>
    apiClient.delete<APIResponse<{ message: string }>>(`/schedulers/${id}`).then((r) => r.data),
};
