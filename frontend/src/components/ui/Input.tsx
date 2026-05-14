'use client';
import React from 'react';

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  'e2e-test-id'?: string;
}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, className = '', 'e2e-test-id': e2eId, ...props }, ref) => (
    <div className="w-full">
      {label && (
        <label className="block text-sm font-medium text-gray-700 mb-1" e2e-test-id={e2eId ? `${e2eId}-label` : undefined}>
          {label}
        </label>
      )}
      <input
        ref={ref}
        className={`w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 ${error ? 'border-red-500' : ''} ${className}`}
        e2e-test-id={e2eId}
        {...props}
      />
      {error && <p className="mt-1 text-xs text-red-600" e2e-test-id={e2eId ? `${e2eId}-error` : undefined}>{error}</p>}
    </div>
  ),
);
Input.displayName = 'Input';
export default Input;
