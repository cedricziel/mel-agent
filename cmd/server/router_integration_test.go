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

// Test the complete router setup that's used in server mode with OpenAPI
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
	workflowEngine := execution.NewDurableExecutionEngine(testDB, mel, "test-server")

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// health endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// readiness endpoint
	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})

	// Use the new combined OpenAPI router
	combinedAPIHandler := httpApi.NewCombinedRouter(testDB, workflowEngine)
	r.Mount("/", combinedAPIHandler)

	// Test that key endpoints are accessible
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "health endpoint",
			method:         "GET",
			path:           "/health",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "readiness endpoint",
			method:         "GET",
			path:           "/ready",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "OpenAPI agents endpoint",
			method:         "GET",
			path:           "/api/agents",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "OpenAPI health endpoint",
			method:         "GET",
			path:           "/api/health",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "WebSocket endpoint accessible",
			method:         "GET",
			path:           "/api/ws/agents/00000000-0000-0000-0000-000000000001",
			expectedStatus: http.StatusBadRequest, // WebSocket upgrade fails but endpoint exists
		},
		{
			name:           "Extensions endpoint",
			method:         "GET",
			path:           "/api/extensions",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Expected status %d for %s %s, got %d",
				tt.expectedStatus, tt.method, tt.path, w.Code)
		})
	}
}

// Test that all essential API endpoints are accessible through the OpenAPI router
func TestOpenAPIEndpointsAccessible(t *testing.T) {
	ctx := context.Background()
	_, testDB, cleanup := testutil.SetupPostgresWithMigrations(ctx, t)
	defer cleanup()

	mel := api.NewMel()
	workflowEngine := execution.NewDurableExecutionEngine(testDB, mel, "test-server")

	// Create OpenAPI router
	router := httpApi.NewOpenAPIRouter(testDB, workflowEngine)

	// Test core OpenAPI endpoints
	coreEndpoints := []struct {
		method string
		path   string
		name   string
	}{
		{"GET", "/api/health", "health check"},
		{"GET", "/api/agents", "list agents"},
		{"GET", "/api/workflows", "list workflows"},
		{"GET", "/api/connections", "list connections"},
		{"GET", "/api/triggers", "list triggers"},
		{"GET", "/api/workers", "list workers"},
		{"GET", "/api/node-types", "list node types"},
		{"GET", "/api/integrations", "list integrations"},
		{"GET", "/api/credential-types", "list credential types"},
		{"GET", "/api/extensions", "list extensions"},
	}

	for _, endpoint := range coreEndpoints {
		t.Run(endpoint.name, func(t *testing.T) {
			req := httptest.NewRequest(endpoint.method, endpoint.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Should be accessible (200) or return proper error codes, not 404
			assert.NotEqual(t, http.StatusNotFound, w.Code,
				"Endpoint %s %s should be accessible, got 404", endpoint.method, endpoint.path)
		})
	}
}
