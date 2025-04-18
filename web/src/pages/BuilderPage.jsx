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
import NodeDetailsPanel from "../components/NodeDetailsPanel";
import "reactflow/dist/style.css";

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

  // Node definitions from server
  const [nodeDefs, setNodeDefs] = useState([]);
  // Test run state
  const [testing, setTesting] = useState(false);
  const [testRunResult, setTestRunResult] = useState(null);
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
  // Execute full-agent test run
  const onTestRun = useCallback(async () => {
    setTesting(true);
    try {
      const res = await axios.post(`/api/agents/${agentId}/runs/test`);
      setTestRunResult(res.data);
    } catch (err) {
      console.error('test run failed', err);
      alert('Test run failed');
    } finally {
      setTesting(false);
    }
  }, [agentId]);
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
  useEffect(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    const url = `${protocol}://${window.location.host}/ws/agents/${agentId}`;
    const ws = new WebSocket(url);
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
      setNodes((nds) => applyNodeChanges(changes, nds));
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({ clientId, type: 'nodesChange', changes }));
      }
    },
    [clientId]
  );
  const onEdgesChange = useCallback(
    (changes) => {
      setEdges((eds) => applyEdgeChanges(changes, eds));
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({ clientId, type: 'edgesChange', changes }));
      }
    },
    [clientId]
  );

  const onConnect = useCallback(
    (params) => {
      setEdges((eds) => addEdge(params, eds));
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({ clientId, type: 'connect', params }));
      }
    },
    [clientId]
  );

  // derive nodeTypes mapping from server definitions, memoized
  const nodeTypes = useMemo(() => {
    const m = {};
    nodeDefs.forEach((def) => {
      // EntryPoint nodes (no input) use TriggerNode
      if (def.entry_point) {
        m[def.type] = TriggerNode;
      } else if (def.branching) {
        m[def.type] = IfNode;
      } else {
        m[def.type] = DefaultNode;
      }
    });
    return m;
  }, [nodeDefs]);

  async function save() {
    const graph = { nodes, edges };
    try {
      await axios.post(`/api/agents/${agentId}/versions`, {
        semantic_version: "0.1.0",
        graph,
        default_params: {},
      });
      alert("Saved!");
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
  const onNodeClick = useCallback((event, node) => {
    setSelectedNodeId(node.id);
  }, []);
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
      <div className="mb-2 flex gap-2">
        <button
          onClick={() => setModalOpen(true)}
          className="px-3 py-1 rounded border"
        >
          + Add Node
        </button>
        <button onClick={save} className="px-3 py-1 bg-indigo-600 text-white rounded">
          Save
        </button>
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
          <ReactFlow
            nodes={nodes}
            edges={edges}
            nodeTypes={nodeTypes}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onNodeClick={onNodeClick}
            onPaneClick={onPaneClick}
            fitView
            attributionPosition="bottom-left"
          >
            <Background />
            <MiniMap />
            <Controls />
          </ReactFlow>
        </div>
      </div>
      {/* Details panel for selected node */}
      {selectedNode && selectedNodeDef && (
        <NodeDetailsPanel
          node={selectedNode}
          nodeDef={selectedNodeDef}
          onChange={(key, value) => {
            setNodes((nds) =>
              nds.map((n) =>
                n.id === selectedNode.id ? { ...n, data: { ...n.data, [key]: value } } : n
              )
            );
          }}
          onExecute={onExecute}
        />
      )}
      {/* Test Run result modal */}
      {testRunResult && (
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
                            className="w-full text-left px-2 py-1 hover:bg-gray-100 rounded"
                            onClick={() => {
                              const id = Date.now().toString();
                              // merge defaults from server if provided
                              const defaults = nt.defaults || {};
                              setNodes((nds) => [
                                ...nds,
                                {
                                  id,
                                  position: { x: 100, y: 100 },
                                  data: { label: nt.label, ...defaults },
                                  type: nt.type,
                                },
                              ]);
                              setModalOpen(false);
                              setSearch('');
                            }}
                          >
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
                onClick={() => setModalOpen(false)}
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
