import { useEffect, useState, useCallback, useMemo, useRef } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import ReactFlow, {
  Background,
  Controls,
  addEdge,
  MiniMap,
} from "reactflow";
import IfNode from "../components/IfNode";
import DefaultNode from "../components/DefaultNode";
import TriggerNode from "../components/TriggerNode";
import HttpRequestNode from "../components/HttpRequestNode";
import NodeDetailsPanel from "../components/NodeDetailsPanel";
import NodeModal from "../components/NodeModal";
import "reactflow/dist/style.css";
import ChatAssistant from "../components/ChatAssistant";
import { useWorkflowState } from "../hooks/useWorkflowState";

function BuilderPage({ agentId }) {
  const navigate = useNavigate();
  
  // Use the new workflow state hook with auto-persistence
  const {
    workflow,
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
    updateWorkflow,
    autoLayout,
    applyNodeChanges,
    applyEdgeChanges,
    testDraftNode,
    deployDraft,
    saveNow,
    saveVersion,
    clearError
  } = useWorkflowState(agentId);

  // UI state
  const [modalOpen, setModalOpen] = useState(false);
  const [nodeModalOpen, setNodeModalOpen] = useState(false);
  const [search, setSearch] = useState('');
  const [sidebarTab, setSidebarTab] = useState(null);
  const [addingFromNodeId, setAddingFromNodeId] = useState(null);
  const [testing, setTesting] = useState(false);
  const [testRunResult, setTestRunResult] = useState(null);
  const [isLiveMode, setIsLiveMode] = useState(false);
  const [selectedNodeId, setSelectedNodeId] = useState(null);
  const [validationErrors, setValidationErrors] = useState({});
  const [viewMode, setViewMode] = useState('editor'); // 'editor', 'executions'
  const [executions, setExecutions] = useState([]);
  const [selectedExecution, setSelectedExecution] = useState(null);
  const [loadingExecutions, setLoadingExecutions] = useState(false);

  // Node definitions and triggers
  const [nodeDefs, setNodeDefs] = useState([]);
  const [triggers, setTriggers] = useState([]);

  // WebSocket for collaboration
  const clientId = useMemo(() => crypto.randomUUID(), []);
  const wsRef = useRef(null);
  const execTimersRef = useRef({});
  
  // Local state for WebSocket updates (separate from workflow state)
  const [wsNodes, setWsNodes] = useState([]);
  const [wsEdges, setWsEdges] = useState([]);

  // Sync workflow state to WebSocket state
  useEffect(() => {
    setWsNodes(nodes);
    setWsEdges(edges);
  }, [nodes, edges]);

  // Load node definitions
  useEffect(() => {
    axios.get('/api/node-types')
      .then((res) => setNodeDefs(res.data))
      .catch((err) => console.error('fetch node-types failed:', err));
  }, []);

  // Load triggers
  useEffect(() => {
    axios.get('/api/triggers')
      .then((res) => setTriggers(res.data))
      .catch((err) => console.error('fetch triggers failed:', err));
  }, []);

  // Load executions when switching to executions mode
  useEffect(() => {
    if (viewMode === 'executions') {
      setLoadingExecutions(true);
      axios.get(`/api/agents/${agentId}/runs`)
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
    return Object.entries(map).map(([category, types]) => ({ category, types }));
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
    () => wsNodes.map((n) => {
      const errs = validationErrors[n.id];
      const hasError = Array.isArray(errs) && errs.length > 0;
      return { ...n, data: { ...n.data, error: hasError } };
    }),
    [wsNodes, validationErrors]
  );

  // WebSocket setup for collaborative editing
  useEffect(() => {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProtocol}//${window.location.host}/api/ws/agents/${agentId}`;
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        if (msg.clientId === clientId) return;

        switch (msg.type) {
          case 'nodeUpdated':
            // Handle individual node updates from other clients
            if (msg.workflowId === agentId) {
              // Update local WebSocket state without triggering API call
              setWsNodes(prev => prev.map(n => 
                n.id === msg.nodeId ? { ...n, ...msg.data } : n
              ));
            }
            break;
          case 'nodeCreated':
            if (msg.workflowId === agentId) {
              setWsNodes(prev => [...prev, msg.node]);
            }
            break;
          case 'nodeDeleted':
            if (msg.workflowId === agentId) {
              setWsNodes(prev => prev.filter(n => n.id !== msg.nodeId));
            }
            break;
          case 'edgeCreated':
            if (msg.workflowId === agentId) {
              setWsEdges(prev => [...prev, msg.edge]);
            }
            break;
          case 'edgeDeleted':
            if (msg.workflowId === agentId) {
              setWsEdges(prev => prev.filter(e => e.id !== msg.edgeId));
            }
            break;
          case 'nodeExecution': {
            // Runtime status updates for Live mode
            const { nodeId, phase } = msg;
            if (phase === 'start') {
              execTimersRef.current[nodeId] = { start: Date.now(), timeoutId: null };
              setWsNodes((nds) =>
                nds.map((n) =>
                  n.id === nodeId ? { ...n, data: { ...n.data, status: 'running' } } : n
                )
              );
            } else if (phase === 'end') {
              const timer = execTimersRef.current[nodeId];
              const now = Date.now();
              const clearStatus = () => {
                setWsNodes((nds) =>
                  nds.map((n) =>
                    n.id === nodeId
                      ? { ...n, data: { ...n.data, status: undefined } }
                      : n
                  )
                );
                delete execTimersRef.current[nodeId];
              };
              if (timer) {
                const elapsed = now - timer.start;
                const remaining = 500 - elapsed;
                if (timer.timeoutId) clearTimeout(timer.timeoutId);
                if (remaining <= 0) {
                  clearStatus();
                } else {
                  const tid = setTimeout(clearStatus, remaining);
                  execTimersRef.current[nodeId].timeoutId = tid;
                }
              } else {
                clearStatus();
              }
            }
            break;
          }
          default:
            console.warn('Unknown message type:', msg.type);
        }
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err);
      }
    };

    ws.onclose = () => { wsRef.current = null; };
    return () => { ws.close(); wsRef.current = null; };
  }, [agentId, clientId]);

  // Broadcast node changes to other clients
  const broadcastNodeChange = useCallback((type, data) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ 
        clientId, 
        type, 
        workflowId: agentId,
        ...data 
      }));
    }
  }, [clientId, agentId]);

  // ReactFlow event handlers with collaborative updates
  const onNodesChange = useCallback(
    (changes) => {
      if (isLiveMode || viewMode === 'executions') return;
      
      // Apply changes locally first
      applyNodeChanges(changes);
      
      // Broadcast to other clients
      changes.forEach(change => {
        switch (change.type) {
          case 'position':
            if (change.position) {
              broadcastNodeChange('nodeUpdated', {
                nodeId: change.id,
                data: { position: change.position }
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
      
      changes.forEach(change => {
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
      
      const edgeId = `edge-${Date.now()}`;
      const newEdge = { ...params, id: edgeId };
      
      try {
        await createEdge(newEdge);
        broadcastNodeChange('edgeCreated', { edge: newEdge });
      } catch (err) {
        console.error('Failed to create edge:', err);
      }
    },
    [createEdge, broadcastNodeChange, isLiveMode, viewMode]
  );

  // Test run functionality
  const onTestRun = useCallback(async () => {
    if (isLiveMode) {
      setWsNodes((nds) => nds.map((n) => ({ ...n, data: { ...n.data, status: undefined } })));
    }
    setTesting(true);
    try {
      const res = await axios.post(`/api/agents/${agentId}/runs/test`);
      if (!isLiveMode) setTestRunResult(res.data);
    } catch (err) {
      console.error('test run failed', err);
      alert('Test run failed');
    } finally {
      setTesting(false);
    }
  }, [agentId, isLiveMode]);

  // Node double-click to rename
  const onNodeDoubleClick = useCallback(async (event, node) => {
    if (isLiveMode) return;
    setSelectedNodeId(node.id);
    setNodeModalOpen(true);
  }, [isLiveMode]);

  // Node click handler for selection (works in both modes)
  const onNodeClick = useCallback((event, node) => {
    setSelectedNodeId(node.id);
    setNodeModalOpen(true);
  }, []);

  // Save functionality with validation
  const save = useCallback(async () => {
    const graph = { nodes, edges };
    
    // Validate nodes before saving
    const errorsMap = {};
    nodes.forEach((n) => {
      if (n.type === 'http_request') {
        const url = n.data.url || '';
        const method = n.data.method || '';
        if (!url.trim()) {
          (errorsMap[n.id] = errorsMap[n.id] || []).push(`Node "${n.data.label || n.id}" is missing a URL`);
        }
        if (!method.trim()) {
          (errorsMap[n.id] = errorsMap[n.id] || []).push(`Node "${n.data.label || n.id}" is missing a method`);
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
            `Async Webhook node \"${n.data.label || n.id}\" must be followed by a Webhook Response node`
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
      alert("Saved!");
      
      // Refresh triggers
      try {
        const res = await axios.get('/api/triggers');
        setTriggers(res.data);
      } catch (err) {
        console.error('refresh triggers failed:', err);
      }
    } catch (err) {
      console.error(err);
      alert("Save failed");
    }
  }, [nodes, edges, saveVersion]);

  // Assistant tool callbacks
  const handleAddNode = useCallback(async ({ type, label, parameters, position }) => {
    const id = Date.now().toString();
    const data = { label: label || type, nodeTypeLabel: label || type, ...(parameters || {}) };
    const pos = position || { x: 100, y: 100 };
    const newNode = { id, type, data, position: pos };
    
    try {
      await createNode(newNode);
      broadcastNodeChange('nodeCreated', { node: newNode });
      return { node_id: id };
    } catch (err) {
      console.error('Failed to add node:', err);
      throw err;
    }
  }, [createNode, broadcastNodeChange]);

  const handleConnectNodes = useCallback(async ({ source_id, target_id }) => {
    const edgeId = `edge-${Date.now()}`;
    const newEdge = { id: edgeId, source: source_id, target: target_id };
    
    try {
      await createEdge(newEdge);
      broadcastNodeChange('edgeCreated', { edge: newEdge });
      return {};
    } catch (err) {
      console.error('Failed to connect nodes:', err);
      throw err;
    }
  }, [createEdge, broadcastNodeChange]);

  const handleGetWorkflow = useCallback(() => {
    return { nodes, edges };
  }, [nodes, edges]);

  // Auto-layout handler
  const handleAutoLayout = useCallback(async () => {
    try {
      await autoLayout();
      alert('Layout updated!');
    } catch (err) {
      console.error('Auto-layout failed:', err);
      alert('Auto-layout failed');
    }
  }, [autoLayout]);

  // Add node from modal
  const handleModalAddNode = useCallback(async (nodeType) => {
    const id = Date.now().toString();
    const nodeDef = nodeDefs.find(def => def.type === nodeType);
    const newNode = {
      id,
      type: nodeType,
      data: { 
        label: nodeDef?.label || nodeType,
        nodeTypeLabel: nodeDef?.label || nodeType
      },
      position: { x: 100, y: 100 }
    };

    try {
      await createNode(newNode);
      broadcastNodeChange('nodeCreated', { node: newNode });
      setModalOpen(false);
    } catch (err) {
      console.error('Failed to add node:', err);
    }
  }, [nodeDefs, createNode, broadcastNodeChange]);

  // Loading state
  if (loading) {
    return <div className="flex items-center justify-center h-screen">Loading workflow...</div>;
  }

  // Error state
  if (error) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-red-500">
          <p>Error loading workflow: {error}</p>
          <button onClick={clearError} className="mt-2 px-4 py-2 bg-red-500 text-white rounded">
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
          nodes={displayedNodes}
          edges={wsEdges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onNodeClick={onNodeClick}
          onNodeDoubleClick={onNodeDoubleClick}
          nodeTypes={{
            default: DefaultNode,
            if: IfNode,
            webhook: (props) => {
              const nodeDef = nodeDefs.find(def => def.type === 'webhook');
              return <TriggerNode {...props} type="webhook" agentId={agentId} icon={nodeDef?.icon} />;
            },
            schedule: (props) => {
              const nodeDef = nodeDefs.find(def => def.type === 'schedule');
              return <TriggerNode {...props} type="schedule" agentId={agentId} icon={nodeDef?.icon} />;
            },
            manual_trigger: (props) => {
              const nodeDef = nodeDefs.find(def => def.type === 'manual_trigger');
              return <TriggerNode {...props} type="manual_trigger" agentId={agentId} icon={nodeDef?.icon} />;
            },
            workflow_trigger: (props) => {
              const nodeDef = nodeDefs.find(def => def.type === 'workflow_trigger');
              return <TriggerNode {...props} type="workflow_trigger" agentId={agentId} icon={nodeDef?.icon} />;
            },
            http_request: HttpRequestNode,
          }}
          fitView
        >
          <Background />
          <Controls />
          <MiniMap />
        </ReactFlow>

        {/* Top toolbar */}
        <div className="absolute top-4 left-4 right-4 flex justify-between items-center">
          <div className="flex gap-2 items-center">
            {/* Draft/Auto-save status */}
            <div className="flex items-center gap-3 mr-4">
              {/* Draft vs Production indicator */}
              <div className={`flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium ${
                isDraft 
                  ? 'bg-yellow-100 text-yellow-800' 
                  : 'bg-green-100 text-green-800'
              }`}>
                <div className={`w-2 h-2 rounded-full ${
                  isDraft ? 'bg-yellow-400' : 'bg-green-400'
                }`}></div>
                {isDraft ? 'Draft' : 'Deployed'}
              </div>
              
              {/* Auto-save status */}
              {isDraft && (
                <div className="flex items-center gap-1 text-xs text-gray-600">
                  {isSaving ? (
                    <>
                      <div className="animate-spin w-3 h-3 border border-blue-500 border-t-transparent rounded-full"></div>
                      Saving...
                    </>
                  ) : saveError ? (
                    <span className="text-red-600">Save failed</span>
                  ) : lastSaved ? (
                    <span>Saved {new Date(lastSaved).toLocaleTimeString()}</span>
                  ) : (
                    <span>Auto-save enabled</span>
                  )}
                </div>
              )}
            </div>
            
            {/* n8n-style toggle switch */}
            <div className="flex bg-gray-100 rounded-lg p-1 mr-4">
              <button
                onClick={() => setViewMode('editor')}
                className={`px-3 py-1 text-sm rounded-md transition-colors ${
                  viewMode === 'editor' 
                    ? 'bg-white text-gray-900 shadow-sm' 
                    : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                Editor
              </button>
              <button
                onClick={() => setViewMode('executions')}
                className={`px-3 py-1 text-sm rounded-md transition-colors ${
                  viewMode === 'executions' 
                    ? 'bg-white text-gray-900 shadow-sm' 
                    : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                Executions
              </button>
            </div>
            
            <button
              onClick={() => setModalOpen(true)}
              disabled={viewMode === 'executions'}
              className={`px-4 py-2 rounded ${
                viewMode === 'executions' 
                  ? 'bg-gray-300 text-gray-500 cursor-not-allowed' 
                  : 'bg-blue-500 text-white'
              }`}
            >
              + Add Node
            </button>
            <button
              onClick={async () => {
                if (isDraft) {
                  // Deploy draft as new version
                  try {
                    await deployDraft('Deployed from draft');
                    alert("Draft deployed successfully!");
                  } catch (err) {
                    alert(`Deploy failed: ${err.message}`);
                  }
                } else {
                  // Traditional save
                  await save();
                }
              }}
              disabled={(!isDirty && !isDraft) || viewMode === 'executions'}
              className={`px-4 py-2 rounded ${
                ((isDirty || isDraft) && viewMode === 'editor') ? 'bg-blue-500 text-white' : 'bg-gray-300 text-gray-500'
              }`}
            >
              {isDraft ? 'Deploy' : 'Save'}
            </button>
            <button
              onClick={onTestRun}
              disabled={testing || viewMode === 'executions'}
              className={`px-4 py-2 rounded ${
                viewMode === 'executions' 
                  ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
                  : 'bg-green-500 text-white'
              }`}
            >
              {testing ? 'Running...' : 'Test Run'}
            </button>
            <button
              onClick={handleAutoLayout}
              disabled={viewMode === 'executions'}
              className={`px-4 py-2 rounded ${
                viewMode === 'executions'
                  ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
                  : 'bg-purple-500 text-white'
              }`}
            >
              Auto Layout
            </button>
          </div>

          <div className="flex gap-2">
            <button
              onClick={() => setSidebarTab(sidebarTab === 'chat' ? null : 'chat')}
              className={`px-4 py-2 rounded ${sidebarTab === 'chat' ? 'bg-blue-500 text-white' : 'bg-gray-300'}`}
            >
              💬 Chat
            </button>
            <button
              onClick={() => setIsLiveMode(!isLiveMode)}
              className={`px-4 py-2 rounded ${isLiveMode ? 'bg-orange-500 text-white' : 'bg-gray-300'}`}
            >
              {isLiveMode ? 'Live Mode' : 'Edit Mode'}
            </button>
          </div>
        </div>
      </div>

      {/* Sidebar */}
      {(sidebarTab || viewMode === 'executions') && (
        <div className="w-80 bg-white border-l shadow-lg h-screen overflow-y-auto">
          {/* Executions panel */}
          {viewMode === 'executions' && (
            <div className="h-full flex flex-col">
              <div className="p-4 border-b">
                <h3 className="text-lg font-semibold">Executions</h3>
                {loadingExecutions && (
                  <div className="text-sm text-gray-500">Loading executions...</div>
                )}
              </div>
              <div className="flex-1 overflow-y-auto">
                {executions.length === 0 && !loadingExecutions ? (
                  <div className="p-4 text-center text-gray-500">
                    No executions found. Run your workflow to see execution history.
                  </div>
                ) : (
                  <div className="p-2">
                    {executions.map((execution) => (
                      <div
                        key={execution.id}
                        onClick={() => setSelectedExecution(execution)}
                        className={`p-3 border rounded mb-2 cursor-pointer hover:bg-gray-50 ${
                          selectedExecution?.id === execution.id ? 'border-blue-500 bg-blue-50' : 'border-gray-200'
                        }`}
                      >
                        <div className="flex items-center justify-between">
                          <span className="text-sm font-medium">
                            {new Date(execution.created_at).toLocaleString()}
                          </span>
                          <span className="px-2 py-1 text-xs bg-green-100 text-green-800 rounded">
                            Completed
                          </span>
                        </div>
                        <div className="text-xs text-gray-500 mt-1">
                          ID: {execution.id.slice(0, 8)}...
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          )}
          
          {sidebarTab === 'details' && selectedNode && selectedNodeDef && viewMode === 'editor' && (
            <NodeDetailsPanel
              node={selectedNode}
              nodeDef={selectedNodeDef}
              readOnly={isLiveMode}
              onChange={(key, value) => {
                updateNode(selectedNodeId, {
                  data: { ...selectedNode.data, [key]: value }
                });
              }}
              onExecute={async (inputData) => {
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
              publicUrl={
                selectedNodeDef.type === 'webhook' && triggersMap[selectedNode.id]
                  ? `${window.location.origin}/api/webhooks/${selectedNode.type}/${triggersMap[selectedNode.id].id}`
                  : undefined
              }
            />
          )}
          {sidebarTab === 'chat' && (
            <ChatAssistant
              inline
              agentId={agentId}
              onAddNode={handleAddNode}
              onConnectNodes={handleConnectNodes}
              onGetWorkflow={handleGetWorkflow}
              onClose={() => setSidebarTab(null)}
            />
          )}
        </div>
      )}

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
                types.some(type => 
                  type.label.toLowerCase().includes(search.toLowerCase()) ||
                  type.type.toLowerCase().includes(search.toLowerCase())
                )
              )
              .map(({ category, types }) => (
                <div key={category} className="mb-4">
                  <h3 className="font-semibold text-sm text-gray-600 mb-2">{category}</h3>
                  {types
                    .filter(type => 
                      type.label.toLowerCase().includes(search.toLowerCase()) ||
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
                          <div className="text-sm text-gray-500">{type.description}</div>
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
            data: { ...selectedNode.data, [key]: value }
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
    </div>
  );
}

export default BuilderPage;