import React from 'react';
import { Handle, Position } from 'reactflow';
import 'reactflow/dist/style.css';

// Trigger node: entry point without input handle
export default function TriggerNode({ data }) {
  return (
    <div className="bg-white border rounded p-2 min-w-[100px]">
      <div className="text-sm font-medium">{data.label}</div>
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