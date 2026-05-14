import apiClient from './client';
import { APIResponse, LoginRequest, LoginResponse, ChangePasswordRequest } from '@/types';

export const authApi = {
  login: (data: LoginRequest) =>
    apiClient.post<APIResponse<LoginResponse>>('/auth/login', data).then((r) => r.data),

  changePassword: (data: ChangePasswordRequest) =>
    apiClient.post<APIResponse<{ message: string }>>('/auth/change-password', data).then((r) => r.data),

  googleLogin: () => {
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';
    window.location.href = `${apiUrl}/auth/google/login`;
  },
};
