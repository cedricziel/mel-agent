import React, { useState, useEffect } from 'react';
import CronEditor from './CronEditor';
import CodeEditor from './CodeEditor';
import DataViewer from './DataViewer';

// Panel to configure node details and preview data flow
export default function NodeDetailsPanel({ node, nodeDef, onChange, onExecute, publicUrl, readOnly }) {
  // All hooks must be called before any conditional returns
  const [connections, setConnections] = useState([]);
  const [credentials, setCredentials] = useState([]);
  const [dynamicOptions, setDynamicOptions] = useState({});
  const [loadingOptions, setLoadingOptions] = useState({});
  const [currentFormData, setCurrentFormData] = useState({});
  const [execInput, setExecInput] = useState('{}');
  const [execOutput, setExecOutput] = useState(null);
  const [execError, setExecError] = useState(null);
  const [running, setRunning] = useState(false);

  // Track current form data locally
  useEffect(() => {
    setCurrentFormData({ ...node?.data });
  }, [node?.data]);

  // Helper to update both local state and parent
  const handleChange = (key, value) => {
    setCurrentFormData(prev => ({ ...prev, [key]: value }));
    onChange(key, value);
  };

  useEffect(() => {
    if (nodeDef && nodeDef.type === 'llm') {
      fetch('/api/connections')
        .then((res) => res.json())
        .then((conns) => setConnections(conns))
        .catch(() => setConnections([]));
    }
  }, [nodeDef]);

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

  // Function to load dynamic options for a specific parameter
  const loadDynamicOptionsForParam = async (paramName) => {
    const param = nodeDef?.parameters?.find(p => p.name === paramName);
    if (!param?.dynamicOptions) return;
    
    try {
      // Set loading state for this specific parameter
      setLoadingOptions(prev => ({ ...prev, [paramName]: true }));
      
      // Build query parameters from current form data
      const queryParams = new URLSearchParams();
      Object.entries(currentFormData || {}).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          queryParams.append(key, value);
        }
      });
      
      const url = `/api/node-types/${nodeDef.type}/parameters/${paramName}/options?${queryParams}`;
      console.log(`Loading dynamic options for ${paramName}:`, url);
      
      const response = await fetch(url);
      
      if (response.ok) {
        const data = await response.json();
        console.log(`Dynamic options for ${paramName}:`, data);
        setDynamicOptions(prev => ({ 
          ...prev, 
          [paramName]: data.options || [] 
        }));
      } else {
        console.log(`No dynamic options for ${paramName}:`, response.status, await response.text());
        setDynamicOptions(prev => ({ 
          ...prev, 
          [paramName]: [] 
        }));
      }
    } catch (error) {
      console.log(`Error loading dynamic options for ${paramName}:`, error);
      setDynamicOptions(prev => ({ 
        ...prev, 
        [paramName]: [] 
      }));
    } finally {
      // Clear loading state for this specific parameter
      setLoadingOptions(prev => ({ ...prev, [paramName]: false }));
    }
  };

  // Load initial dynamic options when credential changes
  useEffect(() => {
    if (!nodeDef || !currentFormData?.credentialId) return;
    
    // Only load database options initially
    loadDynamicOptionsForParam('databaseId');
  }, [nodeDef, currentFormData?.credentialId]);

  // Load table options when database changes
  useEffect(() => {
    if (!nodeDef || !currentFormData?.databaseId) {
      // Clear table options when no database is selected
      setDynamicOptions(prev => ({ ...prev, tableId: [] }));
      return;
    }
    
    loadDynamicOptionsForParam('tableId');
  }, [nodeDef, currentFormData?.databaseId]);

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
        {// only show public URL for webhook nodes
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
        )}
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
  const fallbackKeys = Object.keys(data).filter((k) => k !== 'label' && k !== 'status' && k !== 'lastOutput');
  return (
    <div className="bg-gray-50 p-4 h-full">
      <h2 className="text-lg font-bold mb-4">{nodeDef.label} Details</h2>
      {// only show public URL for webhook nodes
      nodeDef.type === 'webhook' && (
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
            <div className="text-sm text-gray-500">Save version to generate URL</div>
          )}
        </div>
      )}
      {/* Configuration form generated from nodeDef.parameters */}
      <div className="mb-6">
        <h3 className="font-semibold mb-2">Configuration</h3>
        {/* Rename field */}
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
          // parameters may come from schema or legacy defaults
          // Normalize parameters: use server-defined parameters if available, otherwise fallback
          const parameters = Array.isArray(nodeDef.parameters) && nodeDef.parameters.length > 0
            ? nodeDef.parameters
            : fallbackKeys.map((key) => ({
                name: key,
                label: key,
                type: 'string',
                required: false,
                default: data[key] ?? '',
                group: 'General',
                description: '',
                validators: []
              }));
          // filter visible parameters
          const visible = parameters.filter((p) => {
            if (!p.visibilityCondition) return true;
            try {
              // naive evaluation of visibility condition (CEL-like)
              // eslint-disable-next-line no-new-func
              return new Function('data', `with(data) { return ${p.visibilityCondition} }`)(data);
            } catch {
              return true;
            }
          });
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
                const val = currentFormData[p.name] != null ? currentFormData[p.name] : p.default;
                const error = p.required && (val === '' || val == null);
                const baseClass = error ? 'border-red-500' : 'border-gray-300';
                // Use parameterType if available, otherwise fall back to type
                const paramType = p.parameterType || p.type;
                
                // Check if this parameter has dynamic options available or is marked as dynamic
                const hasDynamicOptions = dynamicOptions[p.name] && dynamicOptions[p.name].length > 0;
                const isLoadingOptions = loadingOptions[p.name];
                const isDynamicParameter = p.dynamicOptions;
                
                switch (paramType) {
                  case 'string':
                    // Check for specific UI components based on parameter name/metadata
                    if (p.name === 'cron') {
                      return (
                        <div key={p.name} className="mb-3">
                          <label className="block text-sm mb-1">
                            {p.label}{p.required && <span className="text-red-500">*</span>}
                          </label>
                          <CronEditor
                            value={val}
                            onCronChange={(cron) => handleChange(p.name, cron)}
                          />
                          {error && (
                            <div className="text-xs text-red-600 mt-1">
                              {p.label} is required
                            </div>
                          )}
                        </div>
                      );
                    }
                    
                    // Check for code format (from jsonSchema.format)
                    if (p.jsonSchema && p.jsonSchema.format === 'code') {
                      // Get the language from the node's language parameter
                      const nodeLanguage = currentFormData.language || 'javascript';
                      
                      return (
                        <div key={p.name} className="mb-3">
                          <label className="block text-sm mb-1">
                            {p.label}{p.required && <span className="text-red-500">*</span>}
                          </label>
                          <CodeEditor
                            value={val || ''}
                            onChange={(code) => handleChange(p.name, code)}
                            language={nodeLanguage}
                            height="300px"
                            placeholder={p.description || 'Enter your code here...'}
                            readOnly={readOnly}
                          />
                          {error && (
                            <div className="text-xs text-red-600 mt-1">
                              {p.label} is required
                            </div>
                          )}
                          {p.description && (
                            <div className="text-xs text-gray-500 mt-1">
                              {p.description}
                            </div>
                          )}
                        </div>
                      );
                    }
                    
                    // Special handling for legacy connectionId (LLM nodes)
                    if (p.name === 'connectionId' && connections.length > 0) {
                      return (
                        <div key={p.name} className="mb-3">
                          <label className="block text-sm mb-1">
                            {p.label}{p.required && <span className="text-red-500">*</span>}
                          </label>
                          <select
                            value={val || ''}
                            onChange={(e) => handleChange(p.name, e.target.value)}
                            className={`w-full border rounded px-2 py-1 ${baseClass}`}
                          >
                            <option value="">Select Connection</option>
                            {connections.map((c) => (
                              <option key={c.id} value={c.id}>
                                {c.name}
                              </option>
                            ))}
                          </select>
                          {error && (
                            <div className="text-xs text-red-600 mt-1">
                              {p.label} is required
                            </div>
                          )}
                        </div>
                      );
                    }
                    
                    // Dynamic parameters always show as select (with loading/empty states)
                    if (isDynamicParameter) {
                      return (
                        <div key={p.name} className="mb-3">
                          <label className="block text-sm mb-1">
                            {p.label}{p.required && <span className="text-red-500">*</span>}
                          </label>
                          <select
                            value={val || ''}
                            onChange={(e) => handleChange(p.name, e.target.value)}
                            className={`w-full border rounded px-2 py-1 ${baseClass}`}
                            disabled={isLoadingOptions}
                          >
                            {isLoadingOptions ? (
                              <option value="">Loading...</option>
                            ) : hasDynamicOptions ? (
                              <>
                                <option value="">Select {p.label}</option>
                                {dynamicOptions[p.name].map((option) => (
                                  <option key={option.value} value={option.value}>
                                    {option.label}
                                  </option>
                                ))}
                              </>
                            ) : (
                              <option value="">No options available</option>
                            )}
                          </select>
                          {error && (
                            <div className="text-xs text-red-600 mt-1">
                              {p.label} is required
                            </div>
                          )}
                          {p.description && (
                            <div className="text-xs text-gray-500 mt-1">
                              {p.description}
                            </div>
                          )}
                        </div>
                      );
                    }
                    
                    // Default string input
                    return (
                      <div key={p.name} className="mb-3">
                        <label className="block text-sm mb-1">
                          {p.label}{p.required && <span className="text-red-500">*</span>}
                        </label>
                        <input
                          type="text"
                          value={val || ''}
                          onChange={(e) => handleChange(p.name, e.target.value)}
                          className={`w-full border rounded px-2 py-1 ${baseClass}`}
                          placeholder={p.description}
                        />
                        {error && (
                          <div className="text-xs text-red-600 mt-1">
                            {p.label} is required
                          </div>
                        )}
                        {p.description && (
                          <div className="text-xs text-gray-500 mt-1">
                            {p.description}
                          </div>
                        )}
                      </div>
                    );
                  case 'number':
                  case 'integer':
                    return (
                      <div key={p.name} className="mb-3">
                        <label className="block text-sm mb-1">
                          {p.label}{p.required && <span className="text-red-500">*</span>}
                        </label>
                        <input
                          type="number"
                          value={val || ''}
                          onChange={(e) => handleChange(p.name, parseFloat(e.target.value) || 0)}
                          className={`w-full border rounded px-2 py-1 ${baseClass}`}
                          placeholder={p.description}
                        />
                        {error && (
                          <div className="text-xs text-red-600 mt-1">
                            {p.label} is required
                          </div>
                        )}
                        {p.description && (
                          <div className="text-xs text-gray-500 mt-1">
                            {p.description}
                          </div>
                        )}
                      </div>
                    );
                  case 'boolean':
                    return (
                      <div key={p.name} className="mb-3">
                        <div className="flex items-center">
                          <input
                            type="checkbox"
                            checked={!!val}
                            onChange={(e) => handleChange(p.name, e.target.checked)}
                            className="mr-2"
                          />
                          <label className="text-sm">
                            {p.label}{p.required && <span className="text-red-500">*</span>}
                          </label>
                        </div>
                        {p.description && (
                          <div className="text-xs text-gray-500 mt-1">
                            {p.description}
                          </div>
                        )}
                      </div>
                    );
                  case 'enum':
                    return (
                      <div key={p.name} className="mb-3">
                        <label className="block text-sm mb-1">
                          {p.label}{p.required && <span className="text-red-500">*</span>}
                        </label>
                        <select
                          value={val || ''}
                          onChange={(e) => handleChange(p.name, e.target.value)}
                          className={`w-full border rounded px-2 py-1 ${baseClass}`}
                        >
                          {p.options && p.options.map((opt) => (
                            <option key={opt} value={opt}>{opt}</option>
                          ))}
                        </select>
                        {error && (
                          <div className="text-xs text-red-600 mt-1">
                            {p.label} is required
                          </div>
                        )}
                        {p.description && (
                          <div className="text-xs text-gray-500 mt-1">
                            {p.description}
                          </div>
                        )}
                      </div>
                    );
                  case 'json':
                  case 'object':
                    return (
                      <div key={p.name} className="mb-3">
                        <label className="block text-sm mb-1">
                          {p.label}{p.required && <span className="text-red-500">*</span>}
                        </label>
                        <textarea
                          rows={4}
                          value={typeof val === 'string' ? val : JSON.stringify(val || {}, null, 2)}
                          onChange={(e) => {
                            try {
                              const v = JSON.parse(e.target.value);
                              handleChange(p.name, v);
                            } catch {
                              // ignore parse error, store as string temporarily
                              handleChange(p.name, e.target.value);
                            }
                          }}
                          className={`w-full border rounded px-2 py-1 font-mono text-xs ${baseClass}`}
                          placeholder={p.description}
                        />
                        {error && (
                          <div className="text-xs text-red-600 mt-1">
                            {p.label} is required
                          </div>
                        )}
                        {p.description && (
                          <div className="text-xs text-gray-500 mt-1">
                            {p.description}
                          </div>
                        )}
                      </div>
                    );
                  case 'credential':
                    // Dynamic credential selection
                    const paramCredentials = credentials[p.name] || [];
                    return (
                      <div key={p.name} className="mb-3">
                        <label className="block text-sm mb-1">
                          {p.label}{p.required && <span className="text-red-500">*</span>}
                        </label>
                        <select
                          value={val || ''}
                          onChange={(e) => handleChange(p.name, e.target.value)}
                          className={`w-full border rounded px-2 py-1 ${baseClass}`}
                        >
                          <option value="">Select Credential</option>
                          {paramCredentials.map((cred) => (
                            <option key={cred.id} value={cred.id}>
                              {cred.name} ({cred.integration_name})
                            </option>
                          ))}
                        </select>
                        {error && (
                          <div className="text-xs text-red-600 mt-1">
                            {p.label} is required
                          </div>
                        )}
                        {p.description && (
                          <div className="text-xs text-gray-500 mt-1">
                            {p.description}
                          </div>
                        )}
                      </div>
                    );
                  default:
                    // Fallback for unknown parameter types - treat as string
                    return (
                      <div key={p.name} className="mb-3">
                        <label className="block text-sm mb-1">
                          {p.label}{p.required && <span className="text-red-500">*</span>}
                        </label>
                        <input
                          type="text"
                          value={val || ''}
                          onChange={(e) => handleChange(p.name, e.target.value)}
                          className={`w-full border rounded px-2 py-1 ${baseClass}`}
                          placeholder={p.description}
                        />
                        {error && (
                          <div className="text-xs text-red-600 mt-1">
                            {p.label} is required
                          </div>
                        )}
                        {p.description && (
                          <div className="text-xs text-gray-500 mt-1">
                            {p.description}
                          </div>
                        )}
                      </div>
                    );
                }
              })}
            </div>
          ));
        })()}
      </div>
      <div>
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
      {onExecute && (
        <div className="mt-6">
          <h3 className="font-semibold mb-2">Execution</h3>
          <label className="block text-sm mb-1">Input Data (JSON)</label>
          <textarea
            rows={4}
            value={execInput}
            onChange={(e) => setExecInput(e.target.value)}
            className="w-full border rounded px-2 py-1 text-xs font-mono mb-2"
          />
          <div className="flex items-center gap-2 mb-2">
            <button
              onClick={async () => {
                setExecError(null);
                setExecOutput(null);
                let parsed;
                try {
                  parsed = JSON.parse(execInput);
                } catch (err) {
                  setExecError('Invalid JSON');
                  return;
                }
                setRunning(true);
                try {
                  const out = await onExecute(parsed);
                  setExecOutput(out);
                } catch (err) {
                  setExecError(err.message || 'Execution error');
                } finally {
                  setRunning(false);
                }
              }}
              disabled={running}
              className="px-3 py-1 bg-indigo-600 text-white rounded text-sm"
            >
              {running ? 'Running...' : 'Run Node'}
            </button>
            {execError && <div className="text-red-500 text-sm">{execError}</div>}
          </div>
          {execOutput !== null && (
            <div>
              <div className="font-medium text-sm mb-1">Output</div>
              <div className="bg-white border rounded max-h-32 overflow-auto">
                <DataViewer 
                  data={execOutput} 
                  title="" 
                  searchable={false}
                />
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}