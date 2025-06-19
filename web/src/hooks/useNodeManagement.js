import { useCallback } from 'react';

/**
 * Custom hook for managing node and edge operations in the workflow builder
 * Handles CRUD operations, agent configuration creation, and broadcasting changes
 */
export function useNodeManagement(agentId, broadcastNodeChange, workflowState) {
  const {
    nodes,
    edges,
    createNode,
    updateNode,
    deleteNode,
    createEdge,
    deleteEdge,
  } = workflowState;

  // Handle node deletion with edge cleanup
  const handleNodeDelete = useCallback(
    async (nodeId) => {
      try {
        // Find all edges connected to this node
        const connectedEdges = edges.filter(
          (edge) => edge.source === nodeId || edge.target === nodeId
        );

        // Delete the node first
        await deleteNode(nodeId);
        broadcastNodeChange('nodeDeleted', { nodeId });

        // Then delete all connected edges
        for (const edge of connectedEdges) {
          try {
            await deleteEdge(edge.id);
            broadcastNodeChange('edgeDeleted', { edgeId: edge.id });
          } catch (edgeErr) {
            console.error('Failed to delete edge:', edge.id, edgeErr);
          }
        }
      } catch (err) {
        console.error('Failed to delete node:', err);
      }
    },
    [deleteNode, deleteEdge, broadcastNodeChange, edges]
  );

  // Handle edge deletion
  const handleEdgeDelete = useCallback(
    async (edgeId) => {
      try {
        await deleteEdge(edgeId);
        broadcastNodeChange('edgeDeleted', { edgeId });
      } catch (err) {
        console.error('Failed to delete edge:', err);
      }
    },
    [deleteEdge, broadcastNodeChange]
  );

  // Helper function to create configuration nodes for an agent
  const createAgentConfigurationNodes = useCallback(
    async (agentId, agentPosition) => {
      // Create model configuration node
      const modelId = crypto.randomUUID();
      const modelNode = {
        id: modelId,
        type: 'model',
        data: {
          label: 'Model Config',
          nodeTypeLabel: 'Model Configuration',
          provider: 'openai',
          model: 'gpt-4',
        },
        position: { x: agentPosition.x - 200, y: agentPosition.y - 100 },
      };

      // Create tools configuration node
      const toolsId = crypto.randomUUID();
      const toolsNode = {
        id: toolsId,
        type: 'tools',
        data: {
          label: 'Tools Config',
          nodeTypeLabel: 'Tools Configuration',
          allowCodeExecution: false,
          allowWebSearch: true,
        },
        position: { x: agentPosition.x - 200, y: agentPosition.y },
      };

      // Create memory configuration node
      const memoryId = crypto.randomUUID();
      const memoryNode = {
        id: memoryId,
        type: 'memory',
        data: {
          label: 'Memory Config',
          nodeTypeLabel: 'Memory Configuration',
          memoryType: 'short_term',
          maxMessages: 100,
        },
        position: { x: agentPosition.x - 200, y: agentPosition.y + 100 },
      };

      // Create edges connecting config nodes to agent (using specific handles)
      const modelEdge = {
        id: `edge-model-${crypto.randomUUID()}`,
        source: modelId,
        sourceHandle: 'config-out',
        target: agentId,
        targetHandle: 'model-config',
        type: 'default',
        style: { stroke: '#3b82f6' }, // Blue for model
      };

      const toolsEdge = {
        id: `edge-tools-${crypto.randomUUID()}`,
        source: toolsId,
        sourceHandle: 'config-out',
        target: agentId,
        targetHandle: 'tools-config',
        type: 'default',
        style: { stroke: '#10b981' }, // Green for tools
      };

      const memoryEdge = {
        id: `edge-memory-${crypto.randomUUID()}`,
        source: memoryId,
        sourceHandle: 'config-out',
        target: agentId,
        targetHandle: 'memory-config',
        type: 'default',
        style: { stroke: '#8b5cf6' }, // Purple for memory
      };

      try {
        // Create all configuration nodes in parallel
        await Promise.all([
          createNode(modelNode),
          createNode(toolsNode),
          createNode(memoryNode),
        ]);

        // Create edges in parallel
        await Promise.all([
          createEdge(modelEdge),
          createEdge(toolsEdge),
          createEdge(memoryEdge),
        ]);

        // Update agent node to reference the configuration nodes
        const agentNode = nodes.find((n) => n.id === agentId);
        if (agentNode) {
          await updateNode(agentId, {
            data: {
              ...agentNode.data,
              modelConfig: modelId,
              toolsConfig: toolsId,
              memoryConfig: memoryId,
            },
          });
        }

        // Broadcast all changes
        broadcastNodeChange('nodeCreated', { node: modelNode });
        broadcastNodeChange('nodeCreated', { node: toolsNode });
        broadcastNodeChange('nodeCreated', { node: memoryNode });
        broadcastNodeChange('edgeCreated', { edge: modelEdge });
        broadcastNodeChange('edgeCreated', { edge: toolsEdge });
        broadcastNodeChange('edgeCreated', { edge: memoryEdge });
      } catch (err) {
        console.error('Failed to create agent configuration nodes:', err);
      }
    },
    [createNode, createEdge, updateNode, broadcastNodeChange, nodes]
  );

  // Assistant tool callbacks
  const handleAddNode = useCallback(
    async ({ type, label, parameters, position }) => {
      const id = crypto.randomUUID();
      const data = {
        label: label || type,
        nodeTypeLabel: label || type,
        ...(parameters || {}),
      };
      const pos = position || { x: 100, y: 100 };
      const newNode = { id, type, data, position: pos };

      try {
        await createNode(newNode);
        broadcastNodeChange('nodeCreated', { node: newNode });

        // Auto-create configuration nodes for agent nodes
        if (type === 'agent') {
          await createAgentConfigurationNodes(id, pos);
        }

        return { node_id: id };
      } catch (err) {
        console.error('Failed to add node:', err);
        throw err;
      }
    },
    [createNode, broadcastNodeChange, createAgentConfigurationNodes]
  );

  const handleConnectNodes = useCallback(
    async ({ source_id, target_id }) => {
      const edgeId = `edge-${crypto.randomUUID()}`;
      const newEdge = { id: edgeId, source: source_id, target: target_id };

      try {
        await createEdge(newEdge);
        broadcastNodeChange('edgeCreated', { edge: newEdge });
        return {};
      } catch (err) {
        console.error('Failed to connect nodes:', err);
        throw err;
      }
    },
    [createEdge, broadcastNodeChange]
  );

  const handleGetWorkflow = useCallback(() => {
    return { nodes, edges };
  }, [nodes, edges]);

  // Add node from modal
  const handleModalAddNode = useCallback(
    async (nodeType, nodeDefs) => {
      const id = crypto.randomUUID();
      const nodeDef = nodeDefs.find((def) => def.type === nodeType);
      const newNode = {
        id,
        type: nodeType,
        data: {
          label: nodeDef?.label || nodeType,
          nodeTypeLabel: nodeDef?.label || nodeType,
        },
        position: { x: 100, y: 100 },
      };

      try {
        await createNode(newNode);
        broadcastNodeChange('nodeCreated', { node: newNode });
        return newNode;
      } catch (err) {
        console.error('Failed to add node:', err);
        throw err;
      }
    },
    [createNode, broadcastNodeChange]
  );

  return {
    handleNodeDelete,
    handleEdgeDelete,
    createAgentConfigurationNodes,
    handleAddNode,
    handleConnectNodes,
    handleGetWorkflow,
    handleModalAddNode,
  };
}
