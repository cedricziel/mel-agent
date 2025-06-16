import React from 'react';
import { Handle, Position } from 'reactflow';
import 'reactflow/dist/style.css';
import { HANDLE_TYPES, getHandleColor } from '../utils/connectionTypes';

// Generic node renderer: one input, one output, shows label
export default function DefaultNode({ data, onAddClick, icon, onDelete, id }) {
  const summaryKeys = Object.keys(data).filter(
    (k) =>
      k !== 'label' && k !== 'status' && k !== 'nodeTypeLabel' && k !== 'error'
  );
  const nodeIcon = icon || '📦';
  return (
    <div
      className={`relative bg-white rounded p-2 pl-6 min-w-[100px] ${
        data.error ? 'border-2 border-red-500' : 'border'
      }`}
    >
      <div className="absolute top-1 left-1 text-xs">{nodeIcon}</div>

      {/* Delete button */}
      {onDelete && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onDelete(id);
          }}
          className="absolute -top-2 -right-2 w-4 h-4 text-xs flex items-center justify-center opacity-60 hover:opacity-100 transition-opacity"
          title="Delete node"
        >
          🗑️
        </button>
      )}

      {/* Quick-add button */}
      {onAddClick && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onAddClick();
          }}
          className="absolute top-1 right-1 w-5 h-5 bg-indigo-600 text-white text-xs rounded flex items-center justify-center"
        >
          +
        </button>
      )}
      {/* Status indicator: running */}
      {data.status === 'running' && (
        <div className="absolute top-1 right-1 w-2 h-2 bg-blue-500 rounded-full animate-pulse z-10" />
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
            const display =
              val !== null && typeof val === 'object'
                ? JSON.stringify(val)
                : String(val);
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
        position={Position.Left}
        id="workflow-in"
        style={{
          backgroundColor: getHandleColor(HANDLE_TYPES.WORKFLOW_INPUT),
          left: '-8px',
          top: '50%',
          width: '16px',
          height: '16px',
          border: '2px solid white',
        }}
      />

      {/* Input handle connector line */}
      <div
        className="absolute bg-gray-300"
        style={{
          left: '-8px',
          top: 'calc(50% - 1px)',
          width: '8px',
          height: '2px',
        }}
      />

      <Handle
        type="source"
        position={Position.Right}
        id="workflow-out"
        style={{
          backgroundColor: getHandleColor(HANDLE_TYPES.WORKFLOW_OUTPUT),
          right: '-8px',
          top: '50%',
          width: '16px',
          height: '16px',
          border: '2px solid white',
        }}
      />

      {/* Output handle connector line */}
      <div
        className="absolute bg-gray-300"
        style={{
          right: '-8px',
          top: 'calc(50% - 1px)',
          width: '8px',
          height: '2px',
        }}
      />

      {/* Add button on output handle */}
      {onAddClick && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onAddClick();
          }}
          className="absolute bg-indigo-500 text-white text-xs rounded-full w-4 h-4 flex items-center justify-center hover:bg-indigo-600 transition-colors z-10"
          style={{ right: '-20px', top: 'calc(50% - 8px)' }}
          title="Add Next Node"
        >
          +
        </button>
      )}
    </div>
  );
}
