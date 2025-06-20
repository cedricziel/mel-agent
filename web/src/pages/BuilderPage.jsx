import { useEffect, useState, useCallback, useMemo } from 'react';
import { workflowRunsApi, nodeTypesApi, workflowsApi } from '../api/client';
import ReactFlow, { Background, Controls, MiniMap } from 'reactflow';
import ConfigSelectionDialog from '../components/ConfigSelectionDialog';
import NodeModal from '../components/NodeModal';
import AddNodeModal from '../components/AddNodeModal';
import 'reactflow/dist/style.css';
import WorkflowToolbar from '../components/WorkflowToolbar';
import WorkflowSidebar from '../components/WorkflowSidebar';
import { useWorkflowState } from '../hooks/useWorkflowState';
import { useWebSocket } from '../hooks/useWebSocket';
import { useNodeManagement } from '../hooks/useNodeManagement';
import { useNodeTypes } from '../hooks/useNodeTypes.jsx';
import { useValidation } from '../hooks/useValidation';
import { useAutoLayout } from '../hooks/useAutoLayout';
import { isValidConnection } from '../utils/connectionTypes';
import CustomEdge from '../components/CustomEdge';

function BuilderPage({ agentId }) {
  // Use the new workflow state hook with auto-persistence
  const workflowState = useWorkflowState(agentId);
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
  } = workflowState;

  // UI state
  const [modalOpen, setModalOpen] = useState(false);
  const [nodeModalOpen, setNodeModalOpen] = useState(false);
  const [sidebarTab, setSidebarTab] = useState(null);
  const [testing, setTesting] = useState(false);
  const [isLiveMode, setIsLiveMode] = useState(false);
  const [selectedNodeId, setSelectedNodeId] = useState(null);
  const [viewMode, setViewMode] = useState('editor'); // 'editor', 'executions'
  const [executions, setExecutions] = useState([]);
  const [selectedExecution, setSelectedExecution] = useState(null);
  const [loadingExecutions, setLoadingExecutions] = useState(false);

  // Config selection dialog state
  const [configDialogOpen, setConfigDialogOpen] = useState(false);
  const [configDialogType, setConfigDialogType] = useState(null);
  const [configDialogAgentId, setConfigDialogAgentId] = useState(null);
  const [configDialogHandleId, setConfigDialogHandleId] = useState(null);

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
  } = useNodeManagement(agentId, broadcastNodeChange, workflowState);

  // Open config selection dialog
  const openConfigDialog = useCallback((agentNodeId, configType, handleId) => {
    setConfigDialogAgentId(agentNodeId);
    setConfigDialogType(configType);
    setConfigDialogHandleId(handleId);
    setConfigDialogOpen(true);
  }, []);

  // Node click handler for selection (works in both modes)
  const onNodeClick = useCallback((event, node) => {
    setSelectedNodeId(node.id);
    setNodeModalOpen(true);
  }, []);

  // Use node types hook for node definitions and types
  const {
    nodeDefs,
    triggers,
    nodeTypes,
    categories,
    triggersMap,
    refreshTriggers,
  } = useNodeTypes(agentId, handleNodeDelete, openConfigDialog, onNodeClick);

  // Use validation hook
  const { validationErrors, validateWorkflow, clearValidationErrors } =
    useValidation();

  // Use auto-layout hook
  const { handleAutoLayout } = useAutoLayout(nodes, edges, wsNodes, updateNode);

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

  // Load executions when switching to executions mode
  useEffect(() => {
    if (viewMode === 'executions') {
      setLoadingExecutions(true);
      workflowRunsApi
        .listWorkflowRuns({ workflow_id: agentId })
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

  // Handle escape key to close node details sidebar
  useEffect(() => {
    const handleEscapeKey = (event) => {
      if (event.key === 'Escape') {
        if (sidebarTab === 'details') {
          setSidebarTab(null);
          setSelectedNodeId(null);
        }
      }
    };

    document.addEventListener('keydown', handleEscapeKey);
    return () => document.removeEventListener('keydown', handleEscapeKey);
  }, [sidebarTab]);

  // Selected node and its definition
  const selectedNode = useMemo(
    () => wsNodes.find((n) => n.id === selectedNodeId),
    [wsNodes, selectedNodeId]
  );
  const selectedNodeDef = useMemo(
    () => nodeDefs.find((def) => def.type === selectedNode?.type),
    [nodeDefs, selectedNode]
  );

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
                position: change.position,
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
      await workflowRunsApi.createWorkflowRun({
        workflow_id: agentId,
        input_data: {},
      });
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

  // Save functionality with validation
  const save = useCallback(async () => {
    const isValid = validateWorkflow(nodes, edges);
    if (!isValid) {
      return;
    }

    clearValidationErrors();
    try {
      await saveVersion();
      alert('Saved!');
      await refreshTriggers();
    } catch (err) {
      console.error(err);
      alert('Save failed');
    }
  }, [
    nodes,
    edges,
    validateWorkflow,
    clearValidationErrors,
    saveVersion,
    refreshTriggers,
  ]);

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
          onAutoLayout={async () => {
            const result = await handleAutoLayout();
            if (result.success) {
              alert('Layout updated!');
            } else {
              alert('Auto-layout failed');
            }
          }}
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
              // Use workflow draft node testing API
              const res = await workflowsApi.testWorkflowDraftNode(
                agentId,
                selectedNodeId,
                { testData: inputData }
              );
              return res.data.output || res.data;
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
      <AddNodeModal
        isOpen={modalOpen}
        categories={categories}
        onClose={() => setModalOpen(false)}
        onAddNode={(nodeType) => handleModalAddNode(nodeType, nodeDefs)}
      />

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
            // Execute single node using workflow draft testing
            const res = await workflowsApi.testWorkflowDraftNode(
              agentId,
              selectedNodeId,
              { testData: inputData }
            );
            return res.data.output || res.data;
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
