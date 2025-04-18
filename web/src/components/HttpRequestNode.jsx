import React from 'react';
import { Handle, Position } from 'reactflow';
import 'reactflow/dist/style.css';

// HTTP Request node: sends an HTTP request during workflow execution
export default function HttpRequestNode({ data, onAddClick }) {
  const { label, method, url } = data;
  return (
    <div
      className={
        `relative bg-white rounded p-2 min-w-[140px] ${
          data.error ? 'border-2 border-red-500' : 'border'
        }`
      }
    >
      {/* Node label */}
      <div className="text-sm font-medium">{label}</div>
      {/* Subtitle: node type */}
      {data.nodeTypeLabel && (
        <div className="text-xs text-gray-500 mb-1">{data.nodeTypeLabel}</div>
      )}
      {/* Summary: HTTP method and URL */}
      <div className="text-xs text-gray-600">
        <span className="font-semibold">{method || 'GET'}</span>{' '}
        <span className="truncate block" title={url}>{url}</span>
      </div>
      {/* Quick-add button */}
      {onAddClick && (
        <button
          onClick={(e) => { e.stopPropagation(); onAddClick(); }}
          className="absolute top-1 right-1 w-5 h-5 bg-indigo-600 text-white text-xs rounded flex items-center justify-center"
        >
          +
        </button>
      )}
      {/* Input handle */}
      <Handle
        type="target"
        position={Position.Top}
        id="in"
        className="!bg-gray-600"
      />
      {/* Output handle */}
      <Handle
        type="source"
        position={Position.Bottom}
        id="out"
        className="!bg-gray-600"
      />
    </div>
  );
}