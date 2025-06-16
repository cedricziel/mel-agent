import { useCallback } from 'react';

export function useAutoLayout(nodes, edges, wsNodes, updateNode) {
  const handleAutoLayout = useCallback(async () => {
    try {
      const configNodeTypes = [
        'model',
        'tools',
        'memory',
        'openai_model',
        'anthropic_model',
        'local_memory',
        'workflow_tools',
      ];

      // Separate workflow nodes from configuration nodes
      const workflowNodes = nodes.filter(
        (node) => !configNodeTypes.includes(node.type)
      );
      const configNodes = nodes.filter((node) =>
        configNodeTypes.includes(node.type)
      );

      // Store configuration node positions and their parent relationships
      const configNodeData = new Map();
      configNodes.forEach((configNode) => {
        // Find the agent this config node belongs to by looking at edges
        const targetEdge = edges.find(
          (edge) =>
            edge.source === configNode.id &&
            edge.targetHandle &&
            (edge.targetHandle.includes('config') ||
              edge.targetHandle === 'model-config' ||
              edge.targetHandle === 'tools-config' ||
              edge.targetHandle === 'memory-config')
        );

        if (targetEdge) {
          const agentNode = nodes.find((n) => n.id === targetEdge.target);
          if (agentNode) {
            configNodeData.set(configNode.id, {
              agentId: agentNode.id,
              relativeX: configNode.position.x - agentNode.position.x,
              relativeY: configNode.position.y - agentNode.position.y,
              currentPosition: { ...configNode.position },
            });
          }
        }
      });

      // Create a custom auto-layout that only affects workflow nodes
      const layoutWorkflowNodes = async () => {
        const GRID_SIZE = 200;
        const VERTICAL_SPACING = 150;
        let currentX = 100;
        let currentY = 100;

        // Find trigger nodes (starting points)
        const triggerNodes = workflowNodes.filter((node) =>
          [
            'webhook',
            'schedule',
            'manual_trigger',
            'workflow_trigger',
            'slack',
            'timer',
          ].includes(node.type)
        );

        // Simple left-to-right layout
        const layoutNodes = [
          ...triggerNodes,
          ...workflowNodes.filter(
            (node) =>
              ![
                'webhook',
                'schedule',
                'manual_trigger',
                'workflow_trigger',
                'slack',
                'timer',
              ].includes(node.type)
          ),
        ];

        // Collect all position updates in a batch
        const positionUpdates = [];

        for (let i = 0; i < layoutNodes.length; i++) {
          const node = layoutNodes[i];
          positionUpdates.push({
            id: node.id,
            position: {
              x: currentX,
              y: currentY,
            },
          });

          currentX += GRID_SIZE;
          if ((i + 1) % 4 === 0) {
            // New row every 4 nodes
            currentX = 100;
            currentY += VERTICAL_SPACING;
          }
        }

        // Apply all updates in batches to prevent UI blocking and reduce API calls
        const batchSize = 5; // Update 5 nodes at a time
        for (let i = 0; i < positionUpdates.length; i += batchSize) {
          const batch = positionUpdates.slice(i, i + batchSize);

          // Execute batch updates in parallel
          await Promise.all(
            batch.map(({ id, position }) => updateNode(id, { position }))
          );

          // Add small delay between batches to prevent overwhelming the API
          if (i + batchSize < positionUpdates.length) {
            await new Promise((resolve) => setTimeout(resolve, 50));
          }
        }
      };

      // Apply layout to workflow nodes only
      await layoutWorkflowNodes();

      // Wait a moment for the layout to propagate, then reposition config nodes
      setTimeout(async () => {
        const configUpdates = [];

        configNodeData.forEach((data, configNodeId) => {
          // Get the updated position of the agent node
          const currentNodes = wsNodes.length > 0 ? wsNodes : nodes;
          const agentNode = currentNodes.find((n) => n.id === data.agentId);

          if (agentNode) {
            const newPosition = {
              x: agentNode.position.x + data.relativeX,
              y: agentNode.position.y + data.relativeY,
            };

            configUpdates.push({
              id: configNodeId,
              position: newPosition,
            });
          }
        });

        // Apply config node updates in parallel
        if (configUpdates.length > 0) {
          await Promise.all(
            configUpdates.map(({ id, position }) =>
              updateNode(id, { position })
            )
          );
        }
      }, 500);

      return { success: true };
    } catch (err) {
      console.error('Auto-layout failed:', err);
      return { success: false, error: err.message };
    }
  }, [nodes, edges, wsNodes, updateNode]);

  return {
    handleAutoLayout,
  };
}
