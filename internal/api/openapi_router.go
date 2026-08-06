package api

import (
	"database/sql"
	"net/http"

	"github.com/cedricziel/mel-agent/pkg/execution"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewOpenAPIRouter creates a new router with OpenAPI-generated handlers
func NewOpenAPIRouter(database *sql.DB, engine execution.ExecutionEngine) http.Handler {
	r := chi.NewRouter()

	// Add basic middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Create OpenAPI handlers (empty string means use OPENAI_API_KEY env var)
	handlers := NewOpenAPIHandlers(database, engine, "")

	// Create strict server
	strictHandler := NewStrictHandler(handlers, nil)

	// Mount OpenAPI routes
	HandlerFromMux(strictHandler, r)

	return r
}

// NewCombinedRouter creates a router that combines OpenAPI handlers with essential legacy endpoints
func NewCombinedRouter(database *sql.DB, engine execution.ExecutionEngine) http.Handler {
	r := chi.NewRouter()

	// Add basic middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Create OpenAPI handlers (empty string means use OPENAI_API_KEY env var)
	openAPIHandlers := NewOpenAPIHandlers(database, engine, "")
	strictHandler := NewStrictHandler(openAPIHandlers, nil)

	// Mount all OpenAPI routes (which are prefixed with /api in the spec)
	HandlerFromMux(strictHandler, r)

	// Essential legacy endpoints that aren't covered by OpenAPI
	r.Route("/api", func(r chi.Router) {
		// WebSocket for collaborative updates (not REST, so not in OpenAPI)
		r.Get("/ws/agents/{agentID}", wsHandler)
	})

	return r
}
