import React from 'react';

// Panel to configure node details and preview data flow
export default function NodeDetailsPanel({ node, nodeDef, onChange }) {
  if (!node || !nodeDef) return null;
  const { data } = node;
  const defaults = nodeDef.defaults || {};
  const paramKeys = Object.keys(defaults);
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
        {paramKeys.map((key) => (
          <div key={key} className="mb-3">
            <label className="block text-sm mb-1">{key}</label>
            <input
              type="text"
              value={data[key] || ''}
              onChange={(e) => onChange(key, e.target.value)}
              className="w-full border rounded px-2 py-1"
            />
          </div>
        ))}
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
    </div>
  );
}