import CronEditor from './CronEditor';
import CodeEditor from './CodeEditor';

/**
 * Component to render individual parameter fields based on parameter type
 * @param {Object} param - Parameter definition
 * @param {*} value - Current parameter value
 * @param {Function} onChange - Change handler
 * @param {boolean} readOnly - Whether field is read-only
 * @param {Object} credentials - Available credentials for credential parameters
 * @param {Object} dynamicOptions - Dynamic options for dynamic parameters
 * @param {boolean} isLoadingOptions - Whether options are loading
 * @param {Array} nodes - Available nodes for nodeReference parameters
 * @param {string} selectedNodeId - Currently selected node ID
 * @returns {JSX.Element} Rendered parameter field
 */
export default function ParameterField({
  param,
  value,
  onChange,
  readOnly = false,
  credentials = {},
  dynamicOptions = {},
  isLoadingOptions = false,
  nodes = [],
  selectedNodeId,
}) {
  const error = param.required && (value === '' || value == null);
  const baseClass = error ? 'border-red-500' : 'border-gray-300';
  const paramType = param.parameterType || param.type;

  // Check if this parameter has dynamic options available or is marked as dynamic
  const hasDynamicOptions =
    dynamicOptions[param.name] && dynamicOptions[param.name].length > 0;
  const isDynamicParameter = param.dynamicOptions;

  const renderField = () => {
    switch (paramType) {
      case 'string':
        // Check for specific UI components based on parameter name/metadata
        if (param.name === 'cron') {
          return (
            <CronEditor
              value={value}
              onCronChange={(cron) => onChange(param.name, cron)}
            />
          );
        }

        // Check for code format (from jsonSchema.format)
        if (param.jsonSchema && param.jsonSchema.format === 'code') {
          // Get the language from the node's language parameter
          const nodeLanguage = 'javascript'; // Default, could be passed as prop

          return (
            <CodeEditor
              value={value || ''}
              onChange={(code) => onChange(param.name, code)}
              language={nodeLanguage}
              height="300px"
              placeholder={param.description || 'Enter your code here...'}
              readOnly={readOnly}
            />
          );
        }

        // Special handling for legacy connectionId (LLM nodes)
        if (
          param.name === 'connectionId' &&
          credentials.connections?.length > 0
        ) {
          return (
            <select
              value={value || ''}
              onChange={(e) => onChange(param.name, e.target.value)}
              className={`w-full border rounded px-2 py-1 ${baseClass}`}
            >
              <option value="">Select Connection</option>
              {credentials.connections.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          );
        }

        // Dynamic parameters always show as select (with loading/empty states)
        if (isDynamicParameter) {
          return (
            <select
              value={value || ''}
              onChange={(e) => onChange(param.name, e.target.value)}
              className={`w-full border rounded px-2 py-1 ${baseClass}`}
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

        // Default string input
        return (
          <input
            type="text"
            value={value || ''}
            onChange={(e) => onChange(param.name, e.target.value)}
            className={`w-full border rounded px-2 py-1 ${baseClass}`}
            placeholder={param.description}
          />
        );

      case 'number':
      case 'integer':
        return (
          <input
            type="number"
            value={value || ''}
            onChange={(e) =>
              onChange(param.name, parseFloat(e.target.value) || 0)
            }
            className={`w-full border rounded px-2 py-1 ${baseClass}`}
            placeholder={param.description}
          />
        );

      case 'boolean':
        return (
          <div className="flex items-center">
            <input
              type="checkbox"
              checked={!!value}
              onChange={(e) => onChange(param.name, e.target.checked)}
              className="mr-2"
            />
            <label className="text-sm">
              {param.label}
              {param.required && <span className="text-red-500">*</span>}
            </label>
          </div>
        );

      case 'enum':
        return (
          <select
            value={value || ''}
            onChange={(e) => onChange(param.name, e.target.value)}
            className={`w-full border rounded px-2 py-1 ${baseClass}`}
          >
            {param.options &&
              param.options.map((opt) => (
                <option key={opt} value={opt}>
                  {opt}
                </option>
              ))}
          </select>
        );

      case 'json':
      case 'object':
        return (
          <textarea
            rows={4}
            value={
              typeof value === 'string'
                ? value
                : JSON.stringify(value || {}, null, 2)
            }
            onChange={(e) => {
              try {
                const v = JSON.parse(e.target.value);
                onChange(param.name, v);
              } catch {
                // ignore parse error, store as string temporarily
                onChange(param.name, e.target.value);
              }
            }}
            className={`w-full border rounded px-2 py-1 font-mono text-xs ${baseClass}`}
            placeholder={param.description}
          />
        );

      case 'credential': {
        // Dynamic credential selection
        const paramCredentials = credentials[param.name] || [];
        return (
          <select
            value={value || ''}
            onChange={(e) => onChange(param.name, e.target.value)}
            className={`w-full border rounded px-2 py-1 ${baseClass}`}
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

      case 'nodeReference': {
        // Node reference selection - allows referencing other nodes in the workflow
        const availableNodes = (nodes || []).filter(
          (n) =>
            n.id !== selectedNodeId && // Exclude self-reference
            n.type !== 'manual_trigger' && // Exclude trigger nodes
            n.type !== 'workflow_trigger' &&
            n.type !== 'schedule'
        );

        return (
          <>
            <select
              value={value || ''}
              onChange={(e) => onChange(param.name, e.target.value)}
              className={`w-full border rounded px-2 py-1 ${baseClass}`}
            >
              <option value="">Select Node</option>
              {availableNodes.map((node) => (
                <option key={node.id} value={node.id}>
                  {node.data?.label || node.type} ({node.type})
                </option>
              ))}
            </select>
            {availableNodes.length === 0 && (
              <div className="text-xs text-gray-500 mt-1">
                No compatible nodes available. Add some nodes to the workflow
                first.
              </div>
            )}
          </>
        );
      }

      default:
        // Fallback for unknown parameter types - treat as string
        return (
          <input
            type="text"
            value={value || ''}
            onChange={(e) => onChange(param.name, e.target.value)}
            className={`w-full border rounded px-2 py-1 ${baseClass}`}
            placeholder={param.description}
          />
        );
    }
  };

  // Special case for boolean - different layout
  if (paramType === 'boolean') {
    return (
      <div key={param.name} className="mb-3">
        {renderField()}
        {param.description && (
          <div className="text-xs text-gray-500 mt-1">{param.description}</div>
        )}
      </div>
    );
  }

  return (
    <div key={param.name} className="mb-3">
      <label className="block text-sm mb-1">
        {param.label}
        {param.required && <span className="text-red-500">*</span>}
      </label>
      {renderField()}
      {error && (
        <div className="text-xs text-red-600 mt-1">
          {param.label} is required
        </div>
      )}
      {param.description && (
        <div className="text-xs text-gray-500 mt-1">{param.description}</div>
      )}
    </div>
  );
}
