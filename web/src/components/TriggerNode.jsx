import React from 'react';
import { Handle, Position } from 'reactflow';
import 'reactflow/dist/style.css';

// Trigger node: entry point without input handle
export default function TriggerNode({ data }) {
  const summaryKeys = Object.keys(data).filter(
    (k) => k !== 'label' && k !== 'status'
  );
  return (
    <div className="relative bg-white border rounded p-2 min-w-[100px]">
      {/* Status indicator: running */}
      {data.status === 'running' && (
        <div className="absolute top-1 right-1 w-2 h-2 bg-blue-500 rounded-full animate-pulse" />
      )}
      <div className="text-sm font-medium">{data.label}</div>
      {/* Parameter summary */}
      {summaryKeys.length > 0 && (
        <div className="mt-1 space-y-0.5">
          {summaryKeys.slice(0, 2).map((key) => (
            <div key={key} className="text-xs text-gray-600">
              {key}: {data[key]}
            </div>
          ))}
          {summaryKeys.length > 2 && (
            <div className="text-xs text-gray-400">…</div>
          )}
        </div>
      )}
      {/* Only output handle: no input for trigger nodes */}
      <Handle
        type="source"
        position={Position.Bottom}
        id="out"
        className="!bg-gray-600"
      />
    </div>
  );
}