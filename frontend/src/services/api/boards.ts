import apiClient from './client';
import { APIResponse, Board, BoardCreate, BoardUpdate } from '@/types';

export const boardsApi = {
  list: (projectId?: string) =>
    apiClient.get<APIResponse<Board[]>>('/boards', { params: projectId ? { project_id: projectId } : {} }).then((r) => r.data),

  get: (id: string) =>
    apiClient.get<APIResponse<Board>>(`/boards/${id}`).then((r) => r.data),

  create: (data: BoardCreate) =>
    apiClient.post<APIResponse<Board>>('/boards', data).then((r) => r.data),

  update: (id: string, data: BoardUpdate) =>
    apiClient.put<APIResponse<Board>>(`/boards/${id}`, data).then((r) => r.data),

  delete: (id: string) =>
    apiClient.delete<APIResponse<{ message: string }>>(`/boards/${id}`).then((r) => r.data),
};
