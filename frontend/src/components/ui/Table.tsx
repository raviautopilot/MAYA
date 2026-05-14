'use client';
import React from 'react';

/* eslint-disable @typescript-eslint/no-explicit-any */
interface Column {
  key: string;
  header: string;
  render?: (item: any) => React.ReactNode;
  className?: string;
}

interface TableProps {
  columns: Column[];
  data: any[];
  keyField: string;
  'e2e-test-id'?: string;
  onRowClick?: (item: any) => void;
}

export default function Table({ columns, data, keyField, 'e2e-test-id': e2eId, onRowClick }: TableProps) {
  return (
    <div className="overflow-x-auto" e2e-test-id={e2eId}>
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            {columns.map((col) => (
              <th key={col.key} className={`px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider ${col.className || ''}`}>
                {col.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {data.length === 0 ? (
            <tr>
              <td colSpan={columns.length} className="px-6 py-8 text-center text-gray-400" e2e-test-id={`${e2eId}-empty`}>
                No data found
              </td>
            </tr>
          ) : (
            data.map((item) => (
              <tr
                key={String(item[keyField])}
                className={onRowClick ? 'hover:bg-gray-50 cursor-pointer' : 'hover:bg-gray-50'}
                onClick={() => onRowClick?.(item)}
                e2e-test-id={`${e2eId}-row-${String(item[keyField])}`}
              >
                {columns.map((col) => (
                  <td key={col.key} className={`px-6 py-4 text-sm text-gray-700 ${col.className || ''}`}>
                    {col.render ? col.render(item) : String(item[col.key] ?? '')}
                  </td>
                ))}
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
}
