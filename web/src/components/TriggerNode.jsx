import React from 'react';
import { Handle, Position } from 'reactflow';
import axios from 'axios';
import 'reactflow/dist/style.css';
import { HANDLE_TYPES, getHandleColor } from '../utils/connectionTypes';

// Trigger node: entry point without input handle, with n8n-style curved left side
export default function TriggerNode({
  data,
  onAddClick,
  icon,
  id,
  type,
  agentId,
}) {
  const summaryKeys = Object.keys(data).filter(
    (k) => k !== 'label' && k !== 'status'
  );
  const nodeIcon = icon || '🔔';
  const isManualTrigger = type === 'manual_trigger';

  const handleManualTrigger = async (e) => {
    e.stopPropagation();
    try {
      if (!agentId) {
        alert('Could not determine agent ID');
        return;
      }

      const response = await axios.post(`/api/agents/${agentId}/runs/test`);
      console.log('Manual trigger executed:', response.data);

      // Show visual feedback
      const button = e.target;
      const originalText = button.textContent;
      button.textContent = '✓';
      button.style.backgroundColor = '#10b981';

      setTimeout(() => {
        button.textContent = originalText;
        button.style.backgroundColor = '';
      }, 1500);
    } catch (error) {
      console.error('Failed to trigger manually:', error);
      alert('Failed to trigger workflow');
    }
  };

  return (
    <div className="relative">
      {/* n8n-style curved left side */}
      <div
        className={`
          relative bg-white p-2 pl-6 min-w-[100px]
          ${data.error ? 'border-2 border-red-500' : 'border border-gray-300'}
          rounded-r-lg
        `}
        style={{
          borderTopLeftRadius: '50px',
          borderBottomLeftRadius: '50px',
          borderLeft: 'none',
        }}
      >
        <div className="absolute top-1 left-1 text-xs">{nodeIcon}</div>

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

        {/* Manual trigger button */}
        {isManualTrigger && (
          <button
            onClick={handleManualTrigger}
            className="mt-2 w-full py-1 px-2 bg-green-500 hover:bg-green-600 text-white text-xs rounded flex items-center justify-center gap-1 transition-colors"
            title="Trigger workflow manually"
          >
            ▶️ Trigger
          </button>
        )}

        {/* Quick-add button */}
        {onAddClick && !isManualTrigger && (
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

        {/* Only output handle: no input for trigger nodes */}
        <Handle
          type="source"
          position={Position.Right}
          id="trigger-out"
          style={{
            backgroundColor: getHandleColor(HANDLE_TYPES.TRIGGER_OUTPUT),
          }}
        />
      </div>
    </div>
  );
}
