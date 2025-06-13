// Handle types for typed connections
export const HANDLE_TYPES = {
  // Workflow data flow
  WORKFLOW_INPUT: 'workflow-input',
  WORKFLOW_OUTPUT: 'workflow-output',

  // Configuration connections
  MODEL_CONFIG: 'model-config',
  TOOLS_CONFIG: 'tools-config',
  MEMORY_CONFIG: 'memory-config',

  // Trigger connections
  TRIGGER_OUTPUT: 'trigger-output',

  // Control flow
  CONTROL_INPUT: 'control-input',
  CONTROL_OUTPUT: 'control-output',
};

// Define which handle types can connect to each other
export const CONNECTION_RULES = {
  [HANDLE_TYPES.WORKFLOW_OUTPUT]: [HANDLE_TYPES.WORKFLOW_INPUT],
  [HANDLE_TYPES.TRIGGER_OUTPUT]: [HANDLE_TYPES.WORKFLOW_INPUT],
  [HANDLE_TYPES.CONTROL_OUTPUT]: [HANDLE_TYPES.CONTROL_INPUT],
  [HANDLE_TYPES.MODEL_CONFIG]: [HANDLE_TYPES.MODEL_CONFIG],
  [HANDLE_TYPES.TOOLS_CONFIG]: [HANDLE_TYPES.TOOLS_CONFIG],
  [HANDLE_TYPES.MEMORY_CONFIG]: [HANDLE_TYPES.MEMORY_CONFIG],
};

// Node type to handle configuration mapping
export const NODE_HANDLE_CONFIG = {
  // Standard workflow nodes
  default: {
    inputs: [
      {
        id: 'workflow-in',
        type: HANDLE_TYPES.WORKFLOW_INPUT,
        position: 'left',
      },
    ],
    outputs: [
      {
        id: 'workflow-out',
        type: HANDLE_TYPES.WORKFLOW_OUTPUT,
        position: 'right',
      },
    ],
  },

  // Agent nodes with configuration inputs
  agent: {
    inputs: [
      {
        id: 'workflow-in',
        type: HANDLE_TYPES.WORKFLOW_INPUT,
        position: 'left',
      },
      {
        id: 'model-config',
        type: HANDLE_TYPES.MODEL_CONFIG,
        position: 'bottom',
        offset: '25%',
      },
      {
        id: 'tools-config',
        type: HANDLE_TYPES.TOOLS_CONFIG,
        position: 'bottom',
        offset: '50%',
      },
      {
        id: 'memory-config',
        type: HANDLE_TYPES.MEMORY_CONFIG,
        position: 'bottom',
        offset: '75%',
      },
    ],
    outputs: [
      {
        id: 'workflow-out',
        type: HANDLE_TYPES.WORKFLOW_OUTPUT,
        position: 'right',
      },
    ],
  },

  // Configuration nodes (output only)
  model: {
    inputs: [],
    outputs: [
      { id: 'config-out', type: HANDLE_TYPES.MODEL_CONFIG, position: 'top' },
    ],
  },

  tools: {
    inputs: [],
    outputs: [
      { id: 'config-out', type: HANDLE_TYPES.TOOLS_CONFIG, position: 'top' },
    ],
  },

  memory: {
    inputs: [],
    outputs: [
      { id: 'config-out', type: HANDLE_TYPES.MEMORY_CONFIG, position: 'top' },
    ],
  },

  // Trigger nodes (output only)
  manual_trigger: {
    inputs: [],
    outputs: [
      {
        id: 'trigger-out',
        type: HANDLE_TYPES.TRIGGER_OUTPUT,
        position: 'right',
      },
    ],
  },

  webhook: {
    inputs: [],
    outputs: [
      {
        id: 'trigger-out',
        type: HANDLE_TYPES.TRIGGER_OUTPUT,
        position: 'right',
      },
    ],
  },

  schedule: {
    inputs: [],
    outputs: [
      {
        id: 'trigger-out',
        type: HANDLE_TYPES.TRIGGER_OUTPUT,
        position: 'right',
      },
    ],
  },

  workflow_trigger: {
    inputs: [],
    outputs: [
      {
        id: 'trigger-out',
        type: HANDLE_TYPES.TRIGGER_OUTPUT,
        position: 'right',
      },
    ],
  },

  // Control flow nodes
  if: {
    inputs: [
      {
        id: 'workflow-in',
        type: HANDLE_TYPES.WORKFLOW_INPUT,
        position: 'left',
      },
    ],
    outputs: [
      {
        id: 'true-out',
        type: HANDLE_TYPES.WORKFLOW_OUTPUT,
        position: 'right',
        offset: '30%',
      },
      {
        id: 'false-out',
        type: HANDLE_TYPES.WORKFLOW_OUTPUT,
        position: 'right',
        offset: '70%',
      },
    ],
  },

  for_each: {
    inputs: [
      {
        id: 'workflow-in',
        type: HANDLE_TYPES.WORKFLOW_INPUT,
        position: 'left',
      },
    ],
    outputs: [
      {
        id: 'workflow-out',
        type: HANDLE_TYPES.WORKFLOW_OUTPUT,
        position: 'right',
      },
    ],
  },

  switch_node: {
    inputs: [
      {
        id: 'workflow-in',
        type: HANDLE_TYPES.WORKFLOW_INPUT,
        position: 'left',
      },
    ],
    outputs: [
      {
        id: 'workflow-out',
        type: HANDLE_TYPES.WORKFLOW_OUTPUT,
        position: 'right',
      },
    ],
  },
};

// Validate if a connection between two handles is allowed
export function isValidConnection(
  sourceHandle,
  targetHandle,
  sourceNodeType,
  targetNodeType
) {
  const sourceConfig =
    NODE_HANDLE_CONFIG[sourceNodeType] || NODE_HANDLE_CONFIG.default;
  const targetConfig =
    NODE_HANDLE_CONFIG[targetNodeType] || NODE_HANDLE_CONFIG.default;

  // Find the handle configurations
  const sourceHandleConfig = sourceConfig.outputs?.find(
    (h) => h.id === sourceHandle
  );
  const targetHandleConfig = targetConfig.inputs?.find(
    (h) => h.id === targetHandle
  );

  if (!sourceHandleConfig || !targetHandleConfig) {
    return false;
  }

  // Check if the handle types are compatible
  const allowedTargetTypes = CONNECTION_RULES[sourceHandleConfig.type] || [];
  return allowedTargetTypes.includes(targetHandleConfig.type);
}

// Get handle color based on type
export function getHandleColor(handleType) {
  switch (handleType) {
    case HANDLE_TYPES.WORKFLOW_INPUT:
    case HANDLE_TYPES.WORKFLOW_OUTPUT:
      return '#6b7280'; // gray
    case HANDLE_TYPES.MODEL_CONFIG:
      return '#3b82f6'; // blue
    case HANDLE_TYPES.TOOLS_CONFIG:
      return '#10b981'; // green
    case HANDLE_TYPES.MEMORY_CONFIG:
      return '#8b5cf6'; // purple
    case HANDLE_TYPES.TRIGGER_OUTPUT:
      return '#f59e0b'; // amber
    case HANDLE_TYPES.CONTROL_INPUT:
    case HANDLE_TYPES.CONTROL_OUTPUT:
      return '#ef4444'; // red
    default:
      return '#6b7280';
  }
}

// Get edge style based on handle type
export function getEdgeStyle(handleType) {
  return {
    stroke: getHandleColor(handleType),
    strokeWidth: 2,
  };
}
