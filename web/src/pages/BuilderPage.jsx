import { useEffect, useState, useCallback, useMemo, useRef } from 'react';
import axios from 'axios';
import ReactFlow, { Background, Controls, MiniMap } from 'reactflow';
import IfNode from '../components/IfNode';
import DefaultNode from '../components/DefaultNode';
import TriggerNode from '../components/TriggerNode';
import HttpRequestNode from '../components/HttpRequestNode';
import AgentNode from '../components/AgentNode';
import ModelNode from '../components/ModelNode';
import ToolsNode from '../components/ToolsNode';
import MemoryNode from '../components/MemoryNode';
import OpenAIModelNode from '../components/OpenAIModelNode';
import AnthropicModelNode from '../components/AnthropicModelNode';
import LocalMemoryNode from '../components/LocalMemoryNode';
import ConfigSelectionDialog from '../components/ConfigSelectionDialog';
import NodeDetailsPanel from '../components/NodeDetailsPanel';
import NodeModal from '../components/NodeModal';
import 'reactflow/dist/style.css';
import ChatAssistant from '../components/ChatAssistant';
import WorkflowToolbar from '../components/WorkflowToolbar';
import WorkflowSidebar from '../components/WorkflowSidebar';
import { useWorkflowState } from '../hooks/useWorkflowState';
import { useWebSocket } from '../hooks/useWebSocket';
import { useNodeManagement } from '../hooks/useNodeManagement';
import { isValidConnection } from '../utils/connectionTypes';
import CustomEdge from '../components/CustomEdge';

function BuilderPage({ agentId }) {
  // Use the new workflow state hook with auto-persistence
  const {
    nodes,
    edges,
    loading,
    error,
    isDirty,
    isDraft,
    isSaving,
    lastSaved,
    saveError,
    createNode,
    updateNode,
    deleteNode,
    createEdge,
    deleteEdge,
    applyNodeChanges,
    applyEdgeChanges,
    testDraftNode,
    deployDraft,
    saveVersion,
    clearError,
  } = useWorkflowState(agentId);

  // UI state
  const [modalOpen, setModalOpen] = useState(false);
  const [nodeModalOpen, setNodeModalOpen] = useState(false);
  const [search, setSearch] = useState('');
  const [sidebarTab, setSidebarTab] = useState(null);
  const [testing, setTesting] = useState(false);
  const [isLiveMode, setIsLiveMode] = useState(false);
  const [selectedNodeId, setSelectedNodeId] = useState(null);
  const [validationErrors, setValidationErrors] = useState({});
  const [viewMode, setViewMode] = useState('editor'); // 'editor', 'executions'
  const [executions, setExecutions] = useState([]);
  const [selectedExecution, setSelectedExecution] = useState(null);
  const [loadingExecutions, setLoadingExecutions] = useState(false);

  // Config selection dialog state
  const [configDialogOpen, setConfigDialogOpen] = useState(false);
  const [configDialogType, setConfigDialogType] = useState(null);
  const [configDialogAgentId, setConfigDialogAgentId] = useState(null);
  const [configDialogHandleId, setConfigDialogHandleId] = useState(null);

  // Node definitions and triggers
  const [nodeDefs, setNodeDefs] = useState([]);
  const [triggers, setTriggers] = useState([]);

  // WebSocket for collaboration
  const clientId = useMemo(() => crypto.randomUUID(), []);

  // Local state for WebSocket updates (separate from workflow state)
  const [wsNodes, setWsNodes] = useState([]);
  const [wsEdges, setWsEdges] = useState([]);

  // Sync workflow state to WebSocket state
  useEffect(() => {
    setWsNodes(nodes);
    setWsEdges(edges);
  }, [nodes, edges]);

  // WebSocket callbacks for handling real-time updates
  const handleNodeUpdated = useCallback((nodeId, data) => {
    setWsNodes((prev) =>
      prev.map((n) => (n.id === nodeId ? { ...n, ...data } : n))
    );
  }, []);

  const handleNodeCreated = useCallback((node) => {
    setWsNodes((prev) => [...prev, node]);
  }, []);

  const handleNodeDeleted = useCallback((nodeId) => {
    setWsNodes((prev) => prev.filter((n) => n.id !== nodeId));
  }, []);

  const handleEdgeCreated = useCallback((edge) => {
    setWsEdges((prev) => [...prev, edge]);
  }, []);

  const handleEdgeDeleted = useCallback((edgeId) => {
    setWsEdges((prev) => prev.filter((e) => e.id !== edgeId));
  }, []);

  const handleNodeExecution = useCallback((nodeId, phase) => {
    if (phase === 'start') {
      setWsNodes((nds) =>
        nds.map((n) =>
          n.id === nodeId ? { ...n, data: { ...n.data, status: 'running' } } : n
        )
      );
    } else if (phase === 'end') {
      // Clear status after a short delay for visual feedback
      setTimeout(() => {
        setWsNodes((nds) =>
          nds.map((n) =>
            n.id === nodeId
              ? { ...n, data: { ...n.data, status: undefined } }
              : n
          )
        );
      }, 500);
    }
  }, []);

  // Use WebSocket hook for collaborative editing
  const { broadcastNodeChange } = useWebSocket(agentId, clientId, {
    onNodeUpdated: handleNodeUpdated,
    onNodeCreated: handleNodeCreated,
    onNodeDeleted: handleNodeDeleted,
    onEdgeCreated: handleEdgeCreated,
    onEdgeDeleted: handleEdgeDeleted,
    onNodeExecution: handleNodeExecution,
  });

  // Use node management hook for CRUD operations
  const {
    handleNodeDelete,
    handleEdgeDelete,
    createAgentConfigurationNodes,
    handleAddNode,
    handleConnectNodes,
    handleGetWorkflow,
    handleModalAddNode,
  } = useNodeManagement(broadcastNodeChange);

  // Open config selection dialog
  const openConfigDialog = useCallback((agentNodeId, configType, handleId) => {
    setConfigDialogAgentId(agentNodeId);
    setConfigDialogType(configType);
    setConfigDialogHandleId(handleId);
    setConfigDialogOpen(true);
  }, []);

  // Handle config selection from dialog
  const handleConfigSelection = useCallback(
    async (configOption) => {
      const agentNode = nodes.find((n) => n.id === configDialogAgentId);
      if (!agentNode) return;

      const configId = crypto.randomUUID();

      // Position the new config node below the agent
      const configPosition = {
        x:
          agentNode.position.x -
          50 +
          (configDialogType === 'model'
            ? -100
            : configDialogType === 'tools'
              ? 0
              : 100),
        y: agentNode.position.y + 150,
      };

      const configNode = {
        id: configId,
        type: configOption.type,
        data: {
          label: configOption.label,
          nodeTypeLabel: configOption.label,
          ...configOption.defaultData,
        },
        position: configPosition,
      };

      // Create edge with proper handle IDs
      const configEdge = {
        id: `edge-${configOption.type}-${crypto.randomUUID()}`,
        source: configId,
        sourceHandle: 'config-out',
        target: configDialogAgentId,
        targetHandle: configDialogHandleId,
        type: 'default',
      };

      try {
        // Create the configuration node
        await createNode(configNode);
        await createEdge(configEdge);

        // Update the agent node to reference the new configuration
        const configFieldMap = {
          model: 'modelConfig',
          tools: 'toolsConfig',
          memory: 'memoryConfig',
        };

        await updateNode(configDialogAgentId, {
          data: {
            ...agentNode.data,
            [configFieldMap[configDialogType]]: configId,
          },
        });

        // Broadcast changes
        broadcastNodeChange('nodeCreated', { node: configNode });
        broadcastNodeChange('edgeCreated', { edge: configEdge });
      } catch (err) {
        console.error('Failed to create configuration node:', err);
      }
    },
    [
      nodes,
      createNode,
      createEdge,
      updateNode,
      broadcastNodeChange,
      configDialogAgentId,
      configDialogType,
      configDialogHandleId,
    ]
  );

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
  }, [nodeDefs, agentId, openConfigDialog, handleNodeDelete]);

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

  // Load executions when switching to executions mode
  useEffect(() => {
    if (viewMode === 'executions') {
      setLoadingExecutions(true);
      axios
        .get(`/api/agents/${agentId}/runs`)
        .then((res) => {
          setExecutions(res.data);
          if (res.data.length > 0 && !selectedExecution) {
            setSelectedExecution(res.data[0]);
          }
        })
        .catch((err) => {
          console.error('Failed to load executions:', err);
          setExecutions([]);
        })
        .finally(() => setLoadingExecutions(false));
    }
  }, [viewMode, agentId, selectedExecution]);

  // Node categories
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

  // Selected node and its definition
  const selectedNode = useMemo(
    () => wsNodes.find((n) => n.id === selectedNodeId),
    [wsNodes, selectedNodeId]
  );
  const selectedNodeDef = useMemo(
    () => nodeDefs.find((def) => def.type === selectedNode?.type),
    [nodeDefs, selectedNode]
  );

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

  // Use WebSocket state for display (includes real-time updates from other clients)
  const displayedNodes = useMemo(
    () =>
      wsNodes.map((n) => {
        const errs = validationErrors[n.id];
        const hasError = Array.isArray(errs) && errs.length > 0;
        return { ...n, data: { ...n.data, error: hasError } };
      }),
    [wsNodes, validationErrors]
  );

  // ReactFlow event handlers with collaborative updates
  const onNodesChange = useCallback(
    (changes) => {
      if (isLiveMode || viewMode === 'executions') return;

      // Apply changes locally first
      applyNodeChanges(changes);

      // Broadcast to other clients
      changes.forEach((change) => {
        switch (change.type) {
          case 'position':
            if (change.position) {
              broadcastNodeChange('nodeUpdated', {
                nodeId: change.id,
                data: { position: change.position },
              });
            }
            break;
          case 'remove':
            broadcastNodeChange('nodeDeleted', { nodeId: change.id });
            break;
        }
      });
    },
    [applyNodeChanges, broadcastNodeChange, isLiveMode, viewMode]
  );

  const onEdgesChange = useCallback(
    (changes) => {
      if (isLiveMode || viewMode === 'executions') return;

      applyEdgeChanges(changes);

      changes.forEach((change) => {
        if (change.type === 'remove') {
          broadcastNodeChange('edgeDeleted', { edgeId: change.id });
        }
      });
    },
    [applyEdgeChanges, broadcastNodeChange, isLiveMode, viewMode]
  );

  const onConnect = useCallback(
    async (params) => {
      if (isLiveMode || viewMode === 'executions') return;

      // Find source and target nodes
      const sourceNode = displayedNodes.find((n) => n.id === params.source);
      const targetNode = displayedNodes.find((n) => n.id === params.target);

      if (!sourceNode || !targetNode) {
        console.error('Source or target node not found');
        return;
      }

      // Validate connection type compatibility
      const isValid = isValidConnection(
        params.sourceHandle,
        params.targetHandle,
        sourceNode.type,
        targetNode.type
      );

      if (!isValid) {
        // Silently prevent invalid connections
        return;
      }

      const edgeId = `edge-${crypto.randomUUID()}`;
      const newEdge = { ...params, id: edgeId };

      try {
        await createEdge(newEdge);
        broadcastNodeChange('edgeCreated', { edge: newEdge });
      } catch (err) {
        console.error('Failed to create edge:', err);
      }
    },
    [createEdge, broadcastNodeChange, isLiveMode, viewMode, displayedNodes]
  );

  // Test run functionality
  const onTestRun = useCallback(async () => {
    if (isLiveMode) {
      setWsNodes((nds) =>
        nds.map((n) => ({ ...n, data: { ...n.data, status: undefined } }))
      );
    }
    setTesting(true);
    try {
      await axios.post(`/api/agents/${agentId}/runs/test`);
    } catch (err) {
      console.error('test run failed', err);
      alert('Test run failed');
    } finally {
      setTesting(false);
    }
  }, [agentId, isLiveMode]);

  // Node double-click to rename
  const onNodeDoubleClick = useCallback(
    async (event, node) => {
      if (isLiveMode) return;
      setSelectedNodeId(node.id);
      setNodeModalOpen(true);
    },
    [isLiveMode]
  );

  // Node click handler for selection (works in both modes)
  const onNodeClick = useCallback((event, node) => {
    setSelectedNodeId(node.id);
    setNodeModalOpen(true);
  }, []);

  // Save functionality with validation
  const save = useCallback(async () => {
    // Validate nodes before saving
    const errorsMap = {};
    nodes.forEach((n) => {
      if (n.type === 'http_request') {
        const url = n.data.url || '';
        const method = n.data.method || '';
        if (!url.trim()) {
          (errorsMap[n.id] = errorsMap[n.id] || []).push(
            `Node "${n.data.label || n.id}" is missing a URL`
          );
        }
        if (!method.trim()) {
          (errorsMap[n.id] = errorsMap[n.id] || []).push(
            `Node "${n.data.label || n.id}" is missing a method`
          );
        }
      }
    });

    // Validate async webhook flows
    const nodeMap = Object.fromEntries(nodes.map((n) => [n.id, n]));
    nodes.forEach((n) => {
      if (n.type === 'webhook' && n.data.mode === 'async') {
        const visited = new Set();
        const queue = [n.id];
        let found = false;
        while (queue.length && !found) {
          const curr = queue.shift();
          visited.add(curr);
          edges.forEach((e) => {
            if (e.source === curr) {
              const tgt = e.target;
              if (visited.has(tgt)) return;
              const child = nodeMap[tgt];
              if (child) {
                if (child.type === 'http_response') {
                  found = true;
                  return;
                }
                queue.push(tgt);
              }
            }
          });
        }
        if (!found) {
          (errorsMap[n.id] = errorsMap[n.id] || []).push(
            `Async Webhook node "${n.data.label || n.id}" must be followed by a Webhook Response node`
          );
        }
      }
    });

    if (Object.keys(errorsMap).length > 0) {
      setValidationErrors(errorsMap);
      return;
    }

    setValidationErrors({});
    try {
      await saveVersion();
      alert('Saved!');

      // Refresh triggers
      try {
        const res = await axios.get('/api/triggers');
        setTriggers(res.data);
      } catch (err) {
        console.error('refresh triggers failed:', err);
      }
    } catch (err) {
      console.error(err);
      alert('Save failed');
    }
  }, [nodes, edges, saveVersion]);

  // Auto-layout handler that excludes configuration nodes
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

      // Temporarily remove configuration nodes from the workflow state
      // This prevents them from being sent to the backend auto-layout
      const configNodeIds = configNodes.map((n) => n.id);
      const tempRemovedNodes = [];

      for (const nodeId of configNodeIds) {
        tempRemovedNodes.push({
          nodeId,
          position: nodes.find((n) => n.id === nodeId)?.position,
        });
        // Don't actually delete, just mark them to exclude from layout
      }

      // Create a custom auto-layout that only affects workflow nodes
      // We'll implement a simple client-side layout for now
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

      alert('Layout updated!');
    } catch (err) {
      console.error('Auto-layout failed:', err);
      alert('Auto-layout failed');
    }
  }, [nodes, edges, wsNodes, updateNode]);

  // Guard against undefined agentId (after all hooks)
  if (!agentId) {
    return (
      <div className="flex items-center justify-center h-screen">
        Loading...
      </div>
    );
  }

  // Loading state
  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen">
        Loading workflow...
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-red-500">
          <p>Error loading workflow: {error}</p>
          <button
            onClick={clearError}
            className="mt-2 px-4 py-2 bg-red-500 text-white rounded"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="h-screen flex">
      {/* Main workflow area */}
      <div className="flex-1 relative">
        <ReactFlow
          data-testid="react-flow"
          nodes={displayedNodes}
          edges={wsEdges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onNodeClick={onNodeClick}
          onNodeDoubleClick={onNodeDoubleClick}
          nodeTypes={nodeTypes}
          edgeTypes={{
            default: (props) => (
              <CustomEdge {...props} onDelete={handleEdgeDelete} />
            ),
          }}
          isValidConnection={(connection) => {
            const sourceNode = displayedNodes.find(
              (n) => n.id === connection.source
            );
            const targetNode = displayedNodes.find(
              (n) => n.id === connection.target
            );
            if (!sourceNode || !targetNode) return false;
            return isValidConnection(
              connection.sourceHandle,
              connection.targetHandle,
              sourceNode.type,
              targetNode.type
            );
          }}
          fitView
        >
          <Background />
          <Controls />
          <MiniMap />
        </ReactFlow>

        {/* Top toolbar */}
        <WorkflowToolbar
          isDraft={isDraft}
          isDirty={isDirty}
          isSaving={isSaving}
          lastSaved={lastSaved}
          saveError={saveError}
          viewMode={viewMode}
          testing={testing}
          isLiveMode={isLiveMode}
          sidebarTab={sidebarTab}
          onViewModeChange={setViewMode}
          onAddNode={() => setModalOpen(true)}
          onSave={async () => {
            if (isDraft) {
              // Deploy draft as new version
              try {
                await deployDraft('Deployed from draft');
                alert('Draft deployed successfully!');
              } catch (err) {
                alert(`Deploy failed: ${err.message}`);
              }
            } else {
              // Traditional save
              await save();
            }
          }}
          onTestRun={onTestRun}
          onAutoLayout={handleAutoLayout}
          onToggleSidebar={(tab) =>
            setSidebarTab(sidebarTab === tab ? null : tab)
          }
          onToggleLiveMode={() => setIsLiveMode(!isLiveMode)}
        />
      </div>

      {/* Sidebar */}
      <WorkflowSidebar
        isVisible={sidebarTab || viewMode === 'executions'}
        sidebarTab={sidebarTab}
        viewMode={viewMode}
        selectedNode={selectedNode}
        selectedNodeDef={selectedNodeDef}
        selectedExecution={selectedExecution}
        executions={executions}
        loadingExecutions={loadingExecutions}
        isLiveMode={isLiveMode}
        isDraft={isDraft}
        agentId={agentId}
        triggersMap={triggersMap}
        onExecutionSelect={setSelectedExecution}
        onNodeChange={(key, value) => {
          updateNode(selectedNodeId, {
            data: { ...selectedNode.data, [key]: value },
          });
        }}
        onNodeExecute={async (inputData) => {
          try {
            if (isDraft) {
              // Use draft testing API for instant feedback
              const result = await testDraftNode(selectedNodeId, inputData);
              if (result.success) {
                return result.result;
              } else {
                throw new Error(result.error);
              }
            } else {
              // Fall back to old API for production versions
              const res = await axios.post(
                `/api/agents/${agentId}/nodes/${selectedNodeId}/execute`,
                inputData
              );
              return res.data.output;
            }
          } catch (err) {
            console.error('Execution failed:', err);
            throw err;
          }
        }}
        onChatClose={() => setSidebarTab(null)}
        onAddNode={handleAddNode}
        onConnectNodes={handleConnectNodes}
        onGetWorkflow={handleGetWorkflow}
      />

      {/* Status indicators */}
      {isDirty && (
        <div className="absolute bottom-4 right-4 bg-yellow-500 text-white px-3 py-1 rounded">
          Unsaved changes
        </div>
      )}

      {/* Add Node Modal */}
      {modalOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-96 max-h-96 overflow-y-auto">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-lg font-bold">Add Node</h2>
              <button
                onClick={() => setModalOpen(false)}
                className="text-gray-500 hover:text-gray-700"
              >
                ✕
              </button>
            </div>

            <input
              type="text"
              placeholder="Search nodes..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full border rounded px-3 py-2 mb-4"
            />

            {categories
              .filter(({ types }) =>
                types.some(
                  (type) =>
                    type.label.toLowerCase().includes(search.toLowerCase()) ||
                    type.type.toLowerCase().includes(search.toLowerCase())
                )
              )
              .map(({ category, types }) => (
                <div key={category} className="mb-4">
                  <h3 className="font-semibold text-sm text-gray-600 mb-2">
                    {category}
                  </h3>
                  {types
                    .filter(
                      (type) =>
                        type.label
                          .toLowerCase()
                          .includes(search.toLowerCase()) ||
                        type.type.toLowerCase().includes(search.toLowerCase())
                    )
                    .map((type) => (
                      <button
                        key={type.type}
                        onClick={() => handleModalAddNode(type.type)}
                        className="w-full text-left p-2 hover:bg-gray-100 rounded"
                      >
                        <div className="font-medium">{type.label}</div>
                        {type.description && (
                          <div className="text-sm text-gray-500">
                            {type.description}
                          </div>
                        )}
                      </button>
                    ))}
                </div>
              ))}
          </div>
        </div>
      )}

      {/* Node editing modal */}
      <NodeModal
        node={selectedNode}
        nodeDef={selectedNodeDef}
        nodes={nodes}
        isOpen={nodeModalOpen}
        viewMode={viewMode}
        selectedExecution={selectedExecution}
        agentId={agentId}
        onClose={() => {
          setNodeModalOpen(false);
          setSelectedNodeId(null);
        }}
        onChange={(key, value) => {
          updateNode(selectedNodeId, {
            data: { ...selectedNode.data, [key]: value },
          });
        }}
        onExecute={async (inputData) => {
          try {
            // Execute single node
            const res = await axios.post(
              `/api/agents/${agentId}/nodes/${selectedNodeId}/execute`,
              inputData
            );
            return res.data.output;
          } catch (err) {
            console.error('Execution failed:', err);
            throw err;
          }
        }}
        onSave={save}
      />

      {/* Config Selection Dialog */}
      <ConfigSelectionDialog
        isOpen={configDialogOpen}
        configType={configDialogType}
        onClose={() => setConfigDialogOpen(false)}
        onSelect={handleConfigSelection}
      />
    </div>
  );
}

export default BuilderPage;
