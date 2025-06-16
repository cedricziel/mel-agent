import { useEffect } from 'react';
import DataViewer from './DataViewer';
import NodeConfigurationPanel from './NodeConfigurationPanel';
import NodeExecutionPanel from './NodeExecutionPanel';
import { useNodeModalState } from '../hooks/useNodeModalState';

// Full-screen modal for editing nodes with input/output panels like n8n
export default function NodeModal({
  node,
  nodeDef,
  nodes,
  isOpen,
  onClose,
  onChange,
  onExecute,
  onSave,
  viewMode,
  selectedExecution,
  agentId,
}) {
  // Use the extracted state management hook
  const {
    currentFormData,
    dynamicOptions,
    loadingOptions,
    credentials,
    handleChange,
    inputData,
    outputData,
    setOutputData,
    activeTab,
    setActiveTab,
    loadNodeExecutionData,
  } = useNodeModalState(node, nodeDef, onChange);

  // Load execution data when modal opens in executions mode
  useEffect(() => {
    if (isOpen && selectedExecution && viewMode === 'executions') {
      loadNodeExecutionData(selectedExecution, agentId, viewMode);
    }
  }, [isOpen, selectedExecution, viewMode, agentId, loadNodeExecutionData]);

  if (!isOpen || !node || !nodeDef) return null;

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
            <div className="flex-1 overflow-auto">
              <DataViewer
                data={inputData}
                title="Input Data"
                searchable={true}
              />
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
                <NodeConfigurationPanel
                  node={node}
                  nodeDef={nodeDef}
                  nodes={nodes}
                  currentFormData={currentFormData}
                  dynamicOptions={dynamicOptions}
                  loadingOptions={loadingOptions}
                  credentials={credentials}
                  handleChange={handleChange}
                  viewMode={viewMode}
                  selectedExecution={selectedExecution}
                />
              )}
            </div>
          </div>

          {/* Right Panel - Output Data */}
          <div className="w-1/3 border-l flex flex-col">
            <div className="p-3 border-b bg-gray-50">
              <h3 className="font-medium">Output Data</h3>
              {selectedExecution && viewMode === 'executions' && (
                <p className="text-xs text-gray-500">
                  From execution:{' '}
                  {new Date(selectedExecution.created_at).toLocaleString()}
                </p>
              )}
            </div>
            <div className="flex-1 overflow-auto">
              <DataViewer
                data={outputData}
                title="Output Data"
                searchable={true}
              />
            </div>
          </div>
        </div>

        {/* Bottom Panel - Execution Controls */}
        <NodeExecutionPanel
          viewMode={viewMode}
          selectedExecution={selectedExecution}
          onExecute={onExecute}
          outputData={outputData}
          setOutputData={setOutputData}
        />
      </div>
    </div>
  );
}
