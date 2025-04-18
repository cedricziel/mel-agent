import React from 'react';
import { Handle, Position } from 'reactflow';
import 'reactflow/dist/style.css';

// Generic node renderer: one input, one output, shows label
export default function DefaultNode({ data }) {
  return (
    <div className="bg-white border rounded p-2 min-w-[100px]">
      <div className="text-sm font-medium">{data.label}</div>
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