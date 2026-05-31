import apiClient from './client';
import { APIResponse, Resource, ResourceCreate, ResourceUpdate } from '@/types';

export const resourcesApi = {
  list: () =>
    apiClient.get<APIResponse<Resource[]>>('/resources').then((r) => r.data),

  get: (id: string) =>
    apiClient.get<APIResponse<Resource>>(`/resources/${id}`).then((r) => r.data),

  create: (data: ResourceCreate) =>
    apiClient.post<APIResponse<Resource>>('/resources', data).then((r) => r.data),

  update: (id: string, data: ResourceUpdate) =>
    apiClient.put<APIResponse<Resource>>(`/resources/${id}`, data).then((r) => r.data),

  delete: (id: string) =>
    apiClient.delete<APIResponse<{ message: string }>>(`/resources/${id}`).then((r) => r.data),
};
