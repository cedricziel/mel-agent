import { useEffect, useState, useMemo, useCallback } from 'react';
import { nodeTypesApi, workflowRunsApi } from '../api/client';

export function useRunsData(agentId) {
  const [runs, setRuns] = useState([]);
  const [selectedRunID, setSelectedRunID] = useState(null);
  const [runDetails, setRunDetails] = useState(null);
  const [nodeDefs, setNodeDefs] = useState([]);
  const [selectedRunNodeID, setSelectedRunNodeID] = useState(null);

  // Fetch node definitions for renderer
  useEffect(() => {
    nodeTypesApi
      .listNodeTypes()
      .then((res) => setNodeDefs(res.data))
      .catch((err) => console.error('fetch node-types failed', err));
  }, []);

  // Fetch runs list
  useEffect(() => {
    if (!agentId) return;

    workflowRunsApi
      .listWorkflowRuns({ workflow_id: agentId })
      .then((res) => setRuns(res.data))
      .catch((err) => console.error('fetch runs list failed', err));
  }, [agentId]);

  // Fetch run details when selected
  useEffect(() => {
    if (selectedRunID && agentId) {
      workflowRunsApi
        .getWorkflowRun(selectedRunID)
        .then((res) => setRunDetails(res.data))
        .catch((err) => console.error('fetch run details failed', err));
    }
  }, [selectedRunID, agentId]);

  // Reset selected node when run details change
  useEffect(() => {
    setSelectedRunNodeID(null);
  }, [runDetails]);

  // Derived state
  const selectedRunStep = useMemo(
    () => runDetails?.trace?.find((s) => s.nodeId === selectedRunNodeID),
    [runDetails, selectedRunNodeID]
  );

  const selectedRunNodeDef = useMemo(
    () =>
      nodeDefs.find(
        (def) =>
          def.type ===
          runDetails?.graph?.nodes?.find((n) => n.id === selectedRunNodeID)
            ?.type
      ),
    [nodeDefs, runDetails, selectedRunNodeID]
  );

  const rfNodes = useMemo(() => {
    return runDetails?.graph?.nodes || [];
  }, [runDetails]);

  const rfEdges = useMemo(() => {
    return runDetails?.graph?.edges || [];
  }, [runDetails]);

  // Handlers
  const handleRunSelect = useCallback((runId) => {
    setSelectedRunID(runId);
  }, []);

  const handleNodeSelect = useCallback((nodeId) => {
    setSelectedRunNodeID(nodeId);
  }, []);

  const handlePaneClick = useCallback(() => {
    setSelectedRunNodeID(null);
  }, []);

  return {
    // Data
    runs,
    runDetails,
    nodeDefs,
    selectedRunID,
    selectedRunNodeID,
    selectedRunStep,
    selectedRunNodeDef,
    rfNodes,
    rfEdges,

    // Handlers
    handleRunSelect,
    handleNodeSelect,
    handlePaneClick,
  };
}
