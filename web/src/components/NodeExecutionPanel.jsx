/**
 * Panel for node execution controls and status display
 * @param {Object} props - Component props
 * @param {string} props.viewMode - Current view mode ('editor' or 'executions')
 * @param {Object} props.selectedExecution - Selected execution for executions mode
 * @param {Function} props.onExecute - Handler for executing the node
 * @param {Object} props.outputData - Current output data
 * @param {Function} props.setOutputData - Setter for output data
 * @returns {JSX.Element} The execution panel
 */
export default function NodeExecutionPanel({
  viewMode,
  selectedExecution,
  onExecute,
  outputData,
  setOutputData,
}) {
  const handleTestNode = async () => {
    if (onExecute) {
      try {
        const result = await onExecute({});
        setOutputData(result || {});
      } catch (error) {
        console.error('Node execution error:', error);
        setOutputData({
          error: error.message || 'Execution failed',
        });
      }
    }
  };

  return (
    <div className="border-t p-4 bg-gray-50">
      <div className="flex items-center justify-between">
        {viewMode === 'executions' ? (
          // In executions mode, show execution info
          <div className="flex items-center gap-4">
            <span className="text-sm text-gray-600">
              Viewing execution data from{' '}
              {selectedExecution
                ? new Date(selectedExecution.created_at).toLocaleString()
                : 'unknown time'}
            </span>
          </div>
        ) : (
          // In editor mode, show test controls
          <div className="flex items-center gap-4">
            <button
              onClick={handleTestNode}
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
  );
}
