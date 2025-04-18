import React from 'react';
import { Handle, Position } from 'reactflow';
import 'reactflow/dist/style.css';

// Generic node renderer: one input, one output, shows label
export default function DefaultNode({ data }) {
  const summaryKeys = Object.keys(data).filter(
    (k) => k !== 'label' && k !== 'status'
  );
  return (
    <div
      className={
        `relative bg-white rounded p-2 min-w-[100px] ${
          data.error ? 'border-2 border-red-500' : 'border'
        }`
      }
    >
      {/* Status indicator: running */}
      {data.status === 'running' && (
        <div className="absolute top-1 right-1 w-2 h-2 bg-blue-500 rounded-full animate-pulse" />
      )}
      <div className="text-sm font-medium">{data.label}</div>
      {data.nodeTypeLabel && (
        <div className="text-xs text-gray-500">{data.nodeTypeLabel}</div>
      )}
      {/* Parameter summary */}
      {summaryKeys.length > 0 && (
        <div className="mt-1 space-y-0.5">
          {summaryKeys.slice(0, 2).map((key) => {
            const val = data[key];
            const display = val !== null && typeof val === 'object' ? JSON.stringify(val) : String(val);
            return (
              <div key={key} className="text-xs text-gray-600">
                {key}: {display}
              </div>
            );
          })}
          {summaryKeys.length > 2 && (
            <div className="text-xs text-gray-400">…</div>
          )}
        </div>
      )}
      <Handle
        type="target"
        position={Position.Top}
        id="in"
        className="!bg-gray-600"
      />
      <Handle
        type="source"
        position={Position.Bottom}
        id="out"
        className="!bg-gray-600"
      />
    </div>
  );
}