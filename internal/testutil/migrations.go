package testutil

import (
	"database/sql"
	"io/fs"
	"sort"
	"testing"
	"time"

	"github.com/cedricziel/mel-agent/migrations"
	"github.com/stretchr/testify/require"
)

// ApplyMigrations applies all migrations using the app's built-in migration system.
// This ensures test databases use the exact same migration logic as production.
func ApplyMigrations(t *testing.T, db *sql.DB) {
	t.Helper()

	// Create schema_migrations table if it doesn't exist
	if _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version TEXT PRIMARY KEY,
            applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
        )`); err != nil {
		require.NoError(t, err, "Failed to create schema_migrations table")
	}

	// Read applied versions into a set
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	require.NoError(t, err, "Failed to query applied migrations")
	defer rows.Close()

	applied := map[string]struct{}{}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			require.NoError(t, err, "Failed to scan migration version")
		}
		applied[v] = struct{}{}
	}

	// Collect migration files from embed FS (same as production)
	entries, err := fs.ReadDir(migrations.FS, ".")
	require.NoError(t, err, "Failed to read embedded migrations")

	// Sort by filename to ensure deterministic order
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		name := entry.Name()
		if _, ok := applied[name]; ok {
			continue // already applied
		}

		// Read migration SQL from embedded migrations FS
		sqlBytes, err := migrations.FS.ReadFile(name)
		require.NoError(t, err, "Failed to read migration %s", name)

		// Execute the migration
		if _, err := db.Exec(string(sqlBytes)); err != nil {
			require.NoError(t, err, "Failed to execute migration %s", name)
		}

		// Record the migration as applied
		if _, err := db.Exec(`INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2)`, name, time.Now()); err != nil {
			require.NoError(t, err, "Failed to record migration %s", name)
		}

		t.Logf("✅ Applied migration: %s", name)
	}
}

// ApplyMigrationsWithTestData applies migrations and inserts common test data.
// This is a convenience function for tests that need standard test agents.
func ApplyMigrationsWithTestData(t *testing.T, db *sql.DB) {
	t.Helper()

	// Apply all migrations first
	ApplyMigrations(t, db)

	// Add standard test agents
	agentsSQL := `
	INSERT INTO agents (id, name, description) VALUES 
	('11111111-1111-1111-1111-111111111111', 'Test Agent 1', 'Integration test agent'),
	('22222222-2222-2222-2222-222222222222', 'Test Agent 2', 'Second test agent')
	ON CONFLICT (id) DO NOTHING;
	`

	_, err := db.Exec(agentsSQL)
	require.NoError(t, err, "Failed to insert test agents")
}
