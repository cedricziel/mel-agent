import CodeEditor from './CodeEditor';

/**
 * Panel for configuring node parameters and settings
 * @param {Object} props - Component props
 * @param {Object} props.node - The node being configured
 * @param {Object} props.nodeDef - The node definition with parameters
 * @param {Object} props.nodes - All nodes in the workflow (for node references)
 * @param {Object} props.currentFormData - Current form data state
 * @param {Object} props.dynamicOptions - Dynamic options for parameters
 * @param {Object} props.loadingOptions - Loading state for dynamic options
 * @param {Object} props.credentials - Available credentials for credential parameters
 * @param {Function} props.handleChange - Handler for form field changes
 * @param {string} props.viewMode - Current view mode ('editor' or 'executions')
 * @param {Object} props.selectedExecution - Selected execution for executions mode
 * @returns {JSX.Element} The configuration panel
 */
export default function NodeConfigurationPanel({
  node,
  nodeDef,
  nodes,
  currentFormData,
  dynamicOptions,
  loadingOptions,
  credentials,
  handleChange,
  viewMode,
  selectedExecution,
}) {
  const renderParameterField = (param) => {
    const val =
      currentFormData[param.name] != null
        ? currentFormData[param.name]
        : param.default;
    const error = param.required && (val === '' || val == null);
    const baseClass = error ? 'border-red-500' : 'border-gray-300';
    const paramType = param.parameterType || param.type;

    const hasDynamicOptions =
      dynamicOptions[param.name] && dynamicOptions[param.name].length > 0;
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

        // Check for code format (from jsonSchema.format)
        if (param.jsonSchema && param.jsonSchema.format === 'code') {
          // Get the language from the node's language parameter
          const nodeLanguage = currentFormData.language || 'javascript';

          return (
            <CodeEditor
              value={val || ''}
              onChange={(code) => handleChange(param.name, code)}
              language={nodeLanguage}
              height="400px"
              placeholder={param.description || 'Enter your code here...'}
            />
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

      case 'credential': {
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
      }

      case 'enum':
        return (
          <select
            value={val || ''}
            onChange={(e) => handleChange(param.name, e.target.value)}
            className={`w-full border rounded px-3 py-2 ${baseClass}`}
          >
            {param.options &&
              param.options.map((opt) => (
                <option key={opt} value={opt}>
                  {opt}
                </option>
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

      case 'nodeReference': {
        // Node reference selection - allows referencing other nodes in the workflow
        const availableNodes = (nodes || []).filter(
          (n) =>
            n.id !== node?.id && // Exclude self-reference
            n.type !== 'manual_trigger' && // Exclude trigger nodes
            n.type !== 'workflow_trigger' &&
            n.type !== 'schedule'
        );

        return (
          <select
            value={val || ''}
            onChange={(e) => handleChange(param.name, e.target.value)}
            className={`w-full border rounded px-3 py-2 ${baseClass}`}
          >
            <option value="">Select Node</option>
            {availableNodes.map((availableNode) => (
              <option key={availableNode.id} value={availableNode.id}>
                {availableNode.data?.label || availableNode.type} (
                {availableNode.type})
              </option>
            ))}
          </select>
        );
      }

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
              <label className="block text-sm font-medium mb-1">
                Execution
              </label>
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
          {nodeDef.parameters &&
            nodeDef.parameters.map((param) => (
              <div key={param.name}>
                <label className="block text-sm font-medium mb-1">
                  {param.label}
                  {param.required && (
                    <span className="text-red-500 ml-1">*</span>
                  )}
                </label>
                {renderParameterField(param)}
                {param.description && (
                  <p className="text-xs text-gray-500 mt-1">
                    {param.description}
                  </p>
                )}
              </div>
            ))}
        </>
      )}
    </div>
  );
}
