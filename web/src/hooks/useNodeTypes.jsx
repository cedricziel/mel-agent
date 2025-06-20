import { useMemo, useEffect, useState, useCallback } from 'react';
import { nodeTypesApi } from '../api/nodeTypesApi';
import { triggersApi } from '../api/client';
import DefaultNode from '../components/DefaultNode';
import AgentNode from '../components/AgentNode';
import ModelNode from '../components/ModelNode';
import ToolsNode from '../components/ToolsNode';
import MemoryNode from '../components/MemoryNode';
import OpenAIModelNode from '../components/OpenAIModelNode';
import AnthropicModelNode from '../components/AnthropicModelNode';
import LocalMemoryNode from '../components/LocalMemoryNode';
import IfNode from '../components/IfNode';
import HttpRequestNode from '../components/HttpRequestNode';
import TriggerNode from '../components/TriggerNode';

export function useNodeTypes(
  agentId,
  handleNodeDelete,
  openConfigDialog,
  onNodeClick
) {
  const [nodeDefs, setNodeDefs] = useState([]);
  const [triggers, setTriggers] = useState([]);

  // Stabilize function dependencies to prevent unnecessary re-renders
  const stableHandleNodeDelete = useCallback(
    (nodeId) => {
      if (handleNodeDelete) {
        handleNodeDelete(nodeId);
      }
    },
    [handleNodeDelete]
  );

  const stableOpenConfigDialog = useCallback(
    (nodeId, configType, handleId) => {
      if (openConfigDialog) {
        openConfigDialog(nodeId, configType, handleId);
      }
    },
    [openConfigDialog]
  );

  const stableOnNodeClick = useCallback(
    (event, node) => {
      if (onNodeClick) {
        onNodeClick(event, node);
      }
    },
    [onNodeClick]
  );

  // Load node definitions
  useEffect(() => {
    // Use the new API client that handles filtering and fallbacks
    nodeTypesApi
      .getAllNodeTypes()
      .then((allNodeDefs) => setNodeDefs(allNodeDefs))
      .catch((err) => console.error('fetch node-types failed:', err));
  }, []);

  // Load triggers
  useEffect(() => {
    triggersApi
      .listTriggers()
      .then((res) => setTriggers(res.data))
      .catch((err) => console.error('fetch triggers failed:', err));
  }, []);

  // Dynamically create nodeTypes based on available node definitions
  const nodeTypes = useMemo(() => {
    const types = {
      default: (props) => (
        <DefaultNode {...props} onDelete={stableHandleNodeDelete} />
      ),
      agent: (props) => (
        <AgentNode
          {...props}
          onDelete={stableHandleNodeDelete}
          onAddConfigNode={(configType, handleId) => {
            stableOpenConfigDialog(props.id, configType, handleId);
          }}
        />
      ),
      // Legacy generic config nodes (keeping for backward compatibility)
      model: (props) => (
        <ModelNode
          {...props}
          onDelete={stableHandleNodeDelete}
          onClick={() => stableOnNodeClick(null, props)}
        />
      ),
      tools: (props) => (
        <ToolsNode
          {...props}
          onDelete={stableHandleNodeDelete}
          onClick={() => stableOnNodeClick(null, props)}
        />
      ),
      memory: (props) => (
        <MemoryNode
          {...props}
          onDelete={stableHandleNodeDelete}
          onClick={() => stableOnNodeClick(null, props)}
        />
      ),
      // Specific config nodes
      openai_model: (props) => (
        <OpenAIModelNode
          {...props}
          onDelete={stableHandleNodeDelete}
          onClick={() => stableOnNodeClick(null, props)}
        />
      ),
      anthropic_model: (props) => (
        <AnthropicModelNode
          {...props}
          onDelete={stableHandleNodeDelete}
          onClick={() => stableOnNodeClick(null, props)}
        />
      ),
      local_memory: (props) => (
        <LocalMemoryNode
          {...props}
          onDelete={stableHandleNodeDelete}
          onClick={() => stableOnNodeClick(null, props)}
        />
      ),
      workflow_tools: (props) => (
        <ToolsNode
          {...props}
          onDelete={stableHandleNodeDelete}
          onClick={() => stableOnNodeClick(null, props)}
        />
      ),
      if: (props) => <IfNode {...props} onDelete={stableHandleNodeDelete} />,
      http_request: (props) => (
        <HttpRequestNode {...props} onDelete={stableHandleNodeDelete} />
      ),
    };

    // Add trigger nodes with special rendering
    const triggerTypes = [
      'webhook',
      'schedule',
      'manual_trigger',
      'workflow_trigger',
      'slack',
      'timer',
    ];
    triggerTypes.forEach((type) => {
      types[type] = (props) => {
        const nodeDef = nodeDefs.find((def) => def.type === type);
        return (
          <TriggerNode
            {...props}
            type={type}
            agentId={agentId}
            icon={nodeDef?.icon}
            onDelete={stableHandleNodeDelete}
          />
        );
      };
    });

    // Add all other node types to use DefaultNode (defensive check for nodeDefs)
    if (Array.isArray(nodeDefs)) {
      nodeDefs.forEach((nodeDef) => {
        if (!types[nodeDef.type]) {
          types[nodeDef.type] = (props) => (
            <DefaultNode {...props} onDelete={stableHandleNodeDelete} />
          );
        }
      });
    }

    return types;
  }, [
    nodeDefs,
    agentId,
    stableOpenConfigDialog,
    stableHandleNodeDelete,
    stableOnNodeClick,
  ]);

  // Node categories for the add node modal
  const categories = useMemo(() => {
    const map = {};
    nodeDefs.forEach((def) => {
      if (!map[def.category]) map[def.category] = [];
      map[def.category].push(def);
    });
    return Object.entries(map).map(([category, types]) => ({
      category,
      types,
    }));
  }, [nodeDefs]);

  // Map of nodeId to trigger instance for this agent
  const triggersMap = useMemo(() => {
    const map = {};
    triggers.forEach((t) => {
      if (t.agent_id === agentId && t.node_id) {
        map[t.node_id] = t;
      }
    });
    return map;
  }, [triggers, agentId]);

  // Refresh triggers function
  const refreshTriggers = async () => {
    try {
      const res = await triggersApi.listTriggers();
      setTriggers(res.data);
    } catch (err) {
      console.error('refresh triggers failed:', err);
    }
  };

  return {
    nodeDefs,
    triggers,
    nodeTypes,
    categories,
    triggersMap,
    refreshTriggers,
  };
}
