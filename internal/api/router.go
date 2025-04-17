package api

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/your-org/agentsaas/internal/models"
)

// Handler returns a http.Handler with all API routes mounted.
func Handler() http.Handler {
    r := chi.NewRouter()

    // Example routes demonstrating CRUD pattern for agents.
    r.Get("/agents", listAgents)
    r.Post("/agents", createAgent)

    return r
}

func listAgents(w http.ResponseWriter, r *http.Request) {
    agents := models.DummyAgents()
    writeJSON(w, http.StatusOK, agents)
}

func createAgent(w http.ResponseWriter, r *http.Request) {
    var agent models.Agent
    if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
        return
    }

    // In a real impl we would persist to DB. For now, just echo back.
    agent.ID = "AGENT-PLACEHOLDER-ID"
    writeJSON(w, http.StatusCreated, agent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}
