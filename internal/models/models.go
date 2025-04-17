package models

// Minimal data structures mirroring the design‑doc.

type User struct {
    ID    string `json:"id"`
    Email string `json:"email"`
    Name  string `json:"name"`
    Plan  string `json:"plan"`
}

type Skill struct {
    ID   string `json:"id"`
    Name string `json:"name"`
    Type string `json:"type"` // llm, http_request, etc.
}

type AgentVersion struct {
    ID              string      `json:"id"`
    SemanticVersion string      `json:"semantic_version"`
    Graph           interface{} `json:"graph"` // keep flexible with empty interface for now
    DefaultParams   interface{} `json:"default_params"`
}

type Agent struct {
    ID              string        `json:"id"`
    UserID          string        `json:"user_id"`
    Name            string        `json:"name"`
    Description     string        `json:"description"`
    IsActive        bool          `json:"is_active"`
    LatestVersion   *AgentVersion `json:"latest_version"`
}

// DummyAgents returns placeholder data used by the in‑memory API stub.
func DummyAgents() []Agent {
    return []Agent{
        {
            ID:          "1",
            UserID:      "u1",
            Name:        "Daily Email Summariser",
            Description: "Summarises yesterday's unread e‑mails and posts to Slack",
            IsActive:    true,
            LatestVersion: &AgentVersion{
                ID:              "v1",
                SemanticVersion: "1.0.0",
            },
        },
    }
}
