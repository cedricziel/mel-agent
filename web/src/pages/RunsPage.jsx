import { useEffect, useState } from 'react';
import axios from 'axios';
import { useParams, Link } from 'react-router-dom';

export default function RunsPage() {
  const { id: agentId } = useParams();
  const [runs, setRuns] = useState([]);
  const [selectedRunID, setSelectedRunID] = useState(null);
  const [runDetails, setRunDetails] = useState(null);

  useEffect(() => {
    axios.get(`/api/agents/${agentId}/runs`)
      .then(res => setRuns(res.data))
      .catch(err => console.error('fetch runs list failed', err));
  }, [agentId]);

  useEffect(() => {
    if (selectedRunID) {
      axios.get(`/api/agents/${agentId}/runs/${selectedRunID}`)
        .then(res => setRunDetails(res.data))
        .catch(err => console.error('fetch run details failed', err));
    }
  }, [selectedRunID, agentId]);

  return (
    <div className="flex h-full">
      <div className="w-1/4 border-r p-4 overflow-auto">
        <h2 className="text-xl font-bold mb-4">Runs for Agent {agentId}</h2>
        <ul className="space-y-2">
          {runs.map(run => (
            <li key={run.id}>
              <button
                onClick={() => setSelectedRunID(run.id)}
                className={`w-full text-left px-2 py-1 rounded ${run.id === selectedRunID ? 'bg-gray-200' : 'hover:bg-gray-100'}`}
              >
                {run.created_at}
              </button>
            </li>
          ))}
        </ul>
        <div className="mt-4">
          <Link to={`/agents/${agentId}/edit`} className="text-indigo-600 hover:underline">
            ← Back to Builder
          </Link>
        </div>
      </div>
      <div className="flex-1 p-4 overflow-auto">
        <h2 className="text-xl font-bold mb-4">Run Details</h2>
        {!runDetails && <p className="text-gray-500">Select a run to view details.</p>}
        {runDetails && (
          <div>
            {runDetails.trace && runDetails.trace.map((step, idx) => (
              <div key={idx} className="mb-6">
                <h3 className="font-semibold">Node: {step.nodeId}</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <h4 className="font-medium">Input</h4>
                    <pre className="bg-gray-100 p-2 text-xs overflow-auto">
                      {JSON.stringify(step.input, null, 2)}
                    </pre>
                  </div>
                  <div>
                    <h4 className="font-medium">Output</h4>
                    <pre className="bg-gray-100 p-2 text-xs overflow-auto">
                      {JSON.stringify(step.output, null, 2)}
                    </pre>
                  </div>
                </div>
              </div>
            ))}
            {runDetails.meta && (
              <div className="mt-4">
                <h3 className="font-semibold">Meta</h3>
                <pre className="bg-gray-100 p-2 text-xs overflow-auto">
                  {JSON.stringify(runDetails.meta, null, 2)}
                </pre>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}