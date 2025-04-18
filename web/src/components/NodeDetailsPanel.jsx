import React, { useState } from 'react';
import CronEditor from './CronEditor';

// Panel to configure node details and preview data flow
export default function NodeDetailsPanel({ node, nodeDef, onChange, onExecute, publicUrl }) {
  if (!node || !nodeDef) return null;
  const { data } = node;
  const defaults = nodeDef.defaults || {};
  const paramKeys = Object.keys(defaults);
  // execution state
  const [execInput, setExecInput] = useState('{}');
  const [execOutput, setExecOutput] = useState(null);
  const [execError, setExecError] = useState(null);
  const [running, setRunning] = useState(false);
  return (
    <div className="w-80 bg-gray-50 border-l p-4 overflow-auto">
      <h2 className="text-lg font-bold mb-4">{nodeDef.label} Details</h2>
      {nodeDef.entry_point && (
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
            value={data.label || ''}
            onChange={(e) => onChange('label', e.target.value)}
            className="w-full border rounded px-2 py-1"
          />
        </div>
        {/* Parameter groups */}
        {(() => {
          // filter visible parameters
          const visible = nodeDef.parameters.filter((p) => {
            if (!p.VisibilityCondition) return true;
            try {
              // naive evaluation: treat condition as JS
              // eslint-disable-next-line no-new-func
              return new Function('data', `with(data) { return ${p.VisibilityCondition} }`)(data);
            } catch {
              return true;
            }
          });
          const groups = {};
          visible.forEach((p) => {
            const g = p.Group || 'General';
            groups[g] = groups[g] || [];
            groups[g].push(p);
          });
          return Object.entries(groups).map(([group, params]) => (
            <div key={group} className="mb-4">
              <h4 className="text-sm font-semibold mb-2">{group}</h4>
              {params.map((p) => {
                const val = data[p.Name] != null ? data[p.Name] : p.Default;
                const error = p.Required && (val === '' || val == null);
                const baseClass = error ? 'border-red-500' : 'border-gray-300';
                switch (p.Type) {
                  case 'string':
                    // Cron editor for schedule nodes
                    if (p.Name === 'cron') {
                      return (
                        <div key={p.Name} className="mb-3">
                          <label className="block text-sm mb-1">
                            {p.Label}{p.Required && <span className="text-red-500">*</span>}
                          </label>
                          <CronEditor
                            value={val}
                            onCronChange={(cron) => onChange(p.Name, cron)}
                          />
                          {error && (
                            <div className="text-xs text-red-600 mt-1">
                              {p.Label} is required
                            </div>
                          )}
                        </div>
                      );
                    }
                    return (
                      <div key={p.Name} className="mb-3">
                        <label className="block text-sm mb-1">
                          {p.Label}{p.Required && <span className="text-red-500">*</span>}
                        </label>
                        <input
                          type="text"
                          value={val}
                          onChange={(e) => onChange(p.Name, e.target.value)}
                          className={`w-full border rounded px-2 py-1 ${baseClass}`}
                        />
                        {error && (
                          <div className="text-xs text-red-600 mt-1">
                            {p.Label} is required
                          </div>
                        )}
                      </div>
                    );
                  case 'number':
                    return (
                      <div key={p.Name} className="mb-3">
                        <label className="block text-sm mb-1">
                          {p.Label}{p.Required && <span className="text-red-500">*</span>}
                        </label>
                        <input
                          type="number"
                          value={val}
                          onChange={(e) => onChange(p.Name, parseFloat(e.target.value) || 0)}
                          className={`w-full border rounded px-2 py-1 ${baseClass}`}
                        />
                        {error && (
                          <div className="text-xs text-red-600 mt-1">
                            {p.Label} is required
                          </div>
                        )}
                      </div>
                    );
                  case 'boolean':
                    return (
                      <div key={p.Name} className="flex items-center mb-3">
                        <input
                          type="checkbox"
                          checked={!!val}
                          onChange={(e) => onChange(p.Name, e.target.checked)}
                          className="mr-2"
                        />
                        <label className="text-sm">
                          {p.Label}{p.Required && <span className="text-red-500">*</span>}
                        </label>
                      </div>
                    );
                  case 'enum':
                    return (
                      <div key={p.Name} className="mb-3">
                        <label className="block text-sm mb-1">
                          {p.Label}{p.Required && <span className="text-red-500">*</span>}
                        </label>
                        <select
                          value={val}
                          onChange={(e) => onChange(p.Name, e.target.value)}
                          className={`w-full border rounded px-2 py-1 ${baseClass}`}
                        >
                          {p.Options.map((opt) => (
                            <option key={opt} value={opt}>{opt}</option>
                          ))}
                        </select>
                        {error && (
                          <div className="text-xs text-red-600 mt-1">
                            {p.Label} is required
                          </div>
                        )}
                      </div>
                    );
                  case 'json':
                    return (
                      <div key={p.Name} className="mb-3">
                        <label className="block text-sm mb-1">
                          {p.Label}{p.Required && <span className="text-red-500">*</span>}
                        </label>
                        <textarea
                          rows={4}
                          value={JSON.stringify(val, null, 2)}
                          onChange={(e) => {
                            try {
                              const v = JSON.parse(e.target.value);
                              onChange(p.Name, v);
                            } catch {
                              // ignore parse error
                              onChange(p.Name, e.target.value);
                            }
                          }}
                          className={`w-full border rounded px-2 py-1 font-mono text-xs ${baseClass}`}
                        />
                        {error && (
                          <div className="text-xs text-red-600 mt-1">
                            {p.Label} is required
                          </div>
                        )}
                      </div>
                    );
                  default:
                    return null;
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
              <pre className="bg-white border rounded p-2 text-xs h-32 overflow-auto">
                {JSON.stringify(execOutput, null, 2)}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  );
}