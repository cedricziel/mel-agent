package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	httpApi "github.com/cedricziel/mel-agent/internal/api"
	"github.com/cedricziel/mel-agent/internal/db"
	"github.com/cedricziel/mel-agent/internal/testutil"
	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/cedricziel/mel-agent/pkg/execution"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
)

// Test the complete router setup that's used in server mode
func TestServerRouterIntegration(t *testing.T) {
	ctx := context.Background()
	_, testDB, cleanup := testutil.SetupPostgresWithMigrations(ctx, t)
	defer cleanup()

	// Set the global database connection that the API handlers expect
	originalDB := db.DB
	db.DB = testDB
	defer func() {
		db.DB = originalDB // Restore after test
	}()

	// Create the exact same router setup as in startServer
	mel := api.NewMel()
	workflowEngineFactory := httpApi.InitializeWorkflowEngine(testDB, mel)
	workflowEngine := execution.NewDurableExecutionEngine(testDB, mel, "test-server")

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// health endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// webhook entrypoint for external events
	r.HandleFunc("/webhooks/{provider}/{triggerID}", httpApi.WebhookHandler)

	// FIXED: Create a merged API handler that includes both main API and workflow engine routes
	apiHandler := createMergedAPIHandler(httpApi.Handler(), workflowEngineFactory(workflowEngine))
	r.Mount("/api", apiHandler)

	// Test critical endpoints that should be available
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		description    string
	}{
		{
			name:           "health_check",
			method:         http.MethodGet,
			path:           "/health",
			expectedStatus: http.StatusOK,
			description:    "Health check endpoint",
		},
		{
			name:           "list_agents",
			method:         http.MethodGet,
			path:           "/api/agents",
			expectedStatus: http.StatusOK,
			description:    "Main API - List agents",
		},
		{
			name:           "list_connections",
			method:         http.MethodGet,
			path:           "/api/connections",
			expectedStatus: http.StatusOK,
			description:    "Main API - List connections",
		},
		{
			name:           "list_node_types",
			method:         http.MethodGet,
			path:           "/api/node-types",
			expectedStatus: http.StatusOK,
			description:    "Main API - List node types",
		},
		{
			name:           "list_integrations",
			method:         http.MethodGet,
			path:           "/api/integrations",
			expectedStatus: http.StatusOK,
			description:    "Main API - List integrations",
		},
		{
			name:           "list_workflow_runs",
			method:         http.MethodGet,
			path:           "/api/workflow-runs",
			expectedStatus: http.StatusOK,
			description:    "Workflow Engine - List workflow runs",
		},
		{
			name:           "worker_registration",
			method:         http.MethodPost,
			path:           "/api/workers",
			expectedStatus: http.StatusBadRequest, // No body, but endpoint should exist
			description:    "Worker API - Register worker",
		},
		{
			name:           "list_triggers",
			method:         http.MethodGet,
			path:           "/api/triggers",
			expectedStatus: http.StatusOK,
			description:    "Main API - List triggers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			t.Logf("Testing %s: %s %s", tt.description, tt.method, tt.path)
			t.Logf("Expected status: %d, Got: %d", tt.expectedStatus, w.Code)

			if w.Code == http.StatusNotFound {
				t.Errorf("❌ Route not found: %s %s - %s", tt.method, tt.path, tt.description)
				t.Logf("Response body: %s", w.Body.String())
			} else if w.Code != tt.expectedStatus {
				t.Logf("⚠️  Unexpected status for %s %s: expected %d, got %d", tt.method, tt.path, tt.expectedStatus, w.Code)
				t.Logf("Response body: %s", w.Body.String())
			} else {
				t.Logf("✅ %s %s works correctly", tt.method, tt.path)
			}

			// For this test, we mainly care that routes exist (not 404)
			// Other status codes like 400, 500 are acceptable as they indicate the route handler ran
			assert.NotEqual(t, http.StatusNotFound, w.Code,
				"Route should exist: %s %s (%s)", tt.method, tt.path, tt.description)
		})
	}
}

// Test the routes that should be available from the main API handler
func TestMainAPIRoutes(t *testing.T) {
	ctx := context.Background()
	_, _, cleanup := testutil.SetupPostgresWithMigrations(ctx, t)
	defer cleanup()

	// Test just the main API handler
	handler := httpApi.Handler()

	mainAPIRoutes := []struct {
		method string
		path   string
		desc   string
	}{
		{http.MethodGet, "/agents", "List agents"},
		{http.MethodPost, "/agents", "Create agent"},
		{http.MethodGet, "/connections", "List connections"},
		{http.MethodPost, "/connections", "Create connection"},
		{http.MethodGet, "/triggers", "List triggers"},
		{http.MethodPost, "/triggers", "Create trigger"},
		{http.MethodGet, "/integrations", "List integrations"},
		{http.MethodGet, "/node-types", "List node types"},
		{http.MethodGet, "/credential-types", "List credential types"},
		{http.MethodPost, "/workers", "Register worker"},
		{http.MethodGet, "/extensions", "List extensions"},
	}

	for _, route := range mainAPIRoutes {
		t.Run(route.desc, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.NotEqual(t, http.StatusNotFound, w.Code,
				"Main API route should exist: %s %s", route.method, route.path)

			if w.Code == http.StatusNotFound {
				t.Errorf("❌ Main API route missing: %s %s", route.method, route.path)
			} else {
				t.Logf("✅ Main API route exists: %s %s (status: %d)", route.method, route.path, w.Code)
			}
		})
	}
}

// Test the routes that should be available from the workflow engine
func TestWorkflowEngineRoutes(t *testing.T) {
	ctx := context.Background()
	_, testDB, cleanup := testutil.SetupPostgresWithMigrations(ctx, t)
	defer cleanup()

	// Test just the workflow engine handler
	mel := api.NewMel()
	workflowEngineFactory := httpApi.InitializeWorkflowEngine(testDB, mel)
	workflowEngine := execution.NewDurableExecutionEngine(testDB, mel, "test-server")
	handler := workflowEngineFactory(workflowEngine)

	workflowRoutes := []struct {
		method string
		path   string
		desc   string
	}{
		{http.MethodGet, "/workflow-runs", "List workflow runs"},
		{http.MethodPost, "/workflow-runs", "Create workflow run"},
	}

	for _, route := range workflowRoutes {
		t.Run(route.desc, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.NotEqual(t, http.StatusNotFound, w.Code,
				"Workflow engine route should exist: %s %s", route.method, route.path)

			if w.Code == http.StatusNotFound {
				t.Errorf("❌ Workflow engine route missing: %s %s", route.method, route.path)
			} else {
				t.Logf("✅ Workflow engine route exists: %s %s (status: %d)", route.method, route.path, w.Code)
			}
		})
	}
}

// Test demonstrating the route conflict issue
func TestRouteConflictDemonstration(t *testing.T) {
	ctx := context.Background()
	_, testDB, cleanup := testutil.SetupPostgresWithMigrations(ctx, t)
	defer cleanup()

	// Set the global database connection that the API handlers expect
	// This is necessary because the main API handlers use db.DB global variable
	originalDB := db.DB
	db.DB = testDB
	defer func() {
		db.DB = originalDB // Restore after test
	}()

	mel := api.NewMel()
	workflowEngineFactory := httpApi.InitializeWorkflowEngine(testDB, mel)
	workflowEngine := execution.NewDurableExecutionEngine(testDB, mel, "test-server")

	r := chi.NewRouter()

	// Test the FIXED setup - no more conflicts
	apiHandler := createMergedAPIHandler(httpApi.Handler(), workflowEngineFactory(workflowEngine))
	r.Mount("/api", apiHandler)

	// Test that main API routes are now accessible
	req := httptest.NewRequest(http.MethodGet, "/api/agents", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// This should demonstrate the issue
	t.Logf("Testing /api/agents after route conflict...")
	t.Logf("Status code: %d", w.Code)
	t.Logf("Response: %s", w.Body.String())

	if w.Code == http.StatusNotFound {
		t.Errorf("❌ STILL BROKEN: /api/agents is not accessible")
		t.Logf("The merge handler approach didn't work")
	} else {
		t.Logf("✅ FIXED: /api/agents is now accessible (status: %d)", w.Code)
	}

	// Test that workflow engine routes are accessible
	req2 := httptest.NewRequest(http.MethodGet, "/api/workflow-runs", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	t.Logf("Testing /api/workflow-runs after route conflict...")
	t.Logf("Status code: %d", w2.Code)
	t.Logf("Response: %s", w2.Body.String())

	if w2.Code != http.StatusNotFound {
		t.Logf("✅ Workflow engine routes are also accessible (status: %d)", w2.Code)
	} else {
		t.Errorf("❌ Workflow engine routes are not accessible")
	}
}
