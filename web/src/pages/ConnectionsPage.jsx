import { useEffect, useState } from "react";
import axios from "axios";

function ConnectionsPage() {
  const [connections, setConnections] = useState([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [form, setForm] = useState({ integration_id: "", name: "", apiKey: "" });
  const [integrations, setIntegrations] = useState([]);

  useEffect(() => {
    fetchData();
  }, []);

  async function fetchData() {
    setLoading(true);
    try {
      const [connRes, integRes] = await Promise.all([
        axios.get("/api/connections"),
        axios.get("/api/integrations").catch(() => ({ data: [] })), // integrations endpoint not yet implemented
      ]);
      setConnections(connRes.data);
      setIntegrations(integRes.data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }

  async function submit(e) {
    e.preventDefault();
    // send secret/config as JSON
    const payload = {
      integration_id: form.integration_id,
      name: form.name,
      secret: { api_key: form.apiKey },
      config: {},
    };
    try {
      await axios.post("/api/connections", payload);
      setForm({ integration_id: "", name: "", apiKey: "" });
      setModalOpen(false);
      fetchData();
    } catch (err) {
      console.error(err);
      alert("Failed to create connection");
    }
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-bold">Connections</h2>
        <button
          onClick={() => setModalOpen(true)}
          className="bg-indigo-600 text-white py-1.5 px-3 rounded"
        >
          + New Connection
        </button>
      </div>

      {loading ? (
        <p>Loading…</p>
      ) : (
        <table className="min-w-full bg-white shadow rounded">
          <thead>
            <tr>
              <th className="px-4 py-2 text-left">Name</th>
              <th className="px-4 py-2 text-left">Integration</th>
            </tr>
          </thead>
          <tbody>
            {connections.map((c) => (
              <tr key={c.id} className="border-t">
                <td className="px-4 py-2">{c.name}</td>
                <td className="px-4 py-2 text-gray-600">{c.integration_id}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {modalOpen && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center">
          <form onSubmit={submit} className="bg-white p-6 rounded shadow w-96">
            <h3 className="text-lg font-semibold mb-4">Create Connection</h3>
            <label className="block mb-2 text-sm">Integration</label>
            <select
              className="w-full border rounded px-2 py-1 mb-4"
              value={form.integration_id}
              onChange={(e) => setForm({ ...form, integration_id: e.target.value })}
              required
            >
              <option value="">-- select --</option>
              {integrations.length === 0 && (
                <option value="openai">OpenAI (placeholder)</option>
              )}
              {integrations.map((i) => (
                <option key={i.id} value={i.id}>
                  {i.name}
                </option>
              ))}
            </select>
            <label className="block mb-2 text-sm">Name</label>
            <input
              className="w-full border rounded px-2 py-1 mb-4"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              required
            />
            <label className="block mb-2 text-sm">API Key</label>
            <input
              className="w-full border rounded px-2 py-1 mb-4"
              value={form.apiKey}
              onChange={(e) => setForm({ ...form, apiKey: e.target.value })}
              required
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

export default ConnectionsPage;
