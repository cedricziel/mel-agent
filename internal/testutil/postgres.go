// Package testutil provides reusable testing utilities for database tests.
//
// Example usage:
//
//	func TestSomething(t *testing.T) {
//		ctx := context.Background()
//		_, db, cleanup := testutil.SetupPostgresWithMigrations(ctx, t)
//		defer cleanup()
//
//		// Your test code here...
//	}
//
// For tests that don't need test data:
//
//	func TestOther(t *testing.T) {
//		ctx := context.Background()
//		_, db, cleanup := testutil.SetupPostgresContainer(ctx, t)
//		defer cleanup()
//
//		testutil.ApplyMigrations(t, db)
//		// Your test code here...
//	}
package testutil

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupPostgresContainer creates a PostgreSQL test container with the given context.
// Returns the container, database connection, and cleanup function.
// The cleanup function should be called with defer to ensure proper cleanup.
func SetupPostgresContainer(ctx context.Context, t *testing.T) (*postgres.PostgresContainer, *sql.DB, func()) {
	t.Helper()

	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Minute)),
	)
	require.NoError(t, err, "Failed to start PostgreSQL container")

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "Failed to get connection string")

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err, "Failed to open database connection")

	err = db.Ping()
	require.NoError(t, err, "Failed to ping database")

	// Return cleanup function
	cleanup := func() {
		if db != nil {
			db.Close()
		}
		if pgContainer != nil {
			pgContainer.Terminate(ctx)
		}
	}

	return pgContainer, db, cleanup
}

// SetupPostgresWithMigrations creates a PostgreSQL container and applies all migrations.
// This is a convenience function for tests that need a fully migrated database without test data.
func SetupPostgresWithMigrations(ctx context.Context, t *testing.T) (*postgres.PostgresContainer, *sql.DB, func()) {
	t.Helper()

	pgContainer, db, cleanup := SetupPostgresContainer(ctx, t)

	// Apply all migrations without test data
	ApplyMigrations(t, db)

	return pgContainer, db, cleanup
}

// SetupPostgresWithTestData creates a PostgreSQL container, applies migrations, and loads test data.
// This is for tests that depend on the standard test agents and fixtures.
func SetupPostgresWithTestData(ctx context.Context, t *testing.T) (*postgres.PostgresContainer, *sql.DB, func()) {
	t.Helper()

	pgContainer, db, cleanup := SetupPostgresContainer(ctx, t)

	// Apply all migrations with test data
	ApplyMigrationsWithTestData(t, db)

	return pgContainer, db, cleanup
}
