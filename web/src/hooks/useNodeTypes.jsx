import { useMemo, useEffect, useState } from 'react';
import axios from 'axios';
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

  // Load node definitions
  useEffect(() => {
    axios
      .get('/api/node-types')
      .then((res) => setNodeDefs(res.data))
      .catch((err) => console.error('fetch node-types failed:', err));
  }, []);

  // Load triggers
  useEffect(() => {
    axios
      .get('/api/triggers')
      .then((res) => setTriggers(res.data))
      .catch((err) => console.error('fetch triggers failed:', err));
  }, []);

  // Dynamically create nodeTypes based on available node definitions
  const nodeTypes = useMemo(() => {
    const types = {
      default: (props) => (
        <DefaultNode {...props} onDelete={handleNodeDelete} />
      ),
      agent: (props) => (
        <AgentNode
          {...props}
          onDelete={handleNodeDelete}
          onAddConfigNode={(configType, handleId) => {
            openConfigDialog(props.id, configType, handleId);
          }}
        />
      ),
      // Legacy generic config nodes (keeping for backward compatibility)
      model: (props) => (
        <ModelNode
          {...props}
          onDelete={handleNodeDelete}
          onClick={() => onNodeClick(null, props)}
        />
      ),
      tools: (props) => (
        <ToolsNode
          {...props}
          onDelete={handleNodeDelete}
          onClick={() => onNodeClick(null, props)}
        />
      ),
      memory: (props) => (
        <MemoryNode
          {...props}
          onDelete={handleNodeDelete}
          onClick={() => onNodeClick(null, props)}
        />
      ),
      // Specific config nodes
      openai_model: (props) => (
        <OpenAIModelNode
          {...props}
          onDelete={handleNodeDelete}
          onClick={() => onNodeClick(null, props)}
        />
      ),
      anthropic_model: (props) => (
        <AnthropicModelNode
          {...props}
          onDelete={handleNodeDelete}
          onClick={() => onNodeClick(null, props)}
        />
      ),
      local_memory: (props) => (
        <LocalMemoryNode
          {...props}
          onDelete={handleNodeDelete}
          onClick={() => onNodeClick(null, props)}
        />
      ),
      workflow_tools: (props) => (
        <ToolsNode
          {...props}
          onDelete={handleNodeDelete}
          onClick={() => onNodeClick(null, props)}
        />
      ),
      if: (props) => <IfNode {...props} onDelete={handleNodeDelete} />,
      http_request: (props) => (
        <HttpRequestNode {...props} onDelete={handleNodeDelete} />
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
            onDelete={handleNodeDelete}
          />
        );
      };
    });

    // Add all other node types to use DefaultNode (defensive check for nodeDefs)
    if (Array.isArray(nodeDefs)) {
      nodeDefs.forEach((nodeDef) => {
        if (!types[nodeDef.type]) {
          types[nodeDef.type] = (props) => (
            <DefaultNode {...props} onDelete={handleNodeDelete} />
          );
        }
      });
    }

    return types;
  }, [nodeDefs, agentId, openConfigDialog, handleNodeDelete, onNodeClick]);

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
      const res = await axios.get('/api/triggers');
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
