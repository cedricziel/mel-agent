import { useState, useEffect } from 'react';

// Full-screen modal for editing nodes with input/output panels like n8n
export default function NodeModal({ node, nodeDef, isOpen, onClose, onChange, onExecute, onSave, viewMode, selectedExecution, agentId }) {
  const [currentFormData, setCurrentFormData] = useState({});
  const [inputData, setInputData] = useState({});
  const [outputData, setOutputData] = useState({});
  const [runHistory, setRunHistory] = useState([]);
  const [selectedRun, setSelectedRun] = useState(null);
  const [activeTab, setActiveTab] = useState('config'); // 'config', 'executions'
  const [dynamicOptions, setDynamicOptions] = useState({});
  const [loadingOptions, setLoadingOptions] = useState({});
  const [credentials, setCredentials] = useState({});

  // Track current form data locally
  useEffect(() => {
    setCurrentFormData({ ...node?.data });
  }, [node?.data]);

  // Helper to update both local state and parent
  const handleChange = (key, value) => {
    setCurrentFormData(prev => ({ ...prev, [key]: value }));
    onChange(key, value);
  };

  // Function to load dynamic options for a specific parameter
  const loadDynamicOptionsForParam = async (paramName) => {
    const param = nodeDef?.parameters?.find(p => p.name === paramName);
    if (!param?.dynamicOptions) return;
    
    try {
      setLoadingOptions(prev => ({ ...prev, [paramName]: true }));
      
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
        setDynamicOptions(prev => ({ 
          ...prev, 
          [paramName]: data.options || [] 
        }));
      } else {
        setDynamicOptions(prev => ({ 
          ...prev, 
          [paramName]: [] 
        }));
      }
    } catch (error) {
      console.error(`Error loading dynamic options for ${paramName}:`, error);
      setDynamicOptions(prev => ({ 
        ...prev, 
        [paramName]: [] 
      }));
    } finally {
      setLoadingOptions(prev => ({ ...prev, [paramName]: false }));
    }
  };

  // Load credentials for credential parameters
  useEffect(() => {
    if (!nodeDef || !nodeDef.parameters) return;
    
    const credentialParams = nodeDef.parameters.filter(p => 
      p.type === 'credential' || p.parameterType === 'credential'
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
  }, [nodeDef, currentFormData?.credentialId]);

  useEffect(() => {
    if (!nodeDef || !currentFormData?.databaseId) {
      setDynamicOptions(prev => ({ ...prev, tableId: [] }));
      return;
    }
    loadDynamicOptionsForParam('tableId');
  }, [nodeDef, currentFormData?.databaseId]);

  // Load execution data for this node when in executions mode
  useEffect(() => {
    if (!isOpen || !node || !selectedExecution || viewMode !== 'executions') return;
    
    const loadNodeExecutionData = async () => {
      try {
        // Fetch the full execution details
        const response = await fetch(`/api/agents/${agentId}/runs/${selectedExecution.id}`);
        if (response.ok) {
          const executionData = await response.json();
          
          // Find the trace step for this specific node
          if (executionData.trace) {
            const nodeStep = executionData.trace.find(step => step.nodeId === node.id);
            if (nodeStep) {
              setInputData(nodeStep.input?.[0]?.data || {});
              setOutputData(nodeStep.output?.[0]?.data || {});
            }
          }
        }
      } catch (error) {
        console.error('Error loading node execution data:', error);
      }
    };
    
    loadNodeExecutionData();
  }, [isOpen, node, selectedExecution, viewMode, agentId]);

  if (!isOpen || !node || !nodeDef) return null;

  const renderParameterField = (param) => {
    const val = currentFormData[param.name] != null ? currentFormData[param.name] : param.default;
    const error = param.required && (val === '' || val == null);
    const baseClass = error ? 'border-red-500' : 'border-gray-300';
    const paramType = param.parameterType || param.type;
    
    const hasDynamicOptions = dynamicOptions[param.name] && dynamicOptions[param.name].length > 0;
    const isLoadingOptions = loadingOptions[param.name];
    const isDynamicParameter = param.dynamicOptions;

    switch (paramType) {
      case 'string':
        if (isDynamicParameter) {
          return (
            <select
              value={val || ''}
              onChange={(e) => handleChange(param.name, e.target.value)}
              className={`w-full border rounded px-3 py-2 ${baseClass}`}
              disabled={isLoadingOptions}
            >
              {isLoadingOptions ? (
                <option value="">Loading...</option>
              ) : hasDynamicOptions ? (
                <>
                  <option value="">Select {param.label}</option>
                  {dynamicOptions[param.name].map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </>
              ) : (
                <option value="">No options available</option>
              )}
            </select>
          );
        }
        
        return (
          <input
            type="text"
            value={val || ''}
            onChange={(e) => handleChange(param.name, e.target.value)}
            className={`w-full border rounded px-3 py-2 ${baseClass}`}
            placeholder={param.description}
          />
        );

      case 'credential':
        // Dynamic credential selection
        const paramCredentials = credentials[param.name] || [];
        return (
          <select
            value={val || ''}
            onChange={(e) => handleChange(param.name, e.target.value)}
            className={`w-full border rounded px-3 py-2 ${baseClass}`}
          >
            <option value="">Select Credential</option>
            {paramCredentials.map((cred) => (
              <option key={cred.id} value={cred.id}>
                {cred.name} ({cred.integration_name})
              </option>
            ))}
          </select>
        );

      case 'enum':
        return (
          <select
            value={val || ''}
            onChange={(e) => handleChange(param.name, e.target.value)}
            className={`w-full border rounded px-3 py-2 ${baseClass}`}
          >
            {param.options && param.options.map((opt) => (
              <option key={opt} value={opt}>{opt}</option>
            ))}
          </select>
        );

      case 'boolean':
        return (
          <div className="flex items-center">
            <input
              type="checkbox"
              checked={!!val}
              onChange={(e) => handleChange(param.name, e.target.checked)}
              className="mr-2"
            />
            <span>{param.label}</span>
          </div>
        );

      default:
        return (
          <input
            type="text"
            value={val || ''}
            onChange={(e) => handleChange(param.name, e.target.value)}
            className={`w-full border rounded px-3 py-2 ${baseClass}`}
            placeholder={param.description}
          />
        );
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg w-full h-full max-w-7xl max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b">
          <div className="flex items-center gap-3">
            <span className="text-2xl">{nodeDef.icon}</span>
            <h2 className="text-xl font-semibold">{nodeDef.label}</h2>
            <span className="text-sm text-gray-500">({node.id})</span>
          </div>
          
          <div className="flex items-center gap-2">
            <button
              onClick={onSave}
              disabled={viewMode === 'executions'}
              className={`px-4 py-2 rounded ${
                viewMode === 'executions'
                  ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
                  : 'bg-blue-600 text-white hover:bg-blue-700'
              }`}
            >
              Save
            </button>
            <button
              onClick={onClose}
              className="px-4 py-2 bg-gray-500 text-white rounded hover:bg-gray-600"
            >
              Close
            </button>
          </div>
        </div>

        {/* Main content */}
        <div className="flex-1 flex overflow-hidden">
          {/* Left Panel - Input Data */}
          <div className="w-1/3 border-r flex flex-col">
            <div className="p-3 border-b bg-gray-50">
              <h3 className="font-medium">Input Data</h3>
            </div>
            <div className="flex-1 p-3 overflow-auto">
              <pre className="text-xs bg-gray-100 p-3 rounded">
                {JSON.stringify(inputData, null, 2) || '// No input data'}
              </pre>
            </div>
          </div>

          {/* Center Panel - Node Configuration */}
          <div className="w-1/3 flex flex-col">
            <div className="border-b">
              <div className="flex">
                <button
                  onClick={() => setActiveTab('config')}
                  className={`px-4 py-2 font-medium ${
                    activeTab === 'config' 
                      ? 'border-b-2 border-blue-500 text-blue-600' 
                      : 'text-gray-600 hover:text-gray-800'
                  }`}
                >
                  {viewMode === 'executions' ? 'Node Data' : 'Configuration'}
                </button>
              </div>
            </div>

            <div className="flex-1 p-4 overflow-auto">
              {activeTab === 'config' && (
                <div className="space-y-4">
                  {viewMode === 'executions' ? (
                    // In executions mode, show read-only node info
                    <div className="space-y-4">
                      <div>
                        <label className="block text-sm font-medium mb-1">Node Name</label>
                        <div className="w-full border rounded px-3 py-2 bg-gray-50 text-gray-700">
                          {node.data.label || node.id}
                        </div>
                      </div>
                      <div>
                        <label className="block text-sm font-medium mb-1">Node Type</label>
                        <div className="w-full border rounded px-3 py-2 bg-gray-50 text-gray-700">
                          {nodeDef.label}
                        </div>
                      </div>
                      {selectedExecution && (
                        <div>
                          <label className="block text-sm font-medium mb-1">Execution</label>
                          <div className="w-full border rounded px-3 py-2 bg-gray-50 text-gray-700">
                            {new Date(selectedExecution.created_at).toLocaleString()}
                          </div>
                        </div>
                      )}
                    </div>
                  ) : (
                    // In editor mode, show editable configuration
                    <>
                      {/* Node name */}
                      <div>
                        <label className="block text-sm font-medium mb-1">Node Name</label>
                        <input
                          type="text"
                          value={currentFormData.label || ''}
                          onChange={(e) => handleChange('label', e.target.value)}
                          className="w-full border rounded px-3 py-2"
                          placeholder="Enter node name"
                        />
                      </div>

                      {/* Parameters */}
                      {nodeDef.parameters && nodeDef.parameters.map((param) => (
                        <div key={param.name}>
                          <label className="block text-sm font-medium mb-1">
                            {param.label}
                            {param.required && <span className="text-red-500 ml-1">*</span>}
                          </label>
                          {renderParameterField(param)}
                          {param.description && (
                            <p className="text-xs text-gray-500 mt-1">{param.description}</p>
                          )}
                        </div>
                      ))}
                    </>
                  )}
                </div>
              )}
            </div>
          </div>

          {/* Right Panel - Output Data */}
          <div className="w-1/3 border-l flex flex-col">
            <div className="p-3 border-b bg-gray-50">
              <h3 className="font-medium">Output Data</h3>
              {selectedExecution && viewMode === 'executions' && (
                <p className="text-xs text-gray-500">
                  From execution: {new Date(selectedExecution.created_at).toLocaleString()}
                </p>
              )}
            </div>
            <div className="flex-1 p-3 overflow-auto">
              <pre className="text-xs bg-gray-100 p-3 rounded">
                {JSON.stringify(outputData, null, 2) || '// No output data'}
              </pre>
            </div>
          </div>
        </div>

        {/* Bottom Panel - Execution Controls */}
        <div className="border-t p-4 bg-gray-50">
          <div className="flex items-center justify-between">
            {viewMode === 'executions' ? (
              // In executions mode, show execution info
              <div className="flex items-center gap-4">
                <span className="text-sm text-gray-600">
                  Viewing execution data from {selectedExecution ? new Date(selectedExecution.created_at).toLocaleString() : 'unknown time'}
                </span>
              </div>
            ) : (
              // In editor mode, show test controls
              <div className="flex items-center gap-4">
                <button
                  onClick={() => onExecute && onExecute({})}
                  className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700"
                >
                  Test Node
                </button>
                <span className="text-sm text-gray-600">
                  Run this node with current configuration
                </span>
              </div>
            )}
            <div className="text-xs text-gray-500">
              Mode: {viewMode === 'executions' ? 'Execution View' : 'Editor'}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}