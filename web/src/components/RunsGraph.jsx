import { useMemo } from 'react';
import ReactFlow, { Background, Controls, MiniMap } from 'reactflow';
import 'reactflow/dist/style.css';
import IfNode from './IfNode';
import DefaultNode from './DefaultNode';
import TriggerNode from './TriggerNode';

export default function RunsGraph({
  runDetails,
  rfNodes,
  rfEdges,
  nodeDefs,
  onNodeClick,
  onPaneClick,
}) {
  const nodeTypes = useMemo(() => {
    const m = {};
    nodeDefs.forEach((def) => {
      if (def.entry_point) m[def.type] = TriggerNode;
      else if (def.branching) m[def.type] = IfNode;
      else m[def.type] = DefaultNode;
    });
    return m;
  }, [nodeDefs]);

  if (!runDetails) {
    return (
      <div className="w-1/2 h-full flex items-center justify-center">
        <p className="p-4 text-gray-500">Select a run to view graph.</p>
      </div>
    );
  }

  return (
    <div className="w-1/2 h-full">
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
    </div>
  );
}
