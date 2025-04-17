function App() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-4">
      <h1 className="text-3xl font-bold text-indigo-600">AI Agent SaaS</h1>
      <p className="mt-2 text-gray-700">React + Tailwind + Go backend starter</p>
      <a
        href="/api/agents"
        className="mt-4 underline text-blue-500 hover:text-blue-700"
        target="_blank"
        rel="noreferrer"
      >
        Try GET /api/agents
      </a>
    </div>
  );
}

export default App;
