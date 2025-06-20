package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cedricziel/mel-agent/internal/testutil"
	"github.com/cedricziel/mel-agent/pkg/execution"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenAPIGetHealth tests the health check endpoint
func TestOpenAPIGetHealth(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/health", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
}

// TestOpenAPIListNodeTypes tests the node types endpoint
func TestOpenAPIListNodeTypes(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/node-types", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []NodeType
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Should return at least some node types
	assert.Greater(t, len(response), 0)

	// Check that each node type has required fields
	for _, nodeType := range response {
		assert.NotEmpty(t, nodeType.Id)
		assert.NotEmpty(t, nodeType.Name)
		assert.NotNil(t, nodeType.Kinds)
		assert.Greater(t, len(*nodeType.Kinds), 0)
	}
}

// TestOpenAPIListNodeTypesWithFilter tests filtering node types by kind
func TestOpenAPIListNodeTypesWithFilter(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/node-types?kind=model", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []NodeType
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Should return only model node types
	for _, nodeType := range response {
		assert.Contains(t, *nodeType.Kinds, NodeTypeKindsModel)
	}
}
