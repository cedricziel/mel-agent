import { useEffect, useState, useCallback, useMemo, useRef } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import ReactFlow, {
  Background,
  Controls,
  addEdge,
  applyNodeChanges,
  applyEdgeChanges,
  MiniMap,
} from "reactflow";
import IfNode from "../components/IfNode";
import DefaultNode from "../components/DefaultNode";
import TriggerNode from "../components/TriggerNode";
import HttpRequestNode from "../components/HttpRequestNode";
import NodeDetailsPanel from "../components/NodeDetailsPanel";
import "reactflow/dist/style.css";
import ChatAssistant from "../components/ChatAssistant";

// TODO: node type definitions are fetched from the backend via /api/node-types

function BuilderPage({ agentId }) {
  const navigate = useNavigate();
  // graph state: nodes and edges
  // Graph nodes and edges (loaded from latest agent version)
  const [nodes, setNodes] = useState([]);
  const [edges, setEdges] = useState([]);
  // modal state for selecting new node type
  const [modalOpen, setModalOpen] = useState(false);
  const [search, setSearch] = useState('');
  // Sidebar tab: 'details' | 'chat' | null
  const [sidebarTab, setSidebarTab] = useState(null);
  // const [chatOpen, setChatOpen] = useState(false);  // replaced by sidebarTab

  // Node definitions from server
  const [nodeDefs, setNodeDefs] = useState([]);
  // For quick-add: node we're adding from
  const [addingFromNodeId, setAddingFromNodeId] = useState(null);
  // Trigger instances from server
  const [triggers, setTriggers] = useState([]);
  // Test run state
  const [testing, setTesting] = useState(false);
  const [testRunResult, setTestRunResult] = useState(null);
  // Toggle between Edit and Live modes
  const [isLiveMode, setIsLiveMode] = useState(false);
  const categories = useMemo(() => {
    const map = {};
    nodeDefs.forEach((def) => {
      if (!map[def.category]) map[def.category] = [];
      map[def.category].push(def);
    });
    return Object.entries(map).map(([category, types]) => ({ category, types }));
  }, [nodeDefs]);
  useEffect(() => {
    axios.get('/api/node-types')
      .then((res) => setNodeDefs(res.data))
      .catch((err) => console.error('fetch node-types failed:', err));
  }, []);
  // Fetch trigger instances for this agent
  useEffect(() => {
    axios.get('/api/triggers')
      .then((res) => setTriggers(res.data))
      .catch((err) => console.error('fetch triggers failed:', err));
  }, []);
  // Validation errors per nodeId
  const [validationErrors, setValidationErrors] = useState({});
  // Decorate nodes with error flag for highlighting
  const displayedNodes = useMemo(
    () => nodes.map((n) => {
      const errs = validationErrors[n.id];
      const hasError = Array.isArray(errs) && errs.length > 0;
      return { ...n, data: { ...n.data, error: hasError } };
    }),
    [nodes, validationErrors]
  );
  // Execute full-agent test run (or live observe in Live mode)
  const onTestRun = useCallback(async () => {
    if (isLiveMode) {
      // Clear previous runtime statuses
      setNodes((nds) => nds.map((n) => ({ ...n, data: { ...n.data, status: undefined } })));
    }
    setTesting(true);
    try {
      const res = await axios.post(`/api/agents/${agentId}/runs/test`);
      // Show result only in Edit mode
      if (!isLiveMode) setTestRunResult(res.data);
    } catch (err) {
      console.error('test run failed', err);
      alert('Test run failed');
    } finally {
      setTesting(false);
    }
  }, [agentId, isLiveMode]);
  // Load latest saved graph for this agent
  useEffect(() => {
    axios.get(`/api/agents/${agentId}/versions/latest`)
      .then((res) => {
        const graph = res.data.graph || {};
        setNodes(graph.nodes || []);
        setEdges(graph.edges || []);
      })
      .catch((err) => console.error('fetch agent graph failed:', err));
  }, [agentId]);
  // WebSocket client for collaborative updates
  const clientId = useMemo(() => crypto.randomUUID(), []);
  const wsRef = useRef(null);
  // Track execution timers per node to enforce minimum animation duration
  const execTimersRef = useRef({});
  useEffect(() => {
    // Establish WebSocket connection for live updates (collaborative editing)
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    // Proxy through /api/ws for backend WebSocket endpoint
    const wsUrl = `${wsProtocol}//${window.location.host}/api/ws/agents/${agentId}`;
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;
    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        if (msg.clientId === clientId) return;
        switch (msg.type) {
          case 'nodesChange':
            setNodes((nds) => applyNodeChanges(msg.changes, nds));
            break;
          case 'edgesChange':
            setEdges((eds) => applyEdgeChanges(msg.changes, eds));
            break;
          case 'connect':
            setEdges((eds) => addEdge(msg.params, eds));
            break;
        case 'nodeExecution': {
            // Runtime status updates for Live mode with min animation (500ms)
            const { nodeId, phase } = msg;
            if (phase === 'start') {
                // record start time and set running status
                execTimersRef.current[nodeId] = { start: Date.now(), timeoutId: null };
                setNodes((nds) =>
                    nds.map((n) =>
                        n.id === nodeId ? { ...n, data: { ...n.data, status: 'running' } } : n
                    )
                );
            } else if (phase === 'end') {
                const timer = execTimersRef.current[nodeId];
                const now = Date.now();
                const clearStatus = () => {
                    setNodes((nds) =>
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
                    // no start record, clear immediately
                    clearStatus();
                }
            }
            break;
        }
          default:
        }
      } catch {
        // ignore malformed
      }
    };
    ws.onclose = () => { wsRef.current = null; };
    return () => { ws.close(); wsRef.current = null; };
  }, [agentId, clientId]);
  // handlers for ReactFlow change events
  const onNodesChange = useCallback(
    (changes) => {
      if (isLiveMode) return;
      setNodes((nds) => applyNodeChanges(changes, nds));
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({ clientId, type: 'nodesChange', changes }));
      }
    },
    [clientId, isLiveMode]
  );
  const onEdgesChange = useCallback(
    (changes) => {
      if (isLiveMode) return;
      setEdges((eds) => applyEdgeChanges(changes, eds));
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({ clientId, type: 'edgesChange', changes }));
      }
    },
    [clientId, isLiveMode]
  );

  const onConnect = useCallback(
    (params) => {
      if (isLiveMode) return;
      setEdges((eds) => addEdge(params, eds));
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({ clientId, type: 'connect', params }));
      }
    },
    [clientId, isLiveMode]
  );
  // Tool callback: add a node from assistant
  const handleAddNode = useCallback(({ type, label, parameters, position }) => {
    const id = Date.now().toString();
    const data = { label: label || type, nodeTypeLabel: label || type, ...(parameters || {}) };
    const pos = position || { x: 100, y: 100 };
    const newNode = { id, type, data, position: pos };
    setNodes((nds) => [...nds, newNode]);
    return { node_id: id };
  }, []);
  // Tool callback: connect two nodes from assistant
  const handleConnectNodes = useCallback(({ source_id, target_id }) => {
    setEdges((eds) => addEdge({ source: source_id, target: target_id }, eds));
    return {};
  }, []);
  // Tool callback: get the current workflow graph
  const handleGetWorkflow = useCallback(() => {
    return { nodes, edges };
  }, [nodes, edges]);

  // double-click a node to rename it
  const onNodeDoubleClick = useCallback((event, node) => {
    if (isLiveMode) return;
    const current = node.data.label || '';
    const name = prompt('Enter node name:', current);
    if (name !== null) {
      setNodes((nds) =>
        nds.map((n) =>
          n.id === node.id ? { ...n, data: { ...n.data, label: name } } : n
        )
      );
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        // broadcast the node data change
        const msg = { clientId, type: 'nodesChange', changes: [{ id: node.id, type: 'reset', position: node.position, data: { ...node.data, label: name } }] };
        wsRef.current.send(JSON.stringify(msg));
      }
    }
  }, [clientId, isLiveMode]);

  // derive nodeTypes mapping from server definitions, memoized
  // Wrap node types to inject quick-add handler
  const nodeTypes = useMemo(() => {
    const m = {};
    nodeDefs.forEach((def) => {
      let Comp;
      if (def.entry_point) Comp = TriggerNode;
      else if (def.branching) Comp = IfNode;
      else if (def.type === 'http_request') Comp = HttpRequestNode;
      else Comp = DefaultNode;
      m[def.type] = (props) => (
        <Comp
          {...props}
          icon={def.icon}
          onAddClick={() => {
            setAddingFromNodeId(props.id);
            setModalOpen(true);
          }}
        />
      );
    });
    return m;
  }, [nodeDefs]);

  async function save() {
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
      // TODO: Add other node type validations here
    });
    // Validate async-mode webhook flows have a terminating HTTP Response node
    // Build node lookup
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
      await axios.post(`/api/agents/${agentId}/versions`, {
        semantic_version: "0.1.0",
        graph,
        default_params: {},
      });
      alert("Saved!");
      // Refresh triggers to update public URLs
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
  }
  // state and handlers for node selection/details
  const [selectedNodeId, setSelectedNodeId] = useState(null);
  const selectedNode = useMemo(
    () => nodes.find((n) => n.id === selectedNodeId),
    [nodes, selectedNodeId]
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
  const onNodeClick = useCallback((event, node) => {
    setSelectedNodeId(node.id);
    setSidebarTab('details');
  }, [setSidebarTab]);
  const onPaneClick = useCallback(() => {
    setSelectedNodeId(null);
  }, []);
  // Execute selected node with given input data
  const onExecute = useCallback(
    async (inputData) => {
      if (!selectedNodeId) return;
      try {
        // mark node as running
        setNodes((nds) =>
          nds.map((n) =>
            n.id === selectedNodeId
              ? { ...n, data: { ...n.data, status: 'running' } }
              : n
          )
        );
        const res = await axios.post(
          `/api/agents/${agentId}/nodes/${selectedNodeId}/execute`,
          inputData
        );
        const output = res.data.output;
        // update node with output and mark completed
        setNodes((nds) =>
          nds.map((n) =>
            n.id === selectedNodeId
              ? {
                  ...n,
                  data: { ...n.data, lastOutput: output, status: 'completed' },
                }
              : n
          )
        );
        return output;
      } catch (err) {
        console.error(err);
        alert('Execution failed');
      }
    },
    [agentId, selectedNodeId]
  );

  return (
    <div className="flex" style={{ height: "80vh" }}>
      {/* Main builder area */}
      <div className="flex-1 flex flex-col">
      <div className="mb-2 flex gap-2 items-center">
        {/* Mode toggle */}
        <button
          onClick={() => setIsLiveMode((prev) => !prev)}
          className="px-3 py-1 border rounded"
        >
          {isLiveMode ? 'Switch to Edit' : 'Switch to Live'}
        </button>
        {/* Edit mode controls */}
        {!isLiveMode && (
          <>
            <button
              onClick={() => setModalOpen(true)}
              className="px-3 py-1 rounded border"
            >
              + Add Node
            </button>
            <button
              onClick={save}
              className="px-3 py-1 bg-indigo-600 text-white rounded"
            >
              Save
            </button>
            <button
              onClick={() => setSidebarTab('chat')}
              className="px-3 py-1 bg-purple-600 text-white rounded"
            >
              AI Assistant
            </button>
          </>
        )}
        <button
          onClick={onTestRun}
          disabled={testing}
          className="px-3 py-1 bg-green-600 text-white rounded disabled:opacity-50"
        >
          {testing ? 'Running...' : 'Test Run'}
        </button>
        <button
          onClick={() => navigate(`/agents/${agentId}/runs`)}
          className="px-3 py-1 bg-gray-600 text-white rounded"
        >
          Runs
        </button>
      </div>
        <div className="flex-1">
          {/* Validation errors */}
          {Object.keys(validationErrors).length > 0 && (
            <div className="p-2 mb-2 bg-red-100 text-red-800 rounded">
              <ul className="list-disc list-inside text-sm">
                {Object.entries(validationErrors).flatMap(([nodeId, errs]) =>
                  errs.map((msg, idx) => <li key={`${nodeId}-${idx}`}>{msg}</li>)
                )}
              </ul>
            </div>
          )}
      <ReactFlow
            nodes={displayedNodes}
            edges={edges}
            nodeTypes={nodeTypes}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onNodeClick={onNodeClick}
            onNodeDoubleClick={onNodeDoubleClick}
            onPaneClick={onPaneClick}
            // Disable editing interactions in Live mode
            nodesDraggable={!isLiveMode}
            nodesConnectable={!isLiveMode}
            nodesSelectable={!isLiveMode}
            connectionRadius={60}
            fitView
            attributionPosition="bottom-left"
          >
            <Background />
            <MiniMap />
            <Controls />
          </ReactFlow>
        </div>
      </div>
      {/* Sidebar with tabs */}
      {sidebarTab && (
        <div className="w-80 bg-gray-50 border-l h-full flex flex-col">
          {/* Tabs header */}
          <div className="flex items-center justify-between p-2 border-b">
            <div className="flex space-x-2">
              <button
                className={`px-3 py-1 ${sidebarTab === 'details' ? 'border-b-2 border-indigo-600 text-indigo-600' : 'text-gray-600'}`}
                onClick={() => setSidebarTab('details')}
              >
                Details
              </button>
              <button
                className={`px-3 py-1 ${sidebarTab === 'chat' ? 'border-b-2 border-indigo-600 text-indigo-600' : 'text-gray-600'}`}
                onClick={() => setSidebarTab('chat')}
              >
                Chat
              </button>
            </div>
            <button onClick={() => setSidebarTab(null)} className="text-gray-500 hover:text-gray-800">&times;</button>
          </div>
          {/* Content area */}
          <div className="flex-1 overflow-auto">
            <div style={{ display: sidebarTab === 'details' ? 'block' : 'none' }}>
              {selectedNode && selectedNodeDef && (
                <NodeDetailsPanel
                  node={selectedNode}
                  nodeDef={selectedNodeDef}
                  readOnly={isLiveMode}
                  onChange={(key, value) => {
                    setNodes((nds) =>
                      nds.map((n) =>
                        n.id === selectedNode.id ? { ...n, data: { ...n.data, [key]: value } } : n
                      )
                    );
                  }}
                  onExecute={onExecute}
                  publicUrl={
                    // only webhook nodes have public URLs
                    selectedNodeDef.type === 'webhook' && triggersMap[selectedNode.id]
                      ? `${window.location.origin}/webhooks/${selectedNode.type}/${triggersMap[selectedNode.id].id}`
                      : undefined
                  }
                />
              )}
            </div>
            <div style={{ display: sidebarTab === 'chat' ? 'block' : 'none' }}>
              <ChatAssistant
                inline
                agentId={agentId}
                onAddNode={handleAddNode}
                onConnectNodes={handleConnectNodes}
                onGetWorkflow={handleGetWorkflow}
                onClose={() => setSidebarTab(null)}
              />
            </div>
          </div>
        </div>
      )}
      {/* Test Run result modal (Edit mode only) */}
      {!isLiveMode && testRunResult && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white rounded shadow-lg w-3/4 max-h-full overflow-auto p-4">
            <h2 className="text-lg font-bold mb-2">Test Run Result</h2>
            <pre className="text-xs font-mono bg-gray-100 p-2 rounded h-96 overflow-auto">
              {JSON.stringify(testRunResult, null, 2)}
            </pre>
            <div className="mt-2 text-right">
              <button
                onClick={() => setTestRunResult(null)}
                className="px-3 py-1 bg-indigo-600 text-white rounded"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
      {/* Node type selection dialog */}
      {modalOpen && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white p-4 rounded shadow-lg w-80 max-h-full overflow-auto">
            <input
              type="text"
              placeholder="Search nodes..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full border rounded px-2 py-1 mb-2"
            />
            <div className="space-y-4">
              {categories.map((cat) => {
                const filtered = cat.types.filter(
                  (nt) =>
                    nt.label.toLowerCase().includes(search.toLowerCase()) ||
                    nt.type.toLowerCase().includes(search.toLowerCase())
                );
                if (filtered.length === 0) return null;
                return (
                  <div key={cat.category}>
                    <div className="text-sm font-semibold text-gray-700 mb-1">
                      {cat.category}
                    </div>
                    <ul>
                        {filtered.map((nt) => (
                        <li key={nt.type}>
                          <button
                            className="w-full flex items-center px-2 py-1 hover:bg-gray-100 rounded"
                            onClick={() => {
                              const id = Date.now().toString();
                              // populate parameter defaults from server metadata
                              const defaults = {};
                              if (Array.isArray(nt.parameters)) {
                                nt.parameters.forEach((p) => {
                                  if (p.default !== undefined) {
                                    defaults[p.name] = p.default;
                                  }
                                });
                              }
                              // Determine position: to the right of source if quick-adding
                              let position = { x: 100, y: 100 };
                              if (addingFromNodeId) {
                                const src = nodes.find((n) => n.id === addingFromNodeId);
                                if (src) {
                                  position = { x: src.position.x + 200, y: src.position.y };
                                }
                              }
                              const newNode = { id, position, data: { label: nt.label, nodeTypeLabel: nt.label, ...defaults }, type: nt.type };
                              setNodes((nds) => [...nds, newNode]);
                              if (addingFromNodeId) {
                                setEdges((eds) => addEdge({ source: addingFromNodeId, target: id }, eds));
                                setAddingFromNodeId(null);
                              }
                              setModalOpen(false);
                              setSearch('');
                            }}
                          >
                            {nt.icon && <span className="mr-2">{nt.icon}</span>}
                            {nt.label}
                          </button>
                        </li>
                      ))}
                    </ul>
                  </div>
                );
              })}
            </div>
            <div className="text-right mt-4">
              <button
                onClick={() => { setModalOpen(false); setAddingFromNodeId(null); }}
                className="px-3 py-1 border rounded"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default BuilderPage;
