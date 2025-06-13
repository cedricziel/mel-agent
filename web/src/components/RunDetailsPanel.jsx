import React from 'react';

// Panel for inspecting a node's input/output within a run
export default function RunDetailsPanel({ nodeDef, step }) {
  if (!nodeDef || !step) return null;
  return (
    <div className="h-full flex flex-col">
      <h2 className="text-lg font-bold mb-4">{nodeDef.label}</h2>
      <div className="mb-4">
        <h3 className="font-medium mb-1">Input</h3>
        <pre className="bg-gray-100 p-2 text-xs overflow-auto flex-1">
          {JSON.stringify(step.input, null, 2)}
        </pre>
      </div>
      <div className="flex-1">
        <h3 className="font-medium mb-1">Output</h3>
        <pre className="bg-gray-100 p-2 text-xs overflow-auto flex-1">
          {JSON.stringify(step.output, null, 2)}
        </pre>
      </div>
    </div>
  );
}
