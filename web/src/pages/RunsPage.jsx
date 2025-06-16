import { useParams } from 'react-router-dom';
import RunDetailsPanel from '../components/RunDetailsPanel';
import RunsList from '../components/RunsList';
import RunsGraph from '../components/RunsGraph';
import { useRunsData } from '../hooks/useRunsData';

export default function RunsPage() {
  const { id: agentId } = useParams();

  const {
    runs,
    runDetails,
    nodeDefs,
    selectedRunID,
    selectedRunNodeID,
    selectedRunStep,
    selectedRunNodeDef,
    rfNodes,
    rfEdges,
    handleRunSelect,
    handleNodeSelect,
    handlePaneClick,
  } = useRunsData(agentId);

  // Handlers for ReactFlow
  const onNodeClick = (_, node) => handleNodeSelect(node.id);

  return (
    <div className="flex h-full">
      <RunsList
        agentId={agentId}
        runs={runs}
        selectedRunID={selectedRunID}
        onRunSelect={handleRunSelect}
      />

      <div className="flex-1 flex h-full min-h-0">
        <RunsGraph
          runDetails={runDetails}
          rfNodes={rfNodes}
          rfEdges={rfEdges}
          nodeDefs={nodeDefs}
          onNodeClick={onNodeClick}
          onPaneClick={handlePaneClick}
        />

        <div className="w-1/4 p-4 overflow-auto h-full">
          {runDetails &&
          selectedRunNodeID &&
          selectedRunStep &&
          selectedRunNodeDef ? (
            <RunDetailsPanel
              nodeDef={selectedRunNodeDef}
              step={selectedRunStep}
            />
          ) : (
            <p className="text-gray-500">
              Select a node to inspect inputs/outputs.
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
