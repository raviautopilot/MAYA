import React from 'react';

export function Loader({ className = '' }: { className?: string }) {
  return (
    <div
      className={`animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full ${className}`}
      e2e-test-id="spinner"
    />
  );
}

export function FullPageLoader() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50" e2e-test-id="full-page-loader">
      <Loader />
    </div>
  );
}
