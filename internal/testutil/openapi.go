package testutil

import (
	"context"
	"database/sql"
	"testing"
)

// SetupOpenAPITestDB sets up a test database for OpenAPI testing.
// The caller needs to create the execution engine to avoid import cycles.
// Uses clean database without pre-inserted test data for OpenAPI tests.
func SetupOpenAPITestDB(t *testing.T) (*sql.DB, func()) {
	ctx := context.Background()
	_, db, cleanup := SetupPostgresWithMigrations(ctx, t)

	return db, cleanup
}

// Helper functions for tests

// StringPtr returns a pointer to the given string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to the given int
func IntPtr(i int) *int {
	return &i
}

// BoolPtr returns a pointer to the given bool
func BoolPtr(b bool) *bool {
	return &b
}
