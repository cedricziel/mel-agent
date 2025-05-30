import { Link, Route, Routes, NavLink, useParams } from "react-router-dom";
import AgentsPage from "./pages/AgentsPage";
import ConnectionsPage from "./pages/ConnectionsPage";
import BuilderPageNew from "./pages/BuilderPageNew";
import RunsPage from "./pages/RunsPage";
import ErrorBoundary from "./components/ErrorBoundary";

function App() {
  return (
    <div className="min-h-screen flex flex-col">
      <header className="bg-white shadow">
        <div className="container mx-auto px-4 py-4 flex items-center gap-4">
          <Link to="/" className="text-xl font-bold text-indigo-600">
            Agent SaaS
          </Link>
          <nav className="flex gap-4">
            <NavLink
              to="/agents"
              className={({ isActive }) =>
                isActive ? "text-indigo-600" : "text-gray-600 hover:text-indigo-600"
              }
            >
              Agents
            </NavLink>
            <NavLink
              to="/connections"
              className={({ isActive }) =>
                isActive ? "text-indigo-600" : "text-gray-600 hover:text-indigo-600"
              }
            >
              Connections
            </NavLink>
          </nav>
        </div>
      </header>
      <main className="flex-1 container mx-auto px-4 py-6">
        <Routes>
          <Route path="/" element={<Landing />} />
          <Route path="/agents" element={<AgentsPage />} />
          <Route path="/connections" element={<ConnectionsPage />} />
          <Route
            path="/agents/:id/edit"
            element={
              <ErrorBoundary>
                <AgentBuilderWrapper />
              </ErrorBoundary>
            }
          />
          <Route
            path="/agents/:id/runs"
            element={<RunsPage />}
          />
        </Routes>
      </main>
    </div>
  );
}

function Landing() {
  return (
    <div className="text-center text-gray-700">
      <h1 className="text-3xl font-bold text-indigo-600">Welcome to Agent SaaS</h1>
      <p className="mt-2">Use the navigation to get started.</p>
    </div>
  );
}

export default App;

function AgentBuilderWrapper() {
  const { id } = useParams();
  return <BuilderPageNew agentId={id} />;
}
