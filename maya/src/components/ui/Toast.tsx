'use client';
import React from 'react';
import { useToastStore } from '@/store/toastStore';
import { X, CheckCircle, AlertCircle, Info } from 'lucide-react';

const icons = {
  success: <CheckCircle size={18} className="text-green-500" />,
  error: <AlertCircle size={18} className="text-red-500" />,
  info: <Info size={18} className="text-blue-500" />,
};

const bgColors = {
  success: 'bg-green-50 border-green-200',
  error: 'bg-red-50 border-red-200',
  info: 'bg-blue-50 border-blue-200',
};

export default function ToastContainer() {
  const { toasts, removeToast } = useToastStore();

  return (
    <div className="fixed top-4 right-4 z-[100] space-y-2" e2e-test-id="toast-container">
      {toasts.map((toast) => (
        <div key={toast.id} className={`flex items-center gap-3 px-4 py-3 rounded-lg border shadow-md min-w-[300px] ${bgColors[toast.type]}`} e2e-test-id={`toast-${toast.id}`}>
          {icons[toast.type]}
          <span className="text-sm text-gray-800 flex-1">{toast.message}</span>
          <button onClick={() => removeToast(toast.id)} className="text-gray-400 hover:text-gray-600" e2e-test-id={`toast-close-${toast.id}`}>
            <X size={14} />
          </button>
        </div>
      ))}
    </div>
  );
}
