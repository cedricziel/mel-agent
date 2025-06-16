import { Handle, Position } from 'reactflow';
import 'reactflow/dist/style.css';
import { HANDLE_TYPES, getHandleColor } from '../utils/connectionTypes';

// Specialized node renderer for agent nodes with configuration inputs at the bottom
export default function AgentNode({
  data,
  onAddClick,
  onAddConfigNode,
  onDelete,
  id,
}) {
  const summaryKeys = Object.keys(data).filter(
    (k) =>
      k !== 'label' &&
      k !== 'status' &&
      k !== 'nodeTypeLabel' &&
      k !== 'error' &&
      k !== 'modelConfig' &&
      k !== 'toolsConfig' &&
      k !== 'memoryConfig'
  );

  return (
    <div
      className={`relative bg-white rounded p-2 pl-6 min-w-[120px] ${
        data.error ? 'border-2 border-red-500' : 'border'
      }`}
    >
      <div className="absolute top-1 left-1 text-xs">🤖</div>

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
        <div className="absolute -top-1 -left-1 w-3 h-3 bg-blue-500 rounded-full animate-pulse"></div>
      )}

      {/* Status indicator: error */}
      {data.error && (
        <div className="absolute -top-1 -left-1 w-3 h-3 bg-red-500 rounded-full"></div>
      )}

      <div className="font-semibold text-sm">{data.label}</div>

      {/* Show configuration status */}
      <div className="text-xs text-gray-500 mt-1">
        {data.modelConfig && '📋'}
        {data.toolsConfig && '🔧'}
        {data.memoryConfig && '🧠'}
      </div>

      {/* Show other parameters */}
      {summaryKeys.length > 0 && (
        <div className="mt-1">
          {summaryKeys.slice(0, 2).map((key) => {
            const val = data[key];
            const display =
              val !== null && typeof val === 'object'
                ? JSON.stringify(val)
                : String(val);
            return (
              <div key={key} className="text-xs text-gray-600">
                {key}:{' '}
                {display.length > 20
                  ? display.substring(0, 20) + '...'
                  : display}
              </div>
            );
          })}
          {summaryKeys.length > 2 && (
            <div className="text-xs text-gray-400">…</div>
          )}
        </div>
      )}

      {/* Standard workflow input/output handles */}
      <Handle
        type="target"
        position={Position.Left}
        id="workflow-in"
        style={{
          backgroundColor: getHandleColor(HANDLE_TYPES.WORKFLOW_INPUT),
          top: '50%',
          left: '-8px',
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
          top: '50%',
          right: '-8px',
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

      {/* Configuration input handles at the bottom */}
      <Handle
        type="target"
        position={Position.Bottom}
        id="model-config"
        style={{
          backgroundColor: getHandleColor(HANDLE_TYPES.MODEL_CONFIG),
          left: '25%',
          bottom: '-8px',
          width: '16px',
          height: '16px',
          border: '2px solid white',
        }}
      />
      <Handle
        type="target"
        position={Position.Bottom}
        id="tools-config"
        style={{
          backgroundColor: getHandleColor(HANDLE_TYPES.TOOLS_CONFIG),
          left: '50%',
          bottom: '-8px',
          width: '16px',
          height: '16px',
          border: '2px solid white',
        }}
      />
      <Handle
        type="target"
        position={Position.Bottom}
        id="memory-config"
        style={{
          backgroundColor: getHandleColor(HANDLE_TYPES.MEMORY_CONFIG),
          left: '75%',
          bottom: '-8px',
          width: '16px',
          height: '16px',
          border: '2px solid white',
        }}
      />

      {/* Config handle connector lines */}
      <div
        className="absolute bg-gray-300"
        style={{
          left: 'calc(25% - 1px)',
          bottom: '-8px',
          width: '2px',
          height: '8px',
        }}
      />
      <div
        className="absolute bg-gray-300"
        style={{
          left: 'calc(50% - 1px)',
          bottom: '-8px',
          width: '2px',
          height: '8px',
        }}
      />
      <div
        className="absolute bg-gray-300"
        style={{
          left: 'calc(75% - 1px)',
          bottom: '-8px',
          width: '2px',
          height: '8px',
        }}
      />

      {/* Add buttons for configuration nodes - always visible */}
      {onAddConfigNode && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onAddConfigNode('model', 'model-config');
          }}
          className="absolute bg-blue-500 text-white text-xs rounded-full w-4 h-4 flex items-center justify-center hover:bg-blue-600 transition-colors z-10"
          style={{ left: 'calc(25% - 8px)', bottom: '-20px' }}
          title="Add Model Configuration"
        >
          +
        </button>
      )}

      {onAddConfigNode && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onAddConfigNode('tools', 'tools-config');
          }}
          className="absolute bg-green-500 text-white text-xs rounded-full w-4 h-4 flex items-center justify-center hover:bg-green-600 transition-colors z-10"
          style={{ left: 'calc(50% - 8px)', bottom: '-20px' }}
          title="Add Tools Configuration"
        >
          +
        </button>
      )}

      {onAddConfigNode && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onAddConfigNode('memory', 'memory-config');
          }}
          className="absolute bg-purple-500 text-white text-xs rounded-full w-4 h-4 flex items-center justify-center hover:bg-purple-600 transition-colors z-10"
          style={{ left: 'calc(75% - 8px)', bottom: '-20px' }}
          title="Add Memory Configuration"
        >
          +
        </button>
      )}

      {/* Configuration labels */}
      <div className="absolute -bottom-8 left-0 right-0 flex justify-around text-xs text-gray-400">
        <span>📋</span>
        <span>🔧</span>
        <span>🧠</span>
      </div>
    </div>
  );
}
