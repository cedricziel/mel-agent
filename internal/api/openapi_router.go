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

	// Create OpenAPI handlers
	handlers := NewOpenAPIHandlers(database, engine)

	// Create strict server
	strictHandler := NewStrictHandler(handlers, nil)

	// Mount OpenAPI routes
	HandlerFromMux(strictHandler, r)

	return r
}

// createMergedAPIHandler creates a handler that tries main API first, then falls back to workflow engine
func createMergedAPIHandler(mainHandler http.Handler, workflowHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For workflow-runs paths, use the workflow handler directly
		if len(r.URL.Path) >= 14 && r.URL.Path[:14] == "/workflow-runs" {
			workflowHandler.ServeHTTP(w, r)
			return
		}

		// Try main handler first
		mainHandler.ServeHTTP(w, r)
	})
}

// NewCombinedRouter creates a router that combines legacy and OpenAPI handlers
func NewCombinedRouter(database *sql.DB, engine execution.ExecutionEngine) http.Handler {
	r := chi.NewRouter()

	// Add basic middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Create OpenAPI handlers
	openAPIHandlers := NewOpenAPIHandlers(database, engine)
	strictHandler := NewStrictHandler(openAPIHandlers, nil)

	// Mount OpenAPI routes under /api
	r.Route("/api", func(r chi.Router) {
		HandlerFromMux(strictHandler, r)
	})

	// Legacy workflow runs handler (for backward compatibility)
	workflowFactory := InitializeWorkflowEngine(database, nil) // Mel not needed for this specific handler
	workflowHandler := workflowFactory(engine)
	r.Mount("/api", workflowHandler)

	// Legacy handlers for non-OpenAPI endpoints
	legacyHandler := LegacyHandler()
	r.Mount("/api", createMergedAPIHandler(legacyHandler, workflowHandler))

	return r
}
