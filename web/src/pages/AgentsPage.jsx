import { useEffect, useState } from 'react';
import axios from 'axios';

function AgentsPage() {
  const [agents, setAgents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [form, setForm] = useState({ name: '', description: '' });

  useEffect(() => {
    fetchAgents();
  }, []);

  async function fetchAgents() {
    setLoading(true);
    try {
      const res = await axios.get('/api/agents');
      setAgents(res.data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }

  async function submit(e) {
    e.preventDefault();
    try {
      await axios.post('/api/agents', form);
      setForm({ name: '', description: '' });
      setModalOpen(false);
      fetchAgents();
    } catch (err) {
      console.error(err);
      alert('Failed to create agent');
    }
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-bold">Agents</h2>
        <button
          onClick={() => setModalOpen(true)}
          className="bg-indigo-600 text-white py-1.5 px-3 rounded"
        >
          + New Agent
        </button>
      </div>

      {loading ? (
        <p>Loading…</p>
      ) : (
        <table className="min-w-full bg-white shadow rounded">
          <thead>
            <tr>
              <th className="px-4 py-2 text-left">Name</th>
              <th className="px-4 py-2 text-left">Description</th>
            </tr>
          </thead>
          <tbody>
            {agents.map((a) => (
              <tr
                key={a.id}
                className="border-t cursor-pointer hover:bg-gray-50"
                onClick={() => (window.location.href = `/agents/${a.id}/edit`)}
              >
                <td className="px-4 py-2">{a.name}</td>
                <td className="px-4 py-2 text-gray-600">{a.description}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {modalOpen && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center">
          <form onSubmit={submit} className="bg-white p-6 rounded shadow w-96">
            <h3 className="text-lg font-semibold mb-4">Create Agent</h3>
            <label className="block mb-2 text-sm">Name</label>
            <input
              className="w-full border rounded px-2 py-1 mb-4"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              required
            />
            <label className="block mb-2 text-sm">Description</label>
            <textarea
              className="w-full border rounded px-2 py-1 mb-4"
              value={form.description}
              onChange={(e) =>
                setForm({ ...form, description: e.target.value })
              }
            />
            <div className="flex justify-end gap-2">
              <button
                type="button"
                className="px-3 py-1 rounded border"
                onClick={() => setModalOpen(false)}
              >
                Cancel
              </button>
              <button className="px-3 py-1 bg-indigo-600 text-white rounded">
                Create
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}

export default AgentsPage;
