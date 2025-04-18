import { useEffect, useState, useCallback } from "react";
import axios from "axios";
import ReactFlow, {
  Background,
  Controls,
  addEdge,
  applyNodeChanges,
  applyEdgeChanges,
  MiniMap,
} from "reactflow";
import "reactflow/dist/style.css";

// Available node categories and types for insertion dialog
const NODE_CATEGORIES = [
  {
    category: 'Triggers',
    types: [
      { type: 'timer', label: 'Timer' },
      { type: 'http', label: 'HTTP Request' },
    ],
  },
  {
    category: 'Basic',
    types: [
      { type: 'if', label: 'If' },
      { type: 'switch', label: 'Switch' },
    ],
  },
  {
    category: 'LLM',
    types: [
      { type: 'agent', label: 'Agent' },
    ],
  },
];

function BuilderPage({ agentId }) {
  // graph state: nodes and edges
  const [nodes, setNodes] = useState([
    {
      id: "1",
      position: { x: 250, y: 5 },
      data: { label: "Start" },
      type: "input",
    },
  ]);
  const [edges, setEdges] = useState([]);
  // modal state for selecting new node type
  const [modalOpen, setModalOpen] = useState(false);
  const [search, setSearch] = useState('');

  // handlers for ReactFlow change events
  const onNodesChange = useCallback(
    (changes) => setNodes((nds) => applyNodeChanges(changes, nds)),
    [setNodes]
  );
  const onEdgesChange = useCallback(
    (changes) => setEdges((eds) => applyEdgeChanges(changes, eds)),
    [setEdges]
  );

  const onConnect = useCallback((params) => setEdges((eds) => addEdge(params, eds)), []);

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

  return (
    <div style={{ height: "80vh" }}>
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
      </div>

      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        fitView
      >
        <Background />
        <MiniMap />
        <Controls />
      </ReactFlow>
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
              {NODE_CATEGORIES.map((cat) => {
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
                              setNodes((nds) => [
                                ...nds,
                                {
                                  id,
                                  position: { x: 100, y: 100 },
                                  data: { label: nt.label },
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
