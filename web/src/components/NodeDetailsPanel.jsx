import React, { useState } from 'react';
import CronEditor from './CronEditor';

// Panel to configure node details and preview data flow
export default function NodeDetailsPanel({ node, nodeDef, onChange, onExecute }) {
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
      <div className="mb-6">
        <h3 className="font-semibold mb-2">Configuration</h3>
        <label className="block text-sm mb-1">Label</label>
        <input
          type="text"
          value={data.label || ''}
          onChange={(e) => onChange('label', e.target.value)}
          className="w-full border rounded px-2 py-1 mb-3"
        />
        {paramKeys.map((key) => {
          if (nodeDef.type === 'schedule' && key === 'cron') {
            return (
              <div key={key} className="mb-3">
                <label className="block text-sm mb-1">Schedule</label>
                <CronEditor
                  value={data.cron || ''}
                  onCronChange={(val) => onChange('cron', val)}
                />
              </div>
            );
          }
          return (
            <div key={key} className="mb-3">
              <label className="block text-sm mb-1">{key}</label>
              <input
                type="text"
                value={data[key] || ''}
                onChange={(e) => onChange(key, e.target.value)}
                className="w-full border rounded px-2 py-1"
              />
            </div>
          );
        })}
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