'use client';
import React from 'react';

interface CardProps {
  children: React.ReactNode;
  className?: string;
  'e2e-test-id'?: string;
}

export default function Card({ children, className = '', 'e2e-test-id': e2eId }: CardProps) {
  return (
    <div className={`bg-white rounded-lg border border-gray-200 shadow-sm ${className}`} e2e-test-id={e2eId}>
      {children}
    </div>
  );
}

export function CardHeader({ children, className = '' }: { children: React.ReactNode; className?: string }) {
  return <div className={`px-6 py-4 border-b border-gray-200 ${className}`}>{children}</div>;
}

export function CardBody({ children, className = '' }: { children: React.ReactNode; className?: string }) {
  return <div className={`px-6 py-4 ${className}`}>{children}</div>;
}
