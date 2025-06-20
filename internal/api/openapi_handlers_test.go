package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cedricziel/mel-agent/internal/db"
	"github.com/cedricziel/mel-agent/internal/testutil"
	"github.com/cedricziel/mel-agent/pkg/execution"
	_ "github.com/cedricziel/mel-agent/pkg/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupOpenAPITest sets up a test database and OpenAPI router
func setupOpenAPITest(t *testing.T) (*sql.DB, http.Handler, func()) {
	ctx := context.Background()
	_, testDB, cleanup := testutil.SetupPostgresWithMigrations(ctx, t)

	// Create mock execution engine
	mockEngine := &execution.MockExecutionEngine{}

	// Create OpenAPI router
	router := NewOpenAPIRouter(testDB, mockEngine)

	return testDB, router, cleanup
}

// TestOpenAPISystemHealth tests the OpenAPI health endpoint
func TestOpenAPISystemHealth(t *testing.T) {
	_, router, cleanup := setupOpenAPITest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Status *string `json:"status"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.NotNil(t, response.Status)
	assert.Equal(t, "ok", *response.Status)
}

// TestOpenAPIListNodeTypes tests the OpenAPI node types endpoint
func TestOpenAPIListNodeTypes(t *testing.T) {
	_, router, cleanup := setupOpenAPITest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/node-types", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var nodeTypes []NodeType
	err := json.NewDecoder(w.Body).Decode(&nodeTypes)
	require.NoError(t, err)

	// Should have at least some node types
	assert.NotEmpty(t, nodeTypes)

	// Check that we have some expected node types
	foundSchedule := false
	for _, nodeType := range nodeTypes {
		if nodeType.Id != nil && *nodeType.Id == "schedule" {
			foundSchedule = true
			assert.NotNil(t, nodeType.Name)
			assert.NotNil(t, nodeType.Description)
			break
		}
	}
	assert.True(t, foundSchedule, "should find schedule node type")
}

// TestOpenAPIListNodeTypesWithKindFilter tests the node types endpoint with kind filtering
func TestOpenAPIListNodeTypesWithKindFilter(t *testing.T) {
	_, router, cleanup := setupOpenAPITest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/node-types?kind=model", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var nodeTypes []NodeType
	err := json.NewDecoder(w.Body).Decode(&nodeTypes)
	require.NoError(t, err)

	// All returned node types should have "model" in their kinds
	for _, nodeType := range nodeTypes {
		if nodeType.Kinds != nil {
			foundModel := false
			for _, kind := range *nodeType.Kinds {
				if kind == NodeTypeKindsModel {
					foundModel = true
					break
				}
			}
			assert.True(t, foundModel, "node type %s should have model kind", *nodeType.Id)
		}
	}
}

// TestOpenAPIListNodeTypesWithMultipleKinds tests filtering with multiple kinds
func TestOpenAPIListNodeTypesWithMultipleKinds(t *testing.T) {
	_, router, cleanup := setupOpenAPITest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/node-types?kind=model,action", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var nodeTypes []NodeType
	err := json.NewDecoder(w.Body).Decode(&nodeTypes)
	require.NoError(t, err)

	// All returned node types should have either "model" or "action" in their kinds
	for _, nodeType := range nodeTypes {
		if nodeType.Kinds != nil {
			foundKind := false
			for _, kind := range *nodeType.Kinds {
				if kind == NodeTypeKindsModel || kind == NodeTypeKindsAction {
					foundKind = true
					break
				}
			}
			assert.True(t, foundKind, "node type %s should have model or action kind", *nodeType.Id)
		}
	}
}