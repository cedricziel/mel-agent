import { useState, useEffect } from 'react';
import useCredentials from '../hooks/useCredentials';
import useDynamicOptions from '../hooks/useDynamicOptions';
import ParameterField from './ParameterField';
import NodeExecutionSection from './NodeExecutionSection';

// Panel to configure node details and preview data flow
export default function NodeDetailsPanel({
  node,
  nodeDef,
  onChange,
  onExecute,
  publicUrl,
  readOnly,
  nodes,
  selectedNodeId,
}) {
  // All hooks must be called before any conditional returns
  const [connections, setConnections] = useState([]);
  const [currentFormData, setCurrentFormData] = useState({});

  // Use extracted hooks
  const { credentials, loadingCredentials } = useCredentials(nodeDef);
  const { dynamicOptions, loadingOptions } = useDynamicOptions(
    nodeDef,
    currentFormData
  );

  // Track current form data locally
  useEffect(() => {
    setCurrentFormData({ ...node?.data });
  }, [node?.data]);

  // Helper to update both local state and parent
  const handleChange = (key, value) => {
    setCurrentFormData((prev) => ({ ...prev, [key]: value }));
    onChange(key, value);
  };

  // Load connections for LLM nodes
  useEffect(() => {
    if (nodeDef && nodeDef.type === 'llm') {
      fetch('/api/connections')
        .then((res) => res.json())
        .then((conns) => setConnections(conns))
        .catch(() => setConnections([]));
    }
  }, [nodeDef]);

  // Early returns after all hooks
  if (!node || !nodeDef) return null;

  const { data } = node;

  // In live/read-only mode, show config but disable edits
  if (readOnly) {
    const fallbackKeys = Object.keys(data).filter(
      (k) => k !== 'label' && k !== 'status' && k !== 'lastOutput'
    );
    return (
      <div className="bg-gray-50 p-4 h-full">
        <h2 className="text-lg font-bold mb-4">{nodeDef.label} Details</h2>
        {
          // only show public URL for webhook nodes
          nodeDef.type === 'webhook' && publicUrl && (
            <div className="mb-4">
              <h3 className="font-semibold mb-1">Public URL</h3>
              <input
                type="text"
                readOnly
                value={publicUrl}
                className="w-full border rounded px-2 py-1 text-sm bg-gray-100"
              />
            </div>
          )
        }
        <div className="mb-4">
          <h3 className="font-semibold mb-2">Name</h3>
          <div className="text-sm text-gray-700">{data.label}</div>
        </div>
        <div className="mb-4">
          <h3 className="font-semibold mb-2">Configuration</h3>
          <ul className="list-disc list-inside text-sm text-gray-600">
            {fallbackKeys.map((key) => (
              <li key={key}>
                <strong>{key}:</strong> {JSON.stringify(data[key])}
              </li>
            ))}
            {fallbackKeys.length === 0 && (
              <li className="text-gray-400">No parameters</li>
            )}
          </ul>
        </div>
      </div>
    );
  }

  // fallback data keys
  const fallbackKeys = Object.keys(data).filter(
    (k) => k !== 'label' && k !== 'status' && k !== 'lastOutput'
  );

  return (
    <div className="bg-gray-50 p-4 h-full">
      <h2 className="text-lg font-bold mb-4">{nodeDef.label} Details</h2>

      {/* Public URL for webhook nodes */}
      {nodeDef.type === 'webhook' && (
        <div className="mb-4">
          <h3 className="font-semibold mb-1">Public URL</h3>
          {publicUrl ? (
            <input
              type="text"
              readOnly
              value={publicUrl}
              className="w-full border rounded px-2 py-1 text-sm bg-gray-100"
            />
          ) : (
            <div className="text-sm text-gray-500">
              Save version to generate URL
            </div>
          )}
        </div>
      )}

      {/* Configuration form generated from nodeDef.parameters */}
      <div className="mb-6">
        <h3 className="font-semibold mb-2">Configuration</h3>

        {/* Name field */}
        <div className="mb-4">
          <label className="block text-sm font-medium mb-1">Name</label>
          <input
            type="text"
            value={currentFormData.label || ''}
            onChange={(e) => handleChange('label', e.target.value)}
            className="w-full border rounded px-2 py-1"
          />
        </div>

        {/* Parameter groups */}
        {(() => {
          // Normalize parameters: use server-defined parameters if available, otherwise fallback
          const parameters =
            Array.isArray(nodeDef.parameters) && nodeDef.parameters.length > 0
              ? nodeDef.parameters
              : fallbackKeys.map((key) => ({
                  name: key,
                  label: key,
                  type: 'string',
                  required: false,
                  default: data[key] ?? '',
                  group: 'General',
                  description: '',
                  validators: [],
                }));

          // Filter visible parameters
          const visible = parameters.filter((p) => {
            if (!p.visibilityCondition) return true;
            try {
              // naive evaluation of visibility condition (CEL-like)
              return new Function(
                'data',
                `with(data) { return ${p.visibilityCondition} }`
              )(data);
            } catch {
              return true;
            }
          });

          // Group parameters
          const groups = {};
          visible.forEach((p) => {
            const g = p.group || 'General';
            groups[g] = groups[g] || [];
            groups[g].push(p);
          });

          return Object.entries(groups).map(([group, params]) => (
            <div key={group} className="mb-4">
              <h4 className="text-sm font-semibold mb-2">{group}</h4>
              {params.map((p) => {
                const val =
                  currentFormData[p.name] != null
                    ? currentFormData[p.name]
                    : p.default;

                // Prepare credentials for this parameter
                const paramCredentials = {
                  ...credentials,
                  connections, // Add connections for legacy connectionId support
                };

                return (
                  <ParameterField
                    key={p.name}
                    param={p}
                    value={val}
                    onChange={handleChange}
                    readOnly={readOnly}
                    credentials={paramCredentials}
                    dynamicOptions={dynamicOptions}
                    isLoadingOptions={loadingOptions[p.name]}
                    nodes={nodes}
                    selectedNodeId={selectedNodeId}
                  />
                );
              })}
            </div>
          ));
        })()}
      </div>

      {/* Preview section */}
      <div className="mb-6">
        <h3 className="font-semibold mb-2">Preview</h3>
        <div className="mb-4">
          <div className="font-medium text-sm mb-1">Incoming Data</div>
          <pre className="bg-white border rounded p-2 text-xs h-32 overflow-auto">
            {JSON.stringify({}, null, 2)}
          </pre>
        </div>
        <div>
          <div className="font-medium text-sm mb-1">Outgoing Data</div>
          <pre className="bg-white border rounded p-2 text-xs h-32 overflow-auto">
            {JSON.stringify({}, null, 2)}
          </pre>
        </div>
      </div>

      {/* Execution section */}
      <NodeExecutionSection onExecute={onExecute} />
    </div>
  );
}
