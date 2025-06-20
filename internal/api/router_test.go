package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	// load builder node definitions
	_ "github.com/cedricziel/mel-agent/pkg/nodes"

	"github.com/cedricziel/mel-agent/internal/api"
	"github.com/cedricziel/mel-agent/internal/plugin"
	"github.com/cedricziel/mel-agent/internal/testutil"
	pkgapi "github.com/cedricziel/mel-agent/pkg/api"
	"github.com/cedricziel/mel-agent/pkg/execution"
)

// TestListExtensions verifies the /extensions endpoint.
func TestListExtensions(t *testing.T) {
	ctx := context.Background()
	_, testDB, cleanup := testutil.SetupPostgresWithMigrations(ctx, t)
	defer cleanup()
	mel := pkgapi.NewMel()
	workflowEngine := execution.NewDurableExecutionEngine(testDB, mel, "test-server")
	handler := api.NewCombinedRouter(testDB, workflowEngine)
	req := httptest.NewRequest(http.MethodGet, "/api/extensions", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	var metas []plugin.PluginMeta
	if err := json.NewDecoder(resp.Body).Decode(&metas); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	// Ensure at least one known plugin is present
	found := false
	for _, m := range metas {
		if m.ID == "schedule" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected plugin meta with ID 'schedule' in /extensions response")
	}
}

// TestListNodeTypes verifies the /node-types endpoint.
func TestListNodeTypes(t *testing.T) {
	ctx := context.Background()
	_, testDB, cleanup := testutil.SetupPostgresWithMigrations(ctx, t)
	defer cleanup()
	mel := pkgapi.NewMel()
	workflowEngine := execution.NewDurableExecutionEngine(testDB, mel, "test-server")
	handler := api.NewCombinedRouter(testDB, workflowEngine)
	req := httptest.NewRequest(http.MethodGet, "/api/node-types", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	var types []pkgapi.NodeType
	if err := json.NewDecoder(resp.Body).Decode(&types); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	// Ensure we get a JSON array (length may vary depending on definitions)
	if types == nil {
		t.Errorf("expected a JSON array, got nil")
	}
}
