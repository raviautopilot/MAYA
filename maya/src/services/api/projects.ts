import apiClient from './client';
import { APIResponse, Project, ProjectCreate, ProjectUpdate } from '@/types';

export const projectsApi = {
  list: () =>
    apiClient.get<APIResponse<Project[]>>('/projects').then((r) => r.data),

  get: (id: string) =>
    apiClient.get<APIResponse<Project>>(`/projects/${id}`).then((r) => r.data),

  create: (data: ProjectCreate) =>
    apiClient.post<APIResponse<Project>>('/projects', data).then((r) => r.data),

  update: (id: string, data: ProjectUpdate) =>
    apiClient.put<APIResponse<Project>>(`/projects/${id}`, data).then((r) => r.data),

  delete: (id: string) =>
    apiClient.delete<APIResponse<{ message: string }>>(`/projects/${id}`).then((r) => r.data),
};
