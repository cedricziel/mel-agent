import { Handle, Position } from 'reactflow';
import 'reactflow/dist/style.css';
import { HANDLE_TYPES, getHandleColor } from '../utils/connectionTypes';

export default function MemoryNode({ data, onAddClick }) {
  const summaryKeys = Object.keys(data).filter(
    (k) =>
      k !== 'label' && k !== 'status' && k !== 'nodeTypeLabel' && k !== 'error'
  );

  return (
    <div
      className={`relative bg-purple-50 border-purple-200 rounded p-2 pl-6 min-w-[100px] ${
        data.error ? 'border-2 border-red-500' : 'border-2'
      }`}
    >
      <div className="absolute top-1 left-1 text-xs">🧠</div>

      {/* Quick-add button */}
      {onAddClick && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onAddClick();
          }}
          className="absolute top-1 right-1 w-5 h-5 bg-purple-600 text-white text-xs rounded flex items-center justify-center"
        >
          +
        </button>
      )}

      {/* Status indicators */}
      {data.status === 'running' && (
        <div className="absolute -top-1 -left-1 w-3 h-3 bg-purple-500 rounded-full animate-pulse"></div>
      )}

      {data.error && (
        <div className="absolute -top-1 -left-1 w-3 h-3 bg-red-500 rounded-full"></div>
      )}

      <div className="font-semibold text-sm text-purple-700">{data.label}</div>

      {/* Show key parameters */}
      {summaryKeys.length > 0 && (
        <div className="mt-1">
          {summaryKeys.slice(0, 3).map((key) => {
            const val = data[key];
            const display =
              val !== null && typeof val === 'object'
                ? JSON.stringify(val)
                : String(val);
            return (
              <div key={key} className="text-xs text-purple-600">
                {key}:{' '}
                {display.length > 15
                  ? display.substring(0, 15) + '...'
                  : display}
              </div>
            );
          })}
          {summaryKeys.length > 3 && (
            <div className="text-xs text-purple-400">…</div>
          )}
        </div>
      )}

      {/* Output handle at the top - connects to agent */}
      <Handle
        type="source"
        position={Position.Top}
        id="config-out"
        style={{
          backgroundColor: getHandleColor(HANDLE_TYPES.MEMORY_CONFIG),
          top: '-6px',
        }}
      />
    </div>
  );
}
