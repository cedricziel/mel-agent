import { Handle, Position } from 'reactflow';
import 'reactflow/dist/style.css';
import { HANDLE_TYPES, getHandleColor } from '../utils/connectionTypes';

interface TransformNodeProps {
  data: {
    label?: string;
    nodeTypeLabel?: string;
    expression?: string;
    status?: string;
    error?: boolean;
  };
  id: string;
  icon?: string;
  onAddClick?: () => void;
  onDelete?: (id: string) => void;
}

/** Canvas node for the Transform node: shows the template it will render. */
export default function TransformNode({
  data,
  id,
  icon,
  onAddClick,
  onDelete,
}: TransformNodeProps) {
  const expression = data.expression || '';
  const nodeIcon = icon || '🔄';

  return (
    <div
      className={`relative bg-sky-50 rounded p-2 pl-6 min-w-[160px] max-w-[240px] ${
        data.error ? 'border-2 border-red-500' : 'border border-sky-300'
      }`}
    >
      <div className="absolute top-1 left-1 text-xs">{nodeIcon}</div>

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

      {data.status === 'running' && (
        <div className="absolute top-1 right-1 w-2 h-2 bg-blue-500 rounded-full animate-pulse z-10" />
      )}

      <div className="text-sm font-medium">{data.label}</div>
      {data.nodeTypeLabel && (
        <div className="text-xs text-gray-500 mb-1">{data.nodeTypeLabel}</div>
      )}

      {expression ? (
        <div
          className="text-xs font-mono text-sky-800 bg-white border border-sky-200 rounded px-1 py-0.5 truncate"
          title={expression}
        >
          {expression}
        </div>
      ) : (
        <div className="text-xs text-gray-400 italic">no expression set</div>
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
    </div>
  );
}
