import React from 'react';
import { Handle, Position } from 'reactflow';
import 'reactflow/dist/style.css';

// Custom node for 'if' logic: one input, two outputs (true/false)
export default function IfNode({ id, data, onAddClick }) {
  const condition = data.condition || '';
  return (
    <div
      className={
        `relative bg-yellow-50 rounded p-2 min-w-[120px] ${
          data.error ? 'border-2 border-red-500' : 'border border-yellow-400'
        }`
      }
    >
      {/* Quick-add button */}
      {onAddClick && (
        <button
          onClick={(e) => { e.stopPropagation(); onAddClick(); }}
          className="absolute top-1 right-1 w-5 h-5 bg-indigo-600 text-white text-xs rounded flex items-center justify-center"
        >
          +
        </button>
      )}
      {/* Status indicator: running */}
      {data.status === 'running' && (
        <div className="absolute top-1 right-1 w-2 h-2 bg-blue-500 rounded-full animate-pulse" />
      )}
      <div className="text-sm font-medium">{data.label}</div>
      {data.nodeTypeLabel && (
        <div className="text-xs text-gray-500 mb-1">{data.nodeTypeLabel}</div>
      )}
      <div className="mb-2">
        <input
          type="text"
          readOnly
          value={condition}
          placeholder="condition"
          className="w-full text-xs px-1 py-0.5 border rounded"
        />
      </div>
      {/* Input handle */}
      <Handle
        type="target"
        position={Position.Top}
        id="in"
        className="!bg-yellow-600"
      />
      {/* True/False output handles */}
      <Handle
        type="source"
        position={Position.Bottom}
        id="true"
        className="!bg-green-600"
        style={{ left: '30%' }}
      />
      <Handle
        type="source"
        position={Position.Bottom}
        id="false"
        className="!bg-red-600"
        style={{ left: '70%' }}
      />
    </div>
  );
}