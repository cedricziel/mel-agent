import { useState } from 'react';
import DataViewer from './DataViewer';

/**
 * Component to handle node execution testing with input/output display
 * @param {Function} onExecute - Function to execute the node
 * @returns {JSX.Element} Node execution section
 */
export default function NodeExecutionSection({ onExecute }) {
  const [execInput, setExecInput] = useState('{}');
  const [execOutput, setExecOutput] = useState(null);
  const [execError, setExecError] = useState(null);
  const [running, setRunning] = useState(false);

  const handleExecute = async () => {
    setExecError(null);
    setExecOutput(null);

    let parsed;
    try {
      parsed = JSON.parse(execInput);
    } catch {
      setExecError('Invalid JSON');
      return;
    }

    setRunning(true);
    try {
      const out = await onExecute(parsed);
      setExecOutput(out);
    } catch (error) {
      setExecError(error.message || 'Execution error');
    } finally {
      setRunning(false);
    }
  };

  if (!onExecute) {
    return null;
  }

  return (
    <div className="mt-6">
      <h3 className="font-semibold mb-2">Execution</h3>

      <label className="block text-sm mb-1">Input Data (JSON)</label>
      <textarea
        rows={4}
        value={execInput}
        onChange={(e) => setExecInput(e.target.value)}
        className="w-full border rounded px-2 py-1 text-xs font-mono mb-2"
        placeholder="Enter JSON input data for testing..."
      />

      <div className="flex items-center gap-2 mb-2">
        <button
          onClick={handleExecute}
          disabled={running}
          className="px-3 py-1 bg-indigo-600 text-white rounded text-sm hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {running ? 'Running...' : 'Run Node'}
        </button>

        {execError && <div className="text-red-500 text-sm">{execError}</div>}
      </div>

      {execOutput !== null && (
        <div>
          <div className="font-medium text-sm mb-1">Output</div>
          <div className="bg-white border rounded max-h-32 overflow-auto">
            <DataViewer data={execOutput} title="" searchable={false} />
          </div>
        </div>
      )}
    </div>
  );
}
