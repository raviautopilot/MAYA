'use client';
import React, { useState, useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { useAuthStore } from '@/store/authStore';
import { authApi } from '@/services/api';
import { useToastStore } from '@/store/toastStore';
import { Button, Input, ToastContainer } from '@/components/ui';
import { LogIn } from 'lucide-react';
import type { LoginRequest } from '@/types';

export default function LoginPage() {
  return (
    <Suspense fallback={<div className="min-h-screen flex items-center justify-center bg-gray-50"><div className="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full" /></div>}>
      <LoginContent />
    </Suspense>
  );
}

function LoginContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { setAuth, hydrate, isAuthenticated } = useAuthStore();
  const addToast = useToastStore((s) => s.addToast);
  const [loading, setLoading] = useState(false);

  const { register, handleSubmit, formState: { errors } } = useForm<LoginRequest>();

  useEffect(() => {
    hydrate();
  }, [hydrate]);

  useEffect(() => {
    if (isAuthenticated) router.replace('/dashboard');
  }, [isAuthenticated, router]);

  // Handle Google OAuth callback token
  useEffect(() => {
    const token = searchParams.get('token');
    const email = searchParams.get('email');
    const name = searchParams.get('name');
    if (token) {
      setAuth(token, { email: email || 'user@google.com', name: name || undefined });
      addToast('Logged in with Google!', 'success');
      router.replace('/dashboard');
    }
  }, [searchParams, setAuth, addToast, router]);

  const onSubmit = async (data: LoginRequest) => {
    setLoading(true);
    try {
      const res = await authApi.login(data);
      if (res.data?.token) {
        setAuth(res.data.token, { email: data.email });
        addToast('Login successful!', 'success');
        router.push('/dashboard');
      } else {
        addToast(res.error || 'Login failed', 'error');
      }
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error || 'Login failed';
      addToast(msg, 'error');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50" e2e-test-id="login-page">
      <ToastContainer />
      <div className="w-full max-w-md bg-white rounded-lg shadow-md p-8" e2e-test-id="login-card">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-gray-900" e2e-test-id="login-title">MyKanban</h1>
          <p className="text-gray-500 mt-1">Sign in to your account</p>
        </div>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" e2e-test-id="login-form">
          <Input
            label="Email"
            type="email"
            placeholder="admin@mykanban.local"
            e2e-test-id="login-email-input"
            {...register('email', { required: 'Email is required' })}
            error={errors.email?.message}
          />
          <Input
            label="Password"
            type="password"
            placeholder="••••••••"
            e2e-test-id="login-password-input"
            {...register('password', { required: 'Password is required' })}
            error={errors.password?.message}
          />
          <Button type="submit" loading={loading} className="w-full" e2e-test-id="login-submit-btn">
            <LogIn size={16} /> Sign In
          </Button>
        </form>

        <div className="mt-6">
          <div className="relative">
            <div className="absolute inset-0 flex items-center"><div className="w-full border-t border-gray-200" /></div>
            <div className="relative flex justify-center text-sm"><span className="px-2 bg-white text-gray-400">or</span></div>
          </div>
          <button
            onClick={() => authApi.googleLogin()}
            className="mt-4 w-full flex items-center justify-center gap-2 px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 transition"
            e2e-test-id="login-google-btn"
          >
            <svg className="w-5 h-5" viewBox="0 0 24 24"><path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" fill="#4285F4"/><path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/><path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/><path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/></svg>
            Sign in with Google
          </button>
        </div>
      </div>
    </div>
  );
}
