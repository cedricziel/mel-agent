import { useState, useEffect, useCallback } from 'react';

/**
 * Custom hook for managing NodeModal state including form data, dynamic options, and credentials
 * @param {Object} node - The node being edited
 * @param {Object} nodeDef - The node definition with parameters
 * @param {Function} onChange - Callback for when form data changes
 * @returns {Object} State and handlers for the node modal
 */
export function useNodeModalState(node, nodeDef, onChange) {
  // Form state
  const [currentFormData, setCurrentFormData] = useState({});
  const [dynamicOptions, setDynamicOptions] = useState({});
  const [loadingOptions, setLoadingOptions] = useState({});
  const [credentials, setCredentials] = useState({});

  // Execution state
  const [inputData, setInputData] = useState({});
  const [outputData, setOutputData] = useState({});
  const [runHistory, setRunHistory] = useState([]);
  const [selectedRun, setSelectedRun] = useState(null);

  // UI state
  const [activeTab, setActiveTab] = useState('config');

  // Track current form data locally
  useEffect(() => {
    setCurrentFormData({ ...node?.data });
  }, [node?.data]);

  // Helper to update both local state and parent
  const handleChange = useCallback(
    (key, value) => {
      setCurrentFormData((prev) => ({ ...prev, [key]: value }));
      onChange(key, value);
    },
    [onChange]
  );

  // Function to load dynamic options for a specific parameter
  const loadDynamicOptionsForParam = useCallback(
    async (paramName) => {
      const param = nodeDef?.parameters?.find((p) => p.name === paramName);
      if (!param?.dynamicOptions) return;

      try {
        setLoadingOptions((prev) => ({ ...prev, [paramName]: true }));

        const queryParams = new URLSearchParams();
        Object.entries(currentFormData || {}).forEach(([key, value]) => {
          if (value !== undefined && value !== null && value !== '') {
            queryParams.append(key, value);
          }
        });

        const url = `/api/node-types/${nodeDef.type}/parameters/${paramName}/options?${queryParams}`;
        const response = await fetch(url);

        if (response.ok) {
          const data = await response.json();
          setDynamicOptions((prev) => ({
            ...prev,
            [paramName]: data.options || [],
          }));
        } else {
          setDynamicOptions((prev) => ({
            ...prev,
            [paramName]: [],
          }));
        }
      } catch (error) {
        console.error(`Error loading dynamic options for ${paramName}:`, error);
        setDynamicOptions((prev) => ({
          ...prev,
          [paramName]: [],
        }));
      } finally {
        setLoadingOptions((prev) => ({ ...prev, [paramName]: false }));
      }
    },
    [nodeDef, currentFormData]
  );

  // Load credentials for credential parameters
  useEffect(() => {
    if (!nodeDef || !nodeDef.parameters) return;

    const credentialParams = nodeDef.parameters.filter(
      (p) => p.type === 'credential' || p.parameterType === 'credential'
    );

    if (credentialParams.length > 0) {
      // Load credentials for each credential parameter
      const promises = credentialParams.map(async (param) => {
        try {
          const url = param.credentialType
            ? `/api/credentials?credential_type=${param.credentialType}`
            : '/api/credentials';
          const response = await fetch(url);
          const data = await response.json();
          return { paramName: param.name, credentials: data };
        } catch (error) {
          console.error(`Failed to load credentials for ${param.name}:`, error);
          return { paramName: param.name, credentials: [] };
        }
      });

      Promise.all(promises).then((results) => {
        const credentialsMap = {};
        results.forEach(({ paramName, credentials }) => {
          credentialsMap[paramName] = credentials;
        });
        setCredentials(credentialsMap);
      });
    }
  }, [nodeDef]);

  // Load dynamic options when dependencies change
  useEffect(() => {
    if (!nodeDef || !currentFormData?.credentialId) return;
    loadDynamicOptionsForParam('databaseId');
  }, [nodeDef, currentFormData?.credentialId, loadDynamicOptionsForParam]);

  useEffect(() => {
    if (!nodeDef || !currentFormData?.databaseId) {
      setDynamicOptions((prev) => ({ ...prev, tableId: [] }));
      return;
    }
    loadDynamicOptionsForParam('tableId');
  }, [nodeDef, currentFormData?.databaseId, loadDynamicOptionsForParam]);

  // Load execution data for this node when in executions mode
  const loadNodeExecutionData = useCallback(
    async (selectedExecution, agentId, viewMode) => {
      if (!node || !selectedExecution || viewMode !== 'executions') return;

      try {
        // Fetch the full execution details
        const response = await fetch(
          `/api/agents/${agentId}/runs/${selectedExecution.id}`
        );
        if (response.ok) {
          const executionData = await response.json();

          // Find the trace step for this specific node
          if (executionData.trace) {
            const nodeStep = executionData.trace.find(
              (step) => step.nodeId === node.id
            );
            if (nodeStep) {
              setInputData(nodeStep.input?.[0]?.data || {});
              setOutputData(nodeStep.output?.[0]?.data || {});
            }
          }
        }
      } catch (error) {
        console.error('Error loading node execution data:', error);
      }
    },
    [node]
  );

  return {
    // Form state
    currentFormData,
    dynamicOptions,
    loadingOptions,
    credentials,
    handleChange,
    loadDynamicOptionsForParam,

    // Execution state
    inputData,
    outputData,
    runHistory,
    selectedRun,
    setInputData,
    setOutputData,
    setRunHistory,
    setSelectedRun,
    loadNodeExecutionData,

    // UI state
    activeTab,
    setActiveTab,
  };
}
