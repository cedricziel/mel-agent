import React, { useEffect, useState, useMemo, useCallback } from 'react';
import axios from 'axios';
import { useParams, Link } from 'react-router-dom';
// ReactFlow and node components imported above
import RunDetailsPanel from '../components/RunDetailsPanel';
import ReactFlow, { Background, Controls, MiniMap } from 'reactflow';
import 'reactflow/dist/style.css';
import IfNode from '../components/IfNode';
import DefaultNode from '../components/DefaultNode';
import TriggerNode from '../components/TriggerNode';

export default function RunsPage() {
  const { id: agentId } = useParams();
  const [runs, setRuns] = useState([]);
  const [selectedRunID, setSelectedRunID] = useState(null);
  const [runDetails, setRunDetails] = useState(null);
  const [nodeDefs, setNodeDefs] = useState([]);
  // Graph and selection state
  const [rfNodes, setRfNodes] = useState([]);
  const [rfEdges, setRfEdges] = useState([]);
  const [selectedRunNodeID, setSelectedRunNodeID] = useState(null);

  // Fetch node definitions for renderer
  useEffect(() => {
    axios.get('/api/node-types')
      .then(res => setNodeDefs(res.data))
      .catch(err => console.error('fetch node-types failed', err));
  }, []);
  const nodeTypes = useMemo(() => {
    const m = {};
    nodeDefs.forEach(def => {
      if (def.entry_point) m[def.type] = TriggerNode;
      else if (def.branching) m[def.type] = IfNode;
      else m[def.type] = DefaultNode;
    });
    return m;
  }, [nodeDefs]);

  // derive selected node and its step
  const selectedRunStep = useMemo(
    () => runDetails?.trace?.find(s => s.nodeId === selectedRunNodeID),
    [runDetails, selectedRunNodeID]
  );
  const selectedRunNodeDef = useMemo(
    () => nodeDefs.find(def => def.type === runDetails?.graph?.nodes?.find(n => n.id === selectedRunNodeID)?.type),
    [nodeDefs, runDetails, selectedRunNodeID]
  );

  // set up nodes and edges when run details arrive
  useEffect(() => {
    if (runDetails) {
      const graph = runDetails.graph || {};
      setRfNodes(graph.nodes || []);
      setRfEdges(graph.edges || []);
    } else {
      setRfNodes([]);
      setRfEdges([]);
    }
    setSelectedRunNodeID(null);
  }, [runDetails]);

  // Handlers for ReactFlow
  const onNodeClick = useCallback((_, node) => setSelectedRunNodeID(node.id), []);
  const onPaneClick = useCallback(() => setSelectedRunNodeID(null), []);

  useEffect(() => {
    axios.get(`/api/agents/${agentId}/runs`)
      .then(res => setRuns(res.data))
      .catch(err => console.error('fetch runs list failed', err));
  }, [agentId]);

  useEffect(() => {
    if (selectedRunID) {
      axios.get(`/api/agents/${agentId}/runs/${selectedRunID}`)
        .then(res => setRunDetails(res.data))
        .catch(err => console.error('fetch run details failed', err));
    }
  }, [selectedRunID, agentId]);

  return (
    <div className="flex h-full">
      {/* Runs list */}
      <div className="w-1/4 border-r p-4 overflow-auto h-full">
        <h2 className="text-xl font-bold mb-4">Runs for Agent {agentId}</h2>
        <ul className="space-y-2">
          {runs.map(run => (
            <li key={run.id}>
              <button
                onClick={() => setSelectedRunID(run.id)}
                className={`w-full text-left px-2 py-1 rounded ${run.id === selectedRunID ? 'bg-gray-200' : 'hover:bg-gray-100'}`}
              >
                {run.created_at}
              </button>
            </li>
          ))}
        </ul>
        <div className="mt-4">
          <Link to={`/agents/${agentId}/edit`} className="text-indigo-600 hover:underline">
            ← Back to Builder
          </Link>
        </div>
      </div>

      {/* Graph & Node details */}
      <div className="flex-1 flex h-full min-h-0">
        <div className="w-1/2 h-full">
          {runDetails && (
            <ReactFlow
              nodes={rfNodes}
              edges={rfEdges}
              nodeTypes={nodeTypes}
              onNodeClick={onNodeClick}
              onPaneClick={onPaneClick}
              fitView
              nodesDraggable={false}
              nodesConnectable={false}
              elementsSelectable={false}
            >
              <Background />
              <MiniMap />
              <Controls />
            </ReactFlow>
          )}
          {!runDetails && <p className="p-4 text-gray-500">Select a run to view graph.</p>}
        </div>
        <div className="w-1/2 p-4 overflow-auto h-full">
          {runDetails && selectedRunNodeID && selectedRunStep && selectedRunNodeDef ? (
            <RunDetailsPanel nodeDef={selectedRunNodeDef} step={selectedRunStep} />
          ) : (
            <p className="text-gray-500">Select a node to inspect inputs/outputs.</p>
          )}
        </div>
      </div>
    </div>
  );
}