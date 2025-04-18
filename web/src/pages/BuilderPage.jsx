import { useEffect, useState, useCallback } from "react";
import axios from "axios";
import ReactFlow, {
  Background,
  Controls,
  addEdge,
  MiniMap,
} from "reactflow";
import "reactflow/dist/style.css";

function BuilderPage({ agentId }) {
  const [nodes, setNodes] = useState([
    {
      id: "1",
      position: { x: 250, y: 5 },
      data: { label: "Start" },
      type: "input",
    },
  ]);
  const [edges, setEdges] = useState([]);

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
          onClick={() => {
            const id = Date.now().toString();
            setNodes((nds) => [
              ...nds,
              {
                id,
                position: { x: 100, y: 100 },
                data: { label: "LLM" },
              },
            ]);
          }}
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
        onNodesChange={setNodes}
        onEdgesChange={setEdges}
        onConnect={onConnect}
        fitView
      >
        <Background />
        <MiniMap />
        <Controls />
      </ReactFlow>
    </div>
  );
}

export default BuilderPage;
