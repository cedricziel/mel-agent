import { Link } from 'react-router-dom';

export default function RunsList({
  agentId,
  runs,
  selectedRunID,
  onRunSelect,
}) {
  return (
    <div className="w-1/4 border-r p-4 overflow-auto h-full">
      <h2 className="text-xl font-bold mb-4">Runs for Agent {agentId}</h2>
      <ul className="space-y-2">
        {runs.map((run) => (
          <li key={run.id}>
            <button
              onClick={() => onRunSelect(run.id)}
              className={`w-full text-left px-2 py-1 rounded ${
                run.id === selectedRunID ? 'bg-gray-200' : 'hover:bg-gray-100'
              }`}
            >
              {run.created_at}
            </button>
          </li>
        ))}
      </ul>
      <div className="mt-4">
        <Link
          to={`/agents/${agentId}/edit`}
          className="text-indigo-600 hover:underline"
        >
          ← Back to Builder
        </Link>
      </div>
    </div>
  );
}
