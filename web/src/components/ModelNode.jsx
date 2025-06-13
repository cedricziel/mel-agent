import { Handle, Position } from 'reactflow';
import 'reactflow/dist/style.css';
import { HANDLE_TYPES, getHandleColor } from '../utils/connectionTypes';

export default function ModelNode({ data, onAddClick, onDelete, id }) {
  const summaryKeys = Object.keys(data).filter(
    (k) =>
      k !== 'label' && k !== 'status' && k !== 'nodeTypeLabel' && k !== 'error'
  );

  return (
    <div
      className={`relative bg-blue-50 border-blue-200 rounded-full p-3 w-[120px] h-[120px] flex flex-col items-center justify-center ${
        data.error ? 'border-2 border-red-500' : 'border-2'
      }`}
    >
      <div className="absolute top-2 text-lg">📋</div>

      {/* Delete button */}
      {onDelete && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onDelete(id);
          }}
          className="absolute -top-2 -right-2 w-5 h-5 bg-red-500 hover:bg-red-600 text-white text-xs rounded-full flex items-center justify-center opacity-70 hover:opacity-100 transition-opacity"
          title="Delete node"
        >
          ×
        </button>
      )}

      {/* Quick-add button */}
      {onAddClick && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onAddClick();
          }}
          className="absolute top-1 right-1 w-5 h-5 bg-blue-600 text-white text-xs rounded-full flex items-center justify-center"
        >
          +
        </button>
      )}

      {/* Status indicators */}
      {data.status === 'running' && (
        <div className="absolute -top-1 -left-1 w-3 h-3 bg-blue-500 rounded-full animate-pulse"></div>
      )}

      {data.error && (
        <div className="absolute -top-1 -left-1 w-3 h-3 bg-red-500 rounded-full"></div>
      )}

      <div className="font-semibold text-sm text-blue-700 text-center mt-4">
        {data.label}
      </div>

      {/* Show key parameters */}
      {summaryKeys.length > 0 && (
        <div className="mt-1 text-center">
          {summaryKeys.slice(0, 2).map((key) => {
            const val = data[key];
            const display =
              val !== null && typeof val === 'object'
                ? JSON.stringify(val)
                : String(val);
            return (
              <div key={key} className="text-xs text-blue-600">
                {key}:{' '}
                {display.length > 12
                  ? display.substring(0, 12) + '...'
                  : display}
              </div>
            );
          })}
          {summaryKeys.length > 2 && (
            <div className="text-xs text-blue-400">…</div>
          )}
        </div>
      )}

      {/* Output handle at the top - connects to agent */}
      <Handle
        type="source"
        position={Position.Top}
        id="config-out"
        style={{
          backgroundColor: getHandleColor(HANDLE_TYPES.MODEL_CONFIG),
          top: '-6px',
        }}
      />
    </div>
  );
}
