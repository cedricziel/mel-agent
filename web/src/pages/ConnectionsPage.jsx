import { useEffect, useState } from 'react';
import axios from 'axios';

function ConnectionsPage() {
  const [connections, setConnections] = useState([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [form, setForm] = useState({
    integration_id: '',
    name: '',
    credentials: {},
  });
  const [editingConnection, setEditingConnection] = useState(null);
  const [integrations, setIntegrations] = useState([]);
  const [credentialTypes, setCredentialTypes] = useState([]);
  const [credentialSchema, setCredentialSchema] = useState(null);
  const [testingCredentials, setTestingCredentials] = useState(false);
  const [testResult, setTestResult] = useState(null);

  useEffect(() => {
    fetchData();
  }, []);

  async function fetchData() {
    setLoading(true);
    try {
      const [connRes, integRes, credRes] = await Promise.all([
        axios.get('/api/connections'),
        axios.get('/api/integrations'),
        axios.get('/api/credential-types'),
      ]);
      setConnections(connRes.data);
      setIntegrations(integRes.data);
      setCredentialTypes(credRes.data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }

  async function fetchCredentialSchema(credentialType, skipDefaults = false) {
    try {
      const response = await axios.get(
        `/api/credential-types/schema/${credentialType}`
      );
      const schema = response.data;
      setCredentialSchema(schema);

      // Only populate default values from schema if not editing (to avoid overriding existing values)
      if (!skipDefaults && !editingConnection) {
        const defaults = {};
        const properties = schema.properties || {};
        Object.entries(properties).forEach(([fieldName, fieldSchema]) => {
          if (fieldSchema.default !== undefined) {
            defaults[fieldName] = fieldSchema.default;
          }
        });

        // Update form with defaults
        if (Object.keys(defaults).length > 0) {
          setForm((prevForm) => ({
            ...prevForm,
            credentials: {
              ...prevForm.credentials,
              ...defaults,
            },
          }));
        }
      }
    } catch (err) {
      console.error('Failed to fetch credential schema:', err);
      setCredentialSchema(null);
    }
  }

  async function submit(e) {
    e.preventDefault();

    const selectedIntegration = integrations.find(
      (i) => i.id === form.integration_id
    );
    if (!selectedIntegration) {
      alert('Please select an integration');
      return;
    }

    setTestingCredentials(true);
    setTestResult(null);

    try {
      // First, test the credentials
      if (selectedIntegration.credential_type) {
        await axios.post(
          `/api/credential-types/${selectedIntegration.credential_type}/test`,
          form.credentials
        );
        setTestResult({ success: true, message: 'credentials are valid' });
      }

      // If test passes, save the connection
      const payload = {
        integration_id: form.integration_id,
        name: form.name,
        secret: form.credentials,
        config: {},
        credential_type: selectedIntegration.credential_type,
      };

      if (editingConnection) {
        // Update existing connection
        await axios.put(`/api/connections/${editingConnection.id}`, payload);
      } else {
        // Create new connection
        await axios.post('/api/connections', payload);
      }

      resetForm();
      fetchData();
    } catch (err) {
      console.error(err);
      const errorMessage = err.response?.data?.error || err.message;
      setTestResult({
        success: false,
        message: errorMessage,
      });
      // Don't close modal on error so user can fix the issue
    } finally {
      setTestingCredentials(false);
    }
  }

  function handleIntegrationChange(integrationId) {
    const integration = integrations.find((i) => i.id === integrationId);

    // Find the credential type to get its name for the default connection name
    let defaultName = '';
    if (integration && integration.credential_type && !editingConnection) {
      const credentialType = credentialTypes.find(
        (ct) => ct.type === integration.credential_type
      );
      if (credentialType) {
        defaultName = credentialType.name;
      }
    }

    setForm({
      ...form,
      integration_id: integrationId,
      name: editingConnection ? form.name : defaultName, // Only set default name for new connections
      credentials: editingConnection ? form.credentials : {},
    });

    // Reset test state
    setTestResult(null);

    if (integration && integration.credential_type) {
      fetchCredentialSchema(integration.credential_type);
    } else {
      setCredentialSchema(null);
    }
  }

  function handleCredentialChange(fieldName, value) {
    setForm({
      ...form,
      credentials: {
        ...form.credentials,
        [fieldName]: value,
      },
    });
    // Clear test result when credentials change
    setTestResult(null);
  }

  function resetForm() {
    setForm({ integration_id: '', name: '', credentials: {} });
    setCredentialSchema(null);
    setTestResult(null);
    setEditingConnection(null);
    setModalOpen(false);
  }

  function openCreateModal() {
    resetForm();
    setModalOpen(true);
  }

  async function openEditModal(connection) {
    setEditingConnection(connection);

    // Find the integration for this connection to get its credential type
    const integration = integrations.find(
      (i) => i.id === connection.integration_id
    );

    try {
      // Fetch the full connection details including secret data
      const response = await axios.get(`/api/connections/${connection.id}`);
      const connectionData = response.data;

      // Extract non-sensitive fields from the secret data
      const secret = connectionData.secret || {};
      const nonSensitiveCredentials = {};

      // Define which fields are considered non-sensitive and can be shown
      const nonSensitiveFields = [
        'baseUrl',
        'url',
        'username',
        'email',
        'apiUrl',
        'endpoint',
      ];

      nonSensitiveFields.forEach((field) => {
        if (secret[field]) {
          nonSensitiveCredentials[field] = secret[field];
        }
      });

      setForm({
        integration_id: connection.integration_id,
        name: connection.name,
        credentials: nonSensitiveCredentials,
      });
    } catch (err) {
      console.error('Failed to fetch connection details:', err);
      // Fallback to basic info if fetch fails
      setForm({
        integration_id: connection.integration_id,
        name: connection.name,
        credentials: {},
      });
    }

    if (integration && integration.credential_type) {
      fetchCredentialSchema(integration.credential_type, true); // Skip defaults when editing
    }

    setModalOpen(true);
  }

  async function deleteConnection(connectionId) {
    if (!confirm('Are you sure you want to delete this connection?')) {
      return;
    }

    try {
      await axios.delete(`/api/connections/${connectionId}`);
      fetchData();
    } catch (err) {
      console.error('Failed to delete connection:', err);
      alert(
        'Failed to delete connection: ' +
          (err.response?.data?.error || err.message)
      );
    }
  }

  function renderCredentialFields() {
    if (!credentialSchema) return null;

    const properties = credentialSchema.properties || {};
    const required = credentialSchema.required || [];

    // Define a logical order for common field names
    const fieldOrder = [
      'baseUrl',
      'url',
      'apiKey',
      'api_key',
      'token',
      'username',
      'password',
    ];

    // Get all field names and sort them according to our preferred order
    const allFieldNames = Object.keys(properties);
    const sortedFieldNames = [
      ...fieldOrder.filter((field) => allFieldNames.includes(field)),
      ...allFieldNames.filter((field) => !fieldOrder.includes(field)),
    ];

    // Define which fields are considered sensitive and should be hidden when editing
    const sensitiveFields = [
      'password',
      'apiKey',
      'api_key',
      'token',
      'secret',
      'key',
    ];

    return sortedFieldNames.map((fieldName) => {
      const fieldSchema = properties[fieldName];
      const isRequired = required.includes(fieldName);
      const isSensitive = sensitiveFields.some((sf) =>
        fieldName.toLowerCase().includes(sf.toLowerCase())
      );
      const fieldType =
        fieldSchema.format === 'uri'
          ? 'url'
          : fieldName.toLowerCase().includes('password')
            ? 'password'
            : 'text';

      // For sensitive fields when editing, show placeholder indicating they're hidden
      const placeholder =
        editingConnection && isSensitive
          ? `Enter new ${fieldSchema.title || fieldName} (current value hidden)`
          : fieldSchema.description;

      return (
        <div key={fieldName}>
          <label className="block mb-2 text-sm">
            {fieldSchema.title || fieldName}
            {isRequired && <span className="text-red-500"> *</span>}
            {editingConnection && isSensitive && (
              <span className="text-gray-500 text-xs ml-1">
                (hidden for security)
              </span>
            )}
          </label>
          <input
            type={fieldType}
            className="w-full border rounded px-2 py-1 mb-4"
            placeholder={placeholder}
            value={form.credentials[fieldName] || ''}
            onChange={(e) => handleCredentialChange(fieldName, e.target.value)}
            required={isRequired}
          />
        </div>
      );
    });
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-bold">Connections</h2>
        <button
          onClick={openCreateModal}
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
              <th className="px-4 py-2 text-left">Actions</th>
            </tr>
          </thead>
          <tbody>
            {connections.map((c) => {
              const integration = integrations.find(
                (i) => i.id === c.integration_id
              );
              return (
                <tr key={c.id} className="border-t">
                  <td className="px-4 py-2">{c.name}</td>
                  <td className="px-4 py-2 text-gray-600">
                    {integration?.name || c.integration_id}
                  </td>
                  <td className="px-4 py-2">
                    <div className="flex gap-2">
                      <button
                        onClick={() => openEditModal(c)}
                        className="text-blue-600 hover:text-blue-800 text-sm"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => deleteConnection(c.id)}
                        className="text-red-600 hover:text-red-800 text-sm"
                      >
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}

      {modalOpen && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center">
          <form onSubmit={submit} className="bg-white p-6 rounded shadow w-96">
            <h3 className="text-lg font-semibold mb-4">
              {editingConnection ? 'Edit Connection' : 'Create Connection'}
            </h3>
            <label className="block mb-2 text-sm">Integration</label>
            <select
              className="w-full border rounded px-2 py-1 mb-4"
              value={form.integration_id}
              onChange={(e) => handleIntegrationChange(e.target.value)}
              required
              disabled={editingConnection} // Can't change integration when editing
            >
              <option value="">-- select --</option>
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

            {renderCredentialFields()}

            {editingConnection && (
              <div className="bg-blue-50 border border-blue-200 rounded p-3 mb-4">
                <p className="text-sm text-blue-800">
                  <strong>Editing Connection:</strong> Non-sensitive fields like
                  URLs and usernames are pre-filled. Sensitive fields like
                  passwords are hidden for security - enter new values to update
                  them.
                </p>
              </div>
            )}

            {/* Test Result Display */}
            {testResult && (
              <div
                className={`p-2 rounded text-sm mt-4 ${
                  testResult.success
                    ? 'bg-green-100 text-green-800 border border-green-300'
                    : 'bg-red-100 text-red-800 border border-red-300'
                }`}
              >
                {testResult.success ? '✓ ' : '✗ '}
                {testResult.message}
              </div>
            )}

            <div className="flex justify-end gap-2">
              <button
                type="button"
                className="px-3 py-1 rounded border"
                onClick={resetForm}
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={testingCredentials}
                className="px-3 py-1 bg-indigo-600 text-white rounded disabled:opacity-50"
              >
                {testingCredentials
                  ? editingConnection
                    ? 'Testing & Updating...'
                    : 'Testing & Saving...'
                  : editingConnection
                    ? 'Test & Update'
                    : 'Test & Save'}
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}

export default ConnectionsPage;
