'use client';
import React from 'react';
import Modal from './Modal';
import Button from './Button';

interface ConfirmDialogProps {
  open: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  message: string;
  loading?: boolean;
  e2eTestId?: string;
}

export default function ConfirmDialog({ open, onClose, onConfirm, title, message, loading, e2eTestId }: ConfirmDialogProps) {
  return (
    <Modal open={open} onClose={onClose} title={title} e2eTestId={e2eTestId || 'confirm-dialog'}>
      <p className="text-gray-600 mb-6" e2e-test-id={`${e2eTestId || 'confirm-dialog'}-message`}>{message}</p>
      <div className="flex justify-end gap-3">
        <Button variant="secondary" onClick={onClose} e2e-test-id={`${e2eTestId || 'confirm-dialog'}-cancel-btn`}>Cancel</Button>
        <Button variant="danger" onClick={onConfirm} loading={loading} e2e-test-id={`${e2eTestId || 'confirm-dialog'}-confirm-btn`}>Delete</Button>
      </div>
    </Modal>
  );
}
